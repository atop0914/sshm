package main

import (
	"fmt"
	"os"

	"github.com/sshm/sshm/internal/config"
	"github.com/sshm/sshm/internal/tui"
)

func main() {
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

	// If no TTY, just print info and exit
	if !isTerminal() {
		fmt.Println("\nNo terminal detected, skipping TUI.")
		return
	}

	// Run TUI
	fmt.Println("\nStarting TUI...")
	if err := tui.Run(config.GetDefaultConfigPath()); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func isTerminal() bool {
	return isInteractive()
}

// Placeholder for terminal detection
func isInteractive() bool {
	return true // Always try to run TUI for testing
}
