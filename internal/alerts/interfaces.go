package alerts

import (
	"context"

	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

// AlertManager defines the interface for managing alerts
type AlertManager interface {
	// Initialize the alert manager
	Initialize(ctx context.Context, config AlertConfig) error

	// Process data and evaluate alert rules
	ProcessData(ctx context.Context, data *types.AnalysisResult, mlResults *ml.MLResults) error

	// Fire an alert
	FireAlert(ctx context.Context, alert *Alert) error

	// Resolve an alert
	ResolveAlert(ctx context.Context, alertID string) error

	// Get active alerts
	GetActiveAlerts() ([]*Alert, error)

	// Get alert history
	GetAlertHistory(limit int) ([]*Alert, error)

	// Add or update an alert rule
	UpdateRule(rule AlertRule) error

	// Remove an alert rule
	RemoveRule(ruleID string) error

	// Get all rules
	GetRules() ([]AlertRule, error)

	// Get manager status
	GetStatus() ManagerStatus

	// Close and cleanup
	Close() error
}

// Evaluator defines the interface for evaluating alert conditions
type Evaluator interface {
	// Evaluate conditions against data
	EvaluateConditions(conditions []AlertCondition, data map[string]interface{}) (bool, error)

	// Evaluate a single condition
	EvaluateCondition(condition AlertCondition, data map[string]interface{}) (bool, error)
}

// NotificationHandler defines the interface for sending notifications
type NotificationHandler interface {
	// Send notification
	SendNotification(ctx context.Context, alert *Alert, config interface{}) error

	// Get notification channel type
	GetChannelType() NotificationChannel

	// Test notification configuration
	TestNotification(ctx context.Context, config interface{}) error
}

// Storage defines the interface for storing alerts
type Storage interface {
	// Store an alert
	Store(alert *Alert) error

	// Get alert by ID
	Get(id string) (*Alert, error)

	// Update alert
	Update(alert *Alert) error

	// Delete alert
	Delete(id string) error

	// List alerts with filters
	List(filters map[string]interface{}) ([]*Alert, error)

	// Get active alerts
	GetActive() ([]*Alert, error)

	// Cleanup old alerts
	Cleanup() error

	// Close storage
	Close() error
}

// ManagerStatus represents the status of the alert manager
type ManagerStatus struct {
	IsRunning        bool   `json:"is_running"`
	ActiveAlerts     int    `json:"active_alerts"`
	TotalAlerts      int    `json:"total_alerts"`
	RulesCount       int    `json:"rules_count"`
	LastEvaluation   string `json:"last_evaluation"`
	ProcessingErrors int    `json:"processing_errors"`
	Status           string `json:"status"`
}
