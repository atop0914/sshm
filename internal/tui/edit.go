package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/store"
)

const (
	fieldName      = "name"
	fieldHost      = "host"
	fieldPort      = "port"
	fieldUser      = "user"
	fieldAuthType  = "auth_type"
	fieldIdentity  = "identity"
	fieldPassword  = "password"
	fieldProxy     = "proxy"
	fieldGroup     = "group"
	fieldTags      = "tags"
)

// AuthType represents authentication method
type AuthType string

const (
	AuthPassword  AuthType = "password"
	AuthKey      AuthType = "key"
	AuthAgent    AuthType = "agent"
)

// EditView handles add/edit host form
type EditView struct {
	store        *store.FileStore
	host         *models.Host
	mode         string // "add" or "edit"
	field        string
	values       map[string]string
	securePassword string // password stored separately (not displayed)
	cursor       int
	errors       map[string]string
	saved        bool
	fileBrowser  *FileBrowser
	showBrowser  bool
	existingTags []string
	existingGroups []string
	enterPassword bool // flag to indicate we're entering password
	passwordMasked string // placeholder display for password
}

// FileBrowser handles SSH key file selection
type FileBrowser struct {
	path     string
	files    []os.DirEntry
	cursor   int
	selected string
}

// NewFileBrowser creates a new file browser for selecting SSH keys
func NewFileBrowser(startPath string) *FileBrowser {
	fb := &FileBrowser{
		path: startPath,
	}
	fb.readDir()
	return fb
}

func (fb *FileBrowser) readDir() {
	entries, err := os.ReadDir(fb.path)
	if err != nil {
		fb.files = nil
		return
	}
	
	var files []os.DirEntry
	for _, e := range entries {
		// Show directories and .pem, .key files, or files in ~/.ssh
		if e.IsDir() || 
		   filepath.Ext(e.Name()) == ".pem" ||
		   filepath.Ext(e.Name()) == ".key" ||
		   e.Name() == "known_hosts" ||
		   e.Name() == "authorized_keys" {
			files = append(files, e)
		}
	}
	fb.files = files
}

func (fb *FileBrowser) Up() {
	if fb.cursor > 0 {
		fb.cursor--
	}
}

func (fb *FileBrowser) Down() {
	if fb.cursor < len(fb.files)-1 {
		fb.cursor++
	}
}

func (fb *FileBrowser) Select() bool {
	if fb.cursor < len(fb.files) {
		fb.selected = filepath.Join(fb.path, fb.files[fb.cursor].Name())
		return fb.files[fb.cursor].IsDir()
	}
	return false
}

func (fb *FileBrowser) GetSelectedPath() string {
	return fb.selected
}

func (fb *FileBrowser) View(width int) string {
	var rows []string
	
	// Current path
	pathRow := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render("📁 " + fb.path)
	rows = append(rows, pathRow)
	
	// Parent directory option
	if fb.path != "/" && fb.path != os.Getenv("HOME") {
		parentRow := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render("  ..")
		rows = append(rows, parentRow)
	}
	
	// Files and directories
	for i, e := range fb.files {
		icon := "📄"
		if e.IsDir() {
			icon = "📁"
		}
		
		name := icon + " " + e.Name()
		if i == fb.cursor {
			name = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Render(name + " ◀")
		}
		rows = append(rows, "  "+name)
	}
	
	if len(rows) == 1 {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render("  (empty)"))
	}
	
	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return BorderStyle.Width(width).Render(body)
}

// NewEditView creates a new edit view for adding a host
func NewAddView(s *store.FileStore) *EditView {
	// Collect existing groups and tags for suggestions
	hosts := s.ListHosts()
	groups := collectGroups(hosts)
	tags := collectTags(hosts)
	
	return &EditView{
		store:         s,
		host:          &models.Host{},
		mode:          "add",
		field:         fieldName,
		values:        make(map[string]string),
		errors:        make(map[string]string),
		saved:         false,
		existingGroups: groups,
		existingTags:  tags,
	}
}

// NewEditView creates a new edit view for editing a host
func NewEditView(s *store.FileStore, hostID string) (*EditView, error) {
	host, err := s.GetHost(hostID)
	if err != nil {
		return nil, err
	}
	
	// Collect existing groups and tags for suggestions
	hosts := s.ListHosts()
	groups := collectGroups(hosts)
	tags := collectTags(hosts)
	
	// Determine auth type from host
	authType := string(host.AuthType)
	if authType == "" {
		if host.Identity != "" {
			authType = string(AuthKey)
		} else if host.Password != "" {
			authType = string(AuthPassword)
		} else {
			authType = string(AuthAgent)
		}
	}
	
	return &EditView{
		store:         s,
		host:          &host,
		mode:          "edit",
		field:         fieldName,
		values: map[string]string{
			fieldName:     host.Name,
			fieldHost:     host.Host,
			fieldPort:     strconv.Itoa(host.Port),
			fieldUser:     host.User,
			fieldAuthType: authType,
			fieldIdentity: host.Identity,
			fieldProxy:    host.Proxy,
			fieldGroup:    host.Group,
			fieldTags:     joinTags(host.Tags),
		},
		securePassword: host.Password,
		passwordMasked: "••••••••",
		errors:         make(map[string]string),
		saved:          false,
		existingGroups: groups,
		existingTags:  tags,
	}, nil
}

func collectGroups(hosts []models.Host) []string {
	groupSet := make(map[string]bool)
	for _, h := range hosts {
		if h.Group != "" {
			groupSet[h.Group] = true
		}
	}
	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	return groups
}

func collectTags(hosts []models.Host) []string {
	tagSet := make(map[string]bool)
	for _, h := range hosts {
		for _, t := range h.Tags {
			tagSet[t] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
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
	// Handle file browser if active
	if v.showBrowser {
		return v.handleBrowserKey(msg)
	}
	
	// Handle password entry mode
	if v.enterPassword {
		return v.handlePasswordKey(msg)
	}
	
	switch msg.String() {
	case "up", "k":
		v.prevField()
	case "down", "j":
		v.nextField()
	case "left", "h":
		if v.field == fieldAuthType {
			v.values[fieldAuthType] = string(AuthPassword)
		} else if v.field == fieldIdentity {
			// Open file browser
			v.showBrowser = true
			homeDir := os.Getenv("HOME")
			sshDir := filepath.Join(homeDir, ".ssh")
			if _, err := os.Stat(sshDir); err == nil {
				v.fileBrowser = NewFileBrowser(sshDir)
			} else {
				v.fileBrowser = NewFileBrowser(homeDir)
			}
		} else if v.field == fieldPassword {
			// Enter password entry mode
			v.enterPassword = true
			v.securePassword = ""
			v.passwordMasked = ""
		}
	case "right", "l":
		if v.field == fieldAuthType {
			v.values[fieldAuthType] = string(AuthKey)
		}
	case "enter":
		if v.field == fieldGroup {
			// Toggle group selection from suggestions
			// For now, just save
		}
		return v, v.save()
	case "tab":
		v.nextField()
	case "b": // backspace
		if len(v.values[v.field]) > 0 {
			v.values[v.field] = v.values[v.field][:len(v.values[v.field])-1]
		}
		v.validate()
	case "esc":
		if v.showBrowser {
			v.showBrowser = false
			v.fileBrowser = nil
		} else {
			return v, tea.Quit
		}
	default:
		// Handle input for current field
		if len(msg.String()) == 1 {
			v.handleInput(msg.String())
		}
	}
	return v, nil
}

func (v *EditView) handleBrowserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		v.fileBrowser.Up()
	case "down", "j":
		v.fileBrowser.Down()
	case "enter":
		if v.fileBrowser.Select() {
			// It's a directory, navigate into it
			v.fileBrowser.path = v.fileBrowser.GetSelectedPath()
			v.fileBrowser.readDir()
			v.fileBrowser.selected = ""
		} else {
			// It's a file, select it
			v.values[fieldIdentity] = v.fileBrowser.GetSelectedPath()
			v.showBrowser = false
			v.fileBrowser = nil
		}
	case "esc":
		v.showBrowser = false
		v.fileBrowser = nil
	}
	return v, nil
}

// handlePasswordKey handles secure password entry
func (v *EditView) handlePasswordKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm password
		v.enterPassword = false
		// Set masked display
		if v.securePassword != "" {
			v.passwordMasked = "••••••••"
		}
		v.validate()
	case "esc":
		// Cancel password entry - restore previous value
		v.enterPassword = false
		if v.securePassword != "" {
			v.passwordMasked = "••••••••"
		}
	case "backspace":
		if len(v.securePassword) > 0 {
			v.securePassword = v.securePassword[:len(v.securePassword)-1]
			v.passwordMasked = v.passwordMasked[:len(v.passwordMasked)-1]
		}
	default:
		// Add character to password
		if len(msg.String()) == 1 {
			v.securePassword += msg.String()
			v.passwordMasked += "•"
		}
	}
	return v, nil
}

func (v *EditView) fields() []string {
	return []string{fieldName, fieldHost, fieldPort, fieldUser, fieldAuthType, fieldIdentity, fieldPassword, fieldProxy, fieldGroup, fieldTags}
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
	if v.field == fieldPort || v.field == fieldAuthType {
		return // Don't allow typing in these fields directly
	}
	v.values[v.field] += key
	v.validate()
}

func (v *EditView) validate() {
	v.errors = make(map[string]string)

	// Name validation
	if v.values[fieldName] == "" {
		v.errors[fieldName] = "Name is required"
	} else if len(v.values[fieldName]) > 50 {
		v.errors[fieldName] = "Name too long (max 50 chars)"
	}

	// Host validation
	if v.values[fieldHost] == "" {
		v.errors[fieldHost] = "Host is required"
	}

	// Port validation
	if v.values[fieldPort] == "" {
		v.errors[fieldPort] = "Port is required"
	} else {
		port, err := strconv.Atoi(v.values[fieldPort])
		if err != nil {
			v.errors[fieldPort] = "Port must be a number"
		} else if port < 1 || port > 65535 {
			v.errors[fieldPort] = "Port must be 1-65535"
		}
	}

	// User validation
	if v.values[fieldUser] == "" {
		v.errors[fieldUser] = "User is required"
	}

	// Auth type specific validation
	authType := v.values[fieldAuthType]
	if authType == string(AuthKey) && v.values[fieldIdentity] == "" {
		v.errors[fieldIdentity] = "Key file required for key auth"
	}
	if authType == string(AuthPassword) && v.securePassword == "" {
		v.errors[fieldPassword] = "Password required for password auth"
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

	// Parse tags
	tags := parseTags(v.values[fieldTags])

	// Parse auth type
	authType := models.AuthType(v.values[fieldAuthType])
	if authType == "" {
		if v.securePassword != "" {
			authType = models.AuthTypePassword
		} else if v.values[fieldIdentity] != "" {
			authType = models.AuthTypeKey
		} else {
			authType = models.AuthTypeAgent
		}
	}

	host := models.Host{
		Name:     v.values[fieldName],
		Host:     v.values[fieldHost],
		Port:     port,
		User:     v.values[fieldUser],
		Password: v.securePassword,
		Identity: v.values[fieldIdentity],
		AuthType: authType,
		Proxy:    v.values[fieldProxy],
		Group:    v.values[fieldGroup],
		Tags:     tags,
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

func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	var tags []string
	// Split by comma and clean up
	for _, t := range strings.Split(tagsStr, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// View renders the edit form
func (v *EditView) View() string {
	// Show password entry overlay
	if v.enterPassword {
		return v.renderPasswordEntry()
	}
	
	// Show file browser if active
	if v.showBrowser && v.fileBrowser != nil {
		return v.renderFileBrowser()
	}

	title := "Add Host"
	if v.mode == "edit" {
		title = "Edit Host"
	}

	header := BorderStyle.Width(60).Render(
		TitleStyle.Render(" "+title+" "),
	)

	var fields []string
	for _, f := range v.fields() {
		row := v.renderField(f)
		fields = append(fields, row)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, fields...)
	form := BorderStyle.Width(60).Render(body)

	help := HelpStyle.Render("↑↓ move | type to edit | ← select key file/password | enter: save | esc: cancel")

	return header + "\n\n" + form + "\n\n" + help
}

func (v *EditView) renderPasswordEntry() string {
	header := BorderStyle.Width(60).Render(
		TitleStyle.Render(" Enter Password "),
	)

	body := lipgloss.NewStyle().
		Width(56).
		Render("Password: " + v.passwordMasked + "_")

	form := BorderStyle.Width(60).Render(body)

	help := HelpStyle.Render("type to enter password | enter: confirm | esc: cancel")

	return header + "\n\n" + form + "\n\n" + help
}

func (v *EditView) renderField(f string) string {
	label := f
	value := v.values[f]
	
	switch f {
	case fieldPort:
		label = "Port"
		if value == "" {
			value = "22"
		}
	case fieldIdentity:
		label = "Identity File"
		if value == "" {
			value = "(default: ~/.ssh/id_rsa)"
		}
	case fieldPassword:
		label = "Password"
		if v.enterPassword {
			value = v.passwordMasked + "_"
		} else if v.securePassword != "" {
			value = "•••••••• (← to edit)"
		} else {
			value = "(empty) (← to set)"
		}
	case fieldProxy:
		label = "Proxy Jump"
	case fieldAuthType:
		label = "Auth Type"
		if value == "" {
			value = string(AuthKey)
		}
		value = "[ " + value + " ] (← → to change)"
	case fieldGroup:
		label = "Group"
	case fieldTags:
		label = "Tags"
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
		row += "\n    " + ErrorStyle.Render(v.errors[f])
	}
	
	// Show suggestions for group
	if f == fieldGroup && len(v.existingGroups) > 0 && value == "" {
		suggestions := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render(fmt.Sprintf("    Suggestions: %v", v.existingGroups[:min(3, len(v.existingGroups))]))
		row += "\n" + suggestions
	}

	return row
}

func (v *EditView) renderFileBrowser() string {
	header := BorderStyle.Width(60).Render(
		TitleStyle.Render(" Select SSH Key File "),
	)

	browser := v.fileBrowser.View(56)

	help := HelpStyle.Render("↑↓ navigate | enter: select | esc: cancel")

	return header + "\n\n" + browser + "\n\n" + help
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
