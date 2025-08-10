package interfaces

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestLogWriter(t *testing.T) {
	// Create a mock log integration
	mockIntegration := &mockLogsIntegration{
		logs: make([]LogEntry, 0),
	}

	// Create log writer
	labels := map[string]string{
		"service": "test",
		"env":     "development",
	}
	writer := NewLogWriter(mockIntegration, 2, labels) // level 2 = INFO

	// Test writing log data
	testMessage := "Test log message"
	n, err := writer.Write([]byte(testMessage))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if n != len(testMessage) {
		t.Fatalf("Expected %d bytes written, got %d", len(testMessage), n)
	}

	// Verify log was received
	if len(mockIntegration.logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(mockIntegration.logs))
	}

	log := mockIntegration.logs[0]
	if log.Message != testMessage {
		t.Errorf("Expected message %q, got %q", testMessage, log.Message)
	}

	if log.Labels["service"] != "test" {
		t.Errorf("Expected service label 'test', got %q", log.Labels["service"])
	}

	if log.Labels["env"] != "development" {
		t.Errorf("Expected env label 'development', got %q", log.Labels["env"])
	}
}

func TestLogWriterWithNilIntegration(t *testing.T) {
	writer := NewLogWriter(nil, 2, nil)

	// Should not error with nil integration
	n, err := writer.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Expected no error with nil integration, got %v", err)
	}

	if n != 4 {
		t.Fatalf("Expected 4 bytes written, got %d", n)
	}
}

func TestIntegrationStatus(t *testing.T) {
	status := IntegrationStatus{
		Name:        "test-integration",
		Enabled:     true,
		Connected:   true,
		LastConnect: time.Now(),
		LastError:   "",
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	if status.Name != "test-integration" {
		t.Errorf("Expected name 'test-integration', got %q", status.Name)
	}

	if !status.Enabled {
		t.Error("Expected integration to be enabled")
	}

	if !status.Connected {
		t.Error("Expected integration to be connected")
	}

	if status.Metadata["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %q", status.Metadata["version"])
	}
}

func TestLogEntry(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		Timestamp: now,
		Level:     2, // INFO level
		Message:   "Test message",
		Component: "test-component",
		Labels: map[string]string{
			"environment": "test",
		},
		Fields: map[string]interface{}{
			"user_id": 123,
			"action":  "login",
		},
	}

	if !entry.Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, entry.Timestamp)
	}

	if entry.Level != 2 {
		t.Errorf("Expected level 2, got %d", entry.Level)
	}

	if entry.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %q", entry.Message)
	}

	if entry.Component != "test-component" {
		t.Errorf("Expected component 'test-component', got %q", entry.Component)
	}

	if entry.Labels["environment"] != "test" {
		t.Errorf("Expected environment label 'test', got %q", entry.Labels["environment"])
	}

	if entry.Fields["user_id"] != 123 {
		t.Errorf("Expected user_id field 123, got %v", entry.Fields["user_id"])
	}

	if entry.Fields["action"] != "login" {
		t.Errorf("Expected action field 'login', got %v", entry.Fields["action"])
	}
}

func TestMetricTypes(t *testing.T) {
	tests := []struct {
		metricType MetricType
		expected   string
	}{
		{MetricTypeCounter, "counter"},
		{MetricTypeGauge, "gauge"},
		{MetricTypeHistogram, "histogram"},
		{MetricTypeSummary, "summary"},
	}

	for _, test := range tests {
		if string(test.metricType) != test.expected {
			t.Errorf("Expected metric type %q, got %q", test.expected, string(test.metricType))
		}
	}
}

func TestDashboard(t *testing.T) {
	dashboard := Dashboard{
		ID:          "test-dashboard",
		Title:       "Test Dashboard",
		Description: "A test dashboard",
		FolderID:    "test-folder",
		Tags:        []string{"test", "monitoring"},
		Panels: []Panel{
			{
				ID:    1,
				Title: "Test Panel",
				Type:  "graph",
				Query: "up",
			},
		},
		Variables: []Variable{
			{
				Name:    "environment",
				Type:    "query",
				Query:   "label_values(environment)",
				Default: "production",
			},
		},
		Metadata: map[string]string{
			"author": "test-user",
		},
		Definition: map[string]interface{}{
			"version": 1,
		},
	}

	if dashboard.ID != "test-dashboard" {
		t.Errorf("Expected ID 'test-dashboard', got %q", dashboard.ID)
	}

	if dashboard.Title != "Test Dashboard" {
		t.Errorf("Expected title 'Test Dashboard', got %q", dashboard.Title)
	}

	if len(dashboard.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(dashboard.Tags))
	}

	if dashboard.Tags[0] != "test" || dashboard.Tags[1] != "monitoring" {
		t.Errorf("Expected tags [test, monitoring], got %v", dashboard.Tags)
	}

	if len(dashboard.Panels) != 1 {
		t.Errorf("Expected 1 panel, got %d", len(dashboard.Panels))
	}

	panel := dashboard.Panels[0]
	if panel.Title != "Test Panel" {
		t.Errorf("Expected panel title 'Test Panel', got %q", panel.Title)
	}

	if len(dashboard.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(dashboard.Variables))
	}

	variable := dashboard.Variables[0]
	if variable.Name != "environment" {
		t.Errorf("Expected variable name 'environment', got %q", variable.Name)
	}
}

func TestAlertRule(t *testing.T) {
	rule := AlertRule{
		ID:        "test-alert",
		Name:      "High CPU Usage",
		Query:     "cpu_usage > 90",
		Condition: "threshold",
		Threshold: 90.0,
		Duration:  5 * time.Minute,
		Severity:  "critical",
		Labels: map[string]string{
			"team": "infrastructure",
		},
		Annotations: map[string]string{
			"description": "CPU usage is above 90%",
		},
		Enabled: true,
	}

	if rule.ID != "test-alert" {
		t.Errorf("Expected ID 'test-alert', got %q", rule.ID)
	}

	if rule.Name != "High CPU Usage" {
		t.Errorf("Expected name 'High CPU Usage', got %q", rule.Name)
	}

	if rule.Threshold != 90.0 {
		t.Errorf("Expected threshold 90.0, got %f", rule.Threshold)
	}

	if rule.Duration != 5*time.Minute {
		t.Errorf("Expected duration 5m, got %v", rule.Duration)
	}

	if !rule.Enabled {
		t.Error("Expected alert rule to be enabled")
	}

	if rule.Labels["team"] != "infrastructure" {
		t.Errorf("Expected team label 'infrastructure', got %q", rule.Labels["team"])
	}

	if rule.Annotations["description"] != "CPU usage is above 90%" {
		t.Errorf("Expected description annotation, got %q", rule.Annotations["description"])
	}
}

// Mock implementations for testing

type mockLogsIntegration struct {
	logs []LogEntry
}

func (m *mockLogsIntegration) Initialize(ctx context.Context, config interface{}) error {
	return nil
}

func (m *mockLogsIntegration) IsEnabled() bool {
	return true
}

func (m *mockLogsIntegration) GetName() string {
	return "mock-logs"
}

func (m *mockLogsIntegration) GetStatus() IntegrationStatus {
	return IntegrationStatus{
		Name:      "mock-logs",
		Enabled:   true,
		Connected: true,
	}
}

func (m *mockLogsIntegration) SendMetrics(ctx context.Context, data *types.AnalysisResult) error {
	return nil
}

func (m *mockLogsIntegration) SendLogs(ctx context.Context, logs []LogEntry) error {
	m.logs = append(m.logs, logs...)
	return nil
}

func (m *mockLogsIntegration) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockLogsIntegration) Close() error {
	return nil
}

func (m *mockLogsIntegration) StreamLogs(ctx context.Context, logStream <-chan LogEntry) error {
	return nil
}

func (m *mockLogsIntegration) WriteLogs(ctx context.Context, logs []LogEntry) error {
	return m.SendLogs(ctx, logs)
}

func (m *mockLogsIntegration) SetLogLevel(level slog.Level) {
	// No-op for mock
}

func (m *mockLogsIntegration) AddStaticLabels(labels map[string]string) {
	// No-op for mock
}
