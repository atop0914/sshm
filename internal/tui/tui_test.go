package tui

import (
	"testing"

	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/store"
)

func TestTagColors(t *testing.T) {
	// Test that predefined tag colors exist
	expectedTags := []string{"production", "staging", "development", "local", "database", "web", "backup", "storage", "admin", "default"}

	for _, tag := range expectedTags {
		if _, ok := tagColors[tag]; !ok {
			t.Errorf("Expected tag color for %s", tag)
		}
	}
}

func TestGetHistoryStatsForHost(t *testing.T) {
	// Test with empty store and history
	fileStore := store.NewFileStore("")
	historyStore := store.NewHistoryStore("")
	stats := GetHistoryStatsForHost(fileStore, historyStore, "test-id")
	if stats.TotalConnections != 0 {
		t.Errorf("Expected 0 connections, got %d", stats.TotalConnections)
	}
}

func TestProfileDefault(t *testing.T) {
	profile := models.DefaultProfile()
	if profile.Name != "default" {
		t.Errorf("Expected default profile name, got %s", profile.Name)
	}
	if profile.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", profile.Timeout)
	}
}

func TestAuthType(t *testing.T) {
	if models.AuthTypePassword != "password" {
		t.Errorf("AuthTypePassword should be 'password'")
	}
	if models.AuthTypeKey != "key" {
		t.Errorf("AuthTypeKey should be 'key'")
	}
	if models.AuthTypeAgent != "agent" {
		t.Errorf("AuthTypeAgent should be 'agent'")
	}
}
