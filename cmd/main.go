package main

import (
	"fmt"
	"log"

	"github.com/sshm/sshm/internal/config"
)

func main() {
	fmt.Println("SSH Host Manager (sshm)")
	fmt.Println("========================")

	// Load configuration from default path
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded %d hosts, %d configs\n", len(cfg.Hosts), len(cfg.Config))
	fmt.Printf("Default config path: %s\n", config.GetDefaultConfigPath())

	fmt.Println("\nsshm initialized successfully!")
}
