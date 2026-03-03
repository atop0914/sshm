package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/store"
)

// HistoryView displays connection history
type HistoryView struct {
	list    list.Model
	store   *store.FileStore
	history *store.HistoryStore
	hostID  string // empty string means show all
}

// NewHistoryView creates a new history view
func NewHistoryView(store *store.FileStore, history *store.HistoryStore, hostID string) *HistoryView {
	h := &HistoryView{
		store:   store,
		history: history,
		hostID:  hostID,
	}

	h.refreshList()
	return h
}

func (h *HistoryView) refreshList() {
	var items []list.Item

	if h.hostID != "" {
		// Show history for specific host
		host, err := h.store.GetHost(h.hostID)
		if err == nil {
			entries := h.history.GetHistoryForHost(h.hostID)
			for _, e := range entries {
				items = append(items, historyItem{
					hostName: host.Name,
					entry:    e,
				})
			}
		}
	} else {
		// Show all history with host names
		entries := h.history.GetRecentHistory(50)
		hosts := h.store.ListHosts()
		hostMap := make(map[string]string)
		for _, host := range hosts {
			hostMap[host.ID] = host.Name
		}

		for _, e := range entries {
			hostName := hostMap[e.HostID]
			if hostName == "" {
				hostName = e.HostID
			}
			items = append(items, historyItem{
				hostName: hostName,
				entry:    e,
			})
		}
	}

	if len(items) == 0 {
		items = append(items, list.Item(nil))
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	h.list = l
}

// Init initializes the history view
func (h *HistoryView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the history view
func (h *HistoryView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh
			h.refreshList()
			return h, nil
		case "c":
			// Clear history
			if h.hostID != "" {
				h.history.ClearHistoryForHost(h.hostID)
			} else {
				h.history.ClearHistory()
			}
			h.refreshList()
			return h, nil
		}
	}
	newModel, cmd := h.list.Update(msg)
	h.list = newModel
	return h, cmd
}

// View renders the history view
func (h *HistoryView) View() string {
	var title string
	if h.hostID != "" {
		host, _ := h.store.GetHost(h.hostID)
		title = fmt.Sprintf("Connection History: %s", host.Name)
	} else {
		title = "All Connection History"
	}

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render(title),
	)

	body := h.list.View()

	footer := StatusBar("↑↓ Navigate | r: Refresh | c: Clear | esc: Back")

	return header + "\n\n" + body + "\n\n" + footer
}

// historyItem represents a connection history entry in the list
type historyItem struct {
	hostName string
	entry    models.ConnectionHistory
}

func (i historyItem) FilterValue() string {
	return i.hostName
}

func (i historyItem) Title() string {
	if i.entry.Timestamp.IsZero() {
		return "No connection history"
	}
	status := "✓"
	if !i.entry.Success {
		status = "✗"
	}
	return fmt.Sprintf("%s %s", status, i.hostName)
}

func (i historyItem) Description() string {
	if i.entry.Timestamp.IsZero() {
		return "No connections recorded yet"
	}
	timestamp := i.entry.Timestamp.Format("2006-01-02 15:04:05")
	desc := timestamp
	if i.entry.Duration > 0 {
		desc += fmt.Sprintf(" (%dms)", i.entry.Duration)
	}
	if !i.entry.Success && i.entry.Error != "" {
		desc += " - " + i.entry.Error
	}
	return desc
}

// GetHistoryStatsForHost returns connection statistics for a host
func GetHistoryStatsForHost(store *store.FileStore, history *store.HistoryStore, hostID string) models.HistoryStats {
	stats := history.GetStatsForHost(hostID)
	if stats.TotalConnections == 0 {
		// Get from host's connection count field
		host, err := store.GetHost(hostID)
		if err == nil {
			stats.TotalConnections = host.ConnectionCount
			stats.HostID = hostID
		}
	}
	return stats
}

// RecordConnection records a connection attempt
func RecordConnection(history *store.HistoryStore, store *store.FileStore, hostID string, success bool, errMsg string, durationMs int64) {
	// Record in history
	history.AddConnection(hostID, success, errMsg, durationMs)

	// Update host connection count if successful
	if success {
		host, err := store.GetHost(hostID)
		if err == nil {
			host.ConnectionCount++
			store.UpdateHost(host)
		}
	}
}

// FormatDuration formats milliseconds to human readable string
func FormatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	d := time.Duration(ms) * time.Millisecond
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
