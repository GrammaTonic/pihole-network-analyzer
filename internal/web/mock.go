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

// Implement interfaces.DataSource interface
func (m *MockWebDataSource) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

func (m *MockWebDataSource) Close() error {
	m.connected = false
	return nil
}

func (m *MockWebDataSource) IsConnected() bool {
	return m.connected
}

func (m *MockWebDataSource) GetQueries(ctx context.Context, params interfaces.QueryParams) ([]types.PiholeRecord, error) {
	// Return realistic mock data
	now := time.Now()
	return []types.PiholeRecord{
		{
			ID:        1,
			DateTime:  now.Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2, // Forwarded
			Timestamp: now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:        2,
			DateTime:  now.Add(-50 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "facebook.com",
			Client:    "192.168.1.101",
			QueryType: "A",
			Status:    1, // Blocked
			Timestamp: now.Add(-50 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        3,
			DateTime:  now.Add(-45 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "github.com",
			Client:    "192.168.1.102",
			QueryType: "A",
			Status:    2, // Forwarded
			Timestamp: now.Add(-45 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        4,
			DateTime:  now.Add(-30 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "doubleclick.net",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    1, // Blocked
			Timestamp: now.Add(-30 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        5,
			DateTime:  now.Add(-25 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "stackoverflow.com",
			Client:    "192.168.1.103",
			QueryType: "A",
			Status:    2, // Forwarded
			Timestamp: now.Add(-25 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        6,
			DateTime:  now.Add(-20 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "ads.yahoo.com",
			Client:    "192.168.1.101",
			QueryType: "A",
			Status:    1, // Blocked
			Timestamp: now.Add(-20 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        7,
			DateTime:  now.Add(-15 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "example.com",
			Client:    "192.168.1.104",
			QueryType: "AAAA",
			Status:    2, // Forwarded
			Timestamp: now.Add(-15 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        8,
			DateTime:  now.Add(-10 * time.Minute).Format("2006-01-02 15:04:05"),
			Domain:    "tracker.malware.com",
			Client:    "192.168.1.102",
			QueryType: "A",
			Status:    1, // Blocked
			Timestamp: now.Add(-10 * time.Minute).Format(time.RFC3339),
		},
	}, nil
}

func (m *MockWebDataSource) GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error) {
	return map[string]*types.ClientStats{
		"192.168.1.100": {
			IP:            "192.168.1.100",
			Hostname:      "laptop-john",
			QueryCount:    45,
			DomainCount:   12,
			MACAddress:    "aa:bb:cc:dd:ee:01",
			IsOnline:      true,
			LastSeen:      time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			QueryTypes:    map[int]int{1: 40, 28: 5},
			StatusCodes:   map[int]int{1: 10, 2: 35},
		},
		"192.168.1.101": {
			IP:            "192.168.1.101",
			Hostname:      "phone-alice",
			QueryCount:    32,
			DomainCount:   8,
			MACAddress:    "aa:bb:cc:dd:ee:02",
			IsOnline:      true,
			LastSeen:      time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			QueryTypes:    map[int]int{1: 30, 28: 2},
			StatusCodes:   map[int]int{1: 8, 2: 24},
		},
		"192.168.1.102": {
			IP:            "192.168.1.102",
			Hostname:      "desktop-bob",
			QueryCount:    67,
			DomainCount:   23,
			MACAddress:    "aa:bb:cc:dd:ee:03",
			IsOnline:      false,
			LastSeen:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			QueryTypes:    map[int]int{1: 60, 28: 7},
			StatusCodes:   map[int]int{1: 15, 2: 52},
		},
		"192.168.1.103": {
			IP:            "192.168.1.103",
			Hostname:      "tablet-charlie",
			QueryCount:    18,
			DomainCount:   6,
			MACAddress:    "aa:bb:cc:dd:ee:04",
			IsOnline:      true,
			LastSeen:      time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			QueryTypes:    map[int]int{1: 15, 28: 3},
			StatusCodes:   map[int]int{1: 3, 2: 15},
		},
		"192.168.1.104": {
			IP:            "192.168.1.104",
			Hostname:      "iot-device",
			QueryCount:    8,
			DomainCount:   3,
			MACAddress:    "aa:bb:cc:dd:ee:05",
			IsOnline:      true,
			LastSeen:      time.Now().Add(-30 * time.Second).Format(time.RFC3339),
			QueryTypes:    map[int]int{1: 6, 28: 2},
			StatusCodes:   map[int]int{1: 1, 2: 7},
		},
	}, nil
}

func (m *MockWebDataSource) GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error) {
	return []types.NetworkDevice{
		{
			IP:       "192.168.1.100",
			MAC:      "aa:bb:cc:dd:ee:01",
			Hostname: "laptop-john",
			IsOnline: true,
		},
		{
			IP:       "192.168.1.101",
			MAC:      "aa:bb:cc:dd:ee:02",
			Hostname: "phone-alice",
			IsOnline: true,
		},
		{
			IP:       "192.168.1.102",
			MAC:      "aa:bb:cc:dd:ee:03",
			Hostname: "desktop-bob",
			IsOnline: false,
		},
		{
			IP:       "192.168.1.103",
			MAC:      "aa:bb:cc:dd:ee:04",
			Hostname: "tablet-charlie",
			IsOnline: true,
		},
		{
			IP:       "192.168.1.104",
			MAC:      "aa:bb:cc:dd:ee:05",
			Hostname: "iot-device",
			IsOnline: true,
		},
	}, nil
}

func (m *MockWebDataSource) GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error) {
	return &types.DomainAnalysis{
		TopDomains: []types.DomainCount{
			{Domain: "google.com", Count: 25},
			{Domain: "github.com", Count: 18},
			{Domain: "stackoverflow.com", Count: 12},
		},
		BlockedDomains: []types.DomainCount{
			{Domain: "facebook.com", Count: 15},
			{Domain: "doubleclick.net", Count: 10},
			{Domain: "ads.yahoo.com", Count: 8},
		},
		QueryTypes:     map[string]int{"A": 120, "AAAA": 15, "PTR": 5},
		TotalQueries:   170,
		TotalBlocked:   37,
		BlockedPercent: 21.8,
	}, nil
}

func (m *MockWebDataSource) GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error) {
	return &types.QueryPerformance{
		AverageResponseTime: 45.2,
		TotalQueries:        170,
		QueriesPerSecond:    2.3,
		PeakQueries:         8,
		SlowQueries:         3,
	}, nil
}

func (m *MockWebDataSource) GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error) {
	return &types.ConnectionStatus{
		Connected:    m.connected,
		LastConnect:  time.Now().Format(time.RFC3339),
		ResponseTime: 45.2,
		Metadata: map[string]string{
			"type":        "mock",
			"mode":        "demo",
			"data_source": "mock_web_source",
		},
	}, nil
}

func (m *MockWebDataSource) GetDataSourceType() interfaces.DataSourceType {
	return interfaces.DataSourceTypeAPI
}

func (m *MockWebDataSource) GetConnectionInfo() *interfaces.ConnectionInfo {
	return &interfaces.ConnectionInfo{
		Type:      interfaces.DataSourceTypeAPI,
		Host:      "mock-pihole",
		Port:      80,
		Connected: m.connected,
		Metadata: map[string]interface{}{
			"mock": true,
			"demo": true,
		},
	}
}