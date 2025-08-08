package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pihole-analyzer/internal/config"
	"pihole-analyzer/testing/testutils"
)

func main() {
	testFlag := flag.Bool("test", false, "Run test suite with mock data")
	configFlag := flag.String("config", "", "Configuration file path")
	flag.Parse()

	if !*testFlag {
		fmt.Println("This is the test runner binary. Use --test flag to run tests.")
		fmt.Println("For production use, use the main pihole-analyzer binary.")
		os.Exit(1)
	}

	// Load configuration for test
	configPath := config.GetConfigPath()
	if *configFlag != "" {
		configPath = *configFlag
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Run test mode
	testutils.RunTestMode(cfg)
}
