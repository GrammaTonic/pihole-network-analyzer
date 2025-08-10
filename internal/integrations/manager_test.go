package integrations

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

func TestNewManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	if manager.logger != logger {
		t.Error("Expected logger to be set")
	}

	if manager.integrations == nil {
		t.Error("Expected integrations map to be initialized")
	}

	if manager.initialized {
		t.Error("Expected manager to not be initialized initially")
	}
}

func TestManagerInitialize_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	config := &types.IntegrationsConfig{
		Enabled: false,
	}

	ctx := context.Background()
	err := manager.Initialize(ctx, config)

	if err != nil {
		t.Fatalf("Expected no error when integrations disabled, got %v", err)
	}

	if manager.IsInitialized() {
		t.Error("Expected manager to not be initialized when disabled")
	}
}

func TestManagerInitialize_Enabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	config := &types.IntegrationsConfig{
		Enabled: true,
		Grafana: types.GrafanaConfig{
			Enabled: false, // Keep disabled for test
		},
		Loki: types.LokiConfig{
			Enabled: false, // Keep disabled for test
		},
		Prometheus: types.PrometheusExtConfig{
			Enabled: false, // Keep disabled for test
		},
	}

	ctx := context.Background()
	err := manager.Initialize(ctx, config)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !manager.IsInitialized() {
		t.Error("Expected manager to be initialized")
	}

	if manager.GetConfig() != config {
		t.Error("Expected config to be stored")
	}
}

func TestManagerRegisterIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockIntegration := &mockIntegration{
		name:    "test-integration",
		enabled: true,
	}

	err := manager.RegisterIntegration(mockIntegration)
	if err != nil {
		t.Fatalf("Expected no error registering integration, got %v", err)
	}

	// Try to register the same integration again
	err = manager.RegisterIntegration(mockIntegration)
	if err == nil {
		t.Error("Expected error when registering duplicate integration")
	}
}

func TestManagerGetIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockIntegration := &mockIntegration{
		name:    "test-integration",
		enabled: true,
	}

	manager.RegisterIntegration(mockIntegration)

	// Test getting existing integration
	retrieved, err := manager.GetIntegration("test-integration")
	if err != nil {
		t.Fatalf("Expected no error getting integration, got %v", err)
	}

	if retrieved != mockIntegration {
		t.Error("Expected to get the same integration instance")
	}

	// Test getting non-existent integration
	_, err = manager.GetIntegration("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent integration")
	}
}

func TestManagerGetEnabledIntegrations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	enabledIntegration := &mockIntegration{
		name:    "enabled-integration",
		enabled: true,
	}

	disabledIntegration := &mockIntegration{
		name:    "disabled-integration",
		enabled: false,
	}

	manager.RegisterIntegration(enabledIntegration)
	manager.RegisterIntegration(disabledIntegration)

	enabled := manager.GetEnabledIntegrations()
	if len(enabled) != 1 {
		t.Fatalf("Expected 1 enabled integration, got %d", len(enabled))
	}

	if enabled[0] != enabledIntegration {
		t.Error("Expected to get the enabled integration")
	}
}

func TestManagerSendToAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	// Initialize with disabled config (so SendToAll returns early)
	config := &types.IntegrationsConfig{
		Enabled: false,
	}
	ctx := context.Background()
	manager.Initialize(ctx, config)

	data := &types.AnalysisResult{
		TotalQueries:  100,
		UniqueClients: 10,
	}

	err := manager.SendToAll(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error when sending to disabled integrations, got %v", err)
	}

	// Test with enabled config but no integrations
	config.Enabled = true
	manager.Initialize(ctx, config)

	err = manager.SendToAll(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error when sending to no integrations, got %v", err)
	}
}

func TestManagerTestAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockIntegration := &mockIntegration{
		name:    "test-integration",
		enabled: true,
	}

	manager.RegisterIntegration(mockIntegration)

	// Initialize manager
	config := &types.IntegrationsConfig{
		Enabled: true,
	}
	ctx := context.Background()
	manager.Initialize(ctx, config)

	results := manager.TestAll(ctx)
	if len(results) != 1 {
		t.Fatalf("Expected 1 test result, got %d", len(results))
	}

	if results["test-integration"] != nil {
		t.Errorf("Expected no error for test integration, got %v", results["test-integration"])
	}
}

func TestManagerClose(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockIntegration := &mockIntegration{
		name:    "test-integration",
		enabled: true,
	}

	manager.RegisterIntegration(mockIntegration)

	// Initialize manager
	config := &types.IntegrationsConfig{
		Enabled: true,
	}
	ctx := context.Background()
	manager.Initialize(ctx, config)

	err := manager.Close()
	if err != nil {
		t.Fatalf("Expected no error closing manager, got %v", err)
	}

	if manager.IsInitialized() {
		t.Error("Expected manager to not be initialized after close")
	}

	if !mockIntegration.closed {
		t.Error("Expected mock integration to be closed")
	}
}

func TestManagerSendLogs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockLogIntegration := &mockLogIntegration{
		mockIntegration: mockIntegration{
			name:    "log-integration",
			enabled: true,
		},
		logs: make([]interfaces.LogEntry, 0),
	}

	manager.RegisterIntegration(mockLogIntegration)

	// Initialize manager
	config := &types.IntegrationsConfig{
		Enabled: true,
	}
	ctx := context.Background()
	manager.Initialize(ctx, config)

	logs := []interfaces.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     2, // INFO
			Message:   "Test log message",
			Component: "test",
		},
	}

	err := manager.SendLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Expected no error sending logs, got %v", err)
	}

	if len(mockLogIntegration.logs) != 1 {
		t.Fatalf("Expected 1 log to be sent, got %d", len(mockLogIntegration.logs))
	}

	if mockLogIntegration.logs[0].Message != "Test log message" {
		t.Errorf("Expected log message 'Test log message', got %q", mockLogIntegration.logs[0].Message)
	}
}

func TestManagerGetStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	manager := NewManager(logger)

	mockIntegration := &mockIntegration{
		name:    "test-integration",
		enabled: true,
	}

	manager.RegisterIntegration(mockIntegration)

	status := manager.GetStatus()
	if len(status) != 1 {
		t.Fatalf("Expected 1 status entry, got %d", len(status))
	}

	integrationStatus, exists := status["test-integration"]
	if !exists {
		t.Fatal("Expected status for test-integration")
	}

	if integrationStatus.Name != "test-integration" {
		t.Errorf("Expected status name 'test-integration', got %q", integrationStatus.Name)
	}

	if !integrationStatus.Enabled {
		t.Error("Expected status to show enabled")
	}
}

// Mock implementations for testing

type mockIntegration struct {
	name    string
	enabled bool
	closed  bool
}

func (m *mockIntegration) Initialize(ctx context.Context, config interface{}) error {
	return nil
}

func (m *mockIntegration) IsEnabled() bool {
	return m.enabled
}

func (m *mockIntegration) GetName() string {
	return m.name
}

func (m *mockIntegration) GetStatus() interfaces.IntegrationStatus {
	return interfaces.IntegrationStatus{
		Name:      m.name,
		Enabled:   m.enabled,
		Connected: true,
	}
}

func (m *mockIntegration) SendMetrics(ctx context.Context, data *types.AnalysisResult) error {
	return nil
}

func (m *mockIntegration) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	return nil
}

func (m *mockIntegration) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockIntegration) Close() error {
	m.closed = true
	return nil
}

type mockLogIntegration struct {
	mockIntegration
	logs []interfaces.LogEntry
}

func (m *mockLogIntegration) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	m.logs = append(m.logs, logs...)
	return nil
}

func (m *mockLogIntegration) StreamLogs(ctx context.Context, logStream <-chan interfaces.LogEntry) error {
	return nil
}

func (m *mockLogIntegration) WriteLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	return m.SendLogs(ctx, logs)
}

func (m *mockLogIntegration) SetLogLevel(level slog.Level) {
	// No-op for mock
}

func (m *mockLogIntegration) AddStaticLabels(labels map[string]string) {
	// No-op for mock
}

// Compile-time interface checks
var (
	_ interfaces.MonitoringIntegration = (*mockIntegration)(nil)
	_ interfaces.LogsIntegration      = (*mockLogIntegration)(nil)
)