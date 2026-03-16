package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/ssh"
	"github.com/sshm/sshm/internal/store"
)

// Tag colors for different tag types
var tagColors = map[string]lipgloss.Color{
	"production": lipgloss.Color("203"), // Red
	"staging":     lipgloss.Color("214"), // Orange
	"development": lipgloss.Color("82"),   // Green
	"local":       lipgloss.Color("75"),   // Blue
	"database":   lipgloss.Color("171"),   // Purple
	"web":         lipgloss.Color("69"),   // Magenta
	"backup":      lipgloss.Color("227"),  // Yellow
	"storage":     lipgloss.Color("141"),  // Lavender
	"admin":       lipgloss.Color("205"),  // Pink
	"default":     lipgloss.Color("241"),  // Gray
}

// ListView displays the host list
type ListView struct {
	store       *store.FileStore
	hosts       []models.Host
	filtered    []models.Host
	selected    int
	filterText  string
	cursor      int
	filtering   bool
	height      int
	width       int
	connecting  bool
	connectHost string
	connectErr  string
	pinging     bool // Whether we're currently pinging hosts
	pingMu      sync.Mutex
}

// NewListView creates a new list view
func NewListView(s *store.FileStore) *ListView {
	hosts := s.ListHosts()
	return &ListView{
		store:    s,
		hosts:    hosts,
		filtered: hosts,
		selected: 0,
		filterText: "",
		cursor:   0,
		filtering: false,
	}
}

// Init initializes the list view
func (v *ListView) Init() tea.Cmd {
	// Start pinging hosts in background
	return v.pingHostsCmd()
}

// connectMsg is used to signal connection result
type connectMsg struct {
	host    models.Host
	err     error
	success bool
}

// pingResultMsg is used to signal ping result for a host
type pingResultMsg struct {
	hostID string
	online bool
	err    error
}

// pingHostsCmd returns a command that pings all hosts in the background
func (v *ListView) pingHostsCmd() tea.Cmd {
	return func() tea.Msg {
		hosts := v.store.ListHosts()
		var wg sync.WaitGroup
		results := make(chan pingResultMsg, len(hosts))

		for _, h := range hosts {
			wg.Add(1)
			go func(host models.Host) {
				defer wg.Done()
				online := true
				err := ssh.Ping(host.Host, host.Port)
				if err != nil {
					online = false
				}
				results <- pingResultMsg{hostID: host.ID, online: online, err: err}
			}(h)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results
		for result := range results {
			// Update host status in background (don't block)
			v.updateHostOnlineStatus(result.hostID, result.online)
		}

		return tea.Msg(pingResultMsg{hostID: "", online: false}) // Signal ping complete
	}
}

// updateHostOnlineStatus updates the online status for a host
func (v *ListView) updateHostOnlineStatus(hostID string, online bool) {
	v.pingMu.Lock()
	defer v.pingMu.Unlock()

	for i := range v.hosts {
		if v.hosts[i].ID == hostID {
			v.hosts[i].Online = &online
		}
	}
	for i := range v.filtered {
		if v.filtered[i].ID == hostID {
			v.filtered[i].Online = &online
		}
	}
}

// Update handles messages
func (v *ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return v.handleKey(msg)
	case tea.WindowSizeMsg:
		v.height = msg.Height
		v.width = msg.Width
		return v, nil
	case connectMsg:
		// Handle connection result
		if msg.success {
			if err := ssh.LaunchSSH(msg.host); err != nil {
				v.connectErr = fmt.Sprintf("Failed to connect: %v", err)
				v.connecting = false
			}
			// SSH launched - quit the TUI
			return v, tea.Quit
		}
		// Connection failed
		v.connectErr = msg.err.Error()
		v.connecting = false
		return v, nil
	case pingResultMsg:
		// Ping completed - refresh filtered list to show updated status
		if msg.hostID == "" {
			// This is the completion signal
			v.updateFiltered()
			return v, nil
		}
	}
	return v, nil
}

func (v *ListView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If filtering, handle filter input
	if v.filtering {
		switch msg.String() {
		case "esc":
			v.filtering = false
			v.filterText = ""
			v.updateFiltered()
			v.cursor = 0
		case "enter":
			v.filtering = false
		case "backspace", "delete", "ctrl+h":
			if len(v.filterText) > 0 {
				v.filterText = v.filterText[:len(v.filterText)-1]
				v.updateFiltered()
				if v.cursor >= len(v.filtered) {
					v.cursor = max(0, len(v.filtered)-1)
				}
			}
		default:
			// Add character to filter
			if len(msg.String()) == 1 {
				v.filterText += msg.String()
				v.updateFiltered()
				v.cursor = 0
			}
		}
		return v, nil
	}

	// Normal navigation
	switch msg.String() {
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.filtered)-1 {
			v.cursor++
		}
	case "home", "g":
		v.cursor = 0
	case "end", "G":
		v.cursor = max(0, len(v.filtered)-1)
	case "pageup":
		v.cursor = max(0, v.cursor-5)
	case "pagedown":
		v.cursor = min(len(v.filtered)-1, v.cursor+5)
	case "/":
		v.filtering = true
		v.filterText = ""
	case "enter":
		// Quick Connect: Connect to selected host
		if len(v.filtered) > 0 && v.cursor < len(v.filtered) {
			host := v.filtered[v.cursor]
			// Set connecting state to show progress
			v.connecting = true
			v.connectHost = host.Name
			v.connectErr = ""
			// Return a command to test connection in background
			return v, func() tea.Msg {
				// Test connection first
				if err := ssh.Ping(host.Host, host.Port); err != nil {
					return connectMsg{host: host, err: err, success: false}
				}
				// Connection OK, return success to launch SSH
				return connectMsg{host: host, success: true}
			}
		}
	case "a":
		// Handled by parent App
	case "e":
		// Handled by parent App
	case "d":
		// Handled by parent App
	case "q", "ctrl+c":
		return v, tea.Quit
	}
	return v, nil
}

func (v *ListView) updateFiltered() {
	if v.filterText == "" {
		v.filtered = v.hosts
	} else {
		lowerFilter := strings.ToLower(v.filterText)
		v.filtered = nil
		for _, h := range v.hosts {
			if strings.Contains(strings.ToLower(h.Name), lowerFilter) ||
				strings.Contains(strings.ToLower(h.Host), lowerFilter) ||
				strings.Contains(strings.ToLower(h.User), lowerFilter) ||
				strings.Contains(strings.ToLower(h.Group), lowerFilter) ||
				stringsContainsAny(h.Tags, lowerFilter) {
				v.filtered = append(v.filtered, h)
			}
		}
	}
}

func stringsContainsAny(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// View renders the list
func (v *ListView) View() string {
	// Ensure filtered is up to date
	if v.filterText != "" {
		v.updateFiltered()
	}

	hosts := v.filtered

	// Calculate dimensions
	width := 70
	if v.width > 0 {
		width = v.width - 4
	}
	if width < 50 {
		width = 50
	}

	listHeight := 15
	if v.height > 20 {
		listHeight = v.height - 12
	}
	if listHeight < 5 {
		listHeight = 5
	}

	// Title bar
	titleBar := v.renderTitleBar(width)

	// Search/filter input
	filterBar := v.renderFilterBar(width)

	// Host list
	listContent := v.renderHostList(width, listHeight)

	// Status bar
	statusBar := v.renderStatusBar(width, hosts)

	return titleBar + "\n" + filterBar + "\n\n" + listContent + "\n\n" + statusBar
}

func (v *ListView) renderTitleBar(width int) string {
	title := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(width).
		Align(lipgloss.Center).
		Render(" SSH Host Manager ")

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Render(title)

	return border
}

func (v *ListView) renderFilterBar(width int) string {
	if v.filtering {
		inputStyle := lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(surfaceColor).
			Width(30).
			Padding(0, 1)
		
		filterLabel := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render("Filter: ")
		
		filterInput := inputStyle.Render(v.filterText + "_")
		
		return filterLabel + filterInput
	}

	// Show hint when not filtering
	hint := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Width(width).
		Render("/ to filter | esc to clear")

	return hint
}

func (v *ListView) renderHostList(width, height int) string {
	hosts := v.filtered

	var content string
	if len(hosts) == 0 {
		emptyMsg := BodyStyle.Width(width).Align(lipgloss.Center).Render(
			"No hosts found.\nPress 'a' to add a host.",
		)
		content = BorderStyle.Width(width).Height(height).Render(emptyMsg)
		return content
	}

	// Calculate visible range
	start := v.cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(hosts) {
		end = len(hosts)
		start = max(0, end-height)
	}

	var rows []string
	for i := start; i < end; i++ {
		h := hosts[i]
		row := v.renderHostRow(h, width-2, i == v.cursor)
		rows = append(rows, row)
	}

	listContent := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(strings.Join(rows, "\n"))

	return BorderStyle.Width(width).Height(height).Render(listContent)
}

func (v *ListView) renderHostRow(h models.Host, width int, selected bool) string {
	// Cursor indicator
	cursor := " "
	if selected {
		cursor = "›"
	}

	// Online/offline status indicator
	var statusIndicator string
	if h.Online != nil {
		if *h.Online {
			statusIndicator = "●" // Green dot for online
		} else {
			statusIndicator = "○" // Gray circle for offline
		}
	} else {
		statusIndicator = "◌" // Dotted circle for unknown
	}

	// Host info
	hostInfo := fmt.Sprintf("%s@%s:%d", h.User, h.Host, h.Port)

	// Group info
	groupInfo := ""
	if h.Group != "" {
		groupInfo = "[" + h.Group + "]"
	}

	// Calculate available width for name (subtract status indicator space)
	availableWidth := width - len(cursor) - len(statusIndicator) - len(hostInfo) - len(groupInfo) - 5
	if availableWidth < 10 {
		availableWidth = 10
	}

	// Truncate name if needed
	name := h.Name
	if len(name) > availableWidth {
		name = name[:availableWidth-2] + ".."
	}

	// Render tags
	tagsStr := v.renderTags(h.Tags, availableWidth)

	// Determine status color
	var statusColor lipgloss.Color
	onlineColor, offlineColor, unknownColor := GetStatusColors()
	if h.Online != nil {
		if *h.Online {
			statusColor = onlineColor
		} else {
			statusColor = offlineColor
		}
	} else {
		statusColor = unknownColor
	}

	// Build the row
	var row string
	if selected {
		row = fmt.Sprintf(" %s %s %-*s %s %s %s", cursor, lipgloss.NewStyle().Foreground(statusColor).Render(statusIndicator), availableWidth, name, groupInfo, hostInfo, tagsStr)
		row = SelectedStyle.Width(width).Render(row)
	} else {
		row = fmt.Sprintf(" %s %s %-*s %s %s %s", cursor, lipgloss.NewStyle().Foreground(statusColor).Render(statusIndicator), availableWidth, name, groupInfo, hostInfo, tagsStr)
		row = NormalStyle.Width(width).Render(row)
	}

	return row
}

func (v *ListView) renderTags(tags []string, availableWidth int) string {
	if len(tags) == 0 {
		return ""
	}

	tagBg := GetTagBackground()

	var tagViews []string
	currentWidth := 0

	for _, tag := range tags {
		color, ok := tagColors[tag]
		if !ok {
			color = tagColors["default"]
		}

		tagStyle := lipgloss.NewStyle().
			Foreground(color).
			Background(tagBg).
			Padding(0, 1).
			Render(tag)

		tagWidth := len(tag) + 2
		if currentWidth+tagWidth > availableWidth-10 {
			break // Don't overflow
		}

		tagViews = append(tagViews, tagStyle)
		currentWidth += tagWidth
	}

	return lipgloss.JoinHorizontal(0, tagViews...)
}

func (v *ListView) renderStatusBar(width int, hosts []models.Host) string {
	// Show connection status if connecting or error
	if v.connecting {
		connectMsg := fmt.Sprintf("Connecting to %s...", v.connectHost)
		connectingStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")). // Green
			Render(connectMsg)
		
		helpText := "↑↓ Navigate | Enter: Connect | a: Add | e: Edit | x: Delete | d: Detail | h: History | i: Import | /: Filter | ?: Help | q: Quit"
		help := HelpStyle.Width(width).Render(helpText)
		return help + "\n" + StatusBar(connectingStatus)
	}

	if v.connectErr != "" {
		errorStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")). // Red
			Render("✗ " + v.connectErr)
		
		helpText := "↑↓ Navigate | Enter: Connect | a: Add | e: Edit | x: Delete | d: Detail | h: History | i: Import | /: Filter | ?: Help | q: Quit"
		help := HelpStyle.Width(width).Render(helpText)
		return help + "\n" + StatusBar(errorStatus)
	}

	// Status text
	hostCount := fmt.Sprintf("%d hosts", len(hosts))
	if v.filterText != "" {
		hostCount = fmt.Sprintf("%d / %d hosts", len(hosts), len(v.hosts))
	}

	statusLeft := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render(hostCount)

	// Selected host info
	var statusRight string
	if len(hosts) > 0 && v.cursor < len(hosts) {
		h := hosts[v.cursor]
		statusRight = fmt.Sprintf("[%d/%d] %s", v.cursor+1, len(hosts), h.Name)
	}

	statusRight = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Width(width - len(hostCount) - 5).
		Align(lipgloss.Right).
		Render(statusRight)

	status := statusLeft + statusRight

	helpText := "↑↓ Navigate | Enter: Connect | a: Add | e: Edit | x: Delete | d: Detail | h: History | i: Import | /: Filter | ?: Help | q: Quit"
	
	help := HelpStyle.Width(width).Render(helpText)

	return help + "\n" + StatusBar(status)
}

// Refresh reloads hosts from store and re-pings all hosts
func (v *ListView) Refresh() {
	v.hosts = v.store.ListHosts()
	v.updateFiltered()
	if v.cursor >= len(v.filtered) {
		v.cursor = max(0, len(v.filtered)-1)
	}
	// Re-ping hosts in background
	go v.pingHostsBackground()
}

// pingHostsBackground pings all hosts without blocking (for Refresh)
func (v *ListView) pingHostsBackground() {
	hosts := v.store.ListHosts()
	var wg sync.WaitGroup

	for _, h := range hosts {
		wg.Add(1)
		go func(host models.Host) {
			defer wg.Done()
			online := true
			err := ssh.Ping(host.Host, host.Port)
			if err != nil {
				online = false
			}
			v.updateHostOnlineStatus(host.ID, online)
		}(h)
	}
	wg.Wait()
}

// GetSelectedHost returns the currently selected host
func (v *ListView) GetSelectedHost() *models.Host {
	if len(v.filtered) > 0 && v.cursor < len(v.filtered) {
		h := v.filtered[v.cursor]
		return &h
	}
	return nil
}

// FilterText returns the current filter text
func (v *ListView) FilterText() string {
	return v.filterText
}

// IsFiltering returns whether filter mode is active
func (v *ListView) IsFiltering() bool {
	return v.filtering
}
