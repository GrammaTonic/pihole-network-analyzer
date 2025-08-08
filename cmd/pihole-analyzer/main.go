package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pihole-network-analyzer/internal/analyzer"
	"pihole-network-analyzer/internal/cli"
	"pihole-network-analyzer/internal/colors"
	"pihole-network-analyzer/internal/config"
	"pihole-network-analyzer/internal/network"
	"pihole-network-analyzer/internal/reporting"
	sshpkg "pihole-network-analyzer/internal/ssh"
	"pihole-network-analyzer/internal/types"
)

// Command-line flags are now handled by the CLI package

// quietPrintf prints only if quiet mode is not enabled
func quietPrintf(quiet bool, format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format, args...)
	}
}

func main() {
	// Parse command-line flags using CLI package
	flags := cli.ParseFlags()
	csvFile := cli.GetCSVFile()

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
		fmt.Println("Using mock data from testdata/test.csv")

		// Use the existing test CSV file
		testCSV := "testdata/test.csv"
		if _, err := os.Stat(testCSV); os.IsNotExist(err) {
			fmt.Printf("Test file not found: %s\n", testCSV)
			return
		}

		// Analyze the test data
		clientStats, err := analyzer.AnalyzeDNSDataWithConfig(testCSV, cfg)
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

	// Validate input
	if err := cli.ValidateInput(flags, csvFile); err != nil {
		fmt.Printf("Error: %v\n", err)
		cli.ShowUsage()
		os.Exit(1)
	}

	// Print startup information
	cli.PrintStartupInfo(flags, cfg)

	// Handle Pi-hole analysis
	if *flags.Pihole != "" {
		configFile := *flags.Pihole
		fmt.Printf("Connecting to Pi-hole using config: %s\n", colors.Info(configFile))

		var clientStats map[string]*types.ClientStats
		var err error

		if cfg.TestMode {
			// Use mock Pi-hole database in test mode
			dbFile := filepath.Join("test_data", "mock_pihole.db")
			clientStats, err = sshpkg.AnalyzePiholeDatabase(dbFile)
		} else {
			clientStats, err = sshpkg.AnalyzePiholeData(configFile)
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
		return
	}

	// Default: CSV analysis
	if csvFile == "" {
		if cfg.TestMode {
			csvFile = filepath.Join("test_data", "mock_dns_data.csv")
		} else {
			csvFile = cli.GetDefaultCSVFile()
		}
	}

	// In test mode, use mock CSV file if original file doesn't exist
	if cfg.TestMode && csvFile == "test.csv" {
		csvFile = filepath.Join("test_data", "mock_dns_data.csv")
	}

	quietPrintf(cfg.Quiet, "Analyzing DNS usage data from: %s\n", csvFile)
	quietPrintf(cfg.Quiet, "%s\n", colors.ProcessingIndicator("Processing large file, please wait..."))

	clientStats, err := analyzer.AnalyzeDNSDataWithConfig(csvFile, cfg)
	if err != nil {
		log.Fatalf("Error analyzing data: %v", err)
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
