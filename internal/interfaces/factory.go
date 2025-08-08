package interfaces

import (
	"context"
	"fmt"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// DataSourceFactory creates data source instances based on configuration
type DataSourceFactory struct {
	logger           *logger.Logger
	migrationManager *MigrationManager
}

// NewDataSourceFactory creates a new data source factory
func NewDataSourceFactory(logger *logger.Logger) *DataSourceFactory {
	return &DataSourceFactory{
		logger: logger.Component("datasource-factory"),
	}
}

// CreateDataSource creates the appropriate data source based on configuration
func (f *DataSourceFactory) CreateDataSource(config *types.Config) (DataSource, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Initialize migration manager if not already done
	if f.migrationManager == nil {
		f.migrationManager = NewMigrationManager(config, f.logger)
	}

	// Use migration manager to create data source with migration strategy
	return f.migrationManager.CreateDataSourceWithMigration(context.Background())
}

// CreateDataSourceWithMigration creates data source using migration strategy
func (f *DataSourceFactory) CreateDataSourceWithMigration(ctx context.Context, config *types.Config) (DataSource, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Initialize migration manager
	migrationManager := NewMigrationManager(config, f.logger)

	// Log migration status
	migrationManager.LogMigrationSummary()

	// Create data source with migration strategy applied
	return migrationManager.CreateDataSourceWithMigration(ctx)
}

// CreateAPIDataSource creates an API-based data source (Phase 4 method)
func (f *DataSourceFactory) CreateAPIDataSource(config *types.Config) (DataSource, error) {
	f.logger.Info("Creating Pi-hole API data source")

	// Validate API configuration
	if config.Pihole.APIPassword == "" {
		return nil, fmt.Errorf("API password is required for API data source")
	}

	// Note: This would normally import the pihole package, but that creates import cycles
	// The actual implementation would be in a concrete factory that can import pihole
	return nil, fmt.Errorf("API data source creation must be handled by concrete implementation")
}

// CreateSSHDataSource creates an SSH-based data source (legacy support)
func (f *DataSourceFactory) CreateSSHDataSource(config *types.Config) (DataSource, error) {
	f.logger.Warn("Creating SSH data source (legacy mode)")

	// Validate SSH configuration
	pihole := &config.Pihole
	if pihole.Username == "" {
		return nil, fmt.Errorf("SSH username is required for SSH data source")
	}

	if pihole.Password == "" && pihole.KeyPath == "" {
		return nil, fmt.Errorf("SSH password or key path is required for SSH data source")
	}

	// Note: This would normally import the ssh package, but that creates import cycles
	// The actual implementation would be in a concrete factory that can import ssh
	return nil, fmt.Errorf("SSH data source creation must be handled by concrete implementation")
}

// detectDataSourceType determines the appropriate data source type based on config (legacy method)
func (f *DataSourceFactory) detectDataSourceType(config *types.Config) DataSourceType {
	pihole := &config.Pihole

	// Check explicit configuration preferences
	if pihole.UseSSH {
		f.logger.Info("SSH mode explicitly enabled in configuration")
		return DataSourceTypeSSH
	}

	// Auto-detection logic: prefer API if password is set for API
	if pihole.APIPassword != "" {
		f.logger.Info("API password detected, using API mode")
		return DataSourceTypeAPI
	}

	// Fallback to SSH if SSH credentials are available
	if pihole.Username != "" && (pihole.Password != "" || pihole.KeyPath != "") {
		f.logger.Info("SSH credentials detected, using SSH mode")
		return DataSourceTypeSSH
	}

	// Default to SSH for backward compatibility
	f.logger.Warn("No clear data source preference, defaulting to SSH mode")
	return DataSourceTypeSSH
}

// createAPIDataSource creates an API-based data source
func (f *DataSourceFactory) createAPIDataSource(config *types.Config) (DataSource, error) {
	// Import here to avoid circular dependencies
	return f.createAPIDataSourceImpl(&config.Pihole)
}

// createSSHDataSource creates an SSH-based data source
func (f *DataSourceFactory) createSSHDataSource(config *types.Config) (DataSource, error) {
	// Import here to avoid circular dependencies
	return f.createSSHDataSourceImpl(&config.Pihole)
}

// createAPIDataSourceImpl creates API data source implementation
func (f *DataSourceFactory) createAPIDataSourceImpl(config *types.PiholeConfig) (DataSource, error) {
	// This will be implemented by importing the pihole package
	// For now, return an error to indicate it needs implementation
	return nil, fmt.Errorf("API data source creation not yet implemented in factory")
}

// createSSHDataSourceImpl creates SSH data source implementation
func (f *DataSourceFactory) createSSHDataSourceImpl(config *types.PiholeConfig) (DataSource, error) {
	// This will be implemented by importing the ssh package
	// For now, return an error to indicate it needs implementation
	return nil, fmt.Errorf("SSH data source creation not yet implemented in factory")
}

// ConnectAndValidate creates a data source, connects it, and validates the connection
func (f *DataSourceFactory) ConnectAndValidate(ctx context.Context, config *types.Config) (DataSource, error) {
	dataSource, err := f.CreateDataSource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	// Attempt to connect
	if err := dataSource.Connect(ctx); err != nil {
		// Log connection failure and attempt fallback if appropriate
		f.logger.Error("Primary data source connection failed: %v", err)

		// Attempt fallback if auto-detection was used
		fallbackSource, fallbackErr := f.tryFallback(ctx, config, dataSource.GetDataSourceType())
		if fallbackErr != nil {
			return nil, fmt.Errorf("primary connection failed and fallback unavailable: %w", err)
		}

		f.logger.Info("Successfully connected using fallback data source")
		return fallbackSource, nil
	}

	f.logger.Info("Data source connected successfully: type=%v", dataSource.GetDataSourceType())
	return dataSource, nil
}

// tryFallback attempts to create and connect an alternative data source
func (f *DataSourceFactory) tryFallback(ctx context.Context, config *types.Config, failedType DataSourceType) (DataSource, error) {
	var fallbackType DataSourceType

	// Determine fallback type
	switch failedType {
	case DataSourceTypeAPI:
		fallbackType = DataSourceTypeSSH
	case DataSourceTypeSSH:
		fallbackType = DataSourceTypeAPI
	default:
		return nil, fmt.Errorf("no fallback available for type: %v", failedType)
	}

	f.logger.Info("Attempting fallback to: %v", fallbackType)

	// Create fallback data source
	var fallbackSource DataSource
	var err error

	switch fallbackType {
	case DataSourceTypeAPI:
		fallbackSource, err = f.createAPIDataSource(config)
	case DataSourceTypeSSH:
		fallbackSource, err = f.createSSHDataSource(config)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create fallback data source: %w", err)
	}

	// Attempt fallback connection
	if err := fallbackSource.Connect(ctx); err != nil {
		return nil, fmt.Errorf("fallback connection failed: %w", err)
	}

	return fallbackSource, nil
}

// GetAvailableDataSources returns list of data source types that can be created from config
func (f *DataSourceFactory) GetAvailableDataSources(config *types.Config) []DataSourceType {
	var available []DataSourceType
	pihole := &config.Pihole

	// Check if API is available
	if pihole.APIEnabled || pihole.APIPassword != "" {
		available = append(available, DataSourceTypeAPI)
	}

	// Check if SSH is available
	if pihole.Username != "" && (pihole.Password != "" || pihole.KeyPath != "") {
		available = append(available, DataSourceTypeSSH)
	}

	return available
}

// ValidateConfiguration checks if the configuration supports any data source type
func (f *DataSourceFactory) ValidateConfiguration(config *types.Config) error {
	available := f.GetAvailableDataSources(config)

	if len(available) == 0 {
		return fmt.Errorf("configuration does not support any data source type - need either API password or SSH credentials")
	}

	f.logger.Info("Configuration supports data sources: %v", available)
	return nil
}
