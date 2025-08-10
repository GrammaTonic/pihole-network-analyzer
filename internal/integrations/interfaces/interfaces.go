package interfaces

import (
	"context"
	"io"
	"log/slog"
	"time"

	"pihole-analyzer/internal/types"
)

// MonitoringIntegration defines the interface for all monitoring platform integrations
type MonitoringIntegration interface {
	// Initialize sets up the integration with the provided configuration
	Initialize(ctx context.Context, config interface{}) error

	// IsEnabled returns whether the integration is enabled
	IsEnabled() bool

	// GetName returns the name of the integration
	GetName() string

	// GetStatus returns the current status of the integration
	GetStatus() IntegrationStatus

	// SendMetrics sends metrics data to the monitoring platform
	SendMetrics(ctx context.Context, data *types.AnalysisResult) error

	// SendLogs sends log data to the monitoring platform
	SendLogs(ctx context.Context, logs []LogEntry) error

	// TestConnection tests the connection to the monitoring platform
	TestConnection(ctx context.Context) error

	// Close cleans up resources and closes connections
	Close() error
}

// MetricsIntegration defines interface for metrics-specific integrations
type MetricsIntegration interface {
	MonitoringIntegration

	// PushMetrics pushes metrics to external systems
	PushMetrics(ctx context.Context, metrics map[string]interface{}) error

	// RegisterMetric registers a new metric definition
	RegisterMetric(name, help string, metricType MetricType, labels []string) error

	// SetMetric sets a metric value
	SetMetric(name string, value float64, labels map[string]string) error

	// IncrementCounter increments a counter metric
	IncrementCounter(name string, labels map[string]string) error
}

// LogsIntegration defines interface for logs-specific integrations
type LogsIntegration interface {
	MonitoringIntegration

	// StreamLogs streams logs to external systems
	StreamLogs(ctx context.Context, logStream <-chan LogEntry) error

	// WriteLogs writes a batch of logs
	WriteLogs(ctx context.Context, logs []LogEntry) error

	// SetLogLevel sets the minimum log level to forward
	SetLogLevel(level slog.Level)

	// AddStaticLabels adds static labels to all log entries
	AddStaticLabels(labels map[string]string)
}

// DashboardIntegration defines interface for dashboard management
type DashboardIntegration interface {
	MonitoringIntegration

	// CreateDashboard creates a new dashboard
	CreateDashboard(ctx context.Context, dashboard Dashboard) error

	// UpdateDashboard updates an existing dashboard
	UpdateDashboard(ctx context.Context, dashboard Dashboard) error

	// DeleteDashboard deletes a dashboard
	DeleteDashboard(ctx context.Context, id string) error

	// ListDashboards lists all dashboards
	ListDashboards(ctx context.Context) ([]Dashboard, error)

	// GetDashboard retrieves a specific dashboard
	GetDashboard(ctx context.Context, id string) (*Dashboard, error)
}

// AlertIntegration defines interface for alert management
type AlertIntegration interface {
	MonitoringIntegration

	// CreateAlert creates a new alert rule
	CreateAlert(ctx context.Context, alert AlertRule) error

	// UpdateAlert updates an existing alert rule
	UpdateAlert(ctx context.Context, alert AlertRule) error

	// DeleteAlert deletes an alert rule
	DeleteAlert(ctx context.Context, id string) error

	// ListAlerts lists all alert rules
	ListAlerts(ctx context.Context) ([]AlertRule, error)

	// TriggerAlert manually triggers an alert
	TriggerAlert(ctx context.Context, alert AlertRule, message string) error
}

// IntegrationManager manages all monitoring integrations
type IntegrationManager interface {
	// Initialize sets up all enabled integrations
	Initialize(ctx context.Context, config *types.IntegrationsConfig) error

	// RegisterIntegration registers a new integration
	RegisterIntegration(integration MonitoringIntegration) error

	// GetIntegration retrieves an integration by name
	GetIntegration(name string) (MonitoringIntegration, error)

	// GetEnabledIntegrations returns all enabled integrations
	GetEnabledIntegrations() []MonitoringIntegration

	// SendToAll sends data to all enabled integrations
	SendToAll(ctx context.Context, data *types.AnalysisResult) error

	// TestAll tests all enabled integrations
	TestAll(ctx context.Context) map[string]error

	// Close closes all integrations
	Close() error
}

// IntegrationStatus represents the status of an integration
type IntegrationStatus struct {
	Name        string            `json:"name"`
	Enabled     bool              `json:"enabled"`
	Connected   bool              `json:"connected"`
	LastConnect time.Time         `json:"last_connect"`
	LastError   string            `json:"last_error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// LogEntry represents a structured log entry for forwarding
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     slog.Level             `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	Labels    map[string]string      `json:"labels"`
	Fields    map[string]interface{} `json:"fields"`
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	FolderID    string            `json:"folder_id"`
	Tags        []string          `json:"tags"`
	Panels      []Panel           `json:"panels"`
	Variables   []Variable        `json:"variables"`
	Metadata    map[string]string `json:"metadata"`
	Definition  interface{}       `json:"definition"` // Platform-specific definition
}

// Panel represents a dashboard panel
type Panel struct {
	ID       int                    `json:"id"`
	Title    string                 `json:"title"`
	Type     string                 `json:"type"`
	Query    string                 `json:"query"`
	Settings map[string]interface{} `json:"settings"`
}

// Variable represents a dashboard variable
type Variable struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Query   string      `json:"query"`
	Default interface{} `json:"default"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
}

// DataSource represents a data source configuration
type DataSource struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	URL      string            `json:"url"`
	Settings map[string]string `json:"settings"`
}

// LogWriter is an io.Writer implementation for integration logging
type LogWriter struct {
	integration LogsIntegration
	labels      map[string]string
	level       slog.Level
}

// NewLogWriter creates a new LogWriter for the given integration
func NewLogWriter(integration LogsIntegration, level slog.Level, labels map[string]string) *LogWriter {
	return &LogWriter{
		integration: integration,
		labels:      labels,
		level:       level,
	}
}

// Write implements io.Writer interface
func (lw *LogWriter) Write(p []byte) (n int, err error) {
	if lw.integration == nil {
		return len(p), nil
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     lw.level,
		Message:   string(p),
		Labels:    lw.labels,
		Fields:    make(map[string]interface{}),
	}

	ctx := context.Background()
	if err := lw.integration.WriteLogs(ctx, []LogEntry{entry}); err != nil {
		return 0, err
	}

	return len(p), nil
}

// Compile-time interface checks
var (
	_ io.Writer = (*LogWriter)(nil)
)
