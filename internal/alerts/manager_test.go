package alerts

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

// TestManagerInitialization tests alert manager initialization
func TestManagerInitialization(t *testing.T) {
	tests := []struct {
		name          string
		config        AlertConfig
		expectError   bool
		expectedRules int
	}{
		{
			name:          "default configuration",
			config:        DefaultAlertConfig(),
			expectError:   false,
			expectedRules: 5, // Default rules count
		},
		{
			name: "disabled configuration",
			config: AlertConfig{
				Enabled: false,
			},
			expectError:   false,
			expectedRules: 0,
		},
		{
			name: "custom rules",
			config: AlertConfig{
				Enabled: true,
				Rules: []AlertRule{
					{
						ID:       "test-rule",
						Name:     "Test Rule",
						Enabled:  true,
						Type:     AlertTypeThreshold,
						Severity: SeverityWarning,
						Conditions: []AlertCondition{
							{
								Field:    "test_field",
								Operator: "gt",
								Value:    100,
							},
						},
					},
				},
				Storage: StorageConfig{
					Type:      "memory",
					MaxSize:   100,
					Retention: 24 * time.Hour,
				},
			},
			expectError:   false,
			expectedRules: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logger.New(logger.DefaultConfig())
			manager := NewManager(tt.config, logger)

			ctx := context.Background()
			err := manager.Initialize(ctx, tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.config.Enabled {
				rules, err := manager.GetRules()
				if err != nil {
					t.Errorf("failed to get rules: %v", err)
				}
				if len(rules) != tt.expectedRules {
					t.Errorf("expected %d rules, got %d", tt.expectedRules, len(rules))
				}
			}

			manager.Close()
		})
	}
}

// TestAlertCreation tests alert creation and firing
func TestAlertCreation(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := AlertConfig{
		Enabled: true,
		Storage: StorageConfig{
			Type:      "memory",
			MaxSize:   100,
			Retention: 24 * time.Hour,
		},
		Notifications: NotificationConfig{
			Slack: SlackConfig{Enabled: false},
			Email: EmailConfig{Enabled: false},
		},
	}

	manager := NewManager(config, logger)
	ctx := context.Background()

	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Create test alert
	alert := &Alert{
		ID:          "test-alert-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test Alert",
		Description: "This is a test alert",
		Timestamp:   time.Now(),
		Source:      "test",
		Tags:        []string{"test"},
	}

	// Fire the alert
	if err := manager.FireAlert(ctx, alert); err != nil {
		t.Errorf("failed to fire alert: %v", err)
	}

	// Check active alerts
	activeAlerts, err := manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}

	if len(activeAlerts) != 1 {
		t.Errorf("expected 1 active alert, got %d", len(activeAlerts))
	}

	if activeAlerts[0].ID != alert.ID {
		t.Errorf("expected alert ID %s, got %s", alert.ID, activeAlerts[0].ID)
	}

	// Resolve the alert
	if err := manager.ResolveAlert(ctx, alert.ID); err != nil {
		t.Errorf("failed to resolve alert: %v", err)
	}

	// Check active alerts after resolution
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts after resolution: %v", err)
	}

	if len(activeAlerts) != 0 {
		t.Errorf("expected 0 active alerts after resolution, got %d", len(activeAlerts))
	}
}

// TestMLAnomalyProcessing tests processing of ML anomalies
func TestMLAnomalyProcessing(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := DefaultAlertConfig()
	manager := NewManager(config, logger)

	ctx := context.Background()
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Create mock analysis result
	analysisResult := &types.AnalysisResult{
		TotalQueries:  1000,
		UniqueClients: 10,
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	// Create mock ML results with anomalies
	mlResults := &ml.MLResults{
		Anomalies: []ml.Anomaly{
			{
				ID:          "anomaly-1",
				Type:        ml.AnomalyTypeVolumeSpike,
				Severity:    ml.SeverityHigh,
				Timestamp:   time.Now(),
				Description: "Unusual volume spike detected",
				Score:       0.9,
				Confidence:  0.85,
				ClientIP:    "192.168.1.100",
				Domain:      "suspicious.com",
			},
			{
				ID:          "anomaly-2",
				Type:        ml.AnomalyTypeUnusualDomain,
				Severity:    ml.SeverityMedium,
				Timestamp:   time.Now(),
				Description: "Unusual domain pattern detected",
				Score:       0.7,
				Confidence:  0.75,
<<<<<<< HEAD
				Domain:      "malware.example.com",
=======
				Domain:      "malware.localhost",
>>>>>>> main
			},
		},
		Summary: ml.MLSummary{
			TotalAnomalies:    2,
			HighSeverityCount: 1,
			HealthScore:       75.0,
		},
		ProcessedAt: time.Now(),
	}

	// Process the data
	if err := manager.ProcessData(ctx, analysisResult, mlResults); err != nil {
		t.Errorf("failed to process data: %v", err)
	}

	// Check that alerts were created
	activeAlerts, err := manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}

	// Should have at least 1 alert (high confidence anomaly)
	if len(activeAlerts) == 0 {
		t.Error("expected at least 1 alert from ML anomalies")
	}

	// Check alert properties
	found := false
	for _, alert := range activeAlerts {
		if alert.Type == AlertTypeAnomaly && alert.Source == "ml-engine" {
			found = true
			if alert.Anomaly == nil {
				t.Error("expected anomaly data in ML-generated alert")
			}
			if alert.MLScore != 0.9 && alert.MLScore != 0.7 {
				t.Errorf("unexpected ML score: %f", alert.MLScore)
			}
		}
	}

	if !found {
		t.Error("no ML-generated alert found")
	}
}

// TestRuleEvaluation tests alert rule evaluation
func TestRuleEvaluation(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	// Create custom config with specific rule
	config := AlertConfig{
		Enabled: true,
		Rules: []AlertRule{
			{
				ID:          "volume-test",
				Name:        "Volume Test Rule",
				Description: "Test volume threshold",
				Enabled:     true,
				Type:        AlertTypeThreshold,
				Severity:    SeverityWarning,
				Conditions: []AlertCondition{
					{
						Field:    "total_queries",
						Operator: "gt",
						Value:    500,
					},
				},
				CooldownPeriod: 5 * time.Minute,
				Channels:       []NotificationChannel{ChannelLog},
			},
		},
		Storage: StorageConfig{
			Type:      "memory",
			MaxSize:   100,
			Retention: 24 * time.Hour,
		},
		Notifications: NotificationConfig{
			Slack: SlackConfig{Enabled: false},
			Email: EmailConfig{Enabled: false},
		},
	}

	manager := NewManager(config, logger)
	ctx := context.Background()
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Test data that should trigger the rule
	analysisResult := &types.AnalysisResult{
		TotalQueries:  1000, // > 500, should trigger
		UniqueClients: 10,
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	// Process the data
	if err := manager.ProcessData(ctx, analysisResult, nil); err != nil {
		t.Errorf("failed to process data: %v", err)
	}

	// Check that alert was created
	activeAlerts, err := manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}

	if len(activeAlerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(activeAlerts))
	}

	alert := activeAlerts[0]
	if alert.Type != AlertTypeThreshold {
		t.Errorf("expected threshold alert, got %s", alert.Type)
	}

	// Test data that should NOT trigger the rule
	analysisResult2 := &types.AnalysisResult{
		TotalQueries:  100, // < 500, should not trigger
		UniqueClients: 5,
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	// Reset active alerts
	manager.ResolveAlert(ctx, alert.ID)

	// Process the data
	if err := manager.ProcessData(ctx, analysisResult2, nil); err != nil {
		t.Errorf("failed to process data: %v", err)
	}

	// Check that no new alert was created
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}

	if len(activeAlerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(activeAlerts))
	}
}

// TestManagerStatus tests manager status reporting
func TestManagerStatus(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := DefaultAlertConfig()
	manager := NewManager(config, logger)

	ctx := context.Background()
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Get initial status
	status := manager.GetStatus()
	if !status.IsRunning {
		t.Error("expected manager to be running")
	}
	if status.ActiveAlerts != 0 {
		t.Errorf("expected 0 active alerts, got %d", status.ActiveAlerts)
	}
	if status.RulesCount != 5 { // Default rules
		t.Errorf("expected 5 rules, got %d", status.RulesCount)
	}

	// Fire an alert
	alert := &Alert{
		ID:          "status-test-alert",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Status Test Alert",
		Description: "Test alert for status verification",
		Timestamp:   time.Now(),
		Source:      "test",
	}

	if err := manager.FireAlert(ctx, alert); err != nil {
		t.Errorf("failed to fire alert: %v", err)
	}

	// Check updated status
	status = manager.GetStatus()
	if status.ActiveAlerts != 1 {
		t.Errorf("expected 1 active alert, got %d", status.ActiveAlerts)
	}
	if status.TotalAlerts != 1 {
		t.Errorf("expected 1 total alert, got %d", status.TotalAlerts)
	}
}

// TestRuleManagement tests adding, updating, and removing rules
func TestRuleManagement(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := AlertConfig{
		Enabled: true,
		Rules:   []AlertRule{}, // Start with no rules
		Storage: StorageConfig{
			Type:      "memory",
			MaxSize:   100,
			Retention: 24 * time.Hour,
		},
	}

	manager := NewManager(config, logger)
	ctx := context.Background()
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Check initial rules count
	rules, err := manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 initial rules, got %d", len(rules))
	}

	// Add a rule
	newRule := AlertRule{
		ID:          "test-rule-1",
		Name:        "Test Rule 1",
		Description: "First test rule",
		Enabled:     true,
		Type:        AlertTypeThreshold,
		Severity:    SeverityWarning,
		Conditions: []AlertCondition{
			{
				Field:    "test_field",
				Operator: "gt",
				Value:    100,
			},
		},
		Tags: []string{"test"},
	}

	if err := manager.UpdateRule(newRule); err != nil {
		t.Errorf("failed to add rule: %v", err)
	}

	// Check rules count after adding
	rules, err = manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules after adding: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule after adding, got %d", len(rules))
	}

	// Update the rule
	newRule.Description = "Updated test rule"
	if err := manager.UpdateRule(newRule); err != nil {
		t.Errorf("failed to update rule: %v", err)
	}

	// Verify update
	rules, err = manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules after updating: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule after updating, got %d", len(rules))
	}
	if rules[0].Description != "Updated test rule" {
		t.Errorf("rule description not updated")
	}

	// Remove the rule
	if err := manager.RemoveRule(newRule.ID); err != nil {
		t.Errorf("failed to remove rule: %v", err)
	}

	// Check rules count after removing
	rules, err = manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules after removing: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after removing, got %d", len(rules))
	}

	// Try to remove non-existent rule
	if err := manager.RemoveRule("non-existent"); err == nil {
		t.Error("expected error when removing non-existent rule")
	}
}
