package grafana

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
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if !client.IsEnabled() {
		t.Error("Expected client to be enabled")
	}

	if client.GetName() != "grafana" {
		t.Errorf("Expected name 'grafana', got %q", client.GetName())
	}
}

func TestNewClientDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: false,
	}

	client := NewClient(config, logger)

	if client.IsEnabled() {
		t.Error("Expected client to be disabled")
	}
}

func TestClientInitialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: false, // Disabled to avoid connection attempts
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
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
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)
	status := client.GetStatus()

	if status.Name != "grafana" {
		t.Errorf("Expected status name 'grafana', got %q", status.Name)
	}

	if !status.Enabled {
		t.Error("Expected status to show enabled")
	}
}

func TestClientSendLogs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	logs := []interfaces.LogEntry{
		{
			Message: "Test log message",
		},
	}

	// SendLogs should be a no-op for Grafana (it's for dashboards, not logs)
	err := client.SendLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Expected no error from SendLogs (should be no-op), got %v", err)
	}
}

func TestClientSendMetricsDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
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

func TestClientSendMetricsWithoutAutoProvision(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
		Dashboards: types.DashboardConfig{
			AutoProvision: false, // Disabled to avoid API calls
		},
	}

	client := NewClient(config, logger)
	ctx := context.Background()

	data := &types.AnalysisResult{
		TotalQueries:  100,
		UniqueClients: 10,
	}

	err := client.SendMetrics(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error sending metrics without auto-provision, got %v", err)
	}
}

func TestClientClose(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: false, // Disabled to avoid connection attempts
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
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

func TestCreateMainDashboard(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
		Dashboards: types.DashboardConfig{
			Tags: []string{"pihole", "network", "dns"},
		},
	}

	client := NewClient(config, logger)

	data := &types.AnalysisResult{
		TotalQueries:  100,
		UniqueClients: 10,
	}

	dashboard := client.createMainDashboard(data)

	if dashboard.Title != "Pi-hole Network Analyzer" {
		t.Errorf("Expected title 'Pi-hole Network Analyzer', got %q", dashboard.Title)
	}

	if dashboard.Description != "Comprehensive network analysis and DNS monitoring dashboard" {
		t.Errorf("Unexpected description: %q", dashboard.Description)
	}

	if len(dashboard.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(dashboard.Tags))
	}

	expectedTags := []string{"pihole", "network", "dns"}
	for i, tag := range dashboard.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %q, got %q", expectedTags[i], tag)
		}
	}

	if len(dashboard.Panels) != 6 {
		t.Errorf("Expected 6 panels, got %d", len(dashboard.Panels))
	}

	// Check specific panels
	expectedPanels := []struct {
		id    int
		title string
		pType string
	}{
		{1, "Total DNS Queries", "stat"},
		{2, "Active Clients", "stat"},
		{3, "Queries by Type", "piechart"},
		{4, "Query Response Time", "timeseries"},
		{5, "Top Domains", "table"},
		{6, "Blocked vs Allowed", "timeseries"},
	}

	for i, expected := range expectedPanels {
		if i >= len(dashboard.Panels) {
			t.Fatalf("Panel %d missing", i)
		}

		panel := dashboard.Panels[i]
		if panel.ID != expected.id {
			t.Errorf("Panel %d: expected ID %d, got %d", i, expected.id, panel.ID)
		}

		if panel.Title != expected.title {
			t.Errorf("Panel %d: expected title %q, got %q", i, expected.title, panel.Title)
		}

		if panel.Type != expected.pType {
			t.Errorf("Panel %d: expected type %q, got %q", i, expected.pType, panel.Type)
		}
	}
}

func TestConvertToGrafanaDashboard(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)

	dashboard := interfaces.Dashboard{
		Title:       "Test Dashboard",
		Description: "A test dashboard",
		Tags:        []string{"test", "monitoring"},
		Panels: []interfaces.Panel{
			{
				ID:    1,
				Title: "Test Panel",
				Type:  "graph",
				Query: "up",
			},
		},
	}

	grafanaDashboard := client.convertToGrafanaDashboard(dashboard)

	if grafanaDashboard["title"] != "Test Dashboard" {
		t.Errorf("Expected title 'Test Dashboard', got %v", grafanaDashboard["title"])
	}

	if grafanaDashboard["description"] != "A test dashboard" {
		t.Errorf("Expected description 'A test dashboard', got %v", grafanaDashboard["description"])
	}

	tags, ok := grafanaDashboard["tags"].([]string)
	if !ok {
		t.Fatal("Expected tags to be []string")
	}

	if len(tags) != 2 || tags[0] != "test" || tags[1] != "monitoring" {
		t.Errorf("Expected tags [test, monitoring], got %v", tags)
	}

	panels, ok := grafanaDashboard["panels"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected panels to be []map[string]interface{}")
	}

	if len(panels) != 1 {
		t.Errorf("Expected 1 panel, got %d", len(panels))
	}

	panel := panels[0]
	if panel["id"] != 1 {
		t.Errorf("Expected panel ID 1, got %v", panel["id"])
	}

	if panel["title"] != "Test Panel" {
		t.Errorf("Expected panel title 'Test Panel', got %v", panel["title"])
	}

	if panel["type"] != "graph" {
		t.Errorf("Expected panel type 'graph', got %v", panel["type"])
	}
}

func TestConvertPanels(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)

	panels := []interfaces.Panel{
		{
			ID:    1,
			Title: "Panel 1",
			Type:  "graph",
			Query: "metric1",
		},
		{
			ID:    2,
			Title: "Panel 2",
			Type:  "stat",
			Query: "metric2",
		},
	}

	grafanaPanels := client.convertPanels(panels)

	if len(grafanaPanels) != 2 {
		t.Errorf("Expected 2 panels, got %d", len(grafanaPanels))
	}

	// Check first panel
	panel1 := grafanaPanels[0]
	if panel1["id"] != 1 {
		t.Errorf("Expected panel 1 ID 1, got %v", panel1["id"])
	}

	if panel1["title"] != "Panel 1" {
		t.Errorf("Expected panel 1 title 'Panel 1', got %v", panel1["title"])
	}

	targets1, ok := panel1["targets"].([]map[string]interface{})
	if !ok || len(targets1) != 1 {
		t.Fatal("Expected panel 1 to have 1 target")
	}

	if targets1[0]["expr"] != "metric1" {
		t.Errorf("Expected panel 1 expr 'metric1', got %v", targets1[0]["expr"])
	}

	// Check grid position for first panel (should be at x=0)
	gridPos1, ok := panel1["gridPos"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected panel 1 to have gridPos")
	}

	if gridPos1["x"] != 0 {
		t.Errorf("Expected panel 1 x position 0, got %v", gridPos1["x"])
	}

	// Check second panel
	panel2 := grafanaPanels[1]
	if panel2["id"] != 2 {
		t.Errorf("Expected panel 2 ID 2, got %v", panel2["id"])
	}

	// Check grid position for second panel (should be at x=12)
	gridPos2, ok := panel2["gridPos"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected panel 2 to have gridPos")
	}

	if gridPos2["x"] != 12 {
		t.Errorf("Expected panel 2 x position 12, got %v", gridPos2["x"])
	}
}

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			m:        map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
		{
			m:        map[string]interface{}{},
			key:      "missing",
			expected: "",
		},
	}

	for _, test := range tests {
		result := getStringFromMap(test.m, test.key)
		if result != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, result)
		}
	}
}

func TestGetStringSliceFromMap(t *testing.T) {
	tests := []struct {
		m        map[string]interface{}
		key      string
		expected []string
	}{
		{
			m:        map[string]interface{}{"key": []interface{}{"a", "b", "c"}},
			key:      "key",
			expected: []string{"a", "b", "c"},
		},
		{
			m:        map[string]interface{}{"key": []interface{}{"a", 123, "c"}},
			key:      "key",
			expected: []string{"a", "", "c"},
		},
		{
			m:        map[string]interface{}{"key": "not a slice"},
			key:      "key",
			expected: []string{},
		},
		{
			m:        map[string]interface{}{},
			key:      "missing",
			expected: []string{},
		},
	}

	for _, test := range tests {
		result := getStringSliceFromMap(test.m, test.key)
		if len(result) != len(test.expected) {
			t.Errorf("Expected slice length %d, got %d", len(test.expected), len(result))
			continue
		}

		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("Expected element %d to be %q, got %q", i, expected, result[i])
			}
		}
	}
}

// Test that interface is implemented correctly
func TestGrafanaClientImplementsInterfaces(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &types.GrafanaConfig{
		Enabled: true,
		URL:     "http://localhost:3000",
		APIKey:  "test-api-key",
	}

	client := NewClient(config, logger)

	// Test MonitoringIntegration interface
	var _ interfaces.MonitoringIntegration = client

	// Test DashboardIntegration interface
	var _ interfaces.DashboardIntegration = client
}
