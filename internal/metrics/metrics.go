package metrics

import (
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector holds all the Prometheus metrics for the pihole network analyzer
type Collector struct {
	// Query-related metrics
	totalQueries      prometheus.Counter
	queriesPerSecond  prometheus.Gauge
	queryResponseTime prometheus.Histogram
	queriesByType     *prometheus.CounterVec
	queriesByStatus   *prometheus.CounterVec

	// Client-related metrics
	activeClients prometheus.Gauge
	topClients    *prometheus.GaugeVec
	uniqueClients prometheus.Gauge

	// Domain-related metrics
	topDomains     *prometheus.CounterVec
	blockedDomains prometheus.Counter
	allowedDomains prometheus.Counter

	// Performance metrics
	analysisProcessTime prometheus.Histogram
	piholeAPICallTime   prometheus.Histogram
	errorCount          *prometheus.CounterVec

	// System metrics
	lastAnalysisTime prometheus.Gauge
	analysisDuration prometheus.Histogram
	dataSourceHealth prometheus.Gauge

	// Logger for structured logging
	logger   *slog.Logger
	registry *prometheus.Registry
	mu       sync.RWMutex
}

// New creates a new metrics collector with all the Prometheus metrics
func New(logger *slog.Logger) *Collector {
	registry := prometheus.NewRegistry()

	collector := &Collector{
		logger:   logger,
		registry: registry,

		// Query metrics
		totalQueries: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pihole_analyzer_total_queries",
			Help: "Total number of DNS queries processed by the analyzer",
		}),

		queriesPerSecond: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pihole_analyzer_queries_per_second",
			Help: "Current rate of queries per second being processed",
		}),

		queryResponseTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "pihole_analyzer_query_response_time_seconds",
			Help:    "Time taken to process DNS queries",
			Buckets: prometheus.DefBuckets,
		}),

		queriesByType: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pihole_analyzer_queries_by_type_total",
				Help: "Total number of queries by query type (A, AAAA, PTR, etc.)",
			},
			[]string{"query_type"},
		),

		queriesByStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pihole_analyzer_queries_by_status_total",
				Help: "Total number of queries by status (blocked, allowed, etc.)",
			},
			[]string{"status"},
		),

		// Client metrics
		activeClients: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pihole_analyzer_active_clients",
			Help: "Number of active clients detected in the network",
		}),

		topClients: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pihole_analyzer_top_clients_queries",
				Help: "Query count for top clients",
			},
			[]string{"client_ip", "hostname"},
		),

		uniqueClients: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pihole_analyzer_unique_clients",
			Help: "Number of unique clients seen in the current analysis",
		}),

		// Domain metrics
		topDomains: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pihole_analyzer_top_domains_total",
				Help: "Total queries for top domains",
			},
			[]string{"domain"},
		),

		blockedDomains: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pihole_analyzer_blocked_domains_total",
			Help: "Total number of blocked domain queries",
		}),

		allowedDomains: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pihole_analyzer_allowed_domains_total",
			Help: "Total number of allowed domain queries",
		}),

		// Performance metrics
		analysisProcessTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "pihole_analyzer_analysis_process_time_seconds",
			Help:    "Time taken to process the complete analysis",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
		}),

		piholeAPICallTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "pihole_analyzer_api_call_time_seconds",
			Help:    "Time taken for Pi-hole API calls",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}),

		errorCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pihole_analyzer_errors_total",
				Help: "Total number of errors by error type",
			},
			[]string{"error_type"},
		),

		// System metrics
		lastAnalysisTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pihole_analyzer_last_analysis_timestamp",
			Help: "Unix timestamp of the last successful analysis",
		}),

		analysisDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "pihole_analyzer_analysis_duration_seconds",
			Help:    "Duration of the complete analysis process",
			Buckets: []float64{1.0, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0},
		}),

		dataSourceHealth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pihole_analyzer_data_source_health",
			Help: "Health status of the data source (1 = healthy, 0 = unhealthy)",
		}),
	}

	// Register all metrics with the registry
	registry.MustRegister(
		collector.totalQueries,
		collector.queriesPerSecond,
		collector.queryResponseTime,
		collector.queriesByType,
		collector.queriesByStatus,
		collector.activeClients,
		collector.topClients,
		collector.uniqueClients,
		collector.topDomains,
		collector.blockedDomains,
		collector.allowedDomains,
		collector.analysisProcessTime,
		collector.piholeAPICallTime,
		collector.errorCount,
		collector.lastAnalysisTime,
		collector.analysisDuration,
		collector.dataSourceHealth,
	)

	// Log metrics initialization
	logger.Info("🔍 Prometheus metrics collector initialized",
		slog.String("component", "metrics"),
		slog.Int("metrics_count", 16))

	return collector
}

// RecordTotalQueries increments the total queries counter
func (c *Collector) RecordTotalQueries(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalQueries.Add(count)
	c.logger.Debug("📊 Recorded total queries",
		slog.Float64("count", count))
}

// RecordQueryResponseTime records the time taken to process a query
func (c *Collector) RecordQueryResponseTime(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queryResponseTime.Observe(duration.Seconds())
	c.logger.Debug("⏱️ Recorded query response time",
		slog.Duration("duration", duration))
}

// RecordQueryByType increments the counter for a specific query type
func (c *Collector) RecordQueryByType(queryType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queriesByType.WithLabelValues(queryType).Inc()
	c.logger.Debug("🏷️ Recorded query by type",
		slog.String("query_type", queryType))
}

// RecordQueryByStatus increments the counter for a specific query status
func (c *Collector) RecordQueryByStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queriesByStatus.WithLabelValues(status).Inc()
	c.logger.Debug("📋 Recorded query by status",
		slog.String("status", status))
}

// SetActiveClients sets the number of active clients
func (c *Collector) SetActiveClients(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeClients.Set(count)
	c.logger.Debug("👥 Updated active clients count",
		slog.Float64("count", count))
}

// SetUniqueClients sets the number of unique clients
func (c *Collector) SetUniqueClients(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.uniqueClients.Set(count)
	c.logger.Debug("🆔 Updated unique clients count",
		slog.Float64("count", count))
}

// RecordTopClient records query count for a top client
func (c *Collector) RecordTopClient(clientIP, hostname string, queryCount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topClients.WithLabelValues(clientIP, hostname).Set(queryCount)
	c.logger.Debug("🔝 Recorded top client",
		slog.String("client_ip", clientIP),
		slog.String("hostname", hostname),
		slog.Float64("query_count", queryCount))
}

// RecordTopDomain increments the counter for a top domain
func (c *Collector) RecordTopDomain(domain string, count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topDomains.WithLabelValues(domain).Add(count)
	c.logger.Debug("🌐 Recorded top domain",
		slog.String("domain", domain),
		slog.Float64("count", count))
}

// RecordBlockedDomains increments the blocked domains counter
func (c *Collector) RecordBlockedDomains(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blockedDomains.Add(count)
	c.logger.Debug("🚫 Recorded blocked domains",
		slog.Float64("count", count))
}

// RecordAllowedDomains increments the allowed domains counter
func (c *Collector) RecordAllowedDomains(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.allowedDomains.Add(count)
	c.logger.Debug("✅ Recorded allowed domains",
		slog.Float64("count", count))
}

// RecordAnalysisProcessTime records the time taken for analysis processing
func (c *Collector) RecordAnalysisProcessTime(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.analysisProcessTime.Observe(duration.Seconds())
	c.logger.Info("📈 Recorded analysis process time",
		slog.Duration("duration", duration))
}

// RecordPiholeAPICallTime records the time taken for Pi-hole API calls
func (c *Collector) RecordPiholeAPICallTime(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.piholeAPICallTime.Observe(duration.Seconds())
	c.logger.Debug("🔄 Recorded Pi-hole API call time",
		slog.Duration("duration", duration))
}

// RecordError increments the error counter for a specific error type
func (c *Collector) RecordError(errorType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errorCount.WithLabelValues(errorType).Inc()
	c.logger.Warn("⚠️ Recorded error",
		slog.String("error_type", errorType))
}

// SetLastAnalysisTime sets the timestamp of the last successful analysis
func (c *Collector) SetLastAnalysisTime(timestamp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastAnalysisTime.Set(float64(timestamp.Unix()))
	c.logger.Info("⏰ Updated last analysis timestamp",
		slog.Time("timestamp", timestamp))
}

// RecordAnalysisDuration records the duration of the complete analysis
func (c *Collector) RecordAnalysisDuration(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.analysisDuration.Observe(duration.Seconds())
	c.logger.Info("📊 Recorded analysis duration",
		slog.Duration("duration", duration))
}

// SetDataSourceHealth sets the health status of the data source
func (c *Collector) SetDataSourceHealth(healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value := 0.0
	if healthy {
		value = 1.0
	}
	c.dataSourceHealth.Set(value)
	c.logger.Info("🏥 Updated data source health",
		slog.Bool("healthy", healthy))
}

// SetQueriesPerSecond sets the current rate of queries per second
func (c *Collector) SetQueriesPerSecond(qps float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queriesPerSecond.Set(qps)
	c.logger.Debug("⚡ Updated queries per second",
		slog.Float64("qps", qps))
}

// GetRegistry returns the Prometheus registry with all metrics
func (c *Collector) GetRegistry() *prometheus.Registry {
	return c.registry
}
