package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/clipboard"
	"github.com/sshm/sshm/internal/config"
	"github.com/sshm/sshm/internal/store"
)

// App represents the main TUI application
type App struct {
	store       *store.FileStore
	history     *store.HistoryStore
	listView    *ListView
	editView    *EditView
	historyView *HistoryView
	helpView    *HelpView
	view        string // "list", "add", "edit", "detail", "history", "help"
	quitting    bool
	err         error
	configPath  string
	pendingDelete string // host ID waiting for delete confirmation
}

// New creates a new TUI application
func New(storePath string) (*App, error) {
	s := store.NewFileStore(storePath)
	h := store.NewHistoryStore("")

	// Load config to get theme preference
	cfgPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err == nil && cfg != nil && cfg.Theme != "" {
		InitTheme(cfg.Theme)
	} else {
		// Default to dark theme
		InitTheme("dark")
	}

	return &App{
		store:      s,
		history:    h,
		listView:   NewListView(s),
		helpView:   NewHelpView(),
		view:       "list",
		configPath: cfgPath,
	}, nil
}

// Init initializes the TUI application
func (m *App) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		return m, nil
	}
	return m, nil
}

// View renders the TUI
func (m *App) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Show delete confirmation if pending
	if m.pendingDelete != "" {
		confirmMsg := fmt.Sprintf("Delete this host? Press 'x' or 'y' to confirm, 'n' or 'esc' to cancel.")
		confirmDisplay := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Bold(true).
			Render("⚠️ " + confirmMsg)
		
		baseView := m.listView.View()
		return baseView + "\n\n" + StatusBar(confirmDisplay)
	}

	switch m.view {
	case "list":
		return m.listView.View()
	case "add":
		if m.editView != nil {
			return m.editView.View()
		}
		return m.renderAdd()
	case "edit":
		if m.editView != nil {
			return m.editView.View()
		}
		return m.renderEdit()
	case "detail":
		return m.renderDetail()
	case "history":
		if m.historyView != nil {
			return m.historyView.View()
		}
		return m.renderHistory()
	case "help":
		return m.helpView.View()
	default:
		return m.listView.View()
	}
}

func (m *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to edit view if active
	if m.view == "add" || m.view == "edit" {
		if m.editView != nil {
			model, cmd := m.editView.Update(msg)
			m.editView = model.(*EditView)

			// Check if edit view signaled quit (save completed or cancel)
			if m.editView.saved || msg.String() == "esc" {
				m.view = "list"
				m.editView = nil
				m.listView.Refresh()
			}
			return m, cmd
		}
	}

	// Handle help view
	if m.view == "help" {
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" {
			m.view = "list"
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "?":
		// Show help view
		m.helpView = NewHelpView()
		m.view = "help"
	case "t":
		// Toggle theme
		newTheme := ToggleTheme()
		m.saveThemePreference(newTheme)
	case "i":
		// Import from SSH config
		return m.handleSSHConfigImport()
	case "a":
		// Start add mode
		m.editView = NewAddView(m.store)
		m.view = "add"
	case "e":
		// Start edit mode with selected host
		selectedHost := m.listView.GetSelectedHost()
		if selectedHost != nil {
			editView, err := NewEditView(m.store, selectedHost.ID)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.editView = editView
			m.view = "edit"
		}
	case "d":
		m.view = "detail"
	case "h":
		// Show history view
		m.historyView = NewHistoryView(m.store, m.history, "")
		m.view = "history"
	case "H":
		// Show history for selected host
		selectedHost := m.listView.GetSelectedHost()
		if selectedHost != nil {
			m.historyView = NewHistoryView(m.store, m.history, selectedHost.ID)
			m.view = "history"
		}
	case "c":
		// Copy SSH command to clipboard
		selectedHost := m.listView.GetSelectedHost()
		if selectedHost != nil {
			sshCmd := selectedHost.GenerateSSHCommand()
			if err := clipboard.CopyToClipboard(sshCmd); err != nil {
				m.err = fmt.Errorf("failed to copy to clipboard: %w", err)
			}
		}
	case "x":
		// Delete selected host (with confirmation)
		selectedHost := m.listView.GetSelectedHost()
		if selectedHost != nil {
			if m.pendingDelete == selectedHost.ID {
				// Second press - confirm delete
				if err := m.store.DeleteHost(selectedHost.ID); err != nil {
					m.err = fmt.Errorf("failed to delete host: %w", err)
				} else {
					m.listView.Refresh()
				}
				m.pendingDelete = ""
			} else {
				// First press - ask for confirmation
				m.pendingDelete = selectedHost.ID
			}
		}
	case "y":
		// Confirm delete when pending
		if m.pendingDelete != "" {
			if err := m.store.DeleteHost(m.pendingDelete); err != nil {
				m.err = fmt.Errorf("failed to delete host: %w", err)
			} else {
				m.listView.Refresh()
			}
			m.pendingDelete = ""
		}
	case "n", "esc":
		// Cancel delete confirmation or go back
		m.pendingDelete = ""
		if m.view != "list" {
			m.view = "list"
			m.editView = nil
			m.historyView = nil
		}
	default:
		// Handle navigation keys in list view
		if m.view == "list" {
			model, cmd := m.listView.Update(msg)
			m.listView = model.(*ListView)
			return m, cmd
		}
		// Handle navigation in history view
		if m.view == "history" && m.historyView != nil {
			model, cmd := m.historyView.Update(msg)
			m.historyView = model.(*HistoryView)
			return m, cmd
		}
	}
	return m, nil
}

// handleSSHConfigImport imports hosts from ~/.ssh/config
func (m *App) handleSSHConfigImport() (tea.Model, tea.Cmd) {
	hosts, err := config.ImportFromSSHConfig("")
	if err != nil {
		m.err = fmt.Errorf("failed to import SSH config: %w", err)
		return m, nil
	}

	if len(hosts) == 0 {
		fmt.Println("No new hosts found in ~/.ssh/config")
		return m, nil
	}

	// Add imported hosts to store
	imported := 0
	for _, h := range hosts {
		if err := m.store.AddHost(h); err == nil {
			imported++
		}
	}

	fmt.Printf("Imported %d hosts from ~/.ssh/config\n", imported)
	m.listView.Refresh()
	return m, nil
}

func (m *App) renderList() string {
	hosts := m.store.ListHosts()

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render("SSH Host Manager"),
	)

	var body string
	if len(hosts) == 0 {
		body = BodyStyle.Render("No hosts configured. Press 'a' to add a host.")
	} else {
		body = ""
		for _, h := range hosts {
			body += NormalStyle.Render(fmt.Sprintf("  %s  %s@%s:%d", h.Name, h.User, h.Host, h.Port)) + "\n"
		}
	}

	footer := StatusBar("↑↓ Navigate | a: Add | e: Edit | d: Detail | q: Quit")

	return header + "\n\n" + body + "\n\n" + footer
}

func (m *App) renderAdd() string {
	if m.editView != nil {
		return m.editView.View()
	}

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render("Add New Host"),
	)

	body := BodyStyle.Render("Form to add new host (coming soon)")

	footer := StatusBar("esc: Back | Enter: Save")

	return header + "\n\n" + body + "\n\n" + footer
}

func (m *App) renderEdit() string {
	if m.editView != nil {
		return m.editView.View()
	}

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render("Edit Host"),
	)

	body := BodyStyle.Render("Form to edit host (coming soon)")

	footer := StatusBar("esc: Back | Enter: Save")

	return header + "\n\n" + body + "\n\n" + footer
}

func (m *App) renderDetail() string {
	selectedHost := m.listView.GetSelectedHost()

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render("Host Details"),
	)

	var body string
	if selectedHost == nil {
		body = BodyStyle.Render("No host selected")
	} else {
		stats := GetHistoryStatsForHost(m.store, m.history, selectedHost.ID)
		body = BodyStyle.Render(
			fmt.Sprintf("Name: %s\nHost: %s\nPort: %d\nUser: %s\nIdentity: %s\nProxy: %s\nGroup: %s\n\nConnection Stats:\n  Total: %d\n  Successful: %d\n  Failed: %d\n  Last: %s",
				selectedHost.Name,
				selectedHost.Host,
				selectedHost.Port,
				selectedHost.User,
				selectedHost.Identity,
				selectedHost.Proxy,
				selectedHost.Group,
				stats.TotalConnections,
				stats.SuccessfulConns,
				stats.FailedConns,
				stats.LastConnected.Format("2006-01-02 15:04"),
			),
		)
	}

	footer := StatusBar("esc: Back")

	return header + "\n\n" + body + "\n\n" + footer
}

func (m *App) renderHistory() string {
	if m.historyView != nil {
		return m.historyView.View()
	}

	header := BorderStyle.Width(60).Render(
		HeaderStyle.Render("Connection History"),
	)

	body := BodyStyle.Render("Loading history...")

	footer := StatusBar("↑↓ Navigate | r: Refresh | c: Clear | esc: Back")

	return header + "\n\n" + body + "\n\n" + footer
}

// saveThemePreference saves the theme preference to config file
func (m *App) saveThemePreference(themeName string) {
	cfg, err := config.LoadConfig(m.configPath)
	if err != nil {
		return // Silently fail - theme will work for this session
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.Theme = themeName
	_ = config.SaveConfig(cfg, m.configPath)
}

// Run starts the TUI application
func Run(storePath string) error {
	app, err := New(storePath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func Main() {
	if err := Run(""); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
