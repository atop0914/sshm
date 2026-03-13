package tui

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/theme"
)

// ThemeManager handles theme switching and provides theme-aware styles
type ThemeManager struct {
	current *theme.Theme
	mu      sync.RWMutex
}

// Global theme manager instance
var themeManager = &ThemeManager{
	current: &theme.DarkTheme,
}

// InitTheme initializes the theme manager with saved preference
func InitTheme(savedTheme string) {
	themeManager.SetTheme(savedTheme)
}

// GetTheme returns the current theme
func GetTheme() *theme.Theme {
	return themeManager.GetCurrent()
}

// SetTheme sets the current theme by name (e.g., "dark", "light")
func SetTheme(name string) string {
	return themeManager.SetTheme(name)
}

// ToggleTheme toggles between light and dark themes
func ToggleTheme() string {
	current := themeManager.GetCurrent()
	if current.Name == "dark" {
		return themeManager.SetTheme("light")
	}
	return themeManager.SetTheme("dark")
}

// GetCurrentThemeName returns the name of the current theme
func GetCurrentThemeName() string {
	return themeManager.GetCurrent().Name
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetTheme(name string) string {
	t := theme.GetTheme(name)
	if t == nil {
		t = &theme.DarkTheme
	}
	tm.mu.Lock()
	tm.current = t
	tm.mu.Unlock()
	// Update global style variables
	updateStyles(t)
	return t.Name
}

// GetCurrent returns the current theme
func (tm *ThemeManager) GetCurrent() *theme.Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

// updateStyles updates the global style variables to use the current theme
func updateStyles(t *theme.Theme) {
	// Update color variables
	primaryColor = t.Primary
	secondaryColor = t.Secondary
	successColor = t.Success
	errorColor = t.Error
	backgroundColor = t.Background
	surfaceColor = t.Surface
	borderColor = t.Border

	// Update style functions
	updateStyleFuncs(t)
}

// updateStyleFuncs updates the style functions with the current theme
func updateStyleFuncs(t *theme.Theme) {
	TitleStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1)

	BodyStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Background(t.SelectedBg).
		Bold(true)

	NormalStyle = lipgloss.NewStyle().
		Foreground(t.Text)

	BorderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border)

	InputStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.Surface).
		Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true)
}

// GetStatusColors returns the status colors for the current theme
func GetStatusColors() (online, offline, unknown lipgloss.Color) {
	t := themeManager.GetCurrent()
	return t.StatusOnline, t.StatusOffline, t.StatusUnknown
}

// GetTagBackground returns the tag background color for the current theme
func GetTagBackground() lipgloss.Color {
	return themeManager.GetCurrent().TagBackground
}

// legacy color variables - these are now managed by theme manager
var (
	// Color definitions (updated by theme manager)
	primaryColor     lipgloss.Color
	secondaryColor   lipgloss.Color
	successColor     lipgloss.Color
	errorColor       lipgloss.Color
	backgroundColor  lipgloss.Color
	surfaceColor     lipgloss.Color
	borderColor      lipgloss.Color

	// Styles (updated by theme manager)
	TitleStyle    lipgloss.Style
	HeaderStyle   lipgloss.Style
	BodyStyle     lipgloss.Style
	HelpStyle     lipgloss.Style
	SelectedStyle lipgloss.Style
	NormalStyle   lipgloss.Style
	BorderStyle   lipgloss.Style
	InputStyle    lipgloss.Style
	ErrorStyle    lipgloss.Style
)

// init initializes the default dark theme styles
func init() {
	updateStyles(&theme.DarkTheme)
}

// StatusBar returns a status bar with given text
func StatusBar(text string) string {
	t := themeManager.GetCurrent()
	return lipgloss.NewStyle().
		Background(t.Surface).
		Foreground(t.Secondary).
		Padding(0, 1).
		Render(text)
}
