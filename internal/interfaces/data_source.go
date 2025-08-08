package interfaces

import "pihole-analyzer/internal/types"

// DataSource defines the interface for data analysis sources
type DataSource interface {
	AnalyzeData(configFile string, cfg *types.Config) (map[string]*types.ClientStats, error)
}

// NetworkChecker defines the interface for network status checking
type NetworkChecker interface {
	CheckARPStatus(clientStats map[string]*types.ClientStats) error
}

// TestMode defines the interface for test mode operations
type TestMode interface {
	RunTestMode(cfg *types.Config)
	GetMockDatabasePath() string
}
