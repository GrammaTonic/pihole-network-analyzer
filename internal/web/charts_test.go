package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// TestMockDataSource implements DataSourceProvider for testing
type TestMockDataSource struct {
	AnalysisResult   *types.AnalysisResult
	ConnectionStatus *types.ConnectionStatus
}

func (m *TestMockDataSource) GetAnalysisResult(ctx context.Context) (*types.AnalysisResult, error) {
	return m.AnalysisResult, nil
}

func (m *TestMockDataSource) GetConnectionStatus() *types.ConnectionStatus {
	return m.ConnectionStatus
}

var (
	testTimeStart = time.Date(2025, 8, 10, 14, 0, 0, 0, time.UTC)
	testTimeEnd   = time.Date(2025, 8, 10, 16, 0, 0, 0, time.UTC)
)

func TestChartAPIHandler_HandleTimelineChart(t *testing.T) {
	// Setup
	mockDataSource := &TestMockDataSource{
		AnalysisResult: &types.AnalysisResult{
			TotalQueries:  10000,
			UniqueClients: 25,
			ClientStats: map[string]*types.ClientStats{
				"192.168.1.100": {
					IP:          "192.168.1.100",
					Hostname:    "desktop-pc",
					QueryCount:  5000,
					DomainCount: 150,
					IsOnline:    true,
				},
				"192.168.1.101": {
					IP:          "192.168.1.101",
					Hostname:    "laptop",
					QueryCount:  3000,
					DomainCount: 120,
					IsOnline:    true,
				},
			},
		},
		ConnectionStatus: &types.ConnectionStatus{
			Connected:   true,
			LastConnect: "2025-08-10T15:30:00Z",
		},
	}

	testLogger := logger.New(&logger.Config{Component: "test"})
	chartHandler := NewChartAPIHandler(mockDataSource, testLogger)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "Default parameters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var chartData ChartData
				if err := json.Unmarshal([]byte(body), &chartData); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(chartData.Labels) == 0 {
					t.Error("Expected labels to be present")
				}

				if len(chartData.Datasets) == 0 {
					t.Error("Expected datasets to be present")
				}

				// Verify metadata
				if chartData.Metadata["timeWindow"] != "24h" {
					t.Errorf("Expected default time window to be 24h, got %v", chartData.Metadata["timeWindow"])
				}
			},
		},
		{
			name:           "Custom time window",
			queryParams:    "?window=1h&granularity=15m",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var chartData ChartData
				if err := json.Unmarshal([]byte(body), &chartData); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if chartData.Metadata["timeWindow"] != "1h" {
					t.Errorf("Expected time window to be 1h, got %v", chartData.Metadata["timeWindow"])
				}

				if chartData.Metadata["granularity"] != "15m" {
					t.Errorf("Expected granularity to be 15m, got %v", chartData.Metadata["granularity"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/charts/timeline"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			chartHandler.HandleTimelineChart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
		})
	}
}

func TestChartAPIHandler_HandleClientChart(t *testing.T) {
	// Setup
	mockDataSource := &TestMockDataSource{
		AnalysisResult: &types.AnalysisResult{
			TotalQueries:  10000,
			UniqueClients: 3,
			ClientStats: map[string]*types.ClientStats{
				"192.168.1.100": {
					IP:          "192.168.1.100",
					Hostname:    "desktop-pc",
					QueryCount:  5000,
					DomainCount: 150,
					IsOnline:    true,
				},
				"192.168.1.101": {
					IP:          "192.168.1.101",
					Hostname:    "laptop",
					QueryCount:  3000,
					DomainCount: 120,
					IsOnline:    true,
				},
				"192.168.1.102": {
					IP:          "192.168.1.102",
					Hostname:    "phone",
					QueryCount:  2000,
					DomainCount: 80,
					IsOnline:    false,
				},
			},
		},
		ConnectionStatus: &types.ConnectionStatus{
			Connected:   true,
			LastConnect: "2025-08-10T15:30:00Z",
		},
	}

	testLogger := logger.New(&logger.Config{Component: "test"})
	chartHandler := NewChartAPIHandler(mockDataSource, testLogger)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "Default pie chart",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var chartData ChartData
				if err := json.Unmarshal([]byte(body), &chartData); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if len(chartData.Labels) != 3 {
					t.Errorf("Expected 3 client labels, got %d", len(chartData.Labels))
				}

				if len(chartData.Datasets) != 1 {
					t.Errorf("Expected 1 dataset, got %d", len(chartData.Datasets))
				}

				// Check that desktop-pc is first (highest query count)
				if chartData.Labels[0] != "desktop-pc" {
					t.Errorf("Expected first label to be 'desktop-pc', got %s", chartData.Labels[0])
				}

				// Verify metadata
				if chartData.Metadata["chartType"] != "pie" {
					t.Errorf("Expected chart type to be pie, got %v", chartData.Metadata["chartType"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/charts/clients"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			chartHandler.HandleClientChart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
		})
	}
}

func TestServer_HandleEnhancedDashboard(t *testing.T) {
	// Setup
	mockDataSource := &TestMockDataSource{
		AnalysisResult: &types.AnalysisResult{
			TotalQueries:  5000,
			UniqueClients: 10,
			ClientStats: map[string]*types.ClientStats{
				"192.168.1.100": {
					IP:          "192.168.1.100",
					Hostname:    "test-device",
					QueryCount:  1000,
					DomainCount: 50,
					IsOnline:    true,
				},
			},
		},
		ConnectionStatus: &types.ConnectionStatus{
			Connected:   true,
			LastConnect: "2025-08-10T15:30:00Z",
		},
	}

	testLogger := logger.New(&logger.Config{Component: "test"})
	server, err := NewServer(&Config{Port: 8080, Host: "localhost"}, mockDataSource, testLogger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "GET enhanced dashboard",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				// Check that it contains enhanced dashboard elements
				if !strings.Contains(body, "Enhanced Network Analysis Dashboard") {
					t.Error("Enhanced dashboard title not found")
				}

				if !strings.Contains(body, "chart.js") {
					t.Error("Chart.js library not loaded")
				}

				if !strings.Contains(body, "d3js.org") {
					t.Error("D3.js library not loaded")
				}

				if !strings.Contains(body, "timeline-chart") {
					t.Error("Timeline chart element not found")
				}

				if !strings.Contains(body, "client-chart") {
					t.Error("Client chart element not found")
				}

				if !strings.Contains(body, "topology-chart") {
					t.Error("Topology chart element not found")
				}

				// Check that data is populated
				if !strings.Contains(body, "5000") { // Total queries
					t.Error("Analysis data not populated")
				}
			},
		},
		{
			name:           "POST not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			validateBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/enhanced", nil)
			w := httptest.NewRecorder()

			server.handleEnhancedDashboard(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
		})
	}
}

func TestGenerateTimeLabels(t *testing.T) {
	handler := &ChartAPIHandler{}

	tests := []struct {
		name        string
		granularity string
		expectError bool
		minLabels   int
	}{
		{
			name:        "Valid 1 hour granularity",
			granularity: "1h",
			expectError: false,
			minLabels:   1,
		},
		{
			name:        "Valid 15 minute granularity",
			granularity: "15m",
			expectError: false,
			minLabels:   1,
		},
		{
			name:        "Invalid granularity",
			granularity: "invalid",
			expectError: true,
			minLabels:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := testTimeStart
			end := testTimeEnd

			labels, err := handler.generateTimeLabels(start, end, tt.granularity)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(labels) < tt.minLabels {
				t.Errorf("Expected at least %d labels, got %d", tt.minLabels, len(labels))
			}
		})
	}
}
