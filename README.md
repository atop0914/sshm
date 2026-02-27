# SSH Host Manager (sshm)

A simple and elegant SSH host manager with a terminal user interface (TUI). Manage your SSH hosts with ease using keyboard navigation.

## Features

- 📋 **Host Management** - Add, edit, delete, and search SSH hosts
- 🔍 **Quick Search** - Filter hosts by name, host, or user
- 🎨 **Beautiful TUI** - Terminal interface built with Bubble Tea
- 🔐 **Secure** - Supports key-based SSH authentication
- 💾 **Persistent Storage** - JSON-based local storage
- ⌨️ **Keyboard Navigation** - Full keyboard control

## Installation

### From Source

```bash
git clone https://github.com/atop0914/sshm.git
cd sshm
go build -o sshm ./cmd
```

### Pre-built Binaries

Download from the [Releases](https://github.com/atop0914/sshm/releases) page.

## Quick Start

1. **Run the application:**
   ```bash
   ./sshm
   ```

2. **Add your first host:**
   - Press `a` to open the add host form
   - Fill in the details (name, host, port, user, identity file)
   - Press `Enter` to save

3. **Connect to a host:**
   - Use `↑↓` or `j/k` to navigate
   - Press `Enter` to connect (coming soon)

## Configuration

The application stores hosts in `~/.sshm.json`:

```json
{
  "hosts": [
    {
      "name": "production",
      "host": "192.168.1.100",
      "port": 22,
      "user": "admin",
      "identity": "~/.ssh/id_rsa"
    }
  ]
}
```

### Host Fields

| Field | Required | Description |
|-------|----------|-------------|
| name | Yes | Display name for the host |
| host | Yes | IP address or hostname |
| port | No | SSH port (default: 22) |
| user | Yes | SSH username |
| identity | No | Path to SSH private key |
| proxy | No | Proxy jump host |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑↓` or `j/k` | Navigate host list |
| `a` | Add new host |
| `e` | Edit selected host |
| `d` | View host details |
| `/` | Search/filter hosts |
| `Enter` | Connect to host |
| `q` / `Ctrl+C` | Quit |

## Project Structure

```
sshm/
├── cmd/
│   └── main.go           # Entry point
└── internal/
    ├── config/           # Configuration loading
    ├── models/           # Data models
    ├── store/            # Data persistence
    ├── ssh/              # SSH connection
    └── tui/              # Terminal UI
        ├── app.go        # Main application
        ├── style.go      # Styling definitions
        ├── list.go       # Host list view
        └── edit.go       # Add/Edit form
```

## Development

### Build

```bash
# Build for current platform
go build -o sshm ./cmd

# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o sshm-linux-amd64 ./cmd

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o sshm-darwin-amd64 ./cmd
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH client

## License

MIT
