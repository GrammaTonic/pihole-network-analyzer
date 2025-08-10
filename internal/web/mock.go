package web

import (
	"context"
	"time"

	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/types"
)

// MockWebDataSource provides test data for web interface demonstration
type MockWebDataSource struct {
	connected bool
}

// NewMockWebDataSource creates a mock data source for web testing
func NewMockWebDataSource() *MockWebDataSource {
	return &MockWebDataSource{
		connected: true,
	}
}

// Implement DataSourceProvider interface
func (m *MockWebDataSource) GetAnalysisResult(ctx context.Context) (*types.AnalysisResult, error) {
	return &types.AnalysisResult{
		ClientStats: map[string]*types.ClientStats{
			"192.168.1.100": {
				IP:          "192.168.1.100",
				Hostname:    "desktop-pc",
				QueryCount:  5000,
				DomainCount: 150,
				Domains: map[string]int{
					"google.com":        2000,
					"github.com":        1500,
					"stackoverflow.com": 1500,
				},
				MACAddress: "aa:bb:cc:dd:ee:ff",
				IsOnline:   true,
				LastSeen:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				TopDomains: []types.DomainStat{
					{Domain: "google.com", Count: 2000},
					{Domain: "github.com", Count: 1500},
					{Domain: "stackoverflow.com", Count: 1500},
				},
			},
			"192.168.1.101": {
				IP:          "192.168.1.101",
				Hostname:    "laptop",
				QueryCount:  3000,
				DomainCount: 120,
				Domains: map[string]int{
					"facebook.com": 1500,
					"google.com":   1000,
					"twitter.com":  500,
				},
				MACAddress: "bb:cc:dd:ee:ff:aa",
				IsOnline:   true,
				LastSeen:   time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
				TopDomains: []types.DomainStat{
					{Domain: "facebook.com", Count: 1500},
					{Domain: "google.com", Count: 1000},
					{Domain: "twitter.com", Count: 500},
				},
			},
			"192.168.1.102": {
				IP:          "192.168.1.102",
				Hostname:    "phone",
				QueryCount:  2000,
				DomainCount: 80,
				Domains: map[string]int{
					"instagram.com": 800,
					"google.com":    700,
					"spotify.com":   500,
				},
				MACAddress: "cc:dd:ee:ff:aa:bb",
				IsOnline:   false,
				LastSeen:   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				TopDomains: []types.DomainStat{
					{Domain: "instagram.com", Count: 800},
					{Domain: "google.com", Count: 700},
					{Domain: "spotify.com", Count: 500},
				},
			},
		},
		NetworkDevices: []types.NetworkDevice{
			{
				IP:       "192.168.1.100",
				Hardware: "aa:bb:cc:dd:ee:ff",
				Name:     "desktop-pc",
				MAC:      "aa:bb:cc:dd:ee:ff",
				Hostname: "desktop-pc",
				Type:     "computer",
				IsOnline: true,
			},
			{
				IP:       "192.168.1.101",
				Hardware: "bb:cc:dd:ee:ff:aa",
				Name:     "laptop",
				MAC:      "bb:cc:dd:ee:ff:aa",
				Hostname: "laptop",
				Type:     "computer",
				IsOnline: true,
			},
			{
				IP:       "192.168.1.102",
				Hardware: "cc:dd:ee:ff:aa:bb",
				Name:     "phone",
				MAC:      "cc:dd:ee:ff:aa:bb",
				Hostname: "phone",
				Type:     "mobile",
				IsOnline: false,
			},
		},
		TotalQueries:   10000,
		UniqueClients:  3,
		AnalysisMode:   "mock",
		DataSourceType: "mock",
		Timestamp:      time.Now().Format(time.RFC3339),
		Performance: &types.QueryPerformance{
			AverageResponseTime: 15.5,
			TotalQueries:        10000,
			QueriesPerSecond:    2.78,
			PeakQueries:         45,
			SlowQueries:         12,
		},
	}, nil
}

func (m *MockWebDataSource) GetConnectionStatus() *types.ConnectionStatus {
	return &types.ConnectionStatus{
		Connected:    m.connected,
		LastConnect:  time.Now().Format(time.RFC3339),
		ResponseTime: 45.2,
		Metadata: map[string]string{
			"type":        "mock",
			"mode":        "demo",
			"data_source": "mock_web_source",
		},
	}
}

// ProductionMockDataSource implements interfaces.DataSource for main application demo
type ProductionMockDataSource struct {
	connected bool
}

// NewMockDataSourceForProduction creates a production mock data source for demo purposes
func NewMockDataSourceForProduction() *ProductionMockDataSource {
	return &ProductionMockDataSource{
		connected: true,
	}
}

// Implement interfaces.DataSource interface
func (m *ProductionMockDataSource) Connect(ctx context.Context) error {
	return nil
}

func (m *ProductionMockDataSource) Close() error {
	return nil
}

func (m *ProductionMockDataSource) IsConnected() bool {
	return m.connected
}

func (m *ProductionMockDataSource) GetQueries(ctx context.Context, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	return []types.PiholeRecord{
		{
			ID:        1,
			DateTime:  time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05"),
			Domain:    "example.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: time.Now().Add(-time.Hour).Format(time.RFC3339),
		},
		{
			ID:        2,
			DateTime:  time.Now().Add(-30 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.101",
			QueryType: "AAAA",
			Status:    0,
			Timestamp: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		},
	}, nil
}

func (m *ProductionMockDataSource) GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error) {
	return map[string]*types.ClientStats{
		"192.168.1.100": {
			IP:          "192.168.1.100",
			Hostname:    "demo-desktop",
			QueryCount:  150,
			DomainCount: 45,
		},
	}, nil
}

func (m *ProductionMockDataSource) GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error) {
	return []types.NetworkDevice{}, nil
}

func (m *ProductionMockDataSource) GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error) {
	return &types.DomainAnalysis{}, nil
}

func (m *ProductionMockDataSource) GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error) {
	return &types.QueryPerformance{}, nil
}

func (m *ProductionMockDataSource) GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error) {
	return &types.ConnectionStatus{
		Connected:   m.connected,
		LastConnect: time.Now().Format(time.RFC3339),
	}, nil
}

func (m *ProductionMockDataSource) GetDataSourceType() interfaces.DataSourceType {
	return interfaces.DataSourceTypeAPI
}

func (m *ProductionMockDataSource) GetConnectionInfo() *interfaces.ConnectionInfo {
	return &interfaces.ConnectionInfo{
		Type:      interfaces.DataSourceTypeAPI,
		Connected: m.connected,
	}
}
