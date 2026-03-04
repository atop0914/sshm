package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sshm/sshm/internal/models"
)

// SSHConfigParser parses SSH config files (~/.ssh/config)
type SSHConfigParser struct{}

// NewSSHConfigParser creates a new SSH config parser
func NewSSHConfigParser() *SSHConfigParser {
	return &SSHConfigParser{}
}

// ParseSSHConfig reads and parses ~/.ssh/config file
func (p *SSHConfigParser) ParseSSHConfig(path string) ([]models.Host, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, ".ssh", "config")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Host{}, nil
		}
		return nil, fmt.Errorf("failed to read SSH config: %w", err)
	}

	return p.ParseConfigString(string(data))
}

// ParseConfigString parses SSH config from a string
func (p *SSHConfigParser) ParseConfigString(content string) ([]models.Host, error) {
	var hosts []models.Host
	var currentHost *parsedSSHHost

	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Match Host directive
		if strings.HasPrefix(line, "Host ") {
			// Save previous host if exists
			if currentHost != nil && currentHost.name != "" {
				host := currentHost.toModel()
				if host.Name != "" {
					hosts = append(hosts, host)
				}
			}
			currentHost = &parsedSSHHost{
				name: strings.TrimSpace(strings.TrimPrefix(line, "Host ")),
			}
			continue
		}

		if currentHost == nil {
			continue
		}

		// Parse host options
		if strings.HasPrefix(line, "HostName ") {
			currentHost.hostname = strings.TrimSpace(strings.TrimPrefix(line, "HostName "))
		} else if strings.HasPrefix(line, "Port ") {
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(line, "Port ")), "%d", &currentHost.port)
		} else if strings.HasPrefix(line, "User ") {
			currentHost.user = strings.TrimSpace(strings.TrimPrefix(line, "User "))
		} else if strings.HasPrefix(line, "IdentityFile ") {
			currentHost.identityFile = expandHome(strings.TrimSpace(strings.TrimPrefix(line, "IdentityFile ")))
		} else if strings.HasPrefix(line, "ProxyCommand ") {
			currentHost.proxyCommand = strings.TrimSpace(strings.TrimPrefix(line, "ProxyCommand "))
		} else if strings.HasPrefix(line, "ProxyJump ") {
			currentHost.proxyJump = strings.TrimSpace(strings.TrimPrefix(line, "ProxyJump "))
		} else if strings.HasPrefix(line, "ForwardAgent ") {
			currentHost.forwardAgent = strings.TrimSpace(strings.TrimPrefix(line, "ForwardAgent ")) == "yes"
		} else if strings.HasPrefix(line, "ServerAliveInterval ") {
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(line, "ServerAliveInterval ")), "%d", &currentHost.serverAliveInterval)
		} else if strings.HasPrefix(line, "Match ") {
			// Skip Match blocks as they are conditional
			continue
		}
	}

	// Don't forget the last host
	if currentHost != nil && currentHost.name != "" {
		host := currentHost.toModel()
		if host.Name != "" {
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}

// parsedSSHHost represents a parsed SSH host entry
type parsedSSHHost struct {
	name                string
	hostname            string
	port                int
	user                string
	identityFile        string
	proxyCommand        string
	proxyJump           string
	forwardAgent        bool
	serverAliveInterval int
}

func (h *parsedSSHHost) toModel() models.Host {
	host := h.name
	
	// Use HostName if different from Host alias
	if h.hostname != "" && h.hostname != h.name {
		host = h.hostname
	}

	port := h.port
	if port == 0 {
		port = 22
	}

	return models.Host{
		ID:        uuid.New().String(),
		Name:      h.name,
		Host:      host,
		Port:      port,
		User:      h.user,
		Identity:  h.identityFile,
		Proxy:     h.proxyJump,
		Group:     "imported",
	}
}

// expandHome expands ~ to the user's home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	// Handle $HOME style
	re := regexp.MustCompile(`\$HOME`)
	return re.ReplaceAllString(path, os.Getenv("HOME"))
}

// ImportFromSSHConfig imports hosts from ~/.ssh/config
func ImportFromSSHConfig(storePath string) ([]models.Host, error) {
	parser := NewSSHConfigParser()
	hosts, err := parser.ParseSSHConfig("")
	if err != nil {
		return nil, err
	}

	// If store path provided, also load existing hosts to avoid duplicates
	if storePath != "" {
		existingCfg, err := LoadConfig(storePath)
		if err == nil {
			// Filter out hosts that already exist
			existingHosts := make(map[string]bool)
			for _, h := range existingCfg.Hosts {
				existingHosts[h.Name] = true
			}
			
			var newHosts []models.Host
			for _, h := range hosts {
				if !existingHosts[h.Name] {
					newHosts = append(newHosts, h)
				}
			}
			hosts = newHosts
		}
	}

	return hosts, nil
}
