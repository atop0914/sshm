# SSH Host Manager (sshm)

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Platforms-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=for-the-badge" alt="Platforms">
</p>

A simple and elegant SSH host manager with a beautiful terminal user interface (TUI). Manage your SSH hosts with ease using keyboard navigation.

## Features

- 📋 **Host Management** - Add, edit, delete, and search SSH hosts
- 🔍 **Quick Search** - Filter hosts by name, host, user, group, or tags
- 🎨 **Beautiful TUI** - Terminal interface built with Bubble Tea
- 🔐 **Secure** - Supports key-based SSH authentication
- 💾 **Persistent Storage** - JSON-based local storage
- ⌨️ **Full Keyboard Control** - Navigate and manage without leaving your terminal
- 📥 **SSH Config Import** - Import hosts directly from `~/.ssh/config`
- 📊 **Connection History** - Track your connection attempts and statistics
- 🏷️ **Groups & Tags** - Organize hosts with groups and tags

## Installation

### From Source

```bash
git clone https://github.com/atop0914/sshm.git
cd sshm
go build -o sshm ./cmd
```

### Pre-built Binaries

Download from the [Releases](https://github.com/atop0914/sshm/releases) page.

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux    | x86_64      | `sshm-linux-amd64` |
| macOS    | x86_64      | `sshm-darwin-amd64` |
| macOS    | ARM64       | `sshm-darwin-arm64` |
| Windows  | x86_64      | `sshm-windows-amd64.exe` |

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
   - Press `Enter` to connect

4. **Import from SSH config:**
   - Press `i` to import hosts from `~/.ssh/config`

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑↓` or `j/k` | Navigate host list |
| `Enter` | Connect to selected host |
| `a` | Add new host |
| `e` | Edit selected host |
| `d` | View host details |
| `h` | View connection history (all) |
| `H` | View history for selected host |
| `/` | Filter/search hosts |
| `i` | Import from SSH config |
| `?` | Show help |
| `esc` | Clear filter / Go back |
| `q` / `Ctrl+C` | Quit application |

## Configuration

The application stores hosts in `~/.sshm.json`:

```json
{
  "hosts": [
    {
      "id": "uuid",
      "name": "production",
      "host": "192.168.1.100",
      "port": 22,
      "user": "admin",
      "identity": "~/.ssh/id_rsa",
      "proxy": "",
      "group": "production",
      "tags": ["web", "production"],
      "connection_count": 0
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
| group | No | Group name for organization |
| tags | No | Array of tags |

### SSH Config Import

The SSH config parser supports standard SSH config directives:

```
Host myserver
    HostName 192.168.1.100
    Port 22
    User admin
    IdentityFile ~/.ssh/id_rsa
    ForwardAgent yes
    ServerAliveInterval 60
```

Imported hosts are automatically tagged as "imported" and can be reorganized as needed.

## Connection History

Connection attempts are tracked in `~/.sshm_history.json`:

```json
[
  {
    "host_id": "uuid",
    "timestamp": "2024-01-15T10:30:00Z",
    "success": true,
    "error": "",
    "duration_ms": 150
  }
]
```

Statistics are displayed in the host details view (`d` key).

## Project Structure

```
sshm/
├── cmd/
│   └── main.go           # Entry point
└── internal/
    ├── config/           # Configuration loading & SSH config parsing
    ├── models/           # Data models
    ├── store/            # Data persistence
    ├── ssh/              # SSH connection
    └── tui/              # Terminal UI
        ├── app.go        # Main application
        ├── style.go      # Styling definitions
        ├── list.go       # Host list view
        ├── edit.go       # Add/Edit form
        ├── history.go    # Connection history view
        └── help.go       # Help/usage view
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

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o sshm-windows-amd64.exe ./cmd
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH client
- [Google UUID](https://github.com/google/uuid) - UUID generation

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
