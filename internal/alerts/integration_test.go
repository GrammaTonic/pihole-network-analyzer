package alerts

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

// TestIntegrationAlertSystemWorkflow tests the complete alert system workflow
func TestIntegrationAlertSystemWorkflow(t *testing.T) {
	// Setup
	logger := logger.New(logger.DefaultConfig())
	config := AlertConfig{
		Enabled: true,
		Rules: []AlertRule{
			{
				ID:          "integration-volume-test",
				Name:        "Integration Volume Test",
				Description: "Test rule for integration testing",
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
				CooldownPeriod: 1 * time.Minute,
				Channels:       []NotificationChannel{ChannelLog},
				Tags:           []string{"integration", "test"},
			},
			{
				ID:          "integration-anomaly-test",
				Name:        "Integration ML Anomaly Test",
				Description: "Test rule for ML anomaly integration",
				Enabled:     true,
				Type:        AlertTypeAnomaly,
				Severity:    SeverityCritical,
				Conditions: []AlertCondition{
					{
						Field:    "anomaly_severity",
						Operator: "eq",
						Value:    "critical",
					},
				},
				CooldownPeriod: 30 * time.Second,
				Channels:       []NotificationChannel{ChannelLog},
				Tags:           []string{"ml", "critical"},
			},
		},
		Storage: StorageConfig{
			Type:      "memory",
			MaxSize:   1000,
			Retention: 24 * time.Hour,
		},
		Notifications: NotificationConfig{
			Slack: SlackConfig{Enabled: false},
			Email: EmailConfig{Enabled: false},
		},
		DefaultSeverity: SeverityWarning,
		DefaultCooldown: 5 * time.Minute,
		MaxActiveAlerts: 100,
		AlertRetention:  24 * time.Hour,
	}

	manager := NewManager(config, logger)
	ctx := context.Background()

	// Test 1: Initialize the alert manager
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize alert manager: %v", err)
	}
	defer manager.Close()

	// Verify initial state
	status := manager.GetStatus()
	if !status.IsRunning {
		t.Error("alert manager should be running after initialization")
	}
	if status.ActiveAlerts != 0 {
		t.Errorf("expected 0 active alerts initially, got %d", status.ActiveAlerts)
	}
	if status.RulesCount != 2 {
		t.Errorf("expected 2 rules, got %d", status.RulesCount)
	}

	// Test 2: Process normal data (should not trigger alerts)
	normalAnalysisResult := &types.AnalysisResult{
		TotalQueries:  100, // Below threshold
		UniqueClients: 5,
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	if err := manager.ProcessData(ctx, normalAnalysisResult, nil); err != nil {
		t.Errorf("failed to process normal data: %v", err)
	}

	// Should have no alerts
	activeAlerts, err := manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}
	if len(activeAlerts) != 0 {
		t.Errorf("expected 0 alerts for normal data, got %d", len(activeAlerts))
	}

	// Test 3: Process data that triggers threshold rule
	highVolumeAnalysisResult := &types.AnalysisResult{
		TotalQueries:  1000, // Above threshold (500)
		UniqueClients: 10,
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	if err := manager.ProcessData(ctx, highVolumeAnalysisResult, nil); err != nil {
		t.Errorf("failed to process high volume data: %v", err)
	}

	// Should have 1 threshold alert
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts after threshold: %v", err)
	}
	if len(activeAlerts) != 1 {
		t.Errorf("expected 1 alert for high volume, got %d", len(activeAlerts))
	}

	thresholdAlert := activeAlerts[0]
	if thresholdAlert.Type != AlertTypeThreshold {
		t.Errorf("expected threshold alert, got %s", thresholdAlert.Type)
	}
	if thresholdAlert.Severity != SeverityWarning {
		t.Errorf("expected warning severity, got %s", thresholdAlert.Severity)
	}

	// Test 4: Process ML data with critical anomaly
	mlResults := &ml.MLResults{
		Anomalies: []ml.Anomaly{
			{
				ID:          "integration-anomaly",
				Type:        ml.AnomalyTypeVolumeSpike,
				Severity:    ml.SeverityCritical, // Critical anomaly
				Timestamp:   time.Now(),
				Description: "Critical volume spike detected in integration test",
				Score:       0.95,
				Confidence:  0.9,
				ClientIP:    "192.168.1.100",
				Domain:      "suspicious.com",
			},
		},
		Summary: ml.MLSummary{
			TotalAnomalies:    1,
			HighSeverityCount: 1,
			HealthScore:       60.0,
		},
		ProcessedAt: time.Now(),
	}

	if err := manager.ProcessData(ctx, highVolumeAnalysisResult, mlResults); err != nil {
		t.Errorf("failed to process ML data: %v", err)
	}

	// Should have 2 alerts now (threshold + ML anomaly)
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts after ML processing: %v", err)
	}
	if len(activeAlerts) < 2 {
		t.Errorf("expected at least 2 alerts after ML processing, got %d", len(activeAlerts))
	}

	// Find the ML-generated alert
	var mlAlert *Alert
	for _, alert := range activeAlerts {
		if alert.Source == "ml-engine" {
			mlAlert = alert
			break
		}
	}

	if mlAlert == nil {
		t.Error("expected to find ML-generated alert")
	} else {
		if mlAlert.Type != AlertTypeAnomaly {
			t.Errorf("expected anomaly alert from ML, got %s", mlAlert.Type)
		}
		if mlAlert.Severity != SeverityCritical {
			t.Errorf("expected critical severity from ML, got %s", mlAlert.Severity)
		}
		if mlAlert.Anomaly == nil {
			t.Error("expected anomaly data in ML alert")
		}
		if mlAlert.MLScore != 0.95 {
			t.Errorf("expected ML score 0.95, got %f", mlAlert.MLScore)
		}
	}

	// Test 5: Resolve alerts
	for _, alert := range activeAlerts {
		if err := manager.ResolveAlert(ctx, alert.ID); err != nil {
			t.Errorf("failed to resolve alert %s: %v", alert.ID, err)
		}
	}

	// Should have no active alerts after resolution
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts after resolution: %v", err)
	}
	if len(activeAlerts) != 0 {
		t.Errorf("expected 0 active alerts after resolution, got %d", len(activeAlerts))
	}

	// Test 6: Check alert history
	history, err := manager.GetAlertHistory(10)
	if err != nil {
		t.Errorf("failed to get alert history: %v", err)
	}
	if len(history) < 2 {
		t.Errorf("expected at least 2 alerts in history, got %d", len(history))
	}

	// Test 7: Rule management
	newRule := AlertRule{
		ID:          "dynamic-rule",
		Name:        "Dynamic Test Rule",
		Description: "Dynamically added rule",
		Enabled:     true,
		Type:        AlertTypeThreshold,
		Severity:    SeverityInfo,
		Conditions: []AlertCondition{
			{
				Field:    "unique_clients",
				Operator: "lt",
				Value:    2,
			},
		},
		CooldownPeriod: 1 * time.Minute,
		Channels:       []NotificationChannel{ChannelLog},
		Tags:           []string{"dynamic"},
	}

	if err := manager.UpdateRule(newRule); err != nil {
		t.Errorf("failed to add dynamic rule: %v", err)
	}

	// Verify rule was added
	rules, err := manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules: %v", err)
	}
	if len(rules) != 3 {
		t.Errorf("expected 3 rules after adding dynamic rule, got %d", len(rules))
	}

	// Test the new rule
	lowClientAnalysisResult := &types.AnalysisResult{
		TotalQueries:  50,
		UniqueClients: 1, // Below threshold (2)
		AnalysisMode:  "test",
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	if err := manager.ProcessData(ctx, lowClientAnalysisResult, nil); err != nil {
		t.Errorf("failed to process low client data: %v", err)
	}

	// Should trigger the dynamic rule
	activeAlerts, err = manager.GetActiveAlerts()
	if err != nil {
		t.Errorf("failed to get active alerts for dynamic rule: %v", err)
	}
	if len(activeAlerts) != 1 {
		t.Errorf("expected 1 alert from dynamic rule, got %d", len(activeAlerts))
	}

	// Remove the dynamic rule
	if err := manager.RemoveRule(newRule.ID); err != nil {
		t.Errorf("failed to remove dynamic rule: %v", err)
	}

	// Verify rule was removed
	rules, err = manager.GetRules()
	if err != nil {
		t.Errorf("failed to get rules after removal: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules after removing dynamic rule, got %d", len(rules))
	}

	// Test 8: Final status check
	finalStatus := manager.GetStatus()
	if !finalStatus.IsRunning {
		t.Error("alert manager should still be running")
	}
	if finalStatus.TotalAlerts < 3 {
		t.Errorf("expected at least 3 total alerts processed, got %d", finalStatus.TotalAlerts)
	}
}

// TestIntegrationConfigurationScenarios tests various configuration scenarios
func TestIntegrationConfigurationScenarios(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	scenarios := []struct {
		name        string
		config      AlertConfig
		expectError bool
		description string
	}{
		{
			name:        "minimal configuration",
			config:      AlertConfig{Enabled: true, Storage: StorageConfig{Type: "memory"}},
			expectError: false,
			description: "Should work with minimal configuration",
		},
		{
			name:        "disabled configuration",
			config:      AlertConfig{Enabled: false},
			expectError: false,
			description: "Should work when disabled",
		},
		{
			name: "file storage configuration",
			config: AlertConfig{
				Enabled: true,
				Storage: StorageConfig{
					Type:      "file",
					Path:      "/tmp/test_alerts.json",
					MaxSize:   100,
					Retention: 24 * time.Hour,
				},
			},
			expectError: false,
			description: "Should work with file storage",
		},
		{
			name: "Slack enabled configuration",
			config: AlertConfig{
				Enabled: true,
				Storage: StorageConfig{Type: "memory"},
				Notifications: NotificationConfig{
					Slack: SlackConfig{
						Enabled:    true,
						WebhookURL: "https://hooks.slack.com/test",
						Channel:    "#alerts",
						Username:   "AlertBot",
						IconEmoji:  ":warning:",
						Timeout:    30 * time.Second,
					},
				},
			},
			expectError: false,
			description: "Should work with Slack notifications enabled",
		},
		{
			name: "Email enabled configuration",
			config: AlertConfig{
				Enabled: true,
				Storage: StorageConfig{Type: "memory"},
				Notifications: NotificationConfig{
					Email: EmailConfig{
						Enabled:    true,
						SMTPHost:   "smtp.test.com",
						SMTPPort:   587,
						Username:   "test@test.com",
						Password:   "password",
						From:       "alerts@test.com",
						Recipients: []string{"admin@test.com"},
						UseTLS:     true,
						Timeout:    30 * time.Second,
					},
				},
			},
			expectError: false,
			description: "Should work with email notifications enabled",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			manager := NewManager(scenario.config, logger)
			ctx := context.Background()

			err := manager.Initialize(ctx, scenario.config)

			if scenario.expectError && err == nil {
				t.Errorf("expected error for scenario '%s' but got none", scenario.description)
			}
			if !scenario.expectError && err != nil {
				t.Errorf("unexpected error for scenario '%s': %v", scenario.description, err)
			}

			// Test basic functionality if no error
			if err == nil && scenario.config.Enabled {
				// Try to process some data
				analysisResult := &types.AnalysisResult{
					TotalQueries:  100,
					UniqueClients: 5,
					AnalysisMode:  "test",
					Timestamp:     time.Now().Format(time.RFC3339),
				}

				if err := manager.ProcessData(ctx, analysisResult, nil); err != nil {
					t.Errorf("failed to process data in scenario '%s': %v", scenario.description, err)
				}

				status := manager.GetStatus()
				if scenario.config.Enabled && !status.IsRunning {
					t.Errorf("manager should be running in scenario '%s'", scenario.description)
				}
			}

			manager.Close()
		})
	}
}

// TestIntegrationPerformance tests the performance of the alert system
func TestIntegrationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	logger := logger.New(logger.DefaultConfig())
	config := DefaultAlertConfig()
	manager := NewManager(config, logger)

	ctx := context.Background()
	if err := manager.Initialize(ctx, config); err != nil {
		t.Fatalf("failed to initialize manager: %v", err)
	}
	defer manager.Close()

	// Test processing many alerts
	startTime := time.Now()
	numIterations := 100

	for i := 0; i < numIterations; i++ {
		analysisResult := &types.AnalysisResult{
			TotalQueries:  1000 + i, // Will trigger volume spike rule
			UniqueClients: 10 + i,
			AnalysisMode:  "performance-test",
			Timestamp:     time.Now().Format(time.RFC3339),
		}

		// Create some ML anomalies
		mlResults := &ml.MLResults{
			Anomalies: []ml.Anomaly{
				{
					ID:          "perf-anomaly-" + string(rune(i)),
					Type:        ml.AnomalyTypeVolumeSpike,
					Severity:    ml.SeverityMedium,
					Timestamp:   time.Now(),
					Description: "Performance test anomaly",
					Score:       0.75,
					Confidence:  0.8,
				},
			},
			Summary: ml.MLSummary{
				TotalAnomalies:    1,
				HighSeverityCount: 0,
				HealthScore:       80.0,
			},
			ProcessedAt: time.Now(),
		}

		if err := manager.ProcessData(ctx, analysisResult, mlResults); err != nil {
			t.Errorf("failed to process data in iteration %d: %v", i, err)
		}

		// Resolve alerts periodically to prevent memory buildup
		if i%10 == 0 {
			activeAlerts, err := manager.GetActiveAlerts()
			if err != nil {
				t.Errorf("failed to get active alerts in iteration %d: %v", i, err)
			}
			for _, alert := range activeAlerts {
				manager.ResolveAlert(ctx, alert.ID)
			}
		}
	}

	processingTime := time.Since(startTime)
	t.Logf("Processed %d iterations in %v (%.2f ms/iteration)",
		numIterations, processingTime, float64(processingTime.Nanoseconds())/float64(numIterations)/1e6)

	// Performance assertions
	maxTimePerIteration := 100 * time.Millisecond
	avgTimePerIteration := processingTime / time.Duration(numIterations)
	if avgTimePerIteration > maxTimePerIteration {
		t.Errorf("average processing time %v exceeds maximum %v", avgTimePerIteration, maxTimePerIteration)
	}

	// Check final status
	status := manager.GetStatus()
	if status.ProcessingErrors > 0 {
		t.Errorf("performance test generated %d processing errors", status.ProcessingErrors)
	}

	t.Logf("Final status: %d total alerts, %d active alerts, %d rules",
		status.TotalAlerts, status.ActiveAlerts, status.RulesCount)
}
