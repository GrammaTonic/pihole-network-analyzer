package main

import (
	"fmt"
	"log"
	"os"

	"pihole-analyzer/internal/analyzer"
	"pihole-analyzer/internal/cli"
	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/network"
	"pihole-analyzer/internal/reporting"
	sshpkg "pihole-analyzer/internal/ssh"
)

func main() {
	// Parse command-line flags using CLI package
	flags := cli.ParseFlags()

	// Handle special flags that should exit immediately
	if cli.HandleSpecialFlags(flags) {
		if *flags.PiholeSetup {
			sshpkg.SetupPiholeConfig()
		}
		return
	}

	// Load configuration
	configPath := config.GetConfigPath()
	if *flags.Config != "" {
		configPath = *flags.Config
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Apply command-line flags to configuration
	cli.ApplyFlags(flags, cfg)

	// Validate input - requires Pi-hole config
	if err := cli.ValidateInput(flags); err != nil {
		fmt.Printf("Error: %v\n", err)
		cli.ShowUsage()
		os.Exit(1)
	}

	// Print startup information
	cli.PrintStartupInfo(flags, cfg)

	// Handle Pi-hole analysis
	configFile := *flags.Pihole
	if configFile == "" {
		fmt.Printf("Error: Pi-hole configuration required. Use --pihole <config.json> or --pihole-setup\n")
		cli.ShowUsage()
		os.Exit(1)
	}

	fmt.Printf("Connecting to Pi-hole using config: %s\n", colors.Info(configFile))

	// Analyze Pi-hole data
	clientStats, err := analyzer.AnalyzePiholeData(configFile, cfg)
	if err != nil {
		log.Fatalf("Error analyzing Pi-hole data: %v", err)
	}

	// Check ARP status for all clients
	err = network.CheckARPStatus(clientStats)
	if err != nil {
		fmt.Printf("%s: Could not check ARP status: %v\n", colors.Warning("Warning"), err)
	}

	reporting.DisplayResultsWithConfig(clientStats, cfg)
}
