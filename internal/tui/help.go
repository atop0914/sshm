package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpView displays help and usage information
type HelpView struct {
	width  int
	height int
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	return &HelpView{}
}

// Init initializes the help view
func (v *HelpView) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (v *HelpView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	}
	return v, nil
}

// View renders the help view
func (v *HelpView) View() string {
	width := 70
	if v.width > 0 {
		width = v.width - 4
	}
	if width < 60 {
		width = 60
	}

	header := BorderStyle.Width(width).Render(
		lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Width(width).
			Align(lipgloss.Center).
			Render(" SSH Host Manager - Help "),
	)

	// Keyboard shortcuts
	shortcuts := [][]string{
		{"↑↓ or j/k", "Navigate host list"},
		{"Enter", "Connect to selected host"},
		{"a", "Add new host"},
		{"e", "Edit selected host"},
		{"x", "Delete selected host"},
		{"d", "View host details"},
		{"c", "Copy SSH command to clipboard"},
		{"h", "View connection history (all)"},
		{"H", "View history for selected host"},
		{"t", "Toggle light/dark theme"},
		{"/", "Filter/search hosts"},
		{"backspace/delete", "Delete character in filter"},
		{"esc", "Clear filter / Go back"},
		{"q, Ctrl+C", "Quit application"},
	}

	var shortcutContent string
	for _, s := range shortcuts {
		key := lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Render(s[0])
		desc := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render(s[1])
		shortcutContent += fmt.Sprintf("  %s  %s\n", key, desc)
	}

	shortcutsBox := lipgloss.NewStyle().
		Width(width).
		Render(shortcutContent)

	shortcutsBorder := BorderStyle.Width(width).Render(
		lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Width(width).
			Render(" Keyboard Shortcuts "),
	)

	// Tips section
	tips := []string{
		"• Hosts are saved to ~/.sshm.json",
		"• SSH config can be imported from ~/.ssh/config",
		"• Connection history is tracked in ~/.sshm_history.json",
		"• Use groups to organize hosts (production, staging, etc.)",
		"• Use tags to label hosts (database, web, backup, etc.)",
		"• Use identity files for key-based authentication",
	}

	var tipsContent string
	for _, t := range tips {
		tipsContent += "  " + lipgloss.NewStyle().Foreground(secondaryColor).Render(t) + "\n"
	}

	tipsBox := lipgloss.NewStyle().
		Width(width).
		Render(tipsContent)

	tipsBorder := BorderStyle.Width(width).Render(
		lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Width(width).
			Render(" Tips "),
	)

	footer := StatusBar("esc: Back to list")

	return header + "\n\n" + shortcutsBorder + "\n" + shortcutsBox + "\n" + tipsBorder + "\n" + tipsBox + "\n\n" + footer
}
