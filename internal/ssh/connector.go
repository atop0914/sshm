package ssh

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

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
func (c *Connector) Connect(host models.Host, profile models.Profile) error {
	config, err := c.buildClientConfig(host, profile)
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
func (c *Connector) ConnectWithAuth(host models.Host, profile models.Profile, auth AuthMethod) error {
	config, err := c.buildClientConfigWithAuth(host, profile, auth)
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

// buildClientConfig builds SSH client configuration based on host's AuthType
func (c *Connector) buildClientConfig(host models.Host, profile models.Profile) (*ssh.ClientConfig, error) {
	// Use the host's AuthType if specified
	authType := string(host.AuthType)

	switch authType {
	case string(models.AuthTypePassword):
		if host.Password != "" {
			return c.buildClientConfigWithAuth(host, profile, AuthMethodPassword)
		}
		// Fall through to try other methods if no password
		return nil, fmt.Errorf("password auth selected but no password set")

	case string(models.AuthTypeKey):
		if host.Identity != "" {
			return c.buildClientConfigWithAuth(host, profile, AuthMethodKeyFile)
		}
		// Try default keys if no identity specified
		return c.buildClientConfigWithAuth(host, profile, AuthMethodKeyFile)

	case string(models.AuthTypeAgent):
		return c.buildClientConfigWithAuth(host, profile, AuthMethodSSHAgent)

	default:
		// Legacy behavior: try all methods
		methods := []AuthMethod{AuthMethodPassword, AuthMethodSSHAgent, AuthMethodKeyFile}
		for _, method := range methods {
			config, err := c.buildClientConfigWithAuth(host, profile, method)
			if err == nil && len(config.Auth) > 0 {
				return config, nil
			}
		}
	}

	return nil, fmt.Errorf("no authentication method available")
}

// buildClientConfigWithAuth builds SSH client configuration with specific auth method
func (c *Connector) buildClientConfigWithAuth(host models.Host, profile models.Profile, auth AuthMethod) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(profile.Timeout) * time.Second,
	}

	switch auth {
	case AuthMethodPassword:
		if err := c.addPasswordAuth(config, host.Password); err != nil {
			return nil, err
		}

	case AuthMethodSSHAgent:
		// Try agent auth - may return nil if agent not available (graceful fallback)
		if err := c.addSSHAgentAuth(config); err != nil {
			return nil, err
		}
		// If no auth added (agent not available), try default keys as fallback
		if len(config.Auth) == 0 {
			if err := c.addDefaultKeysAuth(config); err != nil {
				return nil, err
			}
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
// Returns nil if agent is not available (graceful fallback)
func (c *Connector) addSSHAgentAuth(config *ssh.ClientConfig) error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		// Agent not available - return nil to allow fallback to other auth methods
		return nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		// Agent socket not accessible - return nil to allow fallback
		return nil
	}
	defer conn.Close()

	sshAgent := agent.NewClient(conn)
	signers, err := sshAgent.Signers()
	if err != nil || len(signers) == 0 {
		// No keys available from agent - return nil to allow fallback
		return nil
	}

	config.Auth = append(config.Auth, ssh.PublicKeys(signers...))
	return nil
}

// addPasswordAuth adds password authentication
func (c *Connector) addPasswordAuth(config *ssh.ClientConfig, password string) error {
	if password == "" {
		return fmt.Errorf("password is empty")
	}
	config.Auth = append(config.Auth, ssh.Password(password))
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

// LaunchSSH launches an external SSH process using the system ssh command
func LaunchSSH(host models.Host) error {
	// Build ssh command arguments
	args := []string{}
	
	// Add port if non-default
	if host.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", host.Port))
	}
	
	// Add identity file if specified
	if host.Identity != "" {
		expandedPath, err := expandPath(host.Identity)
		if err == nil {
			args = append(args, "-i", expandedPath)
		}
	}
	
	// Add user@host
	args = append(args, fmt.Sprintf("%s@%s", host.User, host.Host))
	
	// Execute the ssh command - use exec.LookPath to find ssh
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh command not found: %w", err)
	}
	
	// Use syscall.Exec to replace the current process
	// This gives control of the terminal to SSH
	err = syscall.Exec(sshPath, append([]string{"ssh"}, args...), os.Environ())
	if err != nil {
		return fmt.Errorf("failed to execute ssh: %w", err)
	}
	
	return nil
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
func ConnectAndInteract(host models.Host, profile models.Profile) error {
	connector := NewConnector()
	defer connector.Close()

	if err := connector.Connect(host, profile); err != nil {
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
