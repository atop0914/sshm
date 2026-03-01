package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/store"
)

const (
	fieldName     = "name"
	fieldHost     = "host"
	fieldPort     = "port"
	fieldUser     = "user"
	fieldIdentity = "identity"
	fieldProxy    = "proxy"
	fieldGroup    = "group"
)

// EditView handles add/edit host form
type EditView struct {
	store    *store.FileStore
	host     *models.Host
	mode     string // "add" or "edit"
	field    string
	values   map[string]string
	cursor   int
	errors   map[string]string
	saved    bool
}

// NewEditView creates a new edit view for adding a host
func NewAddView(s *store.FileStore) *EditView {
	return &EditView{
		store:  s,
		host:   &models.Host{},
		mode:   "add",
		field:  fieldName,
		values: make(map[string]string),
		errors: make(map[string]string),
		saved:  false,
	}
}

// NewEditView creates a new edit view for editing a host
func NewEditView(s *store.FileStore, hostID string) (*EditView, error) {
	host, err := s.GetHost(hostID)
	if err != nil {
		return nil, err
	}
	return &EditView{
		store:  s,
		host:   &host,
		mode:   "edit",
		field:  fieldName,
		values: map[string]string{
			fieldName:     host.Name,
			fieldHost:     host.Host,
			fieldPort:     strconv.Itoa(host.Port),
			fieldUser:     host.User,
			fieldIdentity: host.Identity,
			fieldProxy:    host.Proxy,
			fieldGroup:    host.Group,
		},
		errors: make(map[string]string),
		saved:  false,
	}, nil
}

// Init initializes the edit view
func (v *EditView) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (v *EditView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return v.handleKey(msg)
	case tea.WindowSizeMsg:
		return v, nil
	}
	return v, nil
}

func (v *EditView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		v.prevField()
	case "down", "j":
		v.nextField()
	case "enter":
		return v, v.save()
	case "tab":
		v.nextField()
	case "esc":
		return v, tea.Quit
	default:
		// Handle input for current field
		v.handleInput(msg.String())
	}
	return v, nil
}

func (v *EditView) fields() []string {
	return []string{fieldName, fieldHost, fieldPort, fieldUser, fieldIdentity, fieldProxy, fieldGroup}
}

func (v *EditView) prevField() {
	fields := v.fields()
	for i := range fields {
		if v.field == fields[i] {
			v.field = fields[(i-1+len(fields))%len(fields)]
			break
		}
	}
}

func (v *EditView) nextField() {
	fields := v.fields()
	for i := range fields {
		if v.field == fields[i] {
			v.field = fields[(i+1)%len(fields)]
			break
		}
	}
}

func (v *EditView) handleInput(key string) {
	v.values[v.field] += key
	v.validate()
}

func (v *EditView) validate() {
	v.errors = make(map[string]string)

	if v.values[fieldName] == "" {
		v.errors[fieldName] = "Name is required"
	}
	if v.values[fieldHost] == "" {
		v.errors[fieldHost] = "Host is required"
	}
	if v.values[fieldPort] == "" {
		v.errors[fieldPort] = "Port is required"
	} else if _, err := strconv.Atoi(v.values[fieldPort]); err != nil {
		v.errors[fieldPort] = "Port must be a number"
	}
	if v.values[fieldUser] == "" {
		v.errors[fieldUser] = "User is required"
	}
}

func (v *EditView) save() tea.Cmd {
	v.validate()
	if len(v.errors) > 0 {
		return nil
	}

	port, _ := strconv.Atoi(v.values[fieldPort])
	if port == 0 {
		port = 22
	}

	host := models.Host{
		Name:     v.values[fieldName],
		Host:     v.values[fieldHost],
		Port:     port,
		User:     v.values[fieldUser],
		Identity: v.values[fieldIdentity],
		Proxy:    v.values[fieldProxy],
		Group:    v.values[fieldGroup],
	}

	if v.mode == "add" {
		v.store.AddHost(host)
	} else {
		host.ID = v.host.ID
		v.store.UpdateHost(host)
	}

	v.saved = true
	return func() tea.Msg { return tea.Quit() }
}

// View renders the edit form
func (v *EditView) View() string {
	title := "Add Host"
	if v.mode == "edit" {
		title = "Edit Host"
	}

	header := BorderStyle.Width(60).Render(
		TitleStyle.Render(" "+title+" "),
	)

	var fields []string
	for _, f := range v.fields() {
		label := f
		value := v.values[f]
		if f == fieldPort {
			label = "Port"
			if value == "" {
				value = "22"
			}
		}
		if f == fieldIdentity {
			label = "Identity File"
		}
		if f == fieldProxy {
			label = "Proxy Jump"
		}
		if f == fieldGroup {
			label = "Group"
		}

		row := fmt.Sprintf("  %s: %s", label, value)
		if v.field == f {
			row = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Render(row + "_")
		} else {
			row = NormalStyle.Render(row)
		}

		if v.errors[f] != "" {
			row += " " + ErrorStyle.Render(v.errors[f])
		}

		fields = append(fields, row)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, fields...)
	form := BorderStyle.Width(60).Render(body)

	help := HelpStyle.Render("↑↓ move | type to edit | enter: save | esc: cancel")

	return header + "\n\n" + form + "\n\n" + help
}
