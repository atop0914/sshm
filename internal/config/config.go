package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Hosts  []Host  `json:"hosts"`
	Config []SSHConfig `json:"configs"`
}

type Host struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Identity string `json:"identity,omitempty"`
	Proxy    string `json:"proxy,omitempty"`
}

type SSHConfig struct {
	Name           string `json:"name"`
	IdentityFile   string `json:"identity_file,omitempty"`
	ProxyCommand   string `json:"proxy_command,omitempty"`
	ForwardAgent   bool   `json:"forward_agent,omitempty"`
	ServerAliveInterval int `json:"server_alive_interval,omitempty"`
}

func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sshm.json"
	}
	return filepath.Join(home, ".sshm.json")
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = GetDefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		path = GetDefaultConfigPath()
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
