package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/store"
)

const (
	viewList = "list"
	viewAdd  = "add"
	viewEdit = "edit"
)

// ListView displays the host list
type ListView struct {
	store    *store.FileStore
	hosts    []models.Host
	selected int
	filter   string
	cursor   int
}

// NewListView creates a new list view
func NewListView(s *store.FileStore) *ListView {
	return &ListView{
		store:    s,
		hosts:    s.ListHosts(),
		selected: 0,
		filter:   "",
		cursor:   0,
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
		return v, nil
	}
	return v, nil
}

func (v *ListView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.filteredHosts())-1 {
			v.cursor++
		}
	case "enter":
		// TODO: Open host detail/edit
	case "a":
		// TODO: Switch to add view
	case "d":
		// TODO: Switch to delete confirmation
	case "/":
		// TODO: Focus filter input
	case "q", "ctrl+c":
		return v, tea.Quit
	}
	return v, nil
}

func (v *ListView) filteredHosts() []models.Host {
	if v.filter == "" {
		return v.hosts
	}
	var result []models.Host
	lowerFilter := strings.ToLower(v.filter)
	for _, h := range v.hosts {
		if strings.Contains(strings.ToLower(h.Name), lowerFilter) ||
			strings.Contains(strings.ToLower(h.Host), lowerFilter) ||
			strings.Contains(strings.ToLower(h.User), lowerFilter) {
			result = append(result, h)
		}
	}
	return result
}

// View renders the list
func (v *ListView) View() string {
	hosts := v.filteredHosts()

	// Title
	title := BorderStyle.Width(60).Render(
		TitleStyle.Render(" SSH Host Manager "),
	)

	// Filter bar
	filterBar := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render("/ to filter, esc to clear")

	// Host list
	var listContent string
	if len(hosts) == 0 {
		listContent = BodyStyle.Render("No hosts found. Press 'a' to add a host.")
	} else {
		for i, h := range hosts {
			cursor := " "
			if i == v.cursor {
				cursor = "›"
			}
			row := fmt.Sprintf("  %s %s  %s@%s:%d", cursor, h.Name, h.User, h.Host, h.Port)
			if i == v.cursor {
				listContent += SelectedStyle.Width(58).Render(row) + "\n"
			} else {
				listContent += NormalStyle.Width(58).Render(row) + "\n"
			}
		}
	}

	listBox := BorderStyle.Width(60).Height(12).Render(listContent)

	// Help bar
	help := HelpStyle.Render("↑↓ navigate | a: add | enter: details | /: filter | q: quit")

	return title + "\n" + filterBar + "\n\n" + listBox + "\n\n" + help
}
