package alerts

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

// Manager implements the AlertManager interface
type Manager struct {
	config           AlertConfig
	logger           *logger.Logger
	storage          Storage
	evaluator        Evaluator
	notificationsMux sync.RWMutex
	notifications    map[NotificationChannel]NotificationHandler
	rules            map[string]AlertRule
	rulesMux         sync.RWMutex
	activeAlerts     map[string]*Alert
	activeAlertsMux  sync.RWMutex
	lastEvaluation   time.Time
	processingErrors int
	errorsMux        sync.RWMutex
	isRunning        bool
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewManager creates a new alert manager
func NewManager(config AlertConfig, logger *logger.Logger) *Manager {
	return &Manager{
		config:        config,
		logger:        logger.Component("alert-manager"),
		notifications: make(map[NotificationChannel]NotificationHandler),
		rules:         make(map[string]AlertRule),
		activeAlerts:  make(map[string]*Alert),
	}
}

// Initialize initializes the alert manager
func (m *Manager) Initialize(ctx context.Context, config AlertConfig) error {
	m.config = config
	m.ctx, m.cancel = context.WithCancel(ctx)

	m.logger.GetSlogger().Info("Initializing alert manager", slog.Bool("enabled", config.Enabled))

	if !config.Enabled {
		m.logger.GetSlogger().Info("Alert system is disabled")
		return nil
	}

	// Initialize storage
	var err error
	m.storage, err = m.createStorage(config.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize evaluator
	m.evaluator = NewEvaluator(m.logger)

	// Initialize notification handlers
	if err := m.initializeNotifications(); err != nil {
		return fmt.Errorf("failed to initialize notifications: %w", err)
	}

	// Load rules
	if err := m.loadRules(); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	m.isRunning = true
	m.logger.GetSlogger().Info("Alert manager initialized successfully",
		slog.Int("rules_count", len(m.rules)),
		slog.Int("notification_channels", len(m.notifications)))

	return nil
}

// ProcessData processes analysis data and evaluates alert rules
func (m *Manager) ProcessData(ctx context.Context, data *types.AnalysisResult, mlResults *ml.MLResults) error {
	if !m.config.Enabled || !m.isRunning {
		return nil
	}

	m.logger.GetSlogger().Debug("Processing data for alert evaluation",
		slog.Int("clients", len(data.ClientStats)),
		slog.Int("total_queries", data.TotalQueries))

	m.lastEvaluation = time.Now()

	// Process ML anomalies first
	if mlResults != nil && len(mlResults.Anomalies) > 0 {
		if err := m.processMLAnomalies(ctx, mlResults.Anomalies); err != nil {
			m.incrementErrors()
			m.logger.GetSlogger().Error("Failed to process ML anomalies", slog.String("error", err.Error()))
		}
	}

	// Evaluate all rules against the data
	evaluationData := m.prepareEvaluationData(data, mlResults)

	m.rulesMux.RLock()
	rules := make([]AlertRule, 0, len(m.rules))
	for _, rule := range m.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	m.rulesMux.RUnlock()

	for _, rule := range rules {
		if err := m.evaluateRule(ctx, rule, evaluationData); err != nil {
			m.incrementErrors()
			m.logger.GetSlogger().Error("Failed to evaluate rule",
				slog.String("rule_id", rule.ID),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// processMLAnomalies processes ML detected anomalies and creates alerts
func (m *Manager) processMLAnomalies(ctx context.Context, anomalies []ml.Anomaly) error {
	for _, anomaly := range anomalies {
		// Check if we should create an alert for this anomaly
		if anomaly.Confidence < 0.7 { // Configurable threshold
			continue
		}

		alert := &Alert{
			ID:          generateAlertID(),
			Type:        AlertTypeAnomaly,
			Severity:    m.mapMLSeverityToAlertSeverity(anomaly.Severity),
			Status:      AlertStatusFired,
			Title:       fmt.Sprintf("ML Anomaly Detected: %s", anomaly.Type),
			Description: anomaly.Description,
			Timestamp:   anomaly.Timestamp,
			Source:      "ml-engine",
			ClientIP:    anomaly.ClientIP,
			Domain:      anomaly.Domain,
			Metadata: map[string]interface{}{
				"anomaly_id":   anomaly.ID,
				"anomaly_type": string(anomaly.Type),
				"ml_metadata":  anomaly.Metadata,
			},
			Tags:       []string{"ml", "anomaly", string(anomaly.Type)},
			Anomaly:    &anomaly,
			MLScore:    anomaly.Score,
			Confidence: anomaly.Confidence,
		}

		if err := m.FireAlert(ctx, alert); err != nil {
			m.logger.GetSlogger().Error("Failed to fire ML anomaly alert",
				slog.String("anomaly_id", anomaly.ID),
				slog.String("error", err.Error()))
			return err
		}
	}

	return nil
}

// mapMLSeverityToAlertSeverity maps ML severity to alert severity
func (m *Manager) mapMLSeverityToAlertSeverity(mlSeverity ml.SeverityLevel) AlertSeverity {
	switch mlSeverity {
	case ml.SeverityLow:
		return SeverityInfo
	case ml.SeverityMedium:
		return SeverityWarning
	case ml.SeverityHigh:
		return SeverityError
	case ml.SeverityCritical:
		return SeverityCritical
	default:
		return SeverityWarning
	}
}

// prepareEvaluationData prepares data for rule evaluation
func (m *Manager) prepareEvaluationData(data *types.AnalysisResult, mlResults *ml.MLResults) map[string]interface{} {
	evaluationData := map[string]interface{}{
		"total_queries":    data.TotalQueries,
		"unique_clients":   data.UniqueClients,
		"analysis_mode":    data.AnalysisMode,
		"data_source_type": data.DataSourceType,
		"timestamp":        data.Timestamp,
	}

	// Add performance data if available
	if data.Performance != nil {
		evaluationData["average_response_time"] = data.Performance.AverageResponseTime
		evaluationData["queries_per_second"] = data.Performance.QueriesPerSecond
		evaluationData["slow_queries"] = data.Performance.SlowQueries
	}

	// Add ML results if available
	if mlResults != nil {
		evaluationData["anomaly_count"] = len(mlResults.Anomalies)
		evaluationData["health_score"] = mlResults.Summary.HealthScore
		evaluationData["high_severity_anomalies"] = mlResults.Summary.HighSeverityCount

		// Add individual anomaly data for specific rule evaluation
		for _, anomaly := range mlResults.Anomalies {
			key := fmt.Sprintf("anomaly_%s", anomaly.Type)
			if existing, ok := evaluationData[key]; ok {
				if count, ok := existing.(int); ok {
					evaluationData[key] = count + 1
				}
			} else {
				evaluationData[key] = 1
			}
		}
	}

	return evaluationData
}

// evaluateRule evaluates a single rule against the data
func (m *Manager) evaluateRule(ctx context.Context, rule AlertRule, data map[string]interface{}) error {
	// Check cooldown period
	if m.isRuleInCooldown(rule.ID) {
		return nil
	}

	// Evaluate conditions
	triggered, err := m.evaluator.EvaluateConditions(rule.Conditions, data)
	if err != nil {
		return fmt.Errorf("failed to evaluate conditions for rule %s: %w", rule.ID, err)
	}

	if triggered {
		alert := &Alert{
			ID:          generateAlertID(),
			Type:        rule.Type,
			Severity:    rule.Severity,
			Status:      AlertStatusFired,
			Title:       rule.Name,
			Description: rule.Description,
			Timestamp:   time.Now(),
			Source:      fmt.Sprintf("rule:%s", rule.ID),
			Metadata: map[string]interface{}{
				"rule_id":              rule.ID,
				"rule_name":            rule.Name,
				"triggered_conditions": m.getTriggeredConditions(rule.Conditions, data),
			},
			Tags: append(rule.Tags, "rule-triggered"),
		}

		// Set suppression if configured
		if rule.CooldownPeriod > 0 {
			suppressUntil := time.Now().Add(rule.CooldownPeriod)
			alert.SuppressUntil = &suppressUntil
		}

		return m.FireAlert(ctx, alert)
	}

	return nil
}

// FireAlert fires an alert and sends notifications
func (m *Manager) FireAlert(ctx context.Context, alert *Alert) error {
	m.logger.GetSlogger().Info("Firing alert",
		slog.String("alert_id", alert.ID),
		slog.String("type", string(alert.Type)),
		slog.String("severity", string(alert.Severity)),
		slog.String("title", alert.Title))

	// Store the alert
	if m.storage != nil {
		if err := m.storage.Store(alert); err != nil {
			m.logger.GetSlogger().Error("Failed to store alert", slog.String("error", err.Error()))
			// Continue with notification even if storage fails
		}
	}

	// Add to active alerts
	m.activeAlertsMux.Lock()
	m.activeAlerts[alert.ID] = alert
	m.activeAlertsMux.Unlock()

	// Send notifications
	if err := m.sendNotifications(ctx, alert); err != nil {
		m.logger.GetSlogger().Error("Failed to send notifications",
			slog.String("alert_id", alert.ID),
			slog.String("error", err.Error()))
		return err
	}

	m.logger.GetSlogger().Info("Alert fired successfully", slog.String("alert_id", alert.ID))
	return nil
}

// ResolveAlert resolves an active alert
func (m *Manager) ResolveAlert(ctx context.Context, alertID string) error {
	m.activeAlertsMux.Lock()
	alert, exists := m.activeAlerts[alertID]
	if !exists {
		m.activeAlertsMux.Unlock()
		return fmt.Errorf("alert %s not found in active alerts", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.ResolvedAt = &now

	delete(m.activeAlerts, alertID)
	m.activeAlertsMux.Unlock()

	// Update in storage
	if m.storage != nil {
		if err := m.storage.Update(alert); err != nil {
			m.logger.GetSlogger().Error("Failed to update resolved alert in storage", slog.String("error", err.Error()))
		}
	}

	m.logger.GetSlogger().Info("Alert resolved", slog.String("alert_id", alertID))
	return nil
}

// GetActiveAlerts returns all active alerts
func (m *Manager) GetActiveAlerts() ([]*Alert, error) {
	m.activeAlertsMux.RLock()
	defer m.activeAlertsMux.RUnlock()

	alerts := make([]*Alert, 0, len(m.activeAlerts))
	for _, alert := range m.activeAlerts {
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetAlertHistory returns alert history
func (m *Manager) GetAlertHistory(limit int) ([]*Alert, error) {
	if m.storage == nil {
		return []*Alert{}, nil
	}

	filters := map[string]interface{}{
		"limit": limit,
	}

	return m.storage.List(filters)
}

// UpdateRule adds or updates an alert rule
func (m *Manager) UpdateRule(rule AlertRule) error {
	m.rulesMux.Lock()
	defer m.rulesMux.Unlock()

	m.rules[rule.ID] = rule
	m.logger.GetSlogger().Info("Alert rule updated", slog.String("rule_id", rule.ID), slog.String("name", rule.Name))
	return nil
}

// RemoveRule removes an alert rule
func (m *Manager) RemoveRule(ruleID string) error {
	m.rulesMux.Lock()
	defer m.rulesMux.Unlock()

	if _, exists := m.rules[ruleID]; !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(m.rules, ruleID)
	m.logger.GetSlogger().Info("Alert rule removed", slog.String("rule_id", ruleID))
	return nil
}

// GetRules returns all alert rules
func (m *Manager) GetRules() ([]AlertRule, error) {
	m.rulesMux.RLock()
	defer m.rulesMux.RUnlock()

	rules := make([]AlertRule, 0, len(m.rules))
	for _, rule := range m.rules {
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetStatus returns the manager status
func (m *Manager) GetStatus() ManagerStatus {
	m.activeAlertsMux.RLock()
	activeCount := len(m.activeAlerts)
	m.activeAlertsMux.RUnlock()

	m.rulesMux.RLock()
	rulesCount := len(m.rules)
	m.rulesMux.RUnlock()

	m.errorsMux.RLock()
	errors := m.processingErrors
	m.errorsMux.RUnlock()

	var totalAlerts int
	if m.storage != nil {
		if alerts, err := m.storage.List(map[string]interface{}{}); err == nil {
			totalAlerts = len(alerts)
		}
	}

	status := "healthy"
	if errors > 0 {
		status = "errors"
	}
	if !m.isRunning {
		status = "stopped"
	}

	return ManagerStatus{
		IsRunning:        m.isRunning,
		ActiveAlerts:     activeCount,
		TotalAlerts:      totalAlerts,
		RulesCount:       rulesCount,
		LastEvaluation:   m.lastEvaluation.Format(time.RFC3339),
		ProcessingErrors: errors,
		Status:           status,
	}
}

// Close shuts down the alert manager
func (m *Manager) Close() error {
	m.logger.GetSlogger().Info("Shutting down alert manager")

	m.isRunning = false

	if m.cancel != nil {
		m.cancel()
	}

	if m.storage != nil {
		if err := m.storage.Close(); err != nil {
			m.logger.GetSlogger().Error("Failed to close storage", slog.String("error", err.Error()))
		}
	}

	m.logger.GetSlogger().Info("Alert manager shut down successfully")
	return nil
}

// Helper methods

func (m *Manager) createStorage(config StorageConfig) (Storage, error) {
	switch config.Type {
	case "memory":
		return NewMemoryStorage(config, m.logger), nil
	case "file":
		return NewFileStorage(config, m.logger)
	default:
		return NewMemoryStorage(config, m.logger), nil // Default to memory
	}
}

func (m *Manager) initializeNotifications() error {
	// Initialize Slack notification handler
	if m.config.Notifications.Slack.Enabled {
		slackHandler := NewSlackHandler(m.config.Notifications.Slack, m.logger)
		m.notifications[ChannelSlack] = slackHandler
		m.logger.GetSlogger().Info("Slack notifications enabled")
	}

	// Initialize Email notification handler
	if m.config.Notifications.Email.Enabled {
		emailHandler := NewEmailHandler(m.config.Notifications.Email, m.logger)
		m.notifications[ChannelEmail] = emailHandler
		m.logger.GetSlogger().Info("Email notifications enabled")
	}

	// Always enable log notifications
	logHandler := NewLogHandler(m.logger)
	m.notifications[ChannelLog] = logHandler

	return nil
}

func (m *Manager) loadRules() error {
	for _, rule := range m.config.Rules {
		m.rules[rule.ID] = rule
	}

	m.logger.GetSlogger().Info("Loaded alert rules", slog.Int("count", len(m.rules)))
	return nil
}

func (m *Manager) isRuleInCooldown(ruleID string) bool {
	// Simple implementation - could be enhanced with proper cooldown tracking
	return false
}

func (m *Manager) getTriggeredConditions(conditions []AlertCondition, data map[string]interface{}) []string {
	var triggered []string
	for _, condition := range conditions {
		if result, err := m.evaluator.EvaluateCondition(condition, data); err == nil && result {
			triggered = append(triggered, fmt.Sprintf("%s %s %v", condition.Field, condition.Operator, condition.Value))
		}
	}
	return triggered
}

func (m *Manager) sendNotifications(ctx context.Context, alert *Alert) error {
	// Get the rule to determine which channels to use
	var channels []NotificationChannel
	if ruleID, ok := alert.Metadata["rule_id"].(string); ok {
		m.rulesMux.RLock()
		if rule, exists := m.rules[ruleID]; exists {
			channels = rule.Channels
		}
		m.rulesMux.RUnlock()
	}

	// Default to log notifications if no channels specified
	if len(channels) == 0 {
		channels = []NotificationChannel{ChannelLog}
	}

	var lastErr error
	for _, channel := range channels {
		m.notificationsMux.RLock()
		handler, exists := m.notifications[channel]
		m.notificationsMux.RUnlock()

		if !exists {
			m.logger.GetSlogger().Warn("Notification handler not found", slog.String("channel", string(channel)))
			continue
		}

		// Create notification record
		record := NotificationRecord{
			Channel: channel,
			SentAt:  time.Now(),
		}

		// Send notification
		var config interface{}
		switch channel {
		case ChannelSlack:
			config = m.config.Notifications.Slack
		case ChannelEmail:
			config = m.config.Notifications.Email
		}

		if err := handler.SendNotification(ctx, alert, config); err != nil {
			record.Success = false
			record.Error = err.Error()
			lastErr = err
			m.logger.GetSlogger().Error("Failed to send notification",
				slog.String("channel", string(channel)),
				slog.String("error", err.Error()))
		} else {
			record.Success = true
			m.logger.GetSlogger().Debug("Notification sent successfully", slog.String("channel", string(channel)))
		}

		alert.NotificationsSent = append(alert.NotificationsSent, record)
	}

	return lastErr
}

func (m *Manager) incrementErrors() {
	m.errorsMux.Lock()
	m.processingErrors++
	m.errorsMux.Unlock()
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}
