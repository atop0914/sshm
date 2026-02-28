package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sshm/sshm/internal/models"
)

// Config holds the entire application configuration
// Uses models.Host and models.SSHConfig for type consistency
type Config struct {
	Hosts   []models.Host     `json:"hosts" yaml:"hosts"`
	Configs []models.SSHConfig `json:"configs" yaml:"configs"`
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sshm.json"
	}
	return filepath.Join(home, ".sshm.json")
}

// LoadConfig loads configuration from the specified path
// If path is empty, uses default path
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
	// Try JSON first
	if err := json.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}

	// Fallback to YAML (requires gopkg.in/yaml.v3)
	// For now, return error if JSON fails
	return nil, fmt.Errorf("failed to parse config: %w", err)
}

// SaveConfig saves configuration to the specified path
// If path is empty, uses default path
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

// EnsureDir ensures the directory for the config file exists
func EnsureConfigDir(path string) error {
	if path == "" {
		path = GetDefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0700)
}
