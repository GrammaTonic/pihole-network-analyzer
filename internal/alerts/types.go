package alerts

import (
	"time"

	"pihole-analyzer/internal/ml"
)

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeAnomaly       AlertType = "anomaly"
	AlertTypeThreshold     AlertType = "threshold"
	AlertTypeConnectivity  AlertType = "connectivity"
	AlertTypePerformance   AlertType = "performance"
	AlertTypeSecurity      AlertType = "security"
	AlertTypeConfiguration AlertType = "configuration"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
	SeverityError    AlertSeverity = "error"
)

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusFired      AlertStatus = "fired"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
	AlertStatusPending    AlertStatus = "pending"
)

// Alert represents a system alert
type Alert struct {
	ID          string        `json:"id"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Status      AlertStatus   `json:"status"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Timestamp   time.Time     `json:"timestamp"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`

	// Alert source information
	Source   string `json:"source"` // Component that triggered the alert
	ClientIP string `json:"client_ip,omitempty"`
	Domain   string `json:"domain,omitempty"`

	// Related data
	Metadata map[string]interface{} `json:"metadata"`
	Tags     []string               `json:"tags"`

	// ML Integration
	Anomaly    *ml.Anomaly `json:"anomaly,omitempty"`
	MLScore    float64     `json:"ml_score,omitempty"`
	Confidence float64     `json:"confidence,omitempty"`

	// Notification tracking
	NotificationsSent []NotificationRecord `json:"notifications_sent"`
	SuppressUntil     *time.Time           `json:"suppress_until,omitempty"`
}

// NotificationRecord tracks sent notifications for an alert
type NotificationRecord struct {
	Channel   NotificationChannel `json:"channel"`
	SentAt    time.Time           `json:"sent_at"`
	Success   bool                `json:"success"`
	Error     string              `json:"error,omitempty"`
	MessageID string              `json:"message_id,omitempty"`
}

// NotificationChannel defines where notifications are sent
type NotificationChannel string

const (
	ChannelSlack NotificationChannel = "slack"
	ChannelEmail NotificationChannel = "email"
	ChannelLog   NotificationChannel = "log"
)

// AlertRule defines criteria for triggering alerts
type AlertRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Enabled     bool          `json:"enabled"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`

	// Rule conditions
	Conditions []AlertCondition `json:"conditions"`

	// Suppression settings
	CooldownPeriod time.Duration `json:"cooldown_period"`
	MaxAlerts      int           `json:"max_alerts"`

	// Notification settings
	Channels   []NotificationChannel `json:"channels"`
	Recipients []string              `json:"recipients"`

	// Tags for organization
	Tags []string `json:"tags"`
}

// AlertCondition defines a condition that must be met to trigger an alert
type AlertCondition struct {
	Field      string      `json:"field"`                 // e.g., "query_count", "anomaly_score", "response_time"
	Operator   string      `json:"operator"`              // e.g., "gt", "lt", "eq", "contains"
	Value      interface{} `json:"value"`                 // Threshold value
	TimeWindow string      `json:"time_window,omitempty"` // Duration string
}

// AlertConfig represents the configuration for the alert system
type AlertConfig struct {
	Enabled       bool                   `json:"enabled"`
	Rules         []AlertRule            `json:"rules"`
	Notifications NotificationConfig     `json:"notifications"`
	Storage       StorageConfig          `json:"storage"`
	Performance   AlertPerformanceConfig `json:"performance"`

	// Default settings
	DefaultSeverity AlertSeverity `json:"default_severity"`
	DefaultCooldown time.Duration `json:"default_cooldown"`
	MaxActiveAlerts int           `json:"max_active_alerts"`
	AlertRetention  time.Duration `json:"alert_retention"`
}

// NotificationConfig configures notification channels
type NotificationConfig struct {
	Slack SlackConfig `json:"slack"`
	Email EmailConfig `json:"email"`
}

// SlackConfig configures Slack notifications
type SlackConfig struct {
	Enabled    bool          `json:"enabled"`
	WebhookURL string        `json:"webhook_url"`
	Channel    string        `json:"channel"`
	Username   string        `json:"username"`
	IconEmoji  string        `json:"icon_emoji"`
	Timeout    time.Duration `json:"timeout"`
}

// EmailConfig configures email notifications
type EmailConfig struct {
	Enabled    bool          `json:"enabled"`
	SMTPHost   string        `json:"smtp_host"`
	SMTPPort   int           `json:"smtp_port"`
	Username   string        `json:"username"`
	Password   string        `json:"password"`
	From       string        `json:"from"`
	Recipients []string      `json:"recipients"`
	UseTLS     bool          `json:"use_tls"`
	Timeout    time.Duration `json:"timeout"`
}

// StorageConfig configures alert storage
type StorageConfig struct {
	Type      string        `json:"type"` // "memory", "file", "database"
	Path      string        `json:"path,omitempty"`
	MaxSize   int           `json:"max_size"` // Maximum number of alerts to store
	Retention time.Duration `json:"retention"`
}

// AlertPerformanceConfig configures performance settings
type AlertPerformanceConfig struct {
	MaxConcurrentNotifications int           `json:"max_concurrent_notifications"`
	NotificationTimeout        time.Duration `json:"notification_timeout"`
	BatchSize                  int           `json:"batch_size"`
	EvaluationInterval         time.Duration `json:"evaluation_interval"`
}

// DefaultAlertConfig returns a sensible default configuration
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		Enabled: true,
		Rules:   DefaultAlertRules(),
		Notifications: NotificationConfig{
			Slack: SlackConfig{
				Enabled:   false,
				Channel:   "#alerts",
				Username:  "PiHole-Analyzer",
				IconEmoji: ":warning:",
				Timeout:   30 * time.Second,
			},
			Email: EmailConfig{
				Enabled:  false,
				SMTPPort: 587,
				UseTLS:   true,
				Timeout:  30 * time.Second,
			},
		},
		Storage: StorageConfig{
			Type:      "memory",
			MaxSize:   1000,
			Retention: 7 * 24 * time.Hour, // 7 days
		},
		Performance: AlertPerformanceConfig{
			MaxConcurrentNotifications: 5,
			NotificationTimeout:        30 * time.Second,
			BatchSize:                  10,
			EvaluationInterval:         30 * time.Second,
		},
		DefaultSeverity: SeverityWarning,
		DefaultCooldown: 5 * time.Minute,
		MaxActiveAlerts: 100,
		AlertRetention:  7 * 24 * time.Hour,
	}
}

// DefaultAlertRules returns default alert rules
func DefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			ID:          "high-anomaly-score",
			Name:        "High Anomaly Score Detected",
			Description: "Alert when ML anomaly score exceeds threshold",
			Enabled:     true,
			Type:        AlertTypeAnomaly,
			Severity:    SeverityWarning,
			Conditions: []AlertCondition{
				{
					Field:    "anomaly_score",
					Operator: "gt",
					Value:    0.8,
				},
			},
			CooldownPeriod: 5 * time.Minute,
			Channels:       []NotificationChannel{ChannelLog},
			Tags:           []string{"ml", "anomaly"},
		},
		{
			ID:          "critical-anomaly",
			Name:        "Critical Anomaly Detected",
			Description: "Alert when critical ML anomaly is detected",
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
			CooldownPeriod: 2 * time.Minute,
			Channels:       []NotificationChannel{ChannelLog, ChannelSlack, ChannelEmail},
			Tags:           []string{"ml", "critical"},
		},
		{
			ID:          "volume-spike",
			Name:        "DNS Query Volume Spike",
			Description: "Alert when DNS query volume spikes unexpectedly",
			Enabled:     true,
			Type:        AlertTypeThreshold,
			Severity:    SeverityWarning,
			Conditions: []AlertCondition{
				{
					Field:      "query_count",
					Operator:   "gt",
					Value:      1000,
					TimeWindow: "5m",
				},
			},
			CooldownPeriod: 10 * time.Minute,
			Channels:       []NotificationChannel{ChannelLog},
			Tags:           []string{"volume", "threshold"},
		},
		{
			ID:          "unusual-domain",
			Name:        "Unusual Domain Activity",
			Description: "Alert when unusual domain patterns are detected",
			Enabled:     true,
			Type:        AlertTypeSecurity,
			Severity:    SeverityWarning,
			Conditions: []AlertCondition{
				{
					Field:    "anomaly_type",
					Operator: "eq",
					Value:    "unusual_domain",
				},
			},
			CooldownPeriod: 15 * time.Minute,
			Channels:       []NotificationChannel{ChannelLog, ChannelSlack},
			Tags:           []string{"security", "domain"},
		},
		{
			ID:          "pihole-connectivity",
			Name:        "Pi-hole Connection Issues",
			Description: "Alert when connection to Pi-hole API fails",
			Enabled:     true,
			Type:        AlertTypeConnectivity,
			Severity:    SeverityError,
			Conditions: []AlertCondition{
				{
					Field:    "connection_status",
					Operator: "eq",
					Value:    "disconnected",
				},
			},
			CooldownPeriod: 1 * time.Minute,
			Channels:       []NotificationChannel{ChannelLog, ChannelEmail},
			Tags:           []string{"connectivity", "pihole"},
		},
	}
}
