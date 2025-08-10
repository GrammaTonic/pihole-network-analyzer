package prometheus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

// Client implements enhanced Prometheus integration with push gateway and remote write
type Client struct {
	config     *types.PrometheusExtConfig
	logger     *slog.Logger
	httpClient *http.Client
	registry   *prometheus.Registry
	pusher     *push.Pusher
	enabled    bool
	status     interfaces.IntegrationStatus
	mu         sync.RWMutex
	metrics    map[string]prometheus.Collector
}

// NewClient creates a new enhanced Prometheus integration client
func NewClient(config *types.PrometheusExtConfig, logger *slog.Logger) *Client {
	registry := prometheus.NewRegistry()

	client := &Client{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		registry:   registry,
		enabled:    config.Enabled,
		metrics:    make(map[string]prometheus.Collector),
		status: interfaces.IntegrationStatus{
			Name:    "prometheus",
			Enabled: config.Enabled,
		},
	}

	if config.Enabled && config.PushGateway.Enabled {
		client.setupPushGateway()
	}

	return client
}

// Initialize sets up the enhanced Prometheus integration
func (c *Client) Initialize(ctx context.Context, config interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if promConfig, ok := config.(*types.PrometheusExtConfig); ok {
		c.config = promConfig
		c.enabled = promConfig.Enabled
	}

	if !c.enabled {
		c.logger.Debug("⏭️ Enhanced Prometheus integration disabled",
			slog.String("component", "prometheus"))
		return nil
	}

	c.logger.Info("🚀 Initializing enhanced Prometheus integration",
		slog.String("component", "prometheus"))

	// Setup push gateway if enabled
	if c.config.PushGateway.Enabled {
		c.setupPushGateway()

		// Test push gateway connection
		if err := c.TestConnection(ctx); err != nil {
			c.status.LastError = err.Error()
			return fmt.Errorf("failed to connect to Prometheus push gateway: %w", err)
		}
	}

	// Register default metrics
	if err := c.registerDefaultMetrics(); err != nil {
		return fmt.Errorf("failed to register default metrics: %w", err)
	}

	c.status.Connected = true
	c.status.LastConnect = time.Now()

	c.logger.Info("✅ Enhanced Prometheus integration initialized successfully",
		slog.String("component", "prometheus"))

	return nil
}

// IsEnabled returns whether the integration is enabled
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// GetName returns the name of the integration
func (c *Client) GetName() string {
	return "prometheus"
}

// GetStatus returns the current status of the integration
func (c *Client) GetStatus() interfaces.IntegrationStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// SendMetrics sends metrics data to Prometheus (push gateway or remote write)
func (c *Client) SendMetrics(ctx context.Context, data *types.AnalysisResult) error {
	if !c.enabled {
		return nil
	}

	c.logger.Debug("📊 Sending metrics to Prometheus",
		slog.String("component", "prometheus"),
		slog.Int("total_queries", data.TotalQueries),
		slog.Int("unique_clients", data.UniqueClients))

	// Update metrics based on analysis data
	if err := c.updateMetricsFromAnalysis(data); err != nil {
		return fmt.Errorf("failed to update metrics: %w", err)
	}

	// Push to gateway if enabled
	if c.config.PushGateway.Enabled {
		if err := c.pushToGateway(ctx); err != nil {
			return fmt.Errorf("failed to push to gateway: %w", err)
		}
	}

	// Send to remote write if enabled
	if c.config.RemoteWrite.Enabled {
		if err := c.sendToRemoteWrite(ctx); err != nil {
			return fmt.Errorf("failed to send to remote write: %w", err)
		}
	}

	return nil
}

// SendLogs sends log data (not implemented for Prometheus)
func (c *Client) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	// Prometheus is for metrics, not logs - this is a no-op
	return nil
}

// PushMetrics pushes metrics to external systems
func (c *Client) PushMetrics(ctx context.Context, metrics map[string]interface{}) error {
	if !c.enabled {
		return nil
	}

	c.logger.Debug("📤 Pushing custom metrics to Prometheus",
		slog.String("component", "prometheus"),
		slog.Int("metric_count", len(metrics)))

	// Convert and register custom metrics
	for name, value := range metrics {
		if err := c.setCustomMetric(name, value); err != nil {
			c.logger.Error("❌ Failed to set custom metric",
				slog.String("component", "prometheus"),
				slog.String("metric", name),
				slog.String("error", err.Error()))
			continue
		}
	}

	// Push to gateway if enabled
	if c.config.PushGateway.Enabled {
		return c.pushToGateway(ctx)
	}

	return nil
}

// RegisterMetric registers a new metric definition
func (c *Client) RegisterMetric(name, help string, metricType interfaces.MetricType, labels []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.registerMetricNoLock(name, help, metricType, labels)
}

// registerMetricNoLock registers a metric without acquiring the mutex lock
// This is used internally when the lock is already held
func (c *Client) registerMetricNoLock(name, help string, metricType interfaces.MetricType, labels []string) error {
	if _, exists := c.metrics[name]; exists {
		return fmt.Errorf("metric %s already registered", name)
	}

	var collector prometheus.Collector

	switch metricType {
	case interfaces.MetricTypeCounter:
		if len(labels) == 0 {
			collector = prometheus.NewCounter(prometheus.CounterOpts{
				Name: name,
				Help: help,
			})
		} else {
			collector = prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: name,
				Help: help,
			}, labels)
		}

	case interfaces.MetricTypeGauge:
		if len(labels) == 0 {
			collector = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: name,
				Help: help,
			})
		} else {
			collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: name,
				Help: help,
			}, labels)
		}

	case interfaces.MetricTypeHistogram:
		if len(labels) == 0 {
			collector = prometheus.NewHistogram(prometheus.HistogramOpts{
				Name: name,
				Help: help,
			})
		} else {
			collector = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: name,
				Help: help,
			}, labels)
		}

	default:
		return fmt.Errorf("unsupported metric type: %s", metricType)
	}

	if err := c.registry.Register(collector); err != nil {
		return fmt.Errorf("failed to register metric: %w", err)
	}

	c.metrics[name] = collector

	c.logger.Debug("📋 Registered Prometheus metric",
		slog.String("component", "prometheus"),
		slog.String("name", name),
		slog.String("type", string(metricType)))

	return nil
}

// SetMetric sets a metric value
func (c *Client) SetMetric(name string, value float64, labels map[string]string) error {
	c.mu.RLock()
	collector, exists := c.metrics[name]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("metric %s not found", name)
	}

	switch metric := collector.(type) {
	case prometheus.Gauge:
		metric.Set(value)
	case *prometheus.GaugeVec:
		metric.With(labels).Set(value)
	case prometheus.Counter:
		// Counters can only increase, so we'll treat this as Add
		metric.Add(value)
	case *prometheus.CounterVec:
		metric.With(labels).Add(value)
	default:
		return fmt.Errorf("cannot set value on metric type %T", metric)
	}

	return nil
}

// IncrementCounter increments a counter metric
func (c *Client) IncrementCounter(name string, labels map[string]string) error {
	c.mu.RLock()
	collector, exists := c.metrics[name]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("metric %s not found", name)
	}

	// Use concrete type checking to distinguish between counters and other metrics
	switch collector.(type) {
	case *prometheus.CounterVec:
		metric := collector.(*prometheus.CounterVec)
		metric.With(labels).Inc()
		return nil
	default:
		// Check if it's a single counter by trying to cast to the prometheus interface
		// and verifying it's not another type
		if counter, ok := collector.(prometheus.Counter); ok {
			// But first check if it's actually a gauge disguised as a counter
			typeName := fmt.Sprintf("%T", collector)
			if strings.Contains(typeName, "gauge") || strings.Contains(typeName, "Gauge") {
				return fmt.Errorf("metric %s is a gauge, not a counter", name)
			}
			counter.Inc()
			return nil
		}
		return fmt.Errorf("metric %s is not a counter (type: %T)", name, collector)
	}
}

// TestConnection tests the connection to Prometheus push gateway
func (c *Client) TestConnection(ctx context.Context) error {
	if !c.enabled || !c.config.PushGateway.Enabled {
		return nil
	}

	// Test push gateway with empty metric set
	url := fmt.Sprintf("%s/metrics/job/%s", c.config.PushGateway.URL, c.config.PushGateway.Job)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus push gateway: %w", err)
	}
	defer resp.Body.Close()

	// Push gateway returns 200 even for non-existent metrics
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Prometheus push gateway error: %d %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("✅ Prometheus push gateway connection test successful",
		slog.String("component", "prometheus"))

	return nil
}

// Close cleans up resources and closes connections
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	c.logger.Info("🔌 Closing enhanced Prometheus integration",
		slog.String("component", "prometheus"))

	// Final push if push gateway is enabled
	if c.config.PushGateway.Enabled && c.pusher != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.pushToGateway(ctx); err != nil {
			c.logger.Error("❌ Failed to perform final push",
				slog.String("component", "prometheus"),
				slog.String("error", err.Error()))
		}
	}

	c.enabled = false
	c.status.Connected = false

	c.logger.Info("✅ Enhanced Prometheus integration closed successfully",
		slog.String("component", "prometheus"))

	return nil
}

// setupPushGateway configures the Prometheus push gateway
func (c *Client) setupPushGateway() {
	c.pusher = push.New(c.config.PushGateway.URL, c.config.PushGateway.Job).
		Gatherer(c.registry)

	if c.config.PushGateway.Instance != "" {
		c.pusher = c.pusher.Grouping("instance", c.config.PushGateway.Instance)
	}

	// Add external labels
	for key, value := range c.config.ExternalLabels {
		c.pusher = c.pusher.Grouping(key, value)
	}

	c.logger.Debug("🔧 Configured Prometheus push gateway",
		slog.String("component", "prometheus"),
		slog.String("url", c.config.PushGateway.URL),
		slog.String("job", c.config.PushGateway.Job))
}

// registerDefaultMetrics registers default analysis metrics
func (c *Client) registerDefaultMetrics() error {
	defaultMetrics := []struct {
		name       string
		help       string
		metricType interfaces.MetricType
		labels     []string
	}{
		{"pihole_analyzer_ext_total_queries", "Total number of DNS queries processed", interfaces.MetricTypeCounter, nil},
		{"pihole_analyzer_ext_unique_clients", "Number of unique clients", interfaces.MetricTypeGauge, nil},
		{"pihole_analyzer_ext_analysis_duration", "Duration of analysis process", interfaces.MetricTypeHistogram, nil},
		{"pihole_analyzer_ext_client_queries", "Query count per client", interfaces.MetricTypeGauge, []string{"client", "hostname"}},
		{"pihole_analyzer_ext_domain_queries", "Query count per domain", interfaces.MetricTypeCounter, []string{"domain"}},
	}

	for _, metric := range defaultMetrics {
		if err := c.registerMetricNoLock(metric.name, metric.help, metric.metricType, metric.labels); err != nil {
			return fmt.Errorf("failed to register metric %s: %w", metric.name, err)
		}
	}

	return nil
}

// updateMetricsFromAnalysis updates metrics based on analysis results
func (c *Client) updateMetricsFromAnalysis(data *types.AnalysisResult) error {
	// Update total queries
	if err := c.SetMetric("pihole_analyzer_ext_total_queries", float64(data.TotalQueries), nil); err != nil {
		c.logger.Error("❌ Failed to update total queries metric",
			slog.String("component", "prometheus"),
			slog.String("error", err.Error()))
	}

	// Update unique clients
	if err := c.SetMetric("pihole_analyzer_ext_unique_clients", float64(data.UniqueClients), nil); err != nil {
		c.logger.Error("❌ Failed to update unique clients metric",
			slog.String("component", "prometheus"),
			slog.String("error", err.Error()))
	}

	// Update client-specific metrics
	for clientIP, stats := range data.ClientStats {
		labels := map[string]string{
			"client":   clientIP,
			"hostname": stats.Hostname,
		}

		if err := c.SetMetric("pihole_analyzer_ext_client_queries", float64(stats.QueryCount), labels); err != nil {
			c.logger.Debug("⚠️ Failed to update client queries metric",
				slog.String("component", "prometheus"),
				slog.String("client", clientIP),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// pushToGateway pushes metrics to the push gateway
func (c *Client) pushToGateway(ctx context.Context) error {
	if c.pusher == nil {
		return fmt.Errorf("push gateway not configured")
	}

	start := time.Now()

	if err := c.pusher.Push(); err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}

	c.logger.Debug("📤 Successfully pushed metrics to gateway",
		slog.String("component", "prometheus"),
		slog.Duration("duration", time.Since(start)))

	return nil
}

// sendToRemoteWrite sends metrics to remote write endpoint
func (c *Client) sendToRemoteWrite(ctx context.Context) error {
	if !c.config.RemoteWrite.Enabled {
		return nil
	}

	c.logger.Debug("📡 Sending metrics to remote write endpoint",
		slog.String("component", "prometheus"),
		slog.String("url", c.config.RemoteWrite.URL))

	// This is a simplified implementation
	// In production, you'd use the Prometheus remote write protocol
	// with protobuf and snappy compression

	// For now, we'll just log that this feature needs implementation
	c.logger.Info("🔧 Remote write feature needs full implementation",
		slog.String("component", "prometheus"))

	return nil
}

// setCustomMetric sets a custom metric value
func (c *Client) setCustomMetric(name string, value interface{}) error {
	floatValue := 0.0

	switch v := value.(type) {
	case float64:
		floatValue = v
	case float32:
		floatValue = float64(v)
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			floatValue = parsed
		} else {
			return fmt.Errorf("cannot convert string value to float: %s", v)
		}
	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}

	// Try to find existing metric or create a gauge
	metricName := "pihole_analyzer_ext_custom_" + strings.ReplaceAll(name, ".", "_")

	if _, exists := c.metrics[metricName]; !exists {
		if err := c.RegisterMetric(metricName, "Custom metric: "+name, interfaces.MetricTypeGauge, nil); err != nil {
			return err
		}
	}

	return c.SetMetric(metricName, floatValue, nil)
}

// addAuth adds authentication to the request
func (c *Client) addAuth(req *http.Request) {
	if c.config.PushGateway.Username != "" && c.config.PushGateway.Password != "" {
		req.SetBasicAuth(c.config.PushGateway.Username, c.config.PushGateway.Password)
	}
}

// GetRegistry returns the Prometheus registry
func (c *Client) GetRegistry() *prometheus.Registry {
	return c.registry
}

// Compile-time interface checks
var (
	_ interfaces.MonitoringIntegration = (*Client)(nil)
	_ interfaces.MetricsIntegration    = (*Client)(nil)
)
