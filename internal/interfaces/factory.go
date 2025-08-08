package interfaces

import (
	"context"
	"fmt"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// DataSourceFactory creates API-only data source instances
type DataSourceFactory struct {
	logger *logger.Logger
}

// NewDataSourceFactory creates a new data source factory
func NewDataSourceFactory(logger *logger.Logger) *DataSourceFactory {
	return &DataSourceFactory{
		logger: logger.Component("datasource-factory"),
	}
}

// CreateDataSource creates an API-based data source
func (f *DataSourceFactory) CreateDataSource(config *types.Config) (DataSource, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	return f.CreateAPIDataSource(config)
}

// CreateAPIDataSource creates an API-based data source
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

// ConnectAndValidate creates a data source, connects it, and validates the connection
func (f *DataSourceFactory) ConnectAndValidate(ctx context.Context, config *types.Config) (DataSource, error) {
	dataSource, err := f.CreateDataSource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	// Attempt to connect
	if err := dataSource.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Validate connection
	if err := f.validateDataSource(ctx, dataSource); err != nil {
		dataSource.Close()
		return nil, fmt.Errorf("data source validation failed: %w", err)
	}

	return dataSource, nil
}

// validateDataSource performs basic validation on the data source connection
func (f *DataSourceFactory) validateDataSource(ctx context.Context, dataSource DataSource) error {
	// Test basic query functionality
	_, err := dataSource.GetQueries(ctx, QueryParams{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to execute test query: %w", err)
	}

	f.logger.Info("Data source validation successful")
	return nil
}
