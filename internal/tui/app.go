package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/sshm/sshm/internal/store"
)

// App represents the main TUI application
type App struct {
	store     *store.FileStore
	listView  *ListView
	editView  *EditView
	view      string // "list", "add", "edit", "detail"
	quitting  bool
	err       error
}

// New creates a new TUI application
func New(storePath string) (*App, error) {
	s := store.NewFileStore(storePath)

	return &App{
		store:    s,
		listView: NewListView(s),
		view:     "list",
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

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
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
	case "esc":
		m.view = "list"
		m.editView = nil
	default:
		// Handle navigation keys in list view
		if m.view == "list" {
			model, cmd := m.listView.Update(msg)
			m.listView = model.(*ListView)
			return m, cmd
		}
	}
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
		body = BodyStyle.Render(
			fmt.Sprintf("Name: %s\nHost: %s\nPort: %d\nUser: %s\nIdentity: %s\nProxy: %s\nGroup: %s",
				selectedHost.Name,
				selectedHost.Host,
				selectedHost.Port,
				selectedHost.User,
				selectedHost.Identity,
				selectedHost.Proxy,
				selectedHost.Group,
			),
		)
	}

	footer := StatusBar("esc: Back")

	return header + "\n\n" + body + "\n\n" + footer
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
