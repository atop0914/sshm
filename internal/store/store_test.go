package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sshm/sshm/internal/models"
)

func TestFileStore(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_hosts.json")

	// Test NewFileStore
	store := NewFileStore(tmpFile)
	if store == nil {
		t.Fatal("NewFileStore returned nil")
	}

	// Test Count on empty store
	if store.Count() != 0 {
		t.Errorf("expected count 0, got %d", store.Count())
	}

	// Test AddHost
	host := models.Host{
		ID:     "test-id-1",
		Name:   "test-server",
		Host:   "192.168.1.100",
		Port:   22,
		User:   "admin",
		Group:  "test",
		Tags:   []string{"web", "production"},
	}

	err := store.AddHost(host)
	if err != nil {
		t.Errorf("AddHost failed: %v", err)
	}

	// Test Count after adding
	if store.Count() != 1 {
		t.Errorf("expected count 1, got %d", store.Count())
	}

	// Test ListHosts
	hosts := store.ListHosts()
	if len(hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hosts))
	}

	// Test GetHost
	retrievedHost, err := store.GetHost("test-id-1")
	if err != nil {
		t.Errorf("GetHost failed: %v", err)
	}
	if retrievedHost.Name != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", retrievedHost.Name)
	}

	// Test GetHost with non-existent ID
	_, err = store.GetHost("non-existent")
	if err != ErrHostNotFound {
		t.Errorf("expected ErrHostNotFound, got %v", err)
	}

	// Test UpdateHost
	host.Name = "updated-server"
	err = store.UpdateHost(host)
	if err != nil {
		t.Errorf("UpdateHost failed: %v", err)
	}

	retrievedHost, _ = store.GetHost("test-id-1")
	if retrievedHost.Name != "updated-server" {
		t.Errorf("expected name 'updated-server', got '%s'", retrievedHost.Name)
	}

	// Test DeleteHost
	err = store.DeleteHost("test-id-1")
	if err != nil {
		t.Errorf("DeleteHost failed: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("expected count 0 after delete, got %d", store.Count())
	}

	// Test DeleteHost with non-existent ID
	err = store.DeleteHost("non-existent")
	if err != ErrHostNotFound {
		t.Errorf("expected ErrHostNotFound, got %v", err)
	}
}

func TestSearchHosts(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_search.json")

	store := NewFileStore(tmpFile)

	// Add test hosts
	hosts := []models.Host{
		{ID: "1", Name: "web-server-1", Host: "192.168.1.10", User: "admin", Group: "production", Tags: []string{"web"}},
		{ID: "2", Name: "db-server-1", Host: "192.168.1.20", User: "dbadmin", Group: "production", Tags: []string{"database"}},
		{ID: "3", Name: "dev-server-1", Host: "192.168.1.30", User: "developer", Group: "development", Tags: []string{"web", "dev"}},
	}

	for _, h := range hosts {
		store.AddHost(h)
	}

	// Test search by name
	results := store.SearchHosts("web")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'web', got %d", len(results))
	}

	// Test search by host
	results = store.SearchHosts("192.168.1.20")
	if len(results) != 1 {
		t.Errorf("expected 1 result for IP search, got %d", len(results))
	}

	// Test search by group
	results = store.SearchHosts("production")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'production', got %d", len(results))
	}

	// Test search by tag
	results = store.SearchHosts("database")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'database', got %d", len(results))
	}

	// Test case-insensitive search
	// "admin" matches user field in server-1 (web-server-1) and also appears as substring in "admin" for db-server-1
	results = store.SearchHosts("ADMIN")
	if len(results) != 2 { // matches user="admin" in server-1, and user contains "admin" in db-server-1 (dbadmin contains "admin")
		t.Errorf("expected 2 results for case-insensitive search, got %d", len(results))
	}
}

func TestFilterByTag(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_filter_tag.json")

	store := NewFileStore(tmpFile)

	hosts := []models.Host{
		{ID: "1", Name: "server-1", Tags: []string{"web", "prod"}},
		{ID: "2", Name: "server-2", Tags: []string{"db", "prod"}},
		{ID: "3", Name: "server-3", Tags: []string{"web", "dev"}},
	}

	for _, h := range hosts {
		store.AddHost(h)
	}

	// Test filter by tag
	results := store.FilterByTag("web")
	if len(results) != 2 {
		t.Errorf("expected 2 results for tag 'web', got %d", len(results))
	}

	// Test filter by tag (case insensitive)
	results = store.FilterByTag("PROD")
	if len(results) != 2 {
		t.Errorf("expected 2 results for tag 'PROD', got %d", len(results))
	}
}

func TestFilterByGroup(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_filter_group.json")

	store := NewFileStore(tmpFile)

	hosts := []models.Host{
		{ID: "1", Name: "server-1", Group: "production"},
		{ID: "2", Name: "server-2", Group: "production"},
		{ID: "3", Name: "server-3", Group: "development"},
	}

	for _, h := range hosts {
		store.AddHost(h)
	}

	// Test filter by group
	results := store.FilterByGroup("production")
	if len(results) != 2 {
		t.Errorf("expected 2 results for group 'production', got %d", len(results))
	}

	// Test filter by group (case insensitive)
	results = store.FilterByGroup("DEVELOPMENT")
	if len(results) != 1 {
		t.Errorf("expected 1 result for group 'DEVELOPMENT', got %d", len(results))
	}
}

func TestAddHostGeneratesID(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_auto_id.json")

	store := NewFileStore(tmpFile)

	// Add host without ID
	host := models.Host{
		Name: "auto-id-server",
		Host: "192.168.1.100",
		User: "admin",
	}

	err := store.AddHost(host)
	if err != nil {
		t.Errorf("AddHost failed: %v", err)
	}

	// Verify ID was generated
	hosts := store.ListHosts()
	if len(hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hosts))
	}

	if hosts[0].ID == "" {
		t.Error("expected auto-generated ID, got empty string")
	}
}

func TestDuplicateHost(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_dup.json")

	store := NewFileStore(tmpFile)

	host := models.Host{
		ID:   "same-id",
		Name: "server-1",
		Host: "192.168.1.100",
		User: "admin",
	}

	err := store.AddHost(host)
	if err != nil {
		t.Errorf("first AddHost failed: %v", err)
	}

	// Try to add same ID again
	err = store.AddHost(host)
	if err != ErrHostExists {
		t.Errorf("expected ErrHostExists, got %v", err)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nonexistent.json")

	store := NewFileStore(tmpFile)

	// Should not error, just create empty store
	if store.Count() != 0 {
		t.Errorf("expected empty store, got count %d", store.Count())
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_persist.json")

	// Create store and add hosts
	store1 := NewFileStore(tmpFile)
	host := models.Host{
		ID:   "persist-test",
		Name: "persistent-server",
		Host: "192.168.1.100",
		User: "admin",
	}
	store1.AddHost(host)

	// Create new store from same file
	store2 := NewFileStore(tmpFile)

	// Verify data persisted
	if store2.Count() != 1 {
		t.Errorf("expected 1 host after reload, got %d", store2.Count())
	}

	retrieved, err := store2.GetHost("persist-test")
	if err != nil {
		t.Errorf("failed to get persisted host: %v", err)
	}
	if retrieved.Name != "persistent-server" {
		t.Errorf("expected name 'persistent-server', got '%s'", retrieved.Name)
	}

	// Cleanup
	os.Remove(tmpFile)
}
