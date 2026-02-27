package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color definitions
	primaryColor     = lipgloss.Color("86")
	secondaryColor   = lipgloss.Color("241")
	successColor     = lipgloss.Color("82")
	errorColor       = lipgloss.Color("203")
	backgroundColor  = lipgloss.Color("235")
	surfaceColor     = lipgloss.Color("237")
	borderColor      = lipgloss.Color("240")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	BodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(surfaceColor).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(surfaceColor).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)
)

// StatusBar returns a status bar with given text
func StatusBar(text string) string {
	return lipgloss.NewStyle().
		Background(surfaceColor).
		Foreground(secondaryColor).
		Padding(0, 1).
		Render(text)
}
