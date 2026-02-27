package ssh

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sshm/sshm/internal/models"
	"golang.org/x/crypto/ssh"
)

// Connector handles SSH connections
type Connector struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

// NewConnector creates a new SSH connector
func NewConnector() *Connector {
	return &Connector{}
}

// Connect establishes an SSH connection to the host
func (c *Connector) Connect(host models.Host) error {
	config, err := c.buildClientConfig(host)
	if err != nil {
		return fmt.Errorf("failed to build client config: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", host.Host, host.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.client = client
	c.config = config
	return nil
}

// buildClientConfig builds SSH client configuration
func (c *Connector) buildClientConfig(host models.Host) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Try password authentication first (if we had a password)
	// For now, we'll use key-based auth

	// Key-based authentication
	if host.Identity != "" {
		keyPath, err := expandPath(host.Identity)
		if err != nil {
			return nil, fmt.Errorf("failed to expand identity path: %w", err)
		}

		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read identity file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	} else {
		// Try default SSH keys
		defaultKeys := []string{
			"~/.ssh/id_ed25519",
			"~/.ssh/id_rsa",
			"~/.ssh/id_ecdsa",
			"~/.ssh/id_dsa",
		}

		for _, keyPath := range defaultKeys {
			expandedPath, err := expandPath(keyPath)
			if err != nil {
				continue
			}

			key, err := os.ReadFile(expandedPath)
			if err != nil {
				continue
			}

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				continue
			}

			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	if len(config.Auth) == 0 {
		return nil, fmt.Errorf("no authentication method available")
	}

	return config, nil
}

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if len(path) > 1 && path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}
	return path, nil
}

// IsConnected returns whether the connector has an active connection
func (c *Connector) IsConnected() bool {
	return c.client != nil
}

// GetClient returns the underlying SSH client
func (c *Connector) GetClient() *ssh.Client {
	return c.client
}

// Close closes the SSH connection
func (c *Connector) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ConnectAndInteract connects to host and starts an interactive session
func ConnectAndInteract(host models.Host) error {
	connector := NewConnector()
	defer connector.Close()

	if err := connector.Connect(host); err != nil {
		return err
	}

	session, err := connector.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up terminal
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Get terminal dimensions
	width, height := getTerminalSize()
	err = session.RequestPty("xterm", height, width, modes)
	if err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %w", err)
	}

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return session.Wait()
}

// getTerminalSize returns the terminal width and height
func getTerminalSize() (int, int) {
	width, height, err := terminalSize()
	if err != nil {
		return 80, 24 // default
	}
	return width, height
}

// terminalSize is a placeholder - in real implementation would use
// golang.org/x/sys/unix or winapi to get terminal size
func terminalSize() (int, int, error) {
	return 80, 24, nil
}

// CheckConnection tests if a connection can be established
func CheckConnection(host models.Host) error {
	connector := NewConnector()
	defer connector.Close()

	// Just test TCP connectivity first
	addr := fmt.Sprintf("%s:%d", host.Host, host.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", addr, err)
	}
	conn.Close()

	return nil
}
