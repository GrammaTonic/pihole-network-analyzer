package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// MockDataSourceProvider implements DataSourceProvider for testing
type MockDataSourceProvider struct {
	analysisResult   *types.AnalysisResult
	connectionStatus *types.ConnectionStatus
	shouldError      bool
}

func NewMockDataSourceProvider() *MockDataSourceProvider {
	return &MockDataSourceProvider{
		analysisResult: &types.AnalysisResult{
			TotalQueries:  100,
			UniqueClients: 5,
			ClientStats: map[string]*types.ClientStats{
				"192.168.1.100": {
					IP:          "192.168.1.100",
					Hostname:    "test-device",
					QueryCount:  50,
					DomainCount: 10,
					MACAddress:  "aa:bb:cc:dd:ee:ff",
					IsOnline:    true,
				},
				"192.168.1.101": {
					IP:          "192.168.1.101",
					Hostname:    "laptop",
					QueryCount:  30,
					DomainCount: 8,
					MACAddress:  "11:22:33:44:55:66",
					IsOnline:    false,
				},
			},
			NetworkDevices: []types.NetworkDevice{
				{
					IP:       "192.168.1.100",
					MAC:      "aa:bb:cc:dd:ee:ff",
					Hostname: "test-device",
					IsOnline: true,
				},
			},
			DataSourceType: "mock",
			AnalysisMode:   "test",
			Timestamp:      time.Now().Format(time.RFC3339),
		},
		connectionStatus: &types.ConnectionStatus{
			Connected:    true,
			LastConnect:  time.Now().Format(time.RFC3339),
			ResponseTime: 100.5,
			Metadata: map[string]string{
				"test": "true",
			},
		},
		shouldError: false,
	}
}

func (m *MockDataSourceProvider) GetAnalysisResult(ctx context.Context) (*types.AnalysisResult, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.analysisResult, nil
}

func (m *MockDataSourceProvider) GetConnectionStatus() *types.ConnectionStatus {
	return m.connectionStatus
}

func (m *MockDataSourceProvider) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func TestNewServer(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())

	config := DefaultConfig()
	server, err := NewServer(config, mockProvider, logger)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.config.Port)
	}
}

func TestNewServerWithNilConfig(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())

	server, err := NewServer(nil, mockProvider, logger)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created with default config")
	}

	if server.config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", server.config.Port)
	}
}

func TestHandleDashboard(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	// Check that the response contains expected content
	expectedStrings := []string{
		"Pi-hole Network Analyzer",
		"Total Queries",
		"Unique Clients",
		"Client Statistics",
	}

	for _, expected := range expectedStrings {
		if !contains(body, expected) {
			t.Errorf("Expected response to contain '%s'", expected)
		}
	}
}

func TestHandleDashboardWithError(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	mockProvider.SetError(true)

	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleDashboard(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHandleAPIStatus(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	server.handleAPIStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	var status types.ConnectionStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if !status.Connected {
		t.Error("Expected status to be connected")
	}

	if status.ResponseTime <= 0 {
		t.Error("Expected positive response time")
	}
}

func TestHandleAPIAnalysis(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/analysis", nil)
	w := httptest.NewRecorder()

	server.handleAPIAnalysis(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result types.AnalysisResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if result.TotalQueries != 100 {
		t.Errorf("Expected 100 total queries, got %d", result.TotalQueries)
	}

	if result.UniqueClients != 5 {
		t.Errorf("Expected 5 unique clients, got %d", result.UniqueClients)
	}
}

func TestHandleAPIClients(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/clients?limit=10", nil)
	w := httptest.NewRecorder()

	server.handleAPIClients(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		Clients      []*types.ClientStats `json:"clients"`
		TotalCount   int                  `json:"total_count"`
		LimitApplied int                  `json:"limit_applied"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.TotalCount != 2 {
		t.Errorf("Expected 2 total clients, got %d", response.TotalCount)
	}

	if response.LimitApplied != 10 {
		t.Errorf("Expected limit 10, got %d", response.LimitApplied)
	}

	if len(response.Clients) != 2 {
		t.Errorf("Expected 2 clients in response, got %d", len(response.Clients))
	}
}

func TestHandleHealth(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var health struct {
		Status     string `json:"status"`
		Timestamp  string `json:"timestamp"`
		Connected  bool   `json:"connected"`
		ServerInfo string `json:"server_info"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}

	if !health.Connected {
		t.Error("Expected connected to be true")
	}

	if health.ServerInfo == "" {
		t.Error("Expected non-empty server info")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test POST to dashboard
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	server.handleDashboard(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap with logging middleware
	wrappedHandler := server.loggingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got '%s'", w.Body.String())
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Expected config to be created")
	}

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.Host)
	}

	if config.EnableWeb {
		t.Error("Expected EnableWeb to be false by default")
	}

	if config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected ReadTimeout 10s, got %v", config.ReadTimeout)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && s[0:len(substr)] == substr) ||
		(len(s) > len(substr) && contains(s[1:], substr)))
}

// Benchmark tests
func BenchmarkHandleAPIStatus(b *testing.B) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		w := httptest.NewRecorder()
		server.handleAPIStatus(w, req)
	}
}

func BenchmarkHandleAPIAnalysis(b *testing.B) {
	mockProvider := NewMockDataSourceProvider()
	logger := logger.New(logger.DefaultConfig())
	config := DefaultConfig()

	server, err := NewServer(config, mockProvider, logger)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/analysis", nil)
		w := httptest.NewRecorder()
		server.handleAPIAnalysis(w, req)
	}
}
