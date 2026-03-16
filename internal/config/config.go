package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sshm/sshm/internal/models"
)

// Config holds the entire application configuration
// Uses models.Host and models.SSHConfig for type consistency
type Config struct {
	Hosts    []models.Host      `json:"hosts" yaml:"hosts"`
	Configs  []models.SSHConfig  `json:"configs" yaml:"configs"`
	Profiles []models.Profile   `json:"profiles" yaml:"profiles"`
	Theme    string             `json:"theme" yaml:"theme"`
}

// GetProfile returns the profile for a host, falling back to default if not found
func (c *Config) GetProfile(host models.Host) models.Profile {
	// If host specifies a profile, look it up
	if host.Profile != "" {
		for _, p := range c.Profiles {
			if p.Name == host.Profile {
				return p
			}
		}
	}
	// Fall back to default profile
	return models.DefaultProfile()
}

// AddProfile adds a new profile to the configuration
func (c *Config) AddProfile(profile models.Profile) {
	// Remove existing profile with same name
	c.Profiles = removeProfile(c.Profiles, profile.Name)
	c.Profiles = append(c.Profiles, profile)
}

// RemoveProfile removes a profile by name
func (c *Config) RemoveProfile(name string) {
	c.Profiles = removeProfile(c.Profiles, name)
}

func removeProfile(profiles []models.Profile, name string) []models.Profile {
	result := make([]models.Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.Name != name {
			result = append(result, p)
		}
	}
	return result
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
// Supports both JSON and YAML formats based on file extension
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

	// Detect format from file extension
	if isYAML(path) {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			// Try legacy array format
			if legacyCfg := tryParseLegacyYAML(data); legacyCfg != nil {
				return legacyCfg, nil
			}
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	} else {
		// Try JSON first
		jsonErr := json.Unmarshal(data, &cfg)
		if jsonErr != nil {
			// Try legacy array format
			if legacyCfg := tryParseLegacyJSON(data); legacyCfg != nil {
				return legacyCfg, nil
			}
			// Try YAML as fallback
			yamlErr := yaml.Unmarshal(data, &cfg)
			if yamlErr != nil {
				return nil, fmt.Errorf("failed to parse config (JSON: %v, YAML: %v)", jsonErr, yamlErr)
			}
		}
	}

	return &cfg, nil
}

// tryParseLegacyJSON tries to parse a legacy JSON array format
func tryParseLegacyJSON(data []byte) *Config {
	var hosts []models.Host
	if err := json.Unmarshal(data, &hosts); err == nil && len(hosts) > 0 {
		return &Config{Hosts: hosts}
	}
	return nil
}

// tryParseLegacyYAML tries to parse a legacy YAML array format
func tryParseLegacyYAML(data []byte) *Config {
	var hosts []models.Host
	if err := yaml.Unmarshal(data, &hosts); err == nil && len(hosts) > 0 {
		return &Config{Hosts: hosts}
	}
	return nil
}

// isYAML returns true if the file path has a .yaml or .yml extension
func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// SaveConfig saves configuration to the specified path
// Supports both JSON and YAML formats based on file extension
// If path is empty, uses default path
func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		path = GetDefaultConfigPath()
	}

	var data []byte
	var err error

	// Detect format from file extension
	if isYAML(path) {
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML config: %w", err)
		}
	} else {
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
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
