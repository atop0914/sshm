package store

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sshm/sshm/internal/models"
)

// Store manages host data persistence
type Store struct {
	path   string
	hosts  map[string]models.Host
}

// NewStore creates a new Store instance
func NewStore(path string) *Store {
	s := &Store{
		path:  path,
		hosts: make(map[string]models.Host),
	}
	// Load existing data if file exists
	s.load()
	return s
}

// load reads data from the storage file
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file yet, that's ok
		}
		return fmt.Errorf("failed to read store: %w", err)
	}

	var hosts []models.Host
	// Try JSON first
	if err := json.Unmarshal(data, &hosts); err != nil {
		// Try YAML
		// Note: For YAML support, would need gopkg.in/yaml.v3
		// For now, we support JSON
		return fmt.Errorf("failed to parse store data: %w", err)
	}

	s.hosts = make(map[string]models.Host)
	for _, host := range hosts {
		s.hosts[host.ID] = host
	}

	return nil
}

// save writes data to the storage file
func (s *Store) save() error {
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
func (s *Store) AddHost(host models.Host) error {
	if host.ID == "" {
		host.ID = uuid.New().String()
	}

	if _, exists := s.hosts[host.ID]; exists {
		return fmt.Errorf("host with ID %s already exists", host.ID)
	}

	s.hosts[host.ID] = host
	return s.save()
}

// UpdateHost updates an existing host
func (s *Store) UpdateHost(host models.Host) error {
	if host.ID == "" {
		return fmt.Errorf("host ID is required for update")
	}

	if _, exists := s.hosts[host.ID]; !exists {
		return fmt.Errorf("host with ID %s not found", host.ID)
	}

	s.hosts[host.ID] = host
	return s.save()
}

// DeleteHost removes a host by ID
func (s *Store) DeleteHost(id string) error {
	if _, exists := s.hosts[id]; !exists {
		return fmt.Errorf("host with ID %s not found", id)
	}

	delete(s.hosts, id)
	return s.save()
}

// ListHosts returns all hosts
func (s *Store) ListHosts() []models.Host {
	hosts := make([]models.Host, 0, len(s.hosts))
	for _, host := range s.hosts {
		hosts = append(hosts, host)
	}
	return hosts
}

// SearchHosts searches hosts by query string
func (s *Store) SearchHosts(query string) []models.Host {
	query = strings.ToLower(query)
	var results []models.Host

	for _, host := range s.hosts {
		if strings.Contains(strings.ToLower(host.Name), query) ||
			strings.Contains(strings.ToLower(host.Host), query) ||
			strings.Contains(strings.ToLower(host.User), query) ||
			strings.Contains(strings.ToLower(host.Proxy), query) ||
			containsAny(host.Tags, query) {
			results = append(results, host)
		}
	}

	return results
}

// containsAny checks if any tag contains the query
func containsAny(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetHost returns a host by ID
func (s *Store) GetHost(id string) (models.Host, error) {
	host, exists := s.hosts[id]
	if !exists {
		return models.Host{}, fmt.Errorf("host with ID %s not found", id)
	}
	return host, nil
}
