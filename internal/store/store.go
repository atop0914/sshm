package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sshm/sshm/internal/models"
)

// ErrHostNotFound is returned when a host is not found
var ErrHostNotFound = errors.New("host not found")

// ErrHostExists is returned when adding a host that already exists
var ErrHostExists = errors.New("host already exists")

// StoreInterface defines the interface for host storage
type StoreInterface interface {
	AddHost(host models.Host) error
	UpdateHost(host models.Host) error
	DeleteHost(id string) error
	GetHost(id string) (models.Host, error)
	ListHosts() []models.Host
	SearchHosts(query string) []models.Host
}

// FileStore manages host data persistence in a file
type FileStore struct {
	path  string
	hosts map[string]models.Host
}

// NewFileStore creates a new FileStore instance
func NewFileStore(path string) *FileStore {
	s := &FileStore{
		path:  path,
		hosts: make(map[string]models.Host),
	}
	s.load()
	return s
}

// load reads data from the storage file
func (s *FileStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read store: %w", err)
	}

	var hosts []models.Host
	if err := json.Unmarshal(data, &hosts); err != nil {
		return fmt.Errorf("failed to parse store data: %w", err)
	}

	s.hosts = make(map[string]models.Host)
	for _, host := range hosts {
		s.hosts[host.ID] = host
	}

	return nil
}

// save writes data to the storage file
func (s *FileStore) save() error {
	hosts := s.ListHosts()
	data, err := json.MarshalIndent(hosts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal hosts: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write store: %w", err)
	}

	return nil
}

// AddHost adds a new host to the store
func (s *FileStore) AddHost(host models.Host) error {
	if host.ID == "" {
		host.ID = uuid.New().String()
	}

	if _, exists := s.hosts[host.ID]; exists {
		return ErrHostExists
	}

	s.hosts[host.ID] = host
	return s.save()
}

// UpdateHost updates an existing host
func (s *FileStore) UpdateHost(host models.Host) error {
	if host.ID == "" {
		return fmt.Errorf("host ID is required for update")
	}

	if _, exists := s.hosts[host.ID]; !exists {
		return ErrHostNotFound
	}

	s.hosts[host.ID] = host
	return s.save()
}

// DeleteHost removes a host by ID
func (s *FileStore) DeleteHost(id string) error {
	if _, exists := s.hosts[id]; !exists {
		return ErrHostNotFound
	}

	delete(s.hosts, id)
	return s.save()
}

// ListHosts returns all hosts
func (s *FileStore) ListHosts() []models.Host {
	hosts := make([]models.Host, 0, len(s.hosts))
	for _, host := range s.hosts {
		hosts = append(hosts, host)
	}
	return hosts
}

// SearchHosts searches hosts by query string
func (s *FileStore) SearchHosts(query string) []models.Host {
	query = lower(query)
	var results []models.Host

	for _, host := range s.hosts {
		if contains(lower(host.Name), query) ||
			contains(lower(host.Host), query) ||
			contains(lower(host.User), query) ||
			contains(lower(host.Proxy), query) ||
			contains(lower(host.Group), query) ||
			containsAny(host.Tags, query) {
			results = append(results, host)
		}
	}

	return results
}

// GetHost returns a host by ID
func (s *FileStore) GetHost(id string) (models.Host, error) {
	host, exists := s.hosts[id]
	if !exists {
		return models.Host{}, ErrHostNotFound
	}
	return host, nil
}

// FilterByTag returns hosts that have the specified tag
func (s *FileStore) FilterByTag(tag string) []models.Host {
	tag = lower(tag)
	var results []models.Host

	for _, host := range s.hosts {
		if containsAny(host.Tags, tag) {
			results = append(results, host)
		}
	}

	return results
}

// FilterByGroup returns hosts that belong to the specified group
func (s *FileStore) FilterByGroup(group string) []models.Host {
	group = lower(group)
	var results []models.Host

	for _, host := range s.hosts {
		if host.Group != "" && contains(lower(host.Group), group) {
			results = append(results, host)
		}
	}

	return results
}

// Count returns the number of hosts in the store
func (s *FileStore) Count() int {
	return len(s.hosts)
}

// helper functions
func lower(s string) string {
	return strings.ToLower(s)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func containsAny(tags []string, query string) bool {
	for _, tag := range tags {
		if contains(lower(tag), query) {
			return true
		}
	}
	return false
}
