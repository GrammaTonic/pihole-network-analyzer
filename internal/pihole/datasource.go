package pihole

import (
	"context"
	"fmt"
	"time"

	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// APIDataSource implements the DataSource interface using Pi-hole REST API
type APIDataSource struct {
	client      *Client
	config      *types.PiholeConfig
	logger      *logger.Logger
	connected   bool
	lastError   error
	connectedAt time.Time
}

// NewAPIDataSource creates a new API-based data source
func NewAPIDataSource(config *types.PiholeConfig, logger *logger.Logger) *APIDataSource {
	return &APIDataSource{
		config: config,
		logger: logger.Component("api-datasource"),
	}
}

// Connect establishes connection to Pi-hole API
func (a *APIDataSource) Connect(ctx context.Context) error {
	a.logger.Info("Connecting to Pi-hole via API: host=%s port=%d",
		a.config.Host, a.config.Port)

	// Create API client configuration
	apiConfig := &Config{
		Host:     a.config.Host,
		Port:     a.config.Port,
		Password: a.config.APIPassword,
		UseHTTPS: a.config.UseHTTPS,
		Timeout:  time.Duration(a.config.APITimeout) * time.Second,
	}

	// Create and configure API client
	client := NewClient(apiConfig, a.logger)

	// Test connection with authentication
	if err := client.Authenticate(ctx); err != nil {
		a.lastError = fmt.Errorf("API authentication failed: %w", err)
		a.logger.Error("API authentication failed: %v", err)
		return a.lastError
	}

	a.client = client
	a.connected = true
	a.connectedAt = time.Now()
	a.lastError = nil

	a.logger.Info("API connection established successfully")
	return nil
}

// Close closes the API connection
func (a *APIDataSource) Close() error {
	if a.client != nil {
		ctx := context.Background()
		if err := a.client.Close(ctx); err != nil {
			a.logger.Warn("Error closing API connection: %v", err)
		}
		a.connected = false
		a.logger.Info("API connection closed")
	}
	return nil
}

// IsConnected returns true if connected to Pi-hole API
func (a *APIDataSource) IsConnected() bool {
	return a.connected && a.client != nil
}

// GetQueries retrieves DNS queries using enhanced API with SSH parity
func (a *APIDataSource) GetQueries(ctx context.Context, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected to Pi-hole API")
	}

	a.logger.Info("Using enhanced API query retrieval for exact SSH parity")

	// Convert interface params to pihole params
	piholeParams := QueryParams{
		Limit: params.Limit,
	}

	// Use enhanced query functionality
	records, err := a.client.GetDNSQueries(ctx, piholeParams)
	if err != nil {
		return nil, fmt.Errorf("enhanced API query failed: %w", err)
	}

	a.logger.Info("Enhanced API query complete: record_count=%d", len(records))
	return records, nil
}

// GetClientStats builds client statistics using enhanced API with SSH parity
func (a *APIDataSource) GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected to Pi-hole API")
	}

	a.logger.Info("Building enhanced client statistics with SSH parity")

	// Get queries for analysis (last 7 days) using enhanced method
	params := interfaces.QueryParams{
		StartTime: time.Now().AddDate(0, 0, -7),
		EndTime:   time.Now(),
		Limit:     10000,
	}

	queries, err := a.GetQueries(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get enhanced queries for stats: %w", err)
	}

	// Get network info from API
	networkInfo, err := a.client.GetNetworkInfo(ctx)
	if err != nil {
		a.logger.Warn("Failed to get network info, continuing without: %v", err)
		networkInfo = []ClientInfo{} // Continue without network info
	}

	// Build client statistics from queries
	clientStats := a.buildClientStatsFromQueries(queries, networkInfo)

	a.logger.Info("Enhanced client statistics complete: client_count=%d", len(clientStats))
	return clientStats, nil
}

// buildClientStatsFromQueries builds client statistics from DNS query records
func (a *APIDataSource) buildClientStatsFromQueries(queries []types.PiholeRecord, networkInfo []ClientInfo) map[string]*types.ClientStats {
	clientMap := make(map[string]*types.ClientStats)

	// Create network info lookup map
	networkMap := make(map[string]ClientInfo)
	for _, info := range networkInfo {
		networkMap[info.IP] = info
	}

	// Process each query record
	for _, query := range queries {
		clientIP := query.Client

		// Initialize client stats if not exists
		if _, exists := clientMap[clientIP]; !exists {
			stats := &types.ClientStats{
				Client:      clientIP,
				IP:          clientIP,
				Domains:     make(map[string]int),
				QueryTypes:  make(map[int]int),
				StatusCodes: make(map[int]int),
				TopDomains:  []types.DomainStat{},
			}

			// Add network info if available
			if netInfo, hasNetInfo := networkMap[clientIP]; hasNetInfo {
				stats.Hostname = netInfo.Name
				stats.HWAddr = netInfo.MAC
			}

			clientMap[clientIP] = stats
		}

		// Update statistics
		client := clientMap[clientIP]
		client.TotalQueries++
		client.Domains[query.Domain]++

		// Update unique queries count
		if client.Domains[query.Domain] == 1 {
			client.UniqueQueries++
		}
	}

	return clientMap
}

// GetNetworkInfo retrieves network information via API
func (a *APIDataSource) GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected to Pi-hole API")
	}

	apiNetworkInfo, err := a.client.GetNetworkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("API network request failed: %w", err)
	}

	// Convert API network info to standard format
	var devices []types.NetworkDevice
	for _, device := range apiNetworkInfo {
		devices = append(devices, types.NetworkDevice{
			IP:       device.IP,
			Hardware: device.MAC,
			Name:     device.Name,
		})
	}

	return devices, nil
}

// GetDomainAnalysis analyzes domains using API data
func (a *APIDataSource) GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected to Pi-hole API")
	}

	// Get statistics from API
	stats, err := a.client.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Calculate blocked percentage
	blockedPercent := 0.0
	if stats.DNSQueriesToday > 0 {
		blockedPercent = float64(stats.AdsBlockedToday) / float64(stats.DNSQueriesToday) * 100
	}

	analysis := &types.DomainAnalysis{
		TopDomains:     []types.DomainCount{},
		TotalQueries:   stats.DNSQueriesToday,
		TotalBlocked:   stats.AdsBlockedToday,
		BlockedPercent: blockedPercent,
		QueryTypes:     make(map[string]int),
	}

	return analysis, nil
}

// GetQueryPerformance retrieves performance metrics via API
func (a *APIDataSource) GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected to Pi-hole API")
	}

	stats, err := a.client.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	performance := &types.QueryPerformance{
		AverageResponseTime: 0,
		TotalQueries:        stats.DNSQueriesToday,
		QueriesPerSecond:    0,
		PeakQueries:         0,
		SlowQueries:         0,
	}

	return performance, nil
}

// GetConnectionStatus returns the current API connection status
func (a *APIDataSource) GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error) {
	status := &types.ConnectionStatus{
		Connected:   a.connected,
		LastConnect: a.connectedAt.Format(time.RFC3339),
		Metadata:    make(map[string]string),
	}

	if a.lastError != nil {
		status.LastError = a.lastError.Error()
	}

	return status, nil
}

// GetDataSourceType returns the data source type
func (a *APIDataSource) GetDataSourceType() interfaces.DataSourceType {
	return interfaces.DataSourceTypeAPI
}

// GetConnectionInfo returns connection metadata
func (a *APIDataSource) GetConnectionInfo() *interfaces.ConnectionInfo {
	return &interfaces.ConnectionInfo{
		Type:        interfaces.DataSourceTypeAPI,
		Host:        a.config.Host,
		Port:        a.config.Port,
		Connected:   a.connected,
		LastError:   a.lastError,
		ConnectedAt: a.connectedAt,
		Metadata: map[string]interface{}{
			"api_version": "v1",
			"use_https":   false,
			"timeout":     30,
		},
	}
}
