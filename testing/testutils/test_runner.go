package testutils

import (
	"fmt"

	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/reporting"
	"pihole-analyzer/internal/types"
)

// RunTestMode executes the test mode using mock data
func RunTestMode(cfg *types.Config) {
	fmt.Println(colors.Header("🧪 Running Test Mode"))
	fmt.Println("Using mock Pi-hole database")

	// Generate mock client statistics for testing
	clientStats := generateMockClientStats()

	fmt.Println(colors.Success("✅ Test mode analysis completed"))
	reporting.DisplayResultsWithConfig(clientStats, cfg)
	fmt.Println(colors.Info("Test mode completed successfully"))
}

// generateMockClientStats creates mock client statistics for testing
func generateMockClientStats() map[string]*types.ClientStats {
	return map[string]*types.ClientStats{
		"192.168.2.110": {
			Client:        "192.168.2.110",
			Hostname:      "unknown",
			HWAddr:        "66:78:8b:59:b...",
			IsOnline:      false,
			TotalQueries:  4,
			UniqueQueries: 4,
			AvgReplyTime:  0.0,
			Domains: map[string]int{
				"microsoft.com": 2,
				"google.com":    2,
			},
			QueryTypes:  map[int]int{1: 4},
			StatusCodes: map[int]int{2: 4},
			TopDomains: []types.DomainStat{
				{Domain: "microsoft.com", Count: 2},
				{Domain: "google.com", Count: 2},
			},
		},
		// Add more mock clients as needed
		"192.168.2.111": {
			Client:        "192.168.2.111",
			Hostname:      "test-device",
			HWAddr:        "aa:bb:cc:dd:ee:ff",
			IsOnline:      true,
			TotalQueries:  8,
			UniqueQueries: 6,
			AvgReplyTime:  15.5,
			Domains: map[string]int{
				"example.com": 3,
				"test.com":    5,
			},
			QueryTypes:  map[int]int{1: 6, 28: 2},
			StatusCodes: map[int]int{2: 6, 3: 2},
			TopDomains: []types.DomainStat{
				{Domain: "test.com", Count: 5},
				{Domain: "example.com", Count: 3},
			},
		},
	}
}
