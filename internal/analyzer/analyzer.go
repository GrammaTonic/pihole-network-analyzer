package analyzer

import (
	"context"
	"fmt"
	"log"
	"time"

	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/network"
	"pihole-analyzer/internal/ssh"
	"pihole-analyzer/internal/types"
)

// AnalyzePiholeData performs DNS data analysis from Pi-hole database (legacy SSH-only method)
// Deprecated: Use Phase5Analyzer for modern API-first analysis with migration support
func AnalyzePiholeData(configFile string, config *types.Config) (map[string]*types.ClientStats, error) {
	if !config.Quiet {
		fmt.Println(colors.ProcessingIndicator("Connecting to Pi-hole server..."))
	}

	// Show deprecation warning if using SSH mode
	if config.Pihole.Host != "" && config.Pihole.Username != "" {
		log.Printf("⚠️  Warning: Using deprecated SSH analysis method")
		log.Printf("   Consider using Phase 5 analyzer with API support")
		log.Printf("   Migration guide: docs/migration-ssh-to-api.md")
	}

	clientStats, err := ssh.AnalyzePiholeData(configFile)
	if err != nil {
		return nil, fmt.Errorf("error analyzing Pi-hole data: %v", err)
	}

	if !config.Quiet {
		fmt.Println(colors.ProcessingIndicator("Checking ARP status and resolving hostnames..."))
	}

	// Resolve hostnames and check ARP status
	network.ResolveHostnames(clientStats)
	if err := network.CheckARPStatus(clientStats); err != nil {
		log.Printf("Warning: Could not check ARP status: %v", err)
	}

	return clientStats, nil
}

// EnhancedAnalyzer provides universal analysis logic regardless of data source
type EnhancedAnalyzer struct {
	dataSource interfaces.DataSource
	config     *types.Config
	logger     *logger.Logger
}

// NewEnhancedAnalyzer creates a new analyzer with migration-aware data source
func NewEnhancedAnalyzer(config *types.Config, logger *logger.Logger) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		config: config,
		logger: logger.Component("enhanced-analyzer"),
	}
}

// Initialize creates and connects the appropriate data source based on migration strategy
func (a *EnhancedAnalyzer) Initialize(ctx context.Context) error {
	a.logger.Info("🔄 Initializing enhanced analyzer with migration-aware data source")

	// Create migration-aware data source factory
	factory := interfaces.NewDataSourceFactory(a.logger)

	// Create data source with migration strategy applied
	dataSource, err := factory.CreateDataSource(a.config)
	if err != nil {
		return fmt.Errorf("failed to create data source: %w", err)
	}

	// Connect to data source
	if err := dataSource.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to data source: %w", err)
	}

	a.dataSource = dataSource
	a.logger.Info("✅ Enhanced analyzer initialized successfully")
	return nil
}

// AnalyzeData performs universal DNS data analysis using migration-aware data source
func (a *EnhancedAnalyzer) AnalyzeData(ctx context.Context) (*types.AnalysisResult, error) {
	if a.dataSource == nil {
		return nil, fmt.Errorf("analyzer not initialized - call Initialize() first")
	}

	a.logger.Info("📊 Starting enhanced data analysis")

	// Get DNS queries from data source (API or SSH based on migration mode)
	queries, err := a.dataSource.GetQueries(ctx, interfaces.QueryParams{
		Limit: 10000, // Default limit for analysis
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS queries: %w", err)
	}

	a.logger.Info("Retrieved %d DNS queries for analysis", len(queries))

	// Get client statistics from data source
	clientStats, err := a.dataSource.GetClientStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client statistics: %w", err)
	}

	a.logger.Info("Retrieved statistics for %d clients", len(clientStats))

	// Get network information
	networkDevices, err := a.dataSource.GetNetworkInfo(ctx)
	if err != nil {
		a.logger.Warn("Failed to get network information: %v", err)
		// Continue without network info - not critical
	}

	// Enhance client statistics with network analysis
	a.enhanceWithNetworkAnalysis(clientStats, networkDevices)

	// Create comprehensive analysis result
	result := &types.AnalysisResult{
		ClientStats:    clientStats,
		NetworkDevices: networkDevices,
		TotalQueries:   len(queries),
		UniqueClients:  len(clientStats),
		AnalysisMode:   a.getAnalysisMode(),
		DataSourceType: string(a.dataSource.GetDataSourceType()),
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	a.logger.Info("✅ Enhanced analysis complete: %d clients, %d queries",
		result.UniqueClients, result.TotalQueries)

	return result, nil
}

// Close releases resources used by the analyzer
func (a *EnhancedAnalyzer) Close() error {
	if a.dataSource != nil {
		return a.dataSource.Close()
	}
	return nil
}

// enhanceWithNetworkAnalysis enhances client statistics with network device information
func (a *EnhancedAnalyzer) enhanceWithNetworkAnalysis(clientStats map[string]*types.ClientStats, networkDevices []types.NetworkDevice) {
	if len(networkDevices) == 0 {
		return
	}

	// Create a map of IP to network device for quick lookup
	deviceMap := make(map[string]*types.NetworkDevice)
	for i := range networkDevices {
		deviceMap[networkDevices[i].IP] = &networkDevices[i]
	}

	// Enhance client statistics with network device information
	for clientIP, stats := range clientStats {
		if device, exists := deviceMap[clientIP]; exists {
			stats.Hostname = device.Hostname
			stats.HWAddr = device.Hardware
			stats.IsOnline = device.IsOnline
		}
	}
}

// getAnalysisMode returns the current analysis mode based on configuration
func (a *EnhancedAnalyzer) getAnalysisMode() string {
	if a.config.Pihole.MigrationMode == "api-first" {
		return "Enhanced API-First Analysis"
	} else if a.config.Pihole.MigrationMode == "ssh-only" {
		return "Traditional SSH Analysis"
	}
	return "Automatic Migration Analysis"
}

// AnalyzePiholeDataWithMigration performs modern API-first analysis with migration support
// This is the recommended method for enhanced implementations
func AnalyzePiholeDataWithMigration(ctx context.Context, config *types.Config, appLogger *logger.Logger) (*types.AnalysisResult, error) {
	// Create enhanced analyzer
	enhancedAnalyzer := NewEnhancedAnalyzer(config, appLogger)

	// Initialize analyzer
	if err := enhancedAnalyzer.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced analyzer: %w", err)
	}
	defer enhancedAnalyzer.Close()

	// Perform analysis
	result, err := enhancedAnalyzer.AnalyzeData(ctx)
	if err != nil {
		return nil, fmt.Errorf("enhanced analysis failed: %w", err)
	}

	return result, nil
}

// CreateMigrationAwareDataSource creates a data source based on migration configuration
// This function bridges legacy and modern implementations
func CreateMigrationAwareDataSource(config *types.Config, appLogger *logger.Logger) (interfaces.DataSource, error) {
	factory := interfaces.NewDataSourceFactory(appLogger)

	// Create data source based on configuration
	dataSource, err := factory.CreateDataSource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration-aware data source: %w", err)
	}

	return dataSource, nil
}

// GetQueryTypeName returns human-readable query type name
func GetQueryTypeName(queryType int) string {
	switch queryType {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 12:
		return "PTR"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 28:
		return "AAAA"
	case 33:
		return "SRV"
	case 35:
		return "NAPTR"
	case 39:
		return "DNAME"
	case 41:
		return "OPT"
	case 43:
		return "DS"
	case 46:
		return "RRSIG"
	case 47:
		return "NSEC"
	case 48:
		return "DNSKEY"
	case 50:
		return "NSEC3"
	case 51:
		return "NSEC3PARAM"
	case 52:
		return "TLSA"
	case 257:
		return "CAA"
	default:
		return "Unknown"
	}
}

// GetStatusName returns human-readable status name
func GetStatusName(status int) string {
	switch status {
	case 0:
		return "Unknown"
	case 1:
		return "Blocked (gravity)"
	case 2:
		return "Forwarded"
	case 3:
		return "Cached"
	case 4:
		return "Blocked (regex/wildcard)"
	case 5:
		return "Blocked (exact)"
	case 6:
		return "Blocked (external)"
	case 7:
		return "CNAME"
	case 8:
		return "Retried"
	case 9:
		return "Retried but ignored"
	case 10:
		return "Already forwarded"
	case 11:
		return "Already cached"
	case 12:
		return "Config blocked"
	case 13:
		return "Gravity blocked"
	case 14:
		return "Regex blocked"
	default:
		return "Unknown"
	}
}

// GetStatusCodeFromString converts status strings to status codes
func GetStatusCodeFromString(status string) int {
	switch status {
	case "Blocked (gravity)":
		return 1
	case "Forwarded":
		return 2
	case "Cached":
		return 3
	case "Blocked (regex/wildcard)":
		return 4
	case "Blocked (exact)":
		return 5
	case "Blocked (external)":
		return 6
	case "CNAME":
		return 7
	case "Retried":
		return 8
	case "Retried but ignored":
		return 9
	case "Already forwarded":
		return 10
	case "Already cached":
		return 11
	case "Config blocked":
		return 12
	case "Gravity blocked":
		return 13
	case "Regex blocked":
		return 14
	default:
		return 0
	}
}
