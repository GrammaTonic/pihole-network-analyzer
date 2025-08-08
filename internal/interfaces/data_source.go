package interfaces

import (
	"context"
	"time"

	"pihole-analyzer/internal/types"
)

// DataSource defines the interface for Pi-hole data access via API
type DataSource interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	IsConnected() bool

	// Core data retrieval
	GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
	GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error)
	GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error)
	GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error)

	// Performance and metadata
	GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error)
	GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error)

	// Configuration and metadata
	GetDataSourceType() DataSourceType
	GetConnectionInfo() *ConnectionInfo
}

// QueryParams represents parameters for DNS query requests
type QueryParams struct {
	StartTime    time.Time
	EndTime      time.Time
	ClientFilter string
	DomainFilter string
	Limit        int
	StatusFilter []int // Pi-hole status codes (blocked, allowed, etc.)
	TypeFilter   []int // DNS query types (A, AAAA, PTR, etc.)
}

// DataSourceType represents the type of data source implementation
type DataSourceType string

const (
	DataSourceTypeAPI DataSourceType = "api"
)

// ConnectionInfo provides metadata about the data source connection
type ConnectionInfo struct {
	Type        DataSourceType
	Host        string
	Port        int
	Connected   bool
	LastError   error
	ConnectedAt time.Time
	Metadata    map[string]interface{}
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
