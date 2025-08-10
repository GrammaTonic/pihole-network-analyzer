package loki

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled:      true,
		URL:          "http://localhost:3100",
		BatchSize:    100,
		BatchTimeout: "10s",
		BufferSize:   1000,
	}

	client := NewClient(config, logger)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if !client.IsEnabled() {
		t.Error("Expected client to be enabled")
	}

	if client.GetName() != "loki" {
		t.Errorf("Expected name 'loki', got %q", client.GetName())
	}
}

func TestNewClientDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)

	if client.IsEnabled() {
		t.Error("Expected client to be disabled")
	}
}

func TestClientInitialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled:      false, // Disabled to avoid connection attempts
		URL:          "http://localhost:3100",
		BatchSize:    100,
		BatchTimeout: "10s",
		BufferSize:   1000,
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
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)
	status := client.GetStatus()

	if status.Name != "loki" {
		t.Errorf("Expected status name 'loki', got %q", status.Name)
	}

	if !status.Enabled {
		t.Error("Expected status to show enabled")
	}
}

func TestClientSendMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	data := &types.AnalysisResult{
		TotalQueries:  100,
		UniqueClients: 10,
	}

	// SendMetrics should be a no-op for Loki (it's for logs, not metrics)
	err := client.SendMetrics(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error from SendMetrics (should be no-op), got %v", err)
	}
}

func TestClientSendLogsDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	logs := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     2, // INFO
			Message:   "Test log message",
			Component: "test",
		},
	}

	err := client.SendLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Expected no error sending logs to disabled client, got %v", err)
	}
}

func TestClientWriteLogsDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	logs := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     2, // INFO
			Message:   "Test log message",
			Component: "test",
		},
	}

	err := client.WriteLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Expected no error writing logs to disabled client, got %v", err)
	}
}

func TestClientSetLogLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)

	// This should not error
	client.SetLogLevel(slog.LevelInfo)

	status := client.GetStatus()
	if status.Metadata == nil {
		t.Error("Expected metadata to be set")
	} else if status.Metadata["min_log_level"] != "INFO" {
		t.Errorf("Expected min_log_level 'INFO', got %q", status.Metadata["min_log_level"])
	}
}

func TestClientAddStaticLabels(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
		StaticLabels: map[string]string{
			"service": "pihole-analyzer",
		},
	}

	client := NewClient(config, logger)

	newLabels := map[string]string{
		"environment": "test",
		"version":     "1.0.0",
	}

	client.AddStaticLabels(newLabels)

	// Check that labels were merged
	if client.config.StaticLabels["service"] != "pihole-analyzer" {
		t.Error("Expected existing label to be preserved")
	}

	if client.config.StaticLabels["environment"] != "test" {
		t.Error("Expected new label to be added")
	}

	if client.config.StaticLabels["version"] != "1.0.0" {
		t.Error("Expected new label to be added")
	}
}

func TestClientClose(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: false, // Disabled to avoid connection attempts
		URL:     "http://localhost:3100",
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

func TestFormatLabels(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)

	tests := []struct {
		labels   map[string]string
		expected string
	}{
		{
			labels:   map[string]string{},
			expected: "{}",
		},
		{
			labels: map[string]string{
				"service": "pihole-analyzer",
			},
			expected: `{"service":"pihole-analyzer"}`,
		},
		{
			labels: map[string]string{
				"service": "pihole-analyzer",
				"level":   "info",
			},
			expected: `{"level":"info","service":"pihole-analyzer"}`, // Sorted alphabetically
		},
	}

	for _, test := range tests {
		result := client.formatLabels(test.labels)
		if result != test.expected {
			t.Errorf("Expected labels %q, got %q", test.expected, result)
		}
	}
}

func TestFormatLogLine(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)

	tests := []struct {
		log      interfaces.LogEntry
		expected string
	}{
		{
			log: interfaces.LogEntry{
				Message: "Simple log message",
				Fields:  nil,
			},
			expected: "Simple log message",
		},
		{
			log: interfaces.LogEntry{
				Message: "Log with fields",
				Fields: map[string]interface{}{
					"user_id": 123,
					"action":  "login",
				},
			},
			expected: `Log with fields {"action":"login","user_id":123}`,
		},
	}

	for _, test := range tests {
		result := client.formatLogLine(test.log)
		if result != test.expected {
			t.Errorf("Expected log line %q, got %q", test.expected, result)
		}
	}
}

func TestConvertToLokiStreams(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
		StaticLabels: map[string]string{
			"service": "pihole-analyzer",
		},
		DynamicLabels: []string{"level", "component"},
	}

	client := NewClient(config, logger)

	logs := []interfaces.LogEntry{
		{
			Timestamp: time.Unix(1609459200, 0), // Fixed timestamp for testing
			Level:     2,                        // INFO
			Message:   "Test message 1",
			Component: "test-component",
			Labels: map[string]string{
				"environment": "test",
			},
		},
		{
			Timestamp: time.Unix(1609459260, 0), // Different timestamp
			Level:     3,                        // WARN
			Message:   "Test message 2",
			Component: "test-component",
			Labels: map[string]string{
				"environment": "test",
			},
		},
	}

	streams := client.convertToLokiStreams(logs)

	// Should have two different streams due to different log levels
	expectedStreamCount := 2
	if len(streams) != expectedStreamCount {
		t.Errorf("Expected %d streams, got %d", expectedStreamCount, len(streams))
	}

	// Check that each stream has the correct number of entries
	for labelString, entries := range streams {
		if len(entries) != 1 {
			t.Errorf("Expected 1 entry per stream, got %d for stream %s", len(entries), labelString)
		}

		// Verify timestamp format
		entry := entries[0]
		if entry.Timestamp != "1609459200000000000" && entry.Timestamp != "1609459260000000000" {
			t.Errorf("Unexpected timestamp format: %s", entry.Timestamp)
		}

		// Verify message content
		if entry.Line != "Test message 1" && entry.Line != "Test message 2" {
			t.Errorf("Unexpected log line: %s", entry.Line)
		}
	}
}

// Test that interface is implemented correctly
func TestLokiClientImplementsInterfaces(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.LokiConfig{
		Enabled: true,
		URL:     "http://localhost:3100",
	}

	client := NewClient(config, logger)

	// Test MonitoringIntegration interface
	var _ interfaces.MonitoringIntegration = client

	// Test LogsIntegration interface
	var _ interfaces.LogsIntegration = client
}
