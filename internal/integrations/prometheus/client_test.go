package prometheus

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
		PushGateway: types.PushGatewayConfig{
			Enabled: true,
			URL:     "http://localhost:9091",
			Job:     "pihole-analyzer",
		},
	}

	client := NewClient(config, logger)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if !client.IsEnabled() {
		t.Error("Expected client to be enabled")
	}

	if client.GetName() != "prometheus" {
		t.Errorf("Expected name 'prometheus', got %q", client.GetName())
	}

	if client.GetRegistry() == nil {
		t.Error("Expected registry to be initialized")
	}
}

func TestNewClientDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)

	if client.IsEnabled() {
		t.Error("Expected client to be disabled")
	}
}

func TestClientInitialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: false, // Disabled to avoid connection attempts
		PushGateway: types.PushGatewayConfig{
			Enabled: false,
			URL:     "http://localhost:9091",
			Job:     "pihole-analyzer",
		},
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	err := client.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Expected no error initializing disabled client, got %v", err)
	}
}

func TestClientGetStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
		PushGateway: types.PushGatewayConfig{
			Enabled: true,
			URL:     "http://localhost:9091",
			Job:     "pihole-analyzer",
		},
	}

	client := NewClient(config, logger)
	status := client.GetStatus()

	if status.Name != "prometheus" {
		t.Errorf("Expected status name 'prometheus', got %q", status.Name)
	}

	if !status.Enabled {
		t.Error("Expected status to show enabled")
	}
}

func TestClientSendLogs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	logs := []interfaces.LogEntry{
		{
			Message: "Test log message",
		},
	}

	// SendLogs should be a no-op for Prometheus (it's for metrics, not logs)
	err := client.SendLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Expected no error from SendLogs (should be no-op), got %v", err)
	}
}

func TestClientSendMetricsDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	data := &types.AnalysisResult{
		TotalQueries:  100,
		UniqueClients: 10,
	}

	err := client.SendMetrics(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error sending metrics to disabled client, got %v", err)
	}
}

func TestClientRegisterMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	// Test registering a counter metric
	err := client.RegisterMetric("test_counter", "A test counter", interfaces.MetricTypeCounter, nil)
	if err != nil {
		t.Fatalf("Expected no error registering counter metric, got %v", err)
	}

	// Test registering a gauge metric with labels
	err = client.RegisterMetric("test_gauge", "A test gauge", interfaces.MetricTypeGauge, []string{"label1", "label2"})
	if err != nil {
		t.Fatalf("Expected no error registering gauge metric with labels, got %v", err)
	}

	// Test registering a histogram metric
	err = client.RegisterMetric("test_histogram", "A test histogram", interfaces.MetricTypeHistogram, nil)
	if err != nil {
		t.Fatalf("Expected no error registering histogram metric, got %v", err)
	}

	// Test registering duplicate metric (should fail)
	err = client.RegisterMetric("test_counter", "A duplicate counter", interfaces.MetricTypeCounter, nil)
	if err == nil {
		t.Error("Expected error when registering duplicate metric")
	}

	// Test unsupported metric type
	err = client.RegisterMetric("test_summary", "A test summary", interfaces.MetricTypeSummary, nil)
	if err == nil {
		t.Error("Expected error for unsupported metric type")
	}
}

func TestClientSetMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	// Register test metrics
	client.RegisterMetric("test_gauge", "A test gauge", interfaces.MetricTypeGauge, nil)
	client.RegisterMetric("test_gauge_vec", "A test gauge with labels", interfaces.MetricTypeGauge, []string{"label1"})
	client.RegisterMetric("test_counter", "A test counter", interfaces.MetricTypeCounter, nil)

	// Test setting gauge value
	err := client.SetMetric("test_gauge", 42.0, nil)
	if err != nil {
		t.Fatalf("Expected no error setting gauge metric, got %v", err)
	}

	// Test setting gauge with labels
	labels := map[string]string{"label1": "value1"}
	err = client.SetMetric("test_gauge_vec", 100.0, labels)
	if err != nil {
		t.Fatalf("Expected no error setting gauge metric with labels, got %v", err)
	}

	// Test setting counter (should add to counter)
	err = client.SetMetric("test_counter", 5.0, nil)
	if err != nil {
		t.Fatalf("Expected no error setting counter metric, got %v", err)
	}

	// Test setting non-existent metric
	err = client.SetMetric("non_existent", 1.0, nil)
	if err == nil {
		t.Error("Expected error when setting non-existent metric")
	}
}

func TestClientIncrementCounter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	// Register test metrics
	client.RegisterMetric("test_counter", "A test counter", interfaces.MetricTypeCounter, nil)
	client.RegisterMetric("test_counter_vec", "A test counter with labels", interfaces.MetricTypeCounter, []string{"label1"})
	client.RegisterMetric("test_gauge", "A test gauge", interfaces.MetricTypeGauge, nil)

	// Test incrementing counter
	err := client.IncrementCounter("test_counter", nil)
	if err != nil {
		t.Fatalf("Expected no error incrementing counter, got %v", err)
	}

	// Test incrementing counter with labels
	labels := map[string]string{"label1": "value1"}
	err = client.IncrementCounter("test_counter_vec", labels)
	if err != nil {
		t.Fatalf("Expected no error incrementing counter with labels, got %v", err)
	}

	// Test incrementing non-counter metric
	err = client.IncrementCounter("test_gauge", nil)
	if err == nil {
		t.Error("Expected error when incrementing non-counter metric")
	}

	// Test incrementing non-existent metric
	err = client.IncrementCounter("non_existent", nil)
	if err == nil {
		t.Error("Expected error when incrementing non-existent metric")
	}
}

func TestClientPushMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
		PushGateway: types.PushGatewayConfig{
			Enabled: false, // Disabled to avoid actual push attempts
		},
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	metrics := map[string]interface{}{
		"test.metric1": 42.0,
		"test.metric2": 100,
		"test.metric3": "123.45",
		"test.metric4": "invalid",
	}

	err := client.PushMetrics(ctx, metrics)
	if err != nil {
		t.Fatalf("Expected no error pushing metrics (push gateway disabled), got %v", err)
	}

	// Verify that custom metrics were registered
	_, exists := client.metrics["pihole_analyzer_ext_custom_test_metric1"]
	if !exists {
		t.Error("Expected custom metric to be registered")
	}
}

func TestClientClose(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
		PushGateway: types.PushGatewayConfig{
			Enabled: false, // Disabled to avoid connection attempts
		},
	}

	client := NewClient(config, logger)

	err := client.Close()
	if err != nil {
		t.Fatalf("Expected no error closing client, got %v", err)
	}

	if client.IsEnabled() {
		t.Error("Expected client to be disabled after close")
	}

	status := client.GetStatus()
	if status.Connected {
		t.Error("Expected status to show disconnected after close")
	}
}

func TestSetCustomMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		hasError bool
	}{
		{"test1", 42.0, 42.0, false},
		{"test2", float32(32.5), 32.5, false},
		{"test3", 100, 100.0, false},
		{"test4", int64(200), 200.0, false},
		{"test5", "123.45", 123.45, false},
		{"test6", "invalid", 0.0, true},
		{"test7", []string{"invalid"}, 0.0, true},
	}

	for _, test := range tests {
		err := client.setCustomMetric(test.name, test.value)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for test %s, got nil", test.name)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for test %s, got %v", test.name, err)
			}
		}
	}
}

func TestRegisterDefaultMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	err := client.registerDefaultMetrics()
	if err != nil {
		t.Fatalf("Expected no error registering default metrics, got %v", err)
	}

	// Check that default metrics were registered
	expectedMetrics := []string{
		"pihole_analyzer_ext_total_queries",
		"pihole_analyzer_ext_unique_clients",
		"pihole_analyzer_ext_analysis_duration",
		"pihole_analyzer_ext_client_queries",
		"pihole_analyzer_ext_domain_queries",
	}

	for _, metricName := range expectedMetrics {
		if _, exists := client.metrics[metricName]; !exists {
			t.Errorf("Expected metric %s to be registered", metricName)
		}
	}
}

func TestUpdateMetricsFromAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	// Register default metrics first
	err := client.registerDefaultMetrics()
	if err != nil {
		t.Fatalf("Failed to register default metrics: %v", err)
	}

	data := &types.AnalysisResult{
		TotalQueries:  1000,
		UniqueClients: 25,
		ClientStats: map[string]*types.ClientStats{
			"192.168.1.100": {
				IP:         "192.168.1.100",
				Hostname:   "test-device",
				QueryCount: 150,
			},
			"192.168.1.101": {
				IP:         "192.168.1.101",
				Hostname:   "another-device",
				QueryCount: 75,
			},
		},
	}

	err = client.updateMetricsFromAnalysis(data)
	if err != nil {
		t.Fatalf("Expected no error updating metrics, got %v", err)
	}

	// Since we can't easily read metric values in tests without exposing them,
	// we just verify that the method doesn't error. In a real scenario,
	// you might use a test registry or mock to verify the actual values.
}

func TestClientTestConnectionDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("Expected no error testing connection for disabled client, got %v", err)
	}
}

func TestClientTestConnectionPushGatewayDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
		PushGateway: types.PushGatewayConfig{
			Enabled: false,
		},
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("Expected no error testing connection when push gateway disabled, got %v", err)
	}
}

// Test that interface is implemented correctly
func TestPrometheusClientImplementsInterfaces(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.PrometheusExtConfig{
		Enabled: true,
	}

	client := NewClient(config, logger)

	// Test MonitoringIntegration interface
	var _ interfaces.MonitoringIntegration = client

	// Test MetricsIntegration interface
	var _ interfaces.MetricsIntegration = client
}

// Helper function for debugging
func getMetricNames(client *Client) []string {
	names := make([]string, 0, len(client.metrics))
	for name := range client.metrics {
		names = append(names, name)
	}
	return names
}
