package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"pihole-analyzer/internal/integrations"
	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

// TestIntegrationsManagerBasic tests basic manager functionality without network dependencies
func TestIntegrationsManagerBasic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := integrations.NewManager(logger)

	// Test 1: Manager creation and state
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.IsInitialized() {
		t.Error("Manager should not be initialized initially")
	}

	// Test 2: Initialize with disabled integrations
	config := &types.IntegrationsConfig{
		Enabled: false,
	}

	ctx := context.Background()
	err := manager.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize with disabled config: %v", err)
	}

	if manager.IsInitialized() {
		t.Error("Manager should not be initialized when integrations are disabled")
	}

	// Test 3: Initialize with enabled but individually disabled integrations
	config = &types.IntegrationsConfig{
		Enabled: true,
		Grafana: types.GrafanaConfig{Enabled: false},
		Loki:    types.LokiConfig{Enabled: false},
		Prometheus: types.PrometheusExtConfig{Enabled: false},
	}

	err = manager.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize with individually disabled integrations: %v", err)
	}

	if !manager.IsInitialized() {
		t.Error("Manager should be initialized even with all integrations disabled")
	}

	// Test 4: Send data to disabled integrations (should succeed as no-op)
	analysisData := &types.AnalysisResult{
		TotalQueries:   100,
		UniqueClients:  5,
		AnalysisMode:   "test",
		DataSourceType: "pihole-api",
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	err = manager.SendToAll(ctx, analysisData)
	if err != nil {
		t.Fatalf("SendToAll should succeed with disabled integrations: %v", err)
	}

	// Test 5: Send logs (should succeed as no-op)
	logs := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     slog.LevelInfo,
			Message:   "Test log",
			Component: "test",
		},
	}

	err = manager.SendLogs(ctx, logs)
	if err != nil {
		t.Fatalf("SendLogs should succeed with disabled integrations: %v", err)
	}

	// Test 6: Get status
	status := manager.GetStatus()
	// Should be empty since no integrations are enabled
	if len(status) != 0 {
		t.Errorf("Expected 0 status entries for disabled integrations, got %d", len(status))
	}

	// Test 7: Test connections
	testResults := manager.TestAll(ctx)
	// Should be empty since no integrations are enabled
	if len(testResults) != 0 {
		t.Errorf("Expected 0 test results for disabled integrations, got %d", len(testResults))
	}

	// Test 8: Close manager
	err = manager.Close()
	if err != nil {
		t.Fatalf("Failed to close manager: %v", err)
	}

	if manager.IsInitialized() {
		t.Error("Manager should not be initialized after close")
	}
}

// TestIntegrationsEndToEnd tests the complete integration ecosystem workflow
func TestIntegrationsEndToEnd(t *testing.T) {
	// Setup mock servers for each integration
	grafanaServer := setupMockGrafanaServer(t)
	defer grafanaServer.Close()

	lokiServer := setupMockLokiServer(t)
	defer lokiServer.Close()

	prometheusServer := setupMockPrometheusServer(t)
	defer prometheusServer.Close()

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create integration manager
	manager := integrations.NewManager(logger)

	// Create configuration for all integrations
	config := &types.IntegrationsConfig{
		Enabled: true,
		Grafana: types.GrafanaConfig{
			Enabled: true,
			URL:     grafanaServer.URL,
			APIKey:  "test-grafana-key",
			DataSource: types.DataSourceConfig{
				CreateIfNotExists: true,
				Name:              "pihole-prometheus",
				Type:              "prometheus",
				URL:               prometheusServer.URL,
			},
			Dashboards: types.DashboardConfig{
				AutoProvision: true,
				FolderName:    "Pi-hole",
			},
			Timeout:    30,
			VerifyTLS:  false,
			RetryCount: 3,
		},
		Loki: types.LokiConfig{
			Enabled:      true,
			URL:          lokiServer.URL,
			BatchSize:    100,
			BatchTimeout: "10s",
		},
		Prometheus: types.PrometheusExtConfig{
			Enabled: true,
			PushGateway: types.PushGatewayConfig{
				Enabled: true,
				URL:     prometheusServer.URL,
			},
			ExternalLabels: map[string]string{
				"service":  "pihole-analyzer",
				"instance": "test-instance",
			},
		},
	}

	// Initialize manager
	ctx := context.Background()
	err := manager.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize integration manager: %v", err)
	}

	// Verify manager is initialized
	if !manager.IsInitialized() {
		t.Error("Manager should be initialized")
	}

	// Test individual integrations are registered and enabled
	enabled := manager.GetEnabledIntegrations()
	if len(enabled) != 3 {
		t.Errorf("Expected 3 enabled integrations, got %d", len(enabled))
	}

	// Test connection to all integrations
	testResults := manager.TestAll(ctx)
	for name, err := range testResults {
		if err != nil {
			t.Errorf("Integration %s failed connection test: %v", name, err)
		}
	}

	// Test sending analysis data to all integrations
	analysisData := &types.AnalysisResult{
		TotalQueries:   1000,
		UniqueClients:  25,
		AnalysisMode:   "full",
		DataSourceType: "pihole-api",
		Timestamp:      time.Now().Format(time.RFC3339),
		ClientStats: map[string]*types.ClientStats{
			"192.168.1.100": {
				IP:         "192.168.1.100",
				Hostname:   "test-client",
				QueryCount: 100,
				IsOnline:   true,
			},
		},
	}

	err = manager.SendToAll(ctx, analysisData)
	if err != nil {
		t.Fatalf("Failed to send analysis data to all integrations: %v", err)
	}

	// Test sending logs to log-capable integrations
	logEntries := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     slog.LevelInfo,
			Message:   "Test log message for integration testing",
			Component: "integration-test",
			Labels: map[string]string{
				"test_type": "e2e",
				"severity":  "info",
			},
		},
		{
			Timestamp: time.Now(),
			Level:     slog.LevelWarn,
			Message:   "Warning message for testing",
			Component: "integration-test",
			Labels: map[string]string{
				"test_type": "e2e",
				"severity":  "warning",
			},
		},
	}

	err = manager.SendLogs(ctx, logEntries)
	if err != nil {
		t.Fatalf("Failed to send logs: %v", err)
	}

	// Test getting status from all integrations
	status := manager.GetStatus()
	if len(status) != 3 {
		t.Errorf("Expected 3 integration statuses, got %d", len(status))
	}

	// Verify each integration status
	for name, integrationStatus := range status {
		if !integrationStatus.Enabled {
			t.Errorf("Integration %s should be enabled", name)
		}
		if !integrationStatus.Connected {
			t.Errorf("Integration %s should be connected", name)
		}
	}

	// Test graceful shutdown
	err = manager.Close()
	if err != nil {
		t.Fatalf("Failed to close integration manager: %v", err)
	}

	if manager.IsInitialized() {
		t.Error("Manager should not be initialized after close")
	}
}

// TestIntegrationsWithRealData tests integrations with realistic Pi-hole data
func TestIntegrationsWithRealData(t *testing.T) {
	// Setup mock servers
	grafanaServer := setupMockGrafanaServer(t)
	defer grafanaServer.Close()

	lokiServer := setupMockLokiServer(t)
	defer lokiServer.Close()

	prometheusServer := setupMockPrometheusServer(t)
	defer prometheusServer.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := integrations.NewManager(logger)

	config := &types.IntegrationsConfig{
		Enabled: true,
		Grafana: types.GrafanaConfig{
			Enabled: true,
			URL:     grafanaServer.URL,
			APIKey:  "test-key",
			DataSource: types.DataSourceConfig{
				CreateIfNotExists: true,
				Name:              "pihole-prometheus",
				Type:              "prometheus",
				URL:               prometheusServer.URL,
			},
			Dashboards: types.DashboardConfig{
				AutoProvision: true,
				FolderName:    "Pi-hole",
			},
		},
		Loki: types.LokiConfig{
			Enabled:      true,
			URL:          lokiServer.URL,
			BatchSize:    50,
			BatchTimeout: "5s",
		},
		Prometheus: types.PrometheusExtConfig{
			Enabled: true,
			PushGateway: types.PushGatewayConfig{
				Enabled: true,
				URL:     prometheusServer.URL,
			},
		},
	}

	ctx := context.Background()
	err := manager.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	// Simulate realistic Pi-hole analysis data
	realisticData := &types.AnalysisResult{
		TotalQueries:   5432,
		UniqueClients:  18,
		AnalysisMode:   "comprehensive",
		DataSourceType: "pihole-api",
		Timestamp:      time.Now().Format(time.RFC3339),
		ClientStats: map[string]*types.ClientStats{
			"192.168.1.100": {
				IP:         "192.168.1.100",
				Hostname:   "laptop-john",
				QueryCount: 892,
				IsOnline:   true,
			},
			"192.168.1.101": {
				IP:         "192.168.1.101",
				Hostname:   "phone-jane",
				QueryCount: 567,
				IsOnline:   true,
			},
			"192.168.1.102": {
				IP:         "192.168.1.102",
				Hostname:   "tablet-kids",
				QueryCount: 234,
				IsOnline:   false,
			},
		},
	}

	// Send realistic data
	err = manager.SendToAll(ctx, realisticData)
	if err != nil {
		t.Fatalf("Failed to send realistic data: %v", err)
	}

	// Send realistic logs
	realisticLogs := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     slog.LevelInfo,
			Message:   "Pi-hole analysis completed successfully",
			Component: "analyzer",
			Labels: map[string]string{
				"queries_processed": "5432",
				"clients_found":     "18",
			},
		},
		{
			Timestamp: time.Now(),
			Level:     slog.LevelWarn,
			Message:   "High query volume detected from client 192.168.1.100",
			Component: "anomaly-detector",
			Labels: map[string]string{
				"client_ip":    "192.168.1.100",
				"query_count":  "892",
				"threshold":    "500",
			},
		},
	}

	err = manager.SendLogs(ctx, realisticLogs)
	if err != nil {
		t.Fatalf("Failed to send realistic logs: %v", err)
	}

	manager.Close()
}

// TestIntegrationsErrorHandling tests error scenarios and recovery
func TestIntegrationsErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := integrations.NewManager(logger)

	// Test with all integrations disabled (should succeed)
	config := &types.IntegrationsConfig{
		Enabled: true,
		Grafana: types.GrafanaConfig{
			Enabled: false, // Disabled
		},
		Loki: types.LokiConfig{
			Enabled: false, // Disabled
		},
		Prometheus: types.PrometheusExtConfig{
			Enabled: false, // Disabled
		},
	}

	ctx := context.Background()
	err := manager.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Manager initialization should succeed with disabled integrations: %v", err)
	}

	// Try to send data to disabled integrations
	analysisData := &types.AnalysisResult{
		TotalQueries:   100,
		UniqueClients:  5,
		AnalysisMode:   "test",
		DataSourceType: "pihole-api",
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	// This should succeed (no-op for disabled integrations)
	err = manager.SendToAll(ctx, analysisData)
	if err != nil {
		t.Fatalf("SendToAll should succeed with disabled integrations: %v", err)
	}

	manager.Close()
}

// Helper functions to setup mock servers

func setupMockGrafanaServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/health"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"database": "ok",
				"version":  "test",
			})
		case strings.Contains(r.URL.Path, "/api/datasources"):
			if r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":      1,
					"name":    "pihole-prometheus",
					"message": "Datasource added",
				})
			} else {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]interface{}{})
			}
		case strings.Contains(r.URL.Path, "/api/dashboards"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      1,
				"uid":     "test-uid",
				"status":  "success",
				"version": 1,
			})
		case strings.Contains(r.URL.Path, "/api/folders"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"uid":   "test-folder-uid",
				"title": "Pi-hole",
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func setupMockLokiServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/ready"):
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ready")
		case strings.Contains(r.URL.Path, "/loki/api/v1/push"):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func setupMockPrometheusServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/-/healthy"):
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Prometheus is Healthy.")
		case strings.Contains(r.URL.Path, "/metrics/job"):
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "success")
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}