package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color scheme for the TUI
type Theme struct {
	Name            string
	Primary         lipgloss.Color
	Secondary       lipgloss.Color
	Success         lipgloss.Color
	Error           lipgloss.Color
	Background      lipgloss.Color
	Surface         lipgloss.Color
	Border          lipgloss.Color
	Text            lipgloss.Color
	TextDim         lipgloss.Color
	SelectedBg      lipgloss.Color
	TagBackground   lipgloss.Color
	StatusOnline    lipgloss.Color
	StatusOffline   lipgloss.Color
	StatusUnknown   lipgloss.Color
}

// DarkTheme is the default dark theme
var DarkTheme = Theme{
	Name:          "dark",
	Primary:       lipgloss.Color("86"),      // Bright green
	Secondary:     lipgloss.Color("241"),     // Gray
	Success:       lipgloss.Color("82"),      // Green
	Error:         lipgloss.Color("203"),      // Red
	Background:    lipgloss.Color("235"),      // Dark gray
	Surface:       lipgloss.Color("237"),     // Medium dark gray
	Border:        lipgloss.Color("240"),     // Light gray border
	Text:          lipgloss.Color("252"),     // Off-white
	TextDim:       lipgloss.Color("245"),     // Dimmed text
	SelectedBg:    lipgloss.Color("237"),     // Surface for selected
	TagBackground: lipgloss.Color("236"),    // Slightly lighter than surface
	StatusOnline:  lipgloss.Color("82"),      // Green
	StatusOffline: lipgloss.Color("241"),     // Gray
	StatusUnknown: lipgloss.Color("245"),     // Light gray
}

// LightTheme is the light theme
var LightTheme = Theme{
	Name:          "light",
	Primary:       lipgloss.Color("35"),      // Teal/cyan
	Secondary:     lipgloss.Color("245"),     // Gray
	Success:       lipgloss.Color("70"),      // Green
	Error:         lipgloss.Color("204"),     // Red/coral
	Background:    lipgloss.Color("255"),     // White
	Surface:       lipgloss.Color("254"),     // Off-white
	Border:        lipgloss.Color("250"),     // Light gray border
	Text:          lipgloss.Color("234"),     // Dark text
	TextDim:       lipgloss.Color("245"),     // Dimmed text
	SelectedBg:    lipgloss.Color("254"),     // Light surface for selected
	TagBackground: lipgloss.Color("252"),     // Light gray for tags
	StatusOnline:  lipgloss.Color("70"),      // Green
	StatusOffline: lipgloss.Color("245"),     // Gray
	StatusUnknown: lipgloss.Color("250"),     // Light gray
}

// GetTheme returns a theme by name
func GetTheme(name string) *Theme {
	switch name {
	case "light":
		return &LightTheme
	case "dark":
		fallthrough
	default:
		return &DarkTheme
	}
}
