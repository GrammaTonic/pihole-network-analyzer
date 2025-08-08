package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pihole-analyzer/internal/analyzer"
	"pihole-analyzer/internal/cli"
	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/network"
	"pihole-analyzer/internal/reporting"
	sshpkg "pihole-analyzer/internal/ssh"
	"pihole-analyzer/internal/types"
)

// quietPrintf prints only if quiet mode is not enabled
func quietPrintf(quiet bool, format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format, args...)
	}
}

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

	// Handle test mode first
	if *flags.Test {
		fmt.Println(colors.Header("🧪 Running Test Mode"))
		fmt.Println("Using mock Pi-hole database")

		// Analyze the test data using Pi-hole mock database
		dbFile := filepath.Join("test_data", "mock_pihole.db")
		clientStats, err := sshpkg.AnalyzePiholeDatabase(dbFile)
		if err != nil {
			log.Fatalf("Error analyzing test data: %v", err)
		}

		fmt.Println(colors.Success("✅ Test mode analysis completed"))
		reporting.DisplayResultsWithConfig(clientStats, cfg)
		fmt.Println(colors.Info("Test mode completed successfully"))
		return
	}

	if cfg.TestMode {
		quietPrintf(cfg.Quiet, "🧪 Test Mode Enabled - Using Mock Data\n")
	}

	// Validate input - now requires Pi-hole config or special modes
	if err := cli.ValidateInput(flags); err != nil {
		fmt.Printf("Error: %v\n", err)
		cli.ShowUsage()
		os.Exit(1)
	}

	// Print startup information
	cli.PrintStartupInfo(flags, cfg)

	// Handle Pi-hole analysis (now the only mode)
	configFile := *flags.Pihole
	if configFile == "" {
		fmt.Printf("Error: Pi-hole configuration required. Use --pihole <config.json> or --pihole-setup\n")
		cli.ShowUsage()
		os.Exit(1)
	}

	fmt.Printf("Connecting to Pi-hole using config: %s\n", colors.Info(configFile))

	var clientStats map[string]*types.ClientStats

	if cfg.TestMode {
		// Use mock Pi-hole database in test mode
		dbFile := filepath.Join("test_data", "mock_pihole.db")
		clientStats, err = sshpkg.AnalyzePiholeDatabase(dbFile)
	} else {
		clientStats, err = analyzer.AnalyzePiholeData(configFile, cfg)
	}

	if err != nil {
		log.Fatalf("Error analyzing Pi-hole data: %v", err)
	}

	// Check ARP status for all clients
	if cfg.TestMode {
		fmt.Println("Mock ARP status check skipped in test mode")
	} else {
		err = network.CheckARPStatus(clientStats)
	}
	if err != nil {
		fmt.Printf("%s: Could not check ARP status: %v\n", colors.Warning("Warning"), err)
	}

	reporting.DisplayResultsWithConfig(clientStats, cfg)
}
