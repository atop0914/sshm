# SSH Host Manager (sshm)

A simple SSH host manager that loads host configurations from a JSON file.

## Installation

```bash
go build -o sshm ./cmd
```

## Configuration

Create a `~/.sshm.json` file with your SSH hosts:

```json
{
  "hosts": [
    {
      "name": "server1",
      "host": "192.168.1.100",
      "port": 22,
      "user": "admin",
      "identity": "~/.ssh/id_rsa"
    }
  ],
  "configs": [
    {
      "name": "default",
      "identity_file": "~/.ssh/id_rsa",
      "server_alive_interval": 60
    }
  ]
}
```

## Usage

```bash
# Run the application
./sshm
```

The application loads configuration from `~/.sshm.json` by default.

## Project Structure

```
sshm/
├── cmd/
│   └── main.go       # Entry point
└── internal/
    ├── config/       # Configuration loading/saving
    └── models/       # Data models
```

## License

MIT
