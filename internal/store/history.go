package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sshm/sshm/internal/models"
)

// HistoryStore manages connection history persistence
type HistoryStore struct {
	path     string
	history  []models.ConnectionHistory
}

// NewHistoryStore creates a new HistoryStore instance
func NewHistoryStore(path string) *HistoryStore {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".sshm_history.json")
	}
	s := &HistoryStore{
		path:    path,
		history: make([]models.ConnectionHistory, 0),
	}
	s.load()
	return s
}

// load reads history from the storage file
func (s *HistoryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	var history []models.ConnectionHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("failed to parse history data: %w", err)
	}

	s.history = history
	return nil
}

// save writes history to the storage file
func (s *HistoryStore) save() error {
	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	return nil
}

// AddConnection records a new connection attempt
func (s *HistoryStore) AddConnection(hostID string, success bool, errMsg string, durationMs int64) error {
	entry := models.ConnectionHistory{
		HostID:    hostID,
		Timestamp: time.Now(),
		Success:   success,
		Error:     errMsg,
		Duration:  durationMs,
	}

	s.history = append(s.history, entry)
	return s.save()
}

// GetHistoryForHost returns all connection history for a specific host
func (s *HistoryStore) GetHistoryForHost(hostID string) []models.ConnectionHistory {
	var results []models.ConnectionHistory
	for _, h := range s.history {
		if h.HostID == hostID {
			results = append(results, h)
		}
	}
	// Sort by timestamp descending (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})
	return results
}

// GetRecentHistory returns the most recent connection attempts
func (s *HistoryStore) GetRecentHistory(limit int) []models.ConnectionHistory {
	if limit <= 0 {
		limit = 10
	}
	if limit > len(s.history) {
		limit = len(s.history)
	}

	// Sort by timestamp descending
	sorted := make([]models.ConnectionHistory, len(s.history))
	copy(sorted, s.history)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	return sorted[:limit]
}

// GetStatsForHost returns connection statistics for a specific host
func (s *HistoryStore) GetStatsForHost(hostID string) models.HistoryStats {
	history := s.GetHistoryForHost(hostID)

	stats := models.HistoryStats{
		HostID: hostID,
	}

	for _, h := range history {
		stats.TotalConnections++
		if h.Success {
			stats.SuccessfulConns++
		} else {
			stats.FailedConns++
		}
		if h.Timestamp.After(stats.LastConnected) {
			stats.LastConnected = h.Timestamp
		}
	}

	return stats
}

// GetAllStats returns connection statistics for all hosts
func (s *HistoryStore) GetAllStats() map[string]models.HistoryStats {
	stats := make(map[string]models.HistoryStats)

	for _, h := range s.history {
		if _, ok := stats[h.HostID]; !ok {
			stats[h.HostID] = models.HistoryStats{HostID: h.HostID}
		}

		s := stats[h.HostID]
		s.TotalConnections++
		if h.Success {
			s.SuccessfulConns++
		} else {
			s.FailedConns++
		}
		if h.Timestamp.After(s.LastConnected) {
			s.LastConnected = h.Timestamp
		}
		stats[h.HostID] = s
	}

	return stats
}

// ClearHistory removes all connection history
func (s *HistoryStore) ClearHistory() error {
	s.history = make([]models.ConnectionHistory, 0)
	return s.save()
}

// ClearHistoryForHost removes history for a specific host
func (s *HistoryStore) ClearHistoryForHost(hostID string) error {
	var remaining []models.ConnectionHistory
	for _, h := range s.history {
		if h.HostID != hostID {
			remaining = append(remaining, h)
		}
	}
	s.history = remaining
	return s.save()
}
