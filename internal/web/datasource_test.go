package web

import (
	"context"
	"fmt"
	"testing"
	"time"

	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// MockDataSource implements interfaces.DataSource for testing
type MockDataSource struct {
	records     []types.PiholeRecord
	shouldError bool
}

func NewMockDataSource() *MockDataSource {
	return &MockDataSource{
		records: []types.PiholeRecord{
			{
				ID:        1,
				DateTime:  "2024-01-15 10:30:00",
				Domain:    "example.com",
				Client:    "192.168.1.100",
				QueryType: "A",
				Status:    2,
				Timestamp: "2024-01-15T10:30:00Z",
			},
			{
				ID:        2,
				DateTime:  "2024-01-15 10:31:00",
				Domain:    "google.com",
				Client:    "192.168.1.101",
				QueryType: "AAAA",
				Status:    2,
				Timestamp: "2024-01-15T10:31:00Z",
			},
		},
		shouldError: false,
	}
}

func (m *MockDataSource) GetQueries(ctx context.Context, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock data source error")
	}

	// Apply limit if specified
	records := m.records
	if params.Limit > 0 && len(records) > params.Limit {
		records = records[:params.Limit]
	}

	return records, nil
}

// Implement required interface methods
func (m *MockDataSource) Connect(ctx context.Context) error {
	if m.shouldError {
		return fmt.Errorf("mock connection error")
	}
	return nil
}

func (m *MockDataSource) Close() error {
	return nil
}

func (m *MockDataSource) IsConnected() bool {
	return !m.shouldError
}

func (m *MockDataSource) GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return make(map[string]*types.ClientStats), nil
}

func (m *MockDataSource) GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return []types.NetworkDevice{}, nil
}

func (m *MockDataSource) GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return &types.DomainAnalysis{}, nil
}

func (m *MockDataSource) GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return &types.QueryPerformance{}, nil
}

func (m *MockDataSource) GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return &types.ConnectionStatus{
		Connected:   true,
		LastConnect: time.Now().Format(time.RFC3339),
	}, nil
}

func (m *MockDataSource) GetDataSourceType() interfaces.DataSourceType {
	return interfaces.DataSourceTypeAPI
}

func (m *MockDataSource) GetConnectionInfo() *interfaces.ConnectionInfo {
	return &interfaces.ConnectionInfo{
		Type:      interfaces.DataSourceTypeAPI,
		Connected: !m.shouldError,
	}
}

func (m *MockDataSource) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockDataSource) AddRecord(record types.PiholeRecord) {
	m.records = append(m.records, record)
}

func TestNewDataSourceAdapter(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
	}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adapter == nil {
		t.Fatal("Expected adapter to be created")
	}

	if adapter.dataSource == nil {
		t.Error("Expected adapter to use provided data source")
	}

	if adapter.cacheTTL != 30*time.Second {
		t.Errorf("Expected cache TTL 30s, got %v", adapter.cacheTTL)
	}
}

func TestNewDataSourceAdapterWithNilDataSource(t *testing.T) {
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(nil, config, logger)

	if err == nil {
		t.Fatal("Expected error for nil data source")
	}

	if adapter != nil {
		t.Error("Expected adapter to be nil when error occurs")
	}
}

func TestGetAnalysisResult(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
		Output: types.OutputConfig{
			MaxClients: 20,
			MaxDomains: 10,
		},
	}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	result, err := adapter.GetAnalysisResult(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	if result.DataSourceType != "api" {
		t.Errorf("Expected data source type 'api', got '%s'", result.DataSourceType)
	}

	if result.AnalysisMode != "web" {
		t.Errorf("Expected analysis mode 'web', got '%s'", result.AnalysisMode)
	}

	if result.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}

	if result.Performance == nil {
		t.Error("Expected performance data to be set")
	}
}

func TestGetAnalysisResultWithError(t *testing.T) {
	mockDataSource := NewMockDataSource()
	mockDataSource.SetError(true)

	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	result, err := adapter.GetAnalysisResult(ctx)

	if err == nil {
		t.Fatal("Expected error when data source fails")
	}

	if result != nil {
		t.Error("Expected result to be nil when error occurs")
	}
}

func TestGetAnalysisResultCaching(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Set a very long cache TTL for testing
	adapter.SetCacheTTL(1 * time.Hour)

	ctx := context.Background()

	// First call should fetch fresh data
	result1, err := adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second call should return cached data
	result2, err := adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Results should be the same object (cached)
	if result1 != result2 {
		t.Error("Expected second call to return cached result")
	}
}

func TestGetAnalysisResultCacheExpiry(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Set a very short cache TTL
	adapter.SetCacheTTL(1 * time.Millisecond)

	ctx := context.Background()

	// First call
	result1, err := adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Millisecond)

	// Second call should fetch fresh data
	result2, err := adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Results should be different objects (cache expired)
	if result1 == result2 {
		t.Error("Expected second call to return fresh result after cache expiry")
	}
}

func TestGetConnectionStatus(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	status := adapter.GetConnectionStatus()

	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	if !status.Connected {
		t.Error("Expected status to be connected")
	}

	if status.LastConnect == "" {
		t.Error("Expected last connect time to be set")
	}

	if status.Metadata == nil {
		t.Error("Expected metadata to be set")
	}

	if status.Metadata["data_source_type"] != "pihole_api" {
		t.Errorf("Expected data source type 'pihole_api', got '%s'", status.Metadata["data_source_type"])
	}
}

func TestGetConnectionStatusWithError(t *testing.T) {
	mockDataSource := NewMockDataSource()
	mockDataSource.SetError(true)

	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	status := adapter.GetConnectionStatus()

	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	if status.Connected {
		t.Error("Expected status to be disconnected when data source fails")
	}

	if status.LastError == "" {
		t.Error("Expected last error to be set")
	}

	if status.Metadata["error_type"] != "connection_failed" {
		t.Errorf("Expected error type 'connection_failed', got '%s'", status.Metadata["error_type"])
	}
}

func TestRefreshCache(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Get initial result to populate cache
	ctx := context.Background()
	_, err = adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify cache is populated
	if adapter.lastResult == nil {
		t.Error("Expected cache to be populated")
	}

	// Refresh cache
	adapter.RefreshCache()

	// Verify cache is cleared
	if adapter.lastResult != nil {
		t.Error("Expected cache to be cleared after refresh")
	}

	if !adapter.lastUpdate.IsZero() {
		t.Error("Expected last update time to be reset")
	}
}

func TestSetCacheTTL(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	newTTL := 5 * time.Minute
	adapter.SetCacheTTL(newTTL)

	if adapter.cacheTTL != newTTL {
		t.Errorf("Expected cache TTL %v, got %v", newTTL, adapter.cacheTTL)
	}
}

func TestGetCacheInfo(t *testing.T) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Get cache info before any data is cached
	info := adapter.GetCacheInfo()

	if info["has_cached_result"] != false {
		t.Error("Expected has_cached_result to be false initially")
	}

	if info["cache_valid"] != false {
		t.Error("Expected cache_valid to be false initially")
	}

	if info["cache_ttl_seconds"] != 30.0 {
		t.Errorf("Expected cache TTL 30s, got %v", info["cache_ttl_seconds"])
	}

	// Cache some data
	ctx := context.Background()
	_, err = adapter.GetAnalysisResult(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Get cache info after data is cached
	info = adapter.GetCacheInfo()

	if info["has_cached_result"] != true {
		t.Error("Expected has_cached_result to be true after caching")
	}

	if info["cache_valid"] != true {
		t.Error("Expected cache_valid to be true after caching")
	}

	if info["cache_age_seconds"] == nil {
		t.Error("Expected cache_age_seconds to be set")
	}

	if info["last_update"] == nil {
		t.Error("Expected last_update to be set")
	}
}

// Benchmark tests
func BenchmarkGetAnalysisResult(b *testing.B) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		b.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache for each iteration to benchmark fresh data fetching
		adapter.RefreshCache()
		_, err := adapter.GetAnalysisResult(ctx)
		if err != nil {
			b.Fatalf("Error in benchmark: %v", err)
		}
	}
}

func BenchmarkGetAnalysisResultCached(b *testing.B) {
	mockDataSource := NewMockDataSource()
	config := &types.Config{}
	logger := logger.New(logger.DefaultConfig())

	adapter, err := NewDataSourceAdapter(mockDataSource, config, logger)
	if err != nil {
		b.Fatalf("Failed to create adapter: %v", err)
	}

	// Set very long cache TTL
	adapter.SetCacheTTL(1 * time.Hour)

	ctx := context.Background()

	// Populate cache
	_, err = adapter.GetAnalysisResult(ctx)
	if err != nil {
		b.Fatalf("Failed to populate cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.GetAnalysisResult(ctx)
		if err != nil {
			b.Fatalf("Error in benchmark: %v", err)
		}
	}
}
