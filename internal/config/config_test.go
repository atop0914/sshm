package config

import (
	"os"
	"testing"
)

func getTestFilePath(filename string) string {
	// Project root is 2 levels up from internal/config/
	return "../../" + filename
}

func TestLoadConfigYAML(t *testing.T) {
	// Test loading YAML file
	path := getTestFilePath("test_hosts.yaml")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	if len(cfg.Hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(cfg.Hosts))
	}
}

func TestLoadConfigJSON(t *testing.T) {
	// Test loading JSON file
	path := getTestFilePath("test_hosts.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Failed to load JSON config: %v", err)
	}

	if len(cfg.Hosts) != 5 {
		t.Errorf("Expected 5 hosts, got %d", len(cfg.Hosts))
	}
}

func TestSaveConfigYAML(t *testing.T) {
	// Load existing config
	path := getTestFilePath("test_hosts.yaml")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Save as YAML
	outPath := getTestFilePath("test_output.yaml")
	err = SaveConfig(cfg, outPath)
	if err != nil {
		t.Fatalf("Failed to save YAML config: %v", err)
	}
	defer os.Remove(outPath)

	// Verify it can be loaded back
	cfg2, err := LoadConfig(outPath)
	if err != nil {
		t.Fatalf("Failed to load saved YAML config: %v", err)
	}

	if len(cfg2.Hosts) != len(cfg.Hosts) {
		t.Errorf("Host count mismatch: %d vs %d", len(cfg2.Hosts), len(cfg.Hosts))
	}
}

func TestSaveConfigJSON(t *testing.T) {
	// Load existing config
	path := getTestFilePath("test_hosts.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Save as JSON
	outPath := getTestFilePath("test_output.json")
	err = SaveConfig(cfg, outPath)
	if err != nil {
		t.Fatalf("Failed to save JSON config: %v", err)
	}
	defer os.Remove(outPath)

	// Verify it can be loaded back
	cfg2, err := LoadConfig(outPath)
	if err != nil {
		t.Fatalf("Failed to load saved JSON config: %v", err)
	}

	if len(cfg2.Hosts) != len(cfg.Hosts) {
		t.Errorf("Host count mismatch: %d vs %d", len(cfg2.Hosts), len(cfg.Hosts))
	}
}
