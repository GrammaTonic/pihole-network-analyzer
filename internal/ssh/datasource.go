package ssh

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	_ "modernc.org/sqlite"

	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// SSHDataSource implements the DataSource interface using SSH database access
// This wraps the existing SSH functionality to provide the new interface
type SSHDataSource struct {
	config      *types.PiholeConfig
	logger      *logger.Logger
	connection  *ConnectionManager
	connected   bool
	lastError   error
	connectedAt time.Time
}

// NewSSHDataSource creates a new SSH-based data source
func NewSSHDataSource(config *types.PiholeConfig, logger *logger.Logger) *SSHDataSource {
	return &SSHDataSource{
		config: config,
		logger: logger.Component("ssh-datasource"),
	}
}

// Connect establishes SSH connection to Pi-hole
func (s *SSHDataSource) Connect(ctx context.Context) error {
	s.logger.Info("Connecting to Pi-hole via SSH: host=%s port=%d user=%s",
		s.config.Host, s.config.Port, s.config.Username)

	// Initialize connection manager with existing SSH configuration
	connManager := NewConnectionManager(s.config, WithLogger(&SSHLoggerAdapter{logger: s.logger}))

	// Test connection
	client, err := connManager.Connect(ctx)
	if err != nil {
		s.lastError = fmt.Errorf("SSH connection failed: %w", err)
		s.logger.Error("SSH connection failed: %v", err)
		return s.lastError
	}
	defer client.Close()

	s.connection = connManager
	s.connected = true
	s.connectedAt = time.Now()
	s.lastError = nil

	s.logger.Info("SSH connection established successfully")
	return nil
}

// Close closes the SSH connection
func (s *SSHDataSource) Close() error {
	if s.connection != nil {
		s.connected = false
		s.logger.Info("SSH connection closed")
	}
	return nil
}

// IsConnected returns true if connected to Pi-hole via SSH
func (s *SSHDataSource) IsConnected() bool {
	return s.connected
}

// GetQueries retrieves DNS queries using existing SSH database functionality
func (s *SSHDataSource) GetQueries(ctx context.Context, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Pi-hole")
	}

	s.logger.Info("Fetching DNS queries via SSH: start_time=%s end_time=%s limit=%d",
		params.StartTime.Format(time.RFC3339), params.EndTime.Format(time.RFC3339), params.Limit)

	// Get SSH connection
	client, err := s.connection.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH connection: %w", err)
	}
	defer client.Close()

	// Transfer database file via SSH
	localDBPath := "/tmp/pihole_" + fmt.Sprintf("%d", time.Now().Unix()) + ".db"
	if err := s.transferDatabaseFile(client, s.config.DatabasePath, localDBPath); err != nil {
		return nil, fmt.Errorf("failed to transfer database file: %w", err)
	}
	defer os.Remove(localDBPath) // Clean up temporary file

	// Extract raw records from the database using our custom query
	records, err := s.extractRecordsFromDB(localDBPath, params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract records from database: %w", err)
	}

	s.logger.Info("Successfully fetched queries via SSH: query_count=%d", len(records))
	return records, nil
}

// GetClientStats builds client statistics using SSH data
func (s *SSHDataSource) GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Pi-hole")
	}

	s.logger.Info("Building client statistics via SSH")

	// Get SSH connection
	client, err := s.connection.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH connection: %w", err)
	}
	defer client.Close()

	// Transfer and analyze database using existing functionality
	localDBPath := "/tmp/pihole_stats_" + fmt.Sprintf("%d", time.Now().Unix()) + ".db"
	if err := s.transferDatabaseFile(client, s.config.DatabasePath, localDBPath); err != nil {
		return nil, fmt.Errorf("failed to transfer database file: %w", err)
	}
	defer os.Remove(localDBPath) // Clean up temporary file

	// Use existing AnalyzePiholeDatabase function
	clientStats, err := AnalyzePiholeDatabase(localDBPath)
	if err != nil {
		s.logger.Error("Failed to analyze Pi-hole database via SSH: %v", err)
		return nil, fmt.Errorf("SSH database analysis failed: %w", err)
	}

	s.logger.Info("Client statistics built via SSH: client_count=%d", len(clientStats))
	return clientStats, nil
}

// GetNetworkInfo retrieves network information (limited in SSH mode)
func (s *SSHDataSource) GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error) {
	s.logger.Info("Network info retrieval via SSH (limited functionality)")

	// SSH mode has limited network info - extract from client stats
	clientStats, err := s.GetClientStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client stats for network info: %w", err)
	}

	var devices []types.NetworkDevice
	for ip, stats := range clientStats {
		device := types.NetworkDevice{
			IP:       ip,
			Hardware: stats.HWAddr,
			Name:     stats.Hostname,
			LastSeen: stats.LastSeen,
		}
		devices = append(devices, device)
	}

	s.logger.Info("Network info extracted from SSH data: device_count=%d", len(devices))
	return devices, nil
}

// GetDomainAnalysis analyzes domains from SSH data
func (s *SSHDataSource) GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error) {
	s.logger.Info("Performing domain analysis via SSH")

	clientStats, err := s.GetClientStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client stats for domain analysis: %w", err)
	}

	// Aggregate domain data from all clients
	domainCounts := make(map[string]int)
	totalQueries := 0

	for _, stats := range clientStats {
		for domain, count := range stats.Domains {
			domainCounts[domain] += count
			totalQueries += count
		}
	}

	// Build top domains list
	var topDomains []types.DomainCount
	for domain, count := range domainCounts {
		topDomains = append(topDomains, types.DomainCount{
			Domain: domain,
			Count:  count,
		})
	}

	analysis := &types.DomainAnalysis{
		TopDomains:   topDomains,
		TotalQueries: totalQueries,
		QueryTypes:   make(map[string]int), // Limited in SSH mode
	}

	s.logger.Info("Domain analysis completed via SSH: total_queries=%d unique_domains=%d",
		totalQueries, len(topDomains))
	return analysis, nil
}

// GetQueryPerformance retrieves performance metrics (limited in SSH mode)
func (s *SSHDataSource) GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error) {
	s.logger.Info("Query performance analysis via SSH (limited functionality)")

	clientStats, err := s.GetClientStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client stats for performance: %w", err)
	}

	totalQueries := 0
	totalResponseTime := 0.0

	for _, stats := range clientStats {
		totalQueries += stats.TotalQueries
		totalResponseTime += stats.TotalReplyTime
	}

	avgResponseTime := 0.0
	if totalQueries > 0 {
		avgResponseTime = totalResponseTime / float64(totalQueries)
	}

	performance := &types.QueryPerformance{
		AverageResponseTime: avgResponseTime,
		TotalQueries:        totalQueries,
		QueriesPerSecond:    0, // Not available in SSH mode
		PeakQueries:         0, // Not available in SSH mode
		SlowQueries:         0, // Not available in SSH mode
	}

	return performance, nil
}

// GetConnectionStatus returns the current connection status
func (s *SSHDataSource) GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error) {
	status := &types.ConnectionStatus{
		Connected:   s.connected,
		LastConnect: s.connectedAt.Format(time.RFC3339),
		Metadata:    make(map[string]string),
	}

	if s.lastError != nil {
		status.LastError = s.lastError.Error()
	}

	status.Metadata["host"] = s.config.Host
	status.Metadata["port"] = fmt.Sprintf("%d", s.config.Port)
	status.Metadata["user"] = s.config.Username

	return status, nil
}

// GetDataSourceType returns the data source type
func (s *SSHDataSource) GetDataSourceType() interfaces.DataSourceType {
	return interfaces.DataSourceTypeSSH
}

// GetConnectionInfo returns connection metadata
func (s *SSHDataSource) GetConnectionInfo() *interfaces.ConnectionInfo {
	return &interfaces.ConnectionInfo{
		Type:        interfaces.DataSourceTypeSSH,
		Host:        s.config.Host,
		Port:        s.config.Port,
		Connected:   s.connected,
		LastError:   s.lastError,
		ConnectedAt: s.connectedAt,
		Metadata: map[string]interface{}{
			"username":      s.config.Username,
			"database_path": s.config.DatabasePath,
			"key_file":      s.config.KeyFile,
		},
	}
}

// Helper methods

// filterRecords filters records based on query parameters
func (s *SSHDataSource) filterRecords(records []types.PiholeRecord, params interfaces.QueryParams) []types.PiholeRecord {
	var filtered []types.PiholeRecord

	for _, record := range records {
		// Parse timestamp
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		// Time range filter
		if timestamp.Before(params.StartTime) || timestamp.After(params.EndTime) {
			continue
		}

		// Client filter
		if params.ClientFilter != "" && record.Client != params.ClientFilter {
			continue
		}

		// Domain filter
		if params.DomainFilter != "" && record.Domain != params.DomainFilter {
			continue
		}

		// Status filter
		if len(params.StatusFilter) > 0 {
			found := false
			for _, status := range params.StatusFilter {
				if record.Status == status {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, record)

		// Limit check
		if params.Limit > 0 && len(filtered) >= params.Limit {
			break
		}
	}

	return filtered
}

// buildClientStatsFromRecords builds client statistics from DNS records
func (s *SSHDataSource) buildClientStatsFromRecords(records []types.PiholeRecord) map[string]*types.ClientStats {
	clientStats := make(map[string]*types.ClientStats)

	for _, record := range records {
		client := record.Client
		if client == "" {
			continue
		}

		// Initialize client stats if not exists
		if _, exists := clientStats[client]; !exists {
			clientStats[client] = &types.ClientStats{
				Client:      client,
				IP:          client,
				Domains:     make(map[string]int),
				QueryTypes:  make(map[int]int),
				StatusCodes: make(map[int]int),
			}
		}

		stats := clientStats[client]
		stats.TotalQueries++

		// Domain statistics
		if record.Domain != "" {
			stats.Domains[record.Domain]++
		}

		// Status statistics
		stats.StatusCodes[record.Status]++

		// Hardware address
		if record.HWAddr != "" {
			stats.HWAddr = record.HWAddr
		}

		// Update last seen
		stats.LastSeen = record.DateTime
	}

	// Calculate unique queries for each client
	for _, stats := range clientStats {
		stats.UniqueQueries = len(stats.Domains)
		stats.DomainCount = len(stats.Domains)
	}

	return clientStats
}

// SSHLoggerAdapter adapts the project logger to the SSH logger interface
type SSHLoggerAdapter struct {
	logger *logger.Logger
}

func (a *SSHLoggerAdapter) Info(msg string, args ...interface{}) {
	a.logger.Info(msg, args...)
}

func (a *SSHLoggerAdapter) Warn(msg string, args ...interface{}) {
	a.logger.Warn(msg, args...)
}

func (a *SSHLoggerAdapter) Error(msg string, args ...interface{}) {
	a.logger.Error(msg, args...)
}

// transferDatabaseFile transfers the Pi-hole database file via SSH
func (s *SSHDataSource) transferDatabaseFile(client *ssh.Client, remotePath, localPath string) error {
	// Use the existing FileTransfer functionality
	transfer := NewFileTransfer(client, &SSHLoggerAdapter{logger: s.logger})

	ctx := context.Background()
	options := DefaultTransferOptions()
	options.CreateDirs = true
	options.Overwrite = true

	return transfer.DownloadFile(ctx, remotePath, localPath, options)
}

// extractRecordsFromDB extracts individual DNS records from the database
func (s *SSHDataSource) extractRecordsFromDB(dbPath string, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	db, err := sql.Open("sqlite", dbPath+"?cache=shared&mode=ro")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Build query with parameters
	query := `
		SELECT 
			q.timestamp,
			q.client,
			COALESCE(n.hwaddr, '') as hwaddr,
			q.domain,
			q.status
		FROM queries q
		LEFT JOIN network_addresses na ON q.client = na.ip
		LEFT JOIN network n ON na.network_id = n.id
		WHERE q.timestamp >= ? AND q.timestamp <= ?
	`

	args := []interface{}{
		params.StartTime.Unix(),
		params.EndTime.Unix(),
	}

	// Add filters
	if params.ClientFilter != "" {
		query += " AND q.client = ?"
		args = append(args, params.ClientFilter)
	}

	if params.DomainFilter != "" {
		query += " AND q.domain LIKE ?"
		args = append(args, "%"+params.DomainFilter+"%")
	}

	query += " ORDER BY q.timestamp DESC"

	if params.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, params.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	var records []types.PiholeRecord
	for rows.Next() {
		var record types.PiholeRecord
		var timestamp float64

		err := rows.Scan(
			&timestamp,
			&record.Client,
			&record.HWAddr,
			&record.Domain,
			&record.Status,
		)
		if err != nil {
			s.logger.Warn("Error scanning record: %v", err)
			continue
		}

		// Convert timestamp to string format
		record.Timestamp = fmt.Sprintf("%.0f", timestamp)
		record.DateTime = time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")

		records = append(records, record)
	}

	return records, nil
}
