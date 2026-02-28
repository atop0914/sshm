package ssh

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/sshm/sshm/internal/models"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// AuthMethod represents the authentication method for SSH
type AuthMethod int

const (
	AuthMethodNone AuthMethod = iota
	AuthMethodPassword
	AuthMethodKeyFile
	AuthMethodSSHAgent
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

// ConnectWithAuth connects using specified auth method
func (c *Connector) ConnectWithAuth(host models.Host, auth AuthMethod) error {
	config, err := c.buildClientConfigWithAuth(host, auth)
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
	// Try SSH agent first, then key file, then default keys
	methods := []AuthMethod{AuthMethodSSHAgent, AuthMethodKeyFile}

	for _, method := range methods {
		config, err := c.buildClientConfigWithAuth(host, method)
		if err == nil && len(config.Auth) > 0 {
			return config, nil
		}
	}

	return nil, fmt.Errorf("no authentication method available")
}

// buildClientConfigWithAuth builds SSH client configuration with specific auth method
func (c *Connector) buildClientConfigWithAuth(host models.Host, auth AuthMethod) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	switch auth {
	case AuthMethodPassword:
		// Password auth not implemented - would need to prompt user
		return config, fmt.Errorf("password authentication not implemented")

	case AuthMethodSSHAgent:
		if err := c.addSSHAgentAuth(config); err != nil {
			return nil, err
		}

	case AuthMethodKeyFile, AuthMethodNone:
		if host.Identity != "" {
			if err := c.addKeyFileAuth(config, host.Identity); err != nil {
				return nil, err
			}
		} else {
			// Try default SSH keys
			if err := c.addDefaultKeysAuth(config); err != nil {
				return nil, err
			}
		}
	}

	if len(config.Auth) == 0 {
		return nil, fmt.Errorf("no authentication method available")
	}

	return config, nil
}

// addSSHAgentAuth adds SSH agent authentication
func (c *Connector) addSSHAgentAuth(config *ssh.ClientConfig) error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	sshAgent := agent.NewClient(conn)
	signers, err := sshAgent.Signers()
	if err != nil || len(signers) == 0 {
		return fmt.Errorf("no keys available from SSH agent: %w", err)
	}

	config.Auth = append(config.Auth, ssh.PublicKeys(signers...))
	return nil
}

// addKeyFileAuth adds key file authentication
func (c *Connector) addKeyFileAuth(config *ssh.ClientConfig, keyPath string) error {
	expandedPath, err := expandPath(keyPath)
	if err != nil {
		return fmt.Errorf("failed to expand identity path: %w", err)
	}

	key, err := os.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read identity file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	return nil
}

// addDefaultKeysAuth adds default SSH key authentication
func (c *Connector) addDefaultKeysAuth(config *ssh.ClientConfig) error {
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

	if len(config.Auth) == 0 {
		return fmt.Errorf("no authentication method available")
	}

	return nil
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
func getTerminalSize() (width, height int) {
	width, height, err := getTerminalSizeImpl()
	if err != nil {
		return 80, 24 // default
	}
	return width, height
}

// getTerminalSizeImpl gets terminal size using environment or defaults
func getTerminalSizeImpl() (int, int, error) {
	// Try to use environment variables first (for common cases)
	if w := os.Getenv("COLUMNS"); w != "" {
		if h := os.Getenv("LINES"); h != "" {
			width, err1 := strconv.Atoi(w)
			height, err2 := strconv.Atoi(h)
			if err1 == nil && err2 == nil {
				return width, height, nil
			}
		}
	}

	// Fallback to default
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

// Ping checks if the host is reachable (TCP only)
func Ping(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5e9) // 5 second timeout
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
