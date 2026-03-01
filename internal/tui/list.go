package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/models"
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
	return nil
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
		case "backspace":
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
		// TODO: Connect to selected host
		if len(v.filtered) > 0 && v.cursor < len(v.filtered) {
			fmt.Printf("Selected host: %s\n", v.filtered[v.cursor].Name)
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

	// Host info
	hostInfo := fmt.Sprintf("%s@%s:%d", h.User, h.Host, h.Port)

	// Calculate available width for name
	availableWidth := width - len(cursor) - len(hostInfo) - 4
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

	// Build the row
	var row string
	if selected {
		row = fmt.Sprintf(" %s %-*s %s %s", cursor, availableWidth, name, hostInfo, tagsStr)
		row = SelectedStyle.Width(width).Render(row)
	} else {
		row = fmt.Sprintf(" %s %-*s %s %s", cursor, availableWidth, name, hostInfo, tagsStr)
		row = NormalStyle.Width(width).Render(row)
	}

	return row
}

func (v *ListView) renderTags(tags []string, availableWidth int) string {
	if len(tags) == 0 {
		return ""
	}

	var tagViews []string
	currentWidth := 0

	for _, tag := range tags {
		color, ok := tagColors[tag]
		if !ok {
			color = tagColors["default"]
		}

		tagStyle := lipgloss.NewStyle().
			Foreground(color).
			Background(lipgloss.Color("236")).
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

	return HelpStyle.Width(width).Render(status)
}

// Refresh reloads hosts from store
func (v *ListView) Refresh() {
	v.hosts = v.store.ListHosts()
	v.updateFiltered()
	if v.cursor >= len(v.filtered) {
		v.cursor = max(0, len(v.filtered)-1)
	}
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
