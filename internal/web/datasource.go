package web

import (
	"context"
	"fmt"
	"time"

	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// DataSourceAdapter implements DataSourceProvider by wrapping the existing DataSource interface
type DataSourceAdapter struct {
	dataSource interfaces.DataSource
	logger     *logger.Logger
	lastResult *types.AnalysisResult
	lastStatus *types.ConnectionStatus
	lastUpdate time.Time
	cacheTTL   time.Duration
	config     *types.Config
}

// NewDataSourceAdapter creates a new data source adapter
func NewDataSourceAdapter(dataSource interfaces.DataSource, config *types.Config, logger *logger.Logger) (*DataSourceAdapter, error) {
	if dataSource == nil {
		return nil, fmt.Errorf("dataSource cannot be nil")
	}

	webLogger := logger.Component("web-datasource")

	adapter := &DataSourceAdapter{
		dataSource: dataSource,
		logger:     webLogger,
		cacheTTL:   30 * time.Second, // Cache results for 30 seconds
		config:     config,
	}

	// Initialize connection status
	adapter.updateConnectionStatus()

	webLogger.Info("Data source adapter initialized")
	return adapter, nil
}

// GetAnalysisResult returns cached or fresh analysis results
func (d *DataSourceAdapter) GetAnalysisResult(ctx context.Context) (*types.AnalysisResult, error) {
	// Check if we have cached results that are still valid
	if d.lastResult != nil && time.Since(d.lastUpdate) < d.cacheTTL {
		d.logger.Debug("Returning cached analysis result")
		return d.lastResult, nil
	}

	d.logger.Debug("Fetching fresh analysis data")

	// Get fresh data from the data source
	params := interfaces.QueryParams{
		Limit: 1000, // Reasonable limit for web display
	}

	records, err := d.dataSource.GetQueries(ctx, params)
	if err != nil {
		d.logger.Error("Failed to get queries from data source: %v", err)
		d.updateConnectionStatus() // Update status after error
		return nil, fmt.Errorf("failed to get queries: %w", err)
	}

	d.logger.DebugFields("Retrieved records", map[string]any{
		"count": len(records),
	})

	// Get client statistics from data source
	clientStats, err := d.dataSource.GetClientStats(ctx)
	if err != nil {
		d.logger.Error("Failed to get client statistics: %v", err)
		// Fall back to analyzing records if client stats fail
		clientStats = d.analyzeRecordsToClientStats(records)
	}

	// Get network devices if available
	networkDevices, err := d.dataSource.GetNetworkInfo(ctx)
	if err != nil {
		d.logger.Debug("Failed to get network info (not critical): %v", err)
		networkDevices = []types.NetworkDevice{}
	}

	// Enhance client statistics with network device information
	d.enhanceClientStatsWithNetworkInfo(clientStats, networkDevices)

	// Create analysis result
	result := &types.AnalysisResult{
		ClientStats:    clientStats,
		NetworkDevices: networkDevices,
		TotalQueries:   len(records),
		UniqueClients:  len(clientStats),
		AnalysisMode:   "web",
		DataSourceType: "api",
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	// Add performance data if available
	if result.Performance == nil {
		result.Performance = &types.QueryPerformance{
			TotalQueries:     result.TotalQueries,
			QueriesPerSecond: 0, // Would need time window for calculation
		}
	}

	// Cache the result
	d.lastResult = result
	d.lastUpdate = time.Now()
	d.updateConnectionStatus() // Update status after successful operation

	d.logger.InfoFields("Analysis completed", map[string]any{
		"total_queries":  result.TotalQueries,
		"unique_clients": result.UniqueClients,
		"client_count":   len(result.ClientStats),
		"device_count":   len(result.NetworkDevices),
	})

	return result, nil
}

// GetConnectionStatus returns the current connection status
func (d *DataSourceAdapter) GetConnectionStatus() *types.ConnectionStatus {
	if d.lastStatus == nil {
		d.updateConnectionStatus()
	}
	return d.lastStatus
}

// updateConnectionStatus updates the connection status by testing the data source
func (d *DataSourceAdapter) updateConnectionStatus() {
	d.logger.Debug("Updating connection status")

	status := &types.ConnectionStatus{
		Connected:    false,
		LastConnect:  time.Now().Format(time.RFC3339),
		ResponseTime: 0,
		Metadata:     make(map[string]string),
	}

	// Test connection by trying to get a small amount of data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()

	// Try to get a single record to test connectivity
	params := interfaces.QueryParams{
		Limit: 1,
	}

	_, err := d.dataSource.GetQueries(ctx, params)
	duration := time.Since(start)

	status.ResponseTime = float64(duration.Milliseconds())

	if err != nil {
		status.Connected = false
		status.LastError = err.Error()
		status.Metadata["error_type"] = "connection_failed"
		d.logger.Warn("Connection test failed: %v (duration: %dms)", err, duration.Milliseconds())
	} else {
		status.Connected = true
		status.Metadata["last_test"] = "successful"
		d.logger.Debug("Connection test successful (duration: %dms)", duration.Milliseconds())
	}

	status.Metadata["test_duration_ms"] = fmt.Sprintf("%.2f", status.ResponseTime)
	status.Metadata["data_source_type"] = "pihole_api"

	d.lastStatus = status
}

// analyzeRecordsToClientStats performs basic analysis on the records (fallback method)
func (d *DataSourceAdapter) analyzeRecordsToClientStats(records []types.PiholeRecord) map[string]*types.ClientStats {
	clientStats := make(map[string]*types.ClientStats)

	// Process records to build client statistics
	for _, record := range records {
		client := record.Client
		if client == "" {
			continue
		}

		// Initialize client stats if not exists
		if _, exists := clientStats[client]; !exists {
			clientStats[client] = &types.ClientStats{
				IP:          client,
				QueryCount:  0,
				Domains:     make(map[string]int),
				DomainCount: 0,
				IsOnline:    false,
				Hostname:    "Unknown",
				QueryTypes:  make(map[int]int),
				StatusCodes: make(map[int]int),
			}
		}

		stats := clientStats[client]
		stats.QueryCount++

		// Track domains
		if record.Domain != "" {
			stats.Domains[record.Domain]++
		}

		// Track query types
		if record.QueryType != "" {
			// Convert query type string to int (simplified)
			queryTypeInt := 1 // Default to A record
			switch record.QueryType {
			case "A":
				queryTypeInt = 1
			case "AAAA":
				queryTypeInt = 28
			case "PTR":
				queryTypeInt = 12
			}
			stats.QueryTypes[queryTypeInt]++
		}

		// Track status codes
		stats.StatusCodes[record.Status]++
	}

	// Calculate unique domain counts
	for _, stats := range clientStats {
		stats.DomainCount = len(stats.Domains)
	}

	return clientStats
}

// enhanceClientStatsWithNetworkInfo enhances client statistics with network device information
func (d *DataSourceAdapter) enhanceClientStatsWithNetworkInfo(clientStats map[string]*types.ClientStats, networkDevices []types.NetworkDevice) {
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
			if device.Hostname != "" {
				stats.Hostname = device.Hostname
			}
			if device.MAC != "" {
				stats.MACAddress = device.MAC
			}
			stats.IsOnline = device.IsOnline
		}
	}
}

// RefreshCache forces a cache refresh on next request
func (d *DataSourceAdapter) RefreshCache() {
	d.logger.Info("Forcing cache refresh")
	d.lastResult = nil
	d.lastUpdate = time.Time{}
	d.updateConnectionStatus()
}

// SetCacheTTL sets the cache time-to-live duration
func (d *DataSourceAdapter) SetCacheTTL(ttl time.Duration) {
	d.logger.InfoFields("Setting cache TTL", map[string]any{
		"old_ttl_seconds": d.cacheTTL.Seconds(),
		"new_ttl_seconds": ttl.Seconds(),
	})
	d.cacheTTL = ttl
}

// GetCacheInfo returns information about the cache state
func (d *DataSourceAdapter) GetCacheInfo() map[string]any {
	info := map[string]any{
		"has_cached_result": d.lastResult != nil,
		"cache_ttl_seconds": d.cacheTTL.Seconds(),
		"cache_age_seconds": 0.0,
		"cache_valid":       false,
	}

	if !d.lastUpdate.IsZero() {
		age := time.Since(d.lastUpdate)
		info["cache_age_seconds"] = age.Seconds()
		info["cache_valid"] = age < d.cacheTTL
		info["last_update"] = d.lastUpdate.Format(time.RFC3339)
	}

	return info
}
