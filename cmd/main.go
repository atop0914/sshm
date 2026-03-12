package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sshm/sshm/internal/config"
	"github.com/sshm/sshm/internal/models"
	"github.com/sshm/sshm/internal/tui"
	"gopkg.in/yaml.v3"
)

var (
	exportFormat string
	outputFile   string
)

func init() {
	flag.StringVar(&exportFormat, "format", "json", "Export format: json, yaml, ssh")
	flag.StringVar(&outputFile, "o", "", "Output file (stdout if empty)")
	flag.Usage = func() {
		fmt.Println("Usage: sshm export [options]")
		fmt.Println("")
		fmt.Println("Export hosts to various formats")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

func main() {
	// Check first arg before full parsing
	if len(os.Args) > 1 && os.Args[1] == "export" {
		// Filter out "export" subcommand from args for flag parsing
		filteredArgs := make([]string, 0, len(os.Args)-1)
		filteredArgs = append(filteredArgs, os.Args[0])
		filteredArgs = append(filteredArgs, os.Args[2:]...)
		os.Args = filteredArgs
		flag.Parse()
		runExport()
		return
	}

	// Original TUI mode
	runTUI()
}

func runExport() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var output []byte

	switch exportFormat {
	case "json":
		output, err = exportToJSON(cfg)
	case "yaml":
		output, err = exportToYAML(cfg)
	case "ssh":
		output, err = exportToSSHConfig(cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (use json, yaml, or ssh)\n", exportFormat)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported to %s\n", outputFile)
	} else {
		fmt.Print(string(output))
	}
}

func exportToJSON(cfg *config.Config) ([]byte, error) {
	type exportConfig struct {
		Hosts    []models.Host      `json:"hosts"`
		Configs  []models.SSHConfig `json:"configs"`
		Profiles []models.Profile   `json:"profiles"`
	}

	exp := exportConfig{
		Hosts:    cfg.Hosts,
		Configs:  cfg.Configs,
		Profiles: cfg.Profiles,
	}

	return json.MarshalIndent(exp, "", "  ")
}

func exportToYAML(cfg *config.Config) ([]byte, error) {
	type exportConfig struct {
		Hosts    []models.Host      `yaml:"hosts"`
		Configs  []models.SSHConfig `yaml:"configs"`
		Profiles []models.Profile   `yaml:"profiles"`
	}

	exp := exportConfig{
		Hosts:    cfg.Hosts,
		Configs:  cfg.Configs,
		Profiles: cfg.Profiles,
	}

	return yaml.Marshal(exp)
}

func exportToSSHConfig(cfg *config.Config) ([]byte, error) {
	var lines []string

	// Export hosts as SSH config
	for _, host := range cfg.Hosts {
		lines = append(lines, formatSSHHost(&host)...)
	}

	return []byte(joinLines(lines)), nil
}

func formatSSHHost(host *models.Host) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Host %s", host.Name))

	if host.Host != "" {
		lines = append(lines, fmt.Sprintf("    HostName %s", host.Host))
	}

	if host.Port != 0 && host.Port != 22 {
		lines = append(lines, fmt.Sprintf("    Port %d", host.Port))
	}

	if host.User != "" {
		lines = append(lines, fmt.Sprintf("    User %s", host.User))
	}

	if host.Identity != "" {
		lines = append(lines, fmt.Sprintf("    IdentityFile %s", host.Identity))
	}

	if host.Proxy != "" {
		lines = append(lines, fmt.Sprintf("    ProxyJump %s", host.Proxy))
	}

	// Add group as comment
	if host.Group != "" {
		lines = append(lines, fmt.Sprintf("    # Group: %s", host.Group))
	}

	// Add tags as comment
	if len(host.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("    # Tags: %s", joinStrings(host.Tags, ", ")))
	}

	lines = append(lines, "") // Empty line between hosts

	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func joinStrings(items []string, sep string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}

func runTUI() {
	fmt.Println("SSH Host Manager (sshm)")
	fmt.Println("========================")

	// Load configuration from default path
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d hosts, %d configs\n", len(cfg.Hosts), len(cfg.Configs))
	fmt.Printf("Default config path: %s\n", config.GetDefaultConfigPath())

	// Run TUI
	fmt.Println("\nStarting TUI...")
	if err := tui.Run(config.GetDefaultConfigPath()); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
