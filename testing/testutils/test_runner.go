package testutils

import (
	"fmt"
	"log"
	"path/filepath"

	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/reporting"
	sshpkg "pihole-analyzer/internal/ssh"
	"pihole-analyzer/internal/types"
)

// RunTestMode executes the test mode using mock data
func RunTestMode(cfg *types.Config) {
	fmt.Println(colors.Header("🧪 Running Test Mode"))
	fmt.Println("Using mock Pi-hole database")

	// Analyze the test data using Pi-hole mock database
	dbFile := filepath.Join("testing", "fixtures", "mock_pihole.db")
	clientStats, err := sshpkg.AnalyzePiholeDatabase(dbFile)
	if err != nil {
		log.Fatalf("Error analyzing test data: %v", err)
	}

	fmt.Println(colors.Success("✅ Test mode analysis completed"))
	reporting.DisplayResultsWithConfig(clientStats, cfg)
	fmt.Println(colors.Info("Test mode completed successfully"))
}

// GetMockDatabasePath returns the path to the mock database for test mode
func GetMockDatabasePath() string {
	return filepath.Join("testing", "fixtures", "mock_pihole.db")
}
