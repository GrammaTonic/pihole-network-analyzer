package analyzer

import (
	"context"
	"fmt"
	"time"

	"pihole-analyzer/internal/alerts"
	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/metrics"
	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

// EnhancedAnalyzer provides universal analysis logic regardless of data source
type EnhancedAnalyzer struct {
	dataSource       interfaces.DataSource
	config           *types.Config
	logger           *logger.Logger
	metricsCollector *metrics.Collector
	mlEngine         ml.MLEngine
	alertManager     alerts.AlertManager
}

// NewEnhancedAnalyzer creates a new analyzer with API data source
func NewEnhancedAnalyzer(config *types.Config, logger *logger.Logger, metricsCollector *metrics.Collector) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		config:           config,
		logger:           logger.Component("enhanced-analyzer"),
		metricsCollector: metricsCollector,
	}
}

// Initialize creates and connects the appropriate data source
func (a *EnhancedAnalyzer) Initialize(ctx context.Context) error {
	a.logger.Info("🔄 Initializing enhanced analyzer with API data source")

	// Create API data source factory
	factory := interfaces.NewDataSourceFactory(a.logger)

	// Create data source
	dataSource, err := factory.CreateDataSource(a.config)
	if err != nil {
		return fmt.Errorf("failed to create data source: %w", err)
	}

	// Connect to data source
	if err := dataSource.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to data source: %w", err)
	}

	a.dataSource = dataSource

	// Initialize ML engine if enabled
	if a.config.ML.AnomalyDetection.Enabled || a.config.ML.TrendAnalysis.Enabled {
		a.logger.Info("🧠 Initializing ML engine")

		// Convert config types
		mlConfig := a.convertToMLConfig(a.config.ML)

		a.mlEngine = ml.NewEngine(mlConfig, a.logger.GetSlogger())
		if err := a.mlEngine.Initialize(ctx, mlConfig); err != nil {
			a.logger.Warn("Failed to initialize ML engine: %v", err)
			// Continue without ML - not critical
		} else {
			a.logger.Info("✅ ML engine initialized successfully")
		}
	}

	// Initialize alert manager if enabled
	if a.config.Alerts.Enabled {
		a.logger.Info("🚨 Initializing alert manager")

		// Convert config types
		alertConfig := a.convertToAlertConfig(a.config.Alerts)

		a.alertManager = alerts.NewManager(alertConfig, a.logger)
		if err := a.alertManager.Initialize(ctx, alertConfig); err != nil {
			a.logger.Warn("Failed to initialize alert manager: %v", err)
			// Continue without alerts - not critical
		} else {
			a.logger.Info("✅ Alert manager initialized successfully")
		}
	}

	a.logger.Info("✅ Enhanced analyzer initialized successfully")
	return nil
}

// AnalyzeData performs DNS data analysis using API data source
func (a *EnhancedAnalyzer) AnalyzeData(ctx context.Context) (*types.AnalysisResult, error) {
	if a.dataSource == nil {
		return nil, fmt.Errorf("analyzer not initialized - call Initialize() first")
	}

	// Start timing the analysis
	analysisStart := time.Now()

	a.logger.Info("📊 Starting enhanced data analysis")

	// Set data source health to healthy initially
	if a.metricsCollector != nil {
		a.metricsCollector.SetDataSourceHealth(true)
	}

	// Get DNS queries from data source
	queryStart := time.Now()
	queries, err := a.dataSource.GetQueries(ctx, interfaces.QueryParams{
		Limit: 10000, // Default limit for analysis
	})
	if err != nil {
		if a.metricsCollector != nil {
			a.metricsCollector.SetDataSourceHealth(false)
			a.metricsCollector.RecordError("query_retrieval_failed")
		}
		return nil, fmt.Errorf("failed to get DNS queries: %w", err)
	}

	// Record API call time
	if a.metricsCollector != nil {
		a.metricsCollector.RecordPiholeAPICallTime(time.Since(queryStart))
	}

	a.logger.Info("Retrieved %d DNS queries for analysis", len(queries))

	// Record total queries
	if a.metricsCollector != nil {
		a.metricsCollector.RecordTotalQueries(float64(len(queries)))
	}

	// Get client statistics from data source
	clientStart := time.Now()
	clientStats, err := a.dataSource.GetClientStats(ctx)
	if err != nil {
		if a.metricsCollector != nil {
			a.metricsCollector.SetDataSourceHealth(false)
			a.metricsCollector.RecordError("client_stats_retrieval_failed")
		}
		return nil, fmt.Errorf("failed to get client statistics: %w", err)
	}

	// Record API call time
	if a.metricsCollector != nil {
		a.metricsCollector.RecordPiholeAPICallTime(time.Since(clientStart))
	}

	a.logger.Info("Retrieved statistics for %d clients", len(clientStats))

	// Record client metrics
	if a.metricsCollector != nil {
		a.metricsCollector.SetUniqueClients(float64(len(clientStats)))
		a.metricsCollector.SetActiveClients(float64(a.countActiveClients(clientStats)))
	}

	// Get network information
	networkStart := time.Now()
	networkDevices, err := a.dataSource.GetNetworkInfo(ctx)
	if err != nil {
		a.logger.Warn("Failed to get network information: %v", err)
		if a.metricsCollector != nil {
			a.metricsCollector.RecordError("network_info_retrieval_failed")
		}
		// Continue without network info - not critical
	} else if a.metricsCollector != nil {
		a.metricsCollector.RecordPiholeAPICallTime(time.Since(networkStart))
	}

	// Enhance client statistics with network analysis
	a.enhanceWithNetworkAnalysis(clientStats, networkDevices)

	// Collect detailed metrics from queries and client stats
	if a.metricsCollector != nil {
		a.collectDetailedMetrics(queries, clientStats)
	}

	// Calculate queries per second based on analysis timeframe
	analysisTime := time.Since(analysisStart)
	qps := float64(len(queries)) / analysisTime.Seconds()
	if a.metricsCollector != nil {
		a.metricsCollector.SetQueriesPerSecond(qps)
		a.metricsCollector.RecordAnalysisProcessTime(analysisTime)
		a.metricsCollector.SetLastAnalysisTime(time.Now())
	}

	// Create comprehensive analysis result
	result := &types.AnalysisResult{
		ClientStats:    clientStats,
		NetworkDevices: networkDevices,
		TotalQueries:   len(queries),
		UniqueClients:  len(clientStats),
		AnalysisMode:   a.getAnalysisMode(),
		DataSourceType: string(a.dataSource.GetDataSourceType()),
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	// Process ML analysis if enabled
	var mlResults *ml.MLResults
	if a.mlEngine != nil {
		a.logger.Info("🧠 Running ML analysis")
		mlStart := time.Now()

		var err error
		mlResults, err = a.mlEngine.ProcessData(ctx, queries)
		if err != nil {
			a.logger.Warn("ML analysis failed: %v", err)
			if a.metricsCollector != nil {
				a.metricsCollector.RecordError("ml_analysis_failed")
			}
		} else {
			mlDuration := time.Since(mlStart)
			a.logger.Info("✅ ML analysis complete: %d anomalies detected in %s",
				len(mlResults.Anomalies), mlDuration)

			if a.metricsCollector != nil {
				a.metricsCollector.RecordAnalysisProcessTime(mlDuration)
				// Record ML metrics
				for _, anomaly := range mlResults.Anomalies {
					a.metricsCollector.RecordError("anomaly_detected_" + string(anomaly.Type))
				}
			}
		}
	}

	// Process alerts if enabled
	if a.alertManager != nil {
		a.logger.Info("🚨 Evaluating alert rules")
		alertStart := time.Now()

		if err := a.alertManager.ProcessData(ctx, result, mlResults); err != nil {
			a.logger.Warn("Alert processing failed: %v", err)
			if a.metricsCollector != nil {
				a.metricsCollector.RecordError("alert_processing_failed")
			}
		} else {
			alertDuration := time.Since(alertStart)

			// Get alert status
			alertStatus := a.alertManager.GetStatus()
			a.logger.Info("✅ Alert evaluation complete: %d active alerts in %s",
				alertStatus.ActiveAlerts, alertDuration)

			if a.metricsCollector != nil {
				a.metricsCollector.RecordAnalysisProcessTime(alertDuration)
			}
		}
	}

	// Record final analysis duration
	totalDuration := time.Since(analysisStart)
	if a.metricsCollector != nil {
		a.metricsCollector.RecordAnalysisDuration(totalDuration)
	}

	a.logger.Info("✅ Enhanced analysis complete: %d clients, %d queries in %s",
		result.UniqueClients, result.TotalQueries, totalDuration)

	return result, nil
}

// Close releases resources used by the analyzer
func (a *EnhancedAnalyzer) Close() error {
	var err error

	// Close alert manager
	if a.alertManager != nil {
		if closeErr := a.alertManager.Close(); closeErr != nil {
			a.logger.Warn("Failed to close alert manager: %v", closeErr)
			err = closeErr
		}
	}

	// Close data source
	if a.dataSource != nil {
		if closeErr := a.dataSource.Close(); closeErr != nil {
			a.logger.Warn("Failed to close data source: %v", closeErr)
			err = closeErr
		}
	}

	return err
}

// enhanceWithNetworkAnalysis enhances client statistics with network device information
func (a *EnhancedAnalyzer) enhanceWithNetworkAnalysis(clientStats map[string]*types.ClientStats, networkDevices []types.NetworkDevice) {
	if len(networkDevices) == 0 {
		return
	}

	// Create a map of IP to network device for quick lookup
	deviceMap := make(map[string]*types.NetworkDevice)
	for i := range networkDevices {
		deviceMap[networkDevices[i].IP] = &networkDevices[i]
	}

	// Enhance client statistics with network device information
	for clientIP, stats := range clientStats {
		if device, exists := deviceMap[clientIP]; exists {
			stats.Hostname = device.Hostname
			stats.HWAddr = device.Hardware
			stats.IsOnline = device.IsOnline

			// Record top client metrics
			if a.metricsCollector != nil {
				a.metricsCollector.RecordTopClient(clientIP, device.Hostname, float64(stats.TotalQueries))
			}
		}
	}
}

// countActiveClients counts the number of active clients
func (a *EnhancedAnalyzer) countActiveClients(clientStats map[string]*types.ClientStats) int {
	activeCount := 0
	for _, stats := range clientStats {
		if stats.IsOnline {
			activeCount++
		}
	}
	return activeCount
}

// collectDetailedMetrics collects detailed metrics from queries and client statistics
func (a *EnhancedAnalyzer) collectDetailedMetrics(queries []types.PiholeRecord, clientStats map[string]*types.ClientStats) {
	if a.metricsCollector == nil {
		return
	}

	// Count queries by type and status
	queryTypeCount := make(map[string]int)
	statusCount := make(map[string]int)
	blockedCount := 0
	allowedCount := 0

	for _, query := range queries {
		// Record query type
		queryType := GetQueryTypeName(query.Status) // Note: This might need adjustment based on actual data structure
		queryTypeCount[queryType]++
		a.metricsCollector.RecordQueryByType(queryType)

		// Record query status
		statusName := GetStatusName(query.Status)
		statusCount[statusName]++
		a.metricsCollector.RecordQueryByStatus(statusName)

		// Count blocked vs allowed
		if query.Status == 1 || query.Status == 4 || query.Status == 5 || query.Status == 6 {
			// Blocked statuses
			blockedCount++
		} else if query.Status == 2 || query.Status == 3 {
			// Allowed statuses
			allowedCount++
		}
	}

	// Record domain metrics
	a.metricsCollector.RecordBlockedDomains(float64(blockedCount))
	a.metricsCollector.RecordAllowedDomains(float64(allowedCount))

	// Record top domains from client statistics
	domainCounts := make(map[string]int)
	for _, stats := range clientStats {
		for domain, count := range stats.Domains {
			domainCounts[domain] += count
		}
	}

	// Record top 10 domains
	topDomains := a.getTopDomains(domainCounts, 10)
	for _, domain := range topDomains {
		a.metricsCollector.RecordTopDomain(domain.Domain, float64(domain.Count))
	}
}

// getTopDomains returns the top N domains by query count
func (a *EnhancedAnalyzer) getTopDomains(domainCounts map[string]int, limit int) []types.DomainCount {
	// Convert map to slice for sorting
	domains := make([]types.DomainCount, 0, len(domainCounts))
	for domain, count := range domainCounts {
		domains = append(domains, types.DomainCount{
			Domain: domain,
			Count:  count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(domains)-1; i++ {
		for j := 0; j < len(domains)-i-1; j++ {
			if domains[j].Count < domains[j+1].Count {
				domains[j], domains[j+1] = domains[j+1], domains[j]
			}
		}
	}

	// Return top N domains
	if len(domains) > limit {
		domains = domains[:limit]
	}

	return domains
}

// getAnalysisMode returns the current analysis mode
func (a *EnhancedAnalyzer) getAnalysisMode() string {
	return "API-Only Analysis"
}

// AnalyzePiholeData performs API-based analysis
func AnalyzePiholeData(ctx context.Context, config *types.Config, appLogger *logger.Logger, metricsCollector *metrics.Collector) (*types.AnalysisResult, error) {
	// Create enhanced analyzer
	enhancedAnalyzer := NewEnhancedAnalyzer(config, appLogger, metricsCollector)

	// Initialize analyzer
	if err := enhancedAnalyzer.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced analyzer: %w", err)
	}
	defer enhancedAnalyzer.Close()

	// Perform analysis
	result, err := enhancedAnalyzer.AnalyzeData(ctx)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	return result, nil
}

// CreateDataSource creates a data source for Pi-hole analysis
func CreateDataSource(config *types.Config, appLogger *logger.Logger) (interfaces.DataSource, error) {
	factory := interfaces.NewDataSourceFactory(appLogger)

	// Create data source based on configuration
	dataSource, err := factory.CreateDataSource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	return dataSource, nil
}

// GetQueryTypeName returns human-readable query type name
func GetQueryTypeName(queryType int) string {
	switch queryType {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 12:
		return "PTR"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 28:
		return "AAAA"
	case 33:
		return "SRV"
	case 35:
		return "NAPTR"
	case 39:
		return "DNAME"
	case 41:
		return "OPT"
	case 43:
		return "DS"
	case 46:
		return "RRSIG"
	case 47:
		return "NSEC"
	case 48:
		return "DNSKEY"
	case 50:
		return "NSEC3"
	case 51:
		return "NSEC3PARAM"
	case 52:
		return "TLSA"
	case 257:
		return "CAA"
	default:
		return "Unknown"
	}
}

// GetStatusName returns human-readable status name
func GetStatusName(status int) string {
	switch status {
	case 0:
		return "Unknown"
	case 1:
		return "Blocked (gravity)"
	case 2:
		return "Forwarded"
	case 3:
		return "Cached"
	case 4:
		return "Blocked (regex/wildcard)"
	case 5:
		return "Blocked (exact)"
	case 6:
		return "Blocked (external)"
	case 7:
		return "CNAME"
	case 8:
		return "Retried"
	case 9:
		return "Retried but ignored"
	case 10:
		return "Already forwarded"
	case 11:
		return "Already cached"
	case 12:
		return "Config blocked"
	case 13:
		return "Gravity blocked"
	case 14:
		return "Regex blocked"
	default:
		return "Unknown"
	}
}

// GetStatusCodeFromString converts status strings to status codes
func GetStatusCodeFromString(status string) int {
	switch status {
	case "Blocked (gravity)":
		return 1
	case "Forwarded":
		return 2
	case "Cached":
		return 3
	case "Blocked (regex/wildcard)":
		return 4
	case "Blocked (exact)":
		return 5
	case "Blocked (external)":
		return 6
	case "CNAME":
		return 7
	case "Retried":
		return 8
	case "Retried but ignored":
		return 9
	case "Already forwarded":
		return 10
	case "Already cached":
		return 11
	case "Config blocked":
		return 12
	case "Gravity blocked":
		return 13
	case "Regex blocked":
		return 14
	default:
		return 0
	}
}

// convertToMLConfig converts types.MLConfig to ml.MLConfig
func (a *EnhancedAnalyzer) convertToMLConfig(config types.MLConfig) ml.MLConfig {
	// Parse duration strings
	var windowSize time.Duration
	if config.AnomalyDetection.WindowSize != "" {
		if d, err := time.ParseDuration(config.AnomalyDetection.WindowSize); err == nil {
			windowSize = d
		} else {
			windowSize = 24 * time.Hour // default
		}
	} else {
		windowSize = 24 * time.Hour // default
	}

	var analysisWindow, forecastWindow, retrainInterval, timeoutDuration, cacheDuration time.Duration

	if config.TrendAnalysis.AnalysisWindow != "" {
		if d, err := time.ParseDuration(config.TrendAnalysis.AnalysisWindow); err == nil {
			analysisWindow = d
		} else {
			analysisWindow = 1 * time.Hour // default
		}
	} else {
		analysisWindow = 1 * time.Hour // default
	}

	if config.TrendAnalysis.ForecastWindow != "" {
		if d, err := time.ParseDuration(config.TrendAnalysis.ForecastWindow); err == nil {
			forecastWindow = d
		} else {
			forecastWindow = 30 * time.Minute // default
		}
	} else {
		forecastWindow = 30 * time.Minute // default
	}

	if config.Training.RetrainInterval != "" {
		if d, err := time.ParseDuration(config.Training.RetrainInterval); err == nil {
			retrainInterval = d
		} else {
			retrainInterval = 24 * time.Hour // default
		}
	} else {
		retrainInterval = 24 * time.Hour // default
	}

	if config.Performance.TimeoutDuration != "" {
		if d, err := time.ParseDuration(config.Performance.TimeoutDuration); err == nil {
			timeoutDuration = d
		} else {
			timeoutDuration = 30 * time.Second // default
		}
	} else {
		timeoutDuration = 30 * time.Second // default
	}

	if config.Performance.CacheDuration != "" {
		if d, err := time.ParseDuration(config.Performance.CacheDuration); err == nil {
			cacheDuration = d
		} else {
			cacheDuration = 5 * time.Minute // default
		}
	} else {
		cacheDuration = 5 * time.Minute // default
	}

	// Convert anomaly types
	var anomalyTypes []ml.AnomalyType
	for _, typeStr := range config.AnomalyDetection.AnomalyTypes {
		switch typeStr {
		case "volume_spike":
			anomalyTypes = append(anomalyTypes, ml.AnomalyTypeVolumeSpike)
		case "unusual_domain":
			anomalyTypes = append(anomalyTypes, ml.AnomalyTypeUnusualDomain)
		case "unusual_client":
			anomalyTypes = append(anomalyTypes, ml.AnomalyTypeUnusualClient)
		case "time_pattern":
			anomalyTypes = append(anomalyTypes, ml.AnomalyTypeTimePattern)
		}
	}

	return ml.MLConfig{
		AnomalyDetection: ml.AnomalyDetectionConfig{
			Enabled:       config.AnomalyDetection.Enabled,
			Sensitivity:   config.AnomalyDetection.Sensitivity,
			MinConfidence: config.AnomalyDetection.MinConfidence,
			WindowSize:    windowSize,
			AnomalyTypes:  anomalyTypes,
			Thresholds:    config.AnomalyDetection.Thresholds,
		},
		TrendAnalysis: ml.TrendAnalysisConfig{
			Enabled:         config.TrendAnalysis.Enabled,
			AnalysisWindow:  analysisWindow,
			ForecastWindow:  forecastWindow,
			MinDataPoints:   config.TrendAnalysis.MinDataPoints,
			SmoothingFactor: config.TrendAnalysis.SmoothingFactor,
		},
		Training: ml.TrainingConfig{
			AutoRetrain:     config.Training.AutoRetrain,
			RetrainInterval: retrainInterval,
			MinTrainingSize: config.Training.MinTrainingSize,
			MaxTrainingSize: config.Training.MaxTrainingSize,
			ValidationSplit: config.Training.ValidationSplit,
		},
		Performance: ml.PerformanceConfig{
			MaxConcurrency:  config.Performance.MaxConcurrency,
			TimeoutDuration: timeoutDuration,
			CacheEnabled:    config.Performance.CacheEnabled,
			CacheDuration:   cacheDuration,
			BatchSize:       config.Performance.BatchSize,
		},
	}
}

// convertToAlertConfig converts types.AlertConfig to alerts.AlertConfig
func (a *EnhancedAnalyzer) convertToAlertConfig(config types.AlertConfig) alerts.AlertConfig {
	// Parse duration strings
	var defaultCooldown, alertRetention time.Duration

	if config.DefaultCooldown != "" {
		if d, err := time.ParseDuration(config.DefaultCooldown); err == nil {
			defaultCooldown = d
		} else {
			defaultCooldown = 5 * time.Minute // default
		}
	} else {
		defaultCooldown = 5 * time.Minute // default
	}

	if config.AlertRetention != "" {
		if d, err := time.ParseDuration(config.AlertRetention); err == nil {
			alertRetention = d
		} else {
			alertRetention = 7 * 24 * time.Hour // default
		}
	} else {
		alertRetention = 7 * 24 * time.Hour // default
	}

	// Convert severity
	var defaultSeverity alerts.AlertSeverity
	switch config.DefaultSeverity {
	case "info":
		defaultSeverity = alerts.SeverityInfo
	case "warning":
		defaultSeverity = alerts.SeverityWarning
	case "error":
		defaultSeverity = alerts.SeverityError
	case "critical":
		defaultSeverity = alerts.SeverityCritical
	default:
		defaultSeverity = alerts.SeverityWarning
	}

	// Convert notification config
	var slackTimeout, emailTimeout, notificationTimeout, evaluationInterval time.Duration
	var storageRetention time.Duration

	if config.Notifications.Slack.Timeout != "" {
		if d, err := time.ParseDuration(config.Notifications.Slack.Timeout); err == nil {
			slackTimeout = d
		} else {
			slackTimeout = 30 * time.Second
		}
	} else {
		slackTimeout = 30 * time.Second
	}

	if config.Notifications.Email.Timeout != "" {
		if d, err := time.ParseDuration(config.Notifications.Email.Timeout); err == nil {
			emailTimeout = d
		} else {
			emailTimeout = 30 * time.Second
		}
	} else {
		emailTimeout = 30 * time.Second
	}

	if config.Performance.NotificationTimeout != "" {
		if d, err := time.ParseDuration(config.Performance.NotificationTimeout); err == nil {
			notificationTimeout = d
		} else {
			notificationTimeout = 30 * time.Second
		}
	} else {
		notificationTimeout = 30 * time.Second
	}

	if config.Performance.EvaluationInterval != "" {
		if d, err := time.ParseDuration(config.Performance.EvaluationInterval); err == nil {
			evaluationInterval = d
		} else {
			evaluationInterval = 30 * time.Second
		}
	} else {
		evaluationInterval = 30 * time.Second
	}

	if config.Storage.Retention != "" {
		if d, err := time.ParseDuration(config.Storage.Retention); err == nil {
			storageRetention = d
		} else {
			storageRetention = 24 * time.Hour
		}
	} else {
		storageRetention = 24 * time.Hour
	}

	// Convert rules
	var rules []alerts.AlertRule
	for _, rule := range config.Rules {
		var cooldown time.Duration
		if rule.CooldownPeriod != "" {
			if d, err := time.ParseDuration(rule.CooldownPeriod); err == nil {
				cooldown = d
			} else {
				cooldown = 5 * time.Minute
			}
		} else {
			cooldown = 5 * time.Minute
		}

		var severity alerts.AlertSeverity
		switch rule.Severity {
		case "info":
			severity = alerts.SeverityInfo
		case "warning":
			severity = alerts.SeverityWarning
		case "error":
			severity = alerts.SeverityError
		case "critical":
			severity = alerts.SeverityCritical
		default:
			severity = alerts.SeverityWarning
		}

		// Convert conditions
		var conditions []alerts.AlertCondition
		for _, condition := range rule.Conditions {
			conditions = append(conditions, alerts.AlertCondition{
				Field:      condition.Field,
				Operator:   condition.Operator,
				Value:      condition.Value,
				TimeWindow: condition.TimeWindow,
			})
		}

		// Convert alert type
		var alertType alerts.AlertType
		switch rule.Type {
		case "anomaly":
			alertType = alerts.AlertTypeAnomaly
		case "threshold":
			alertType = alerts.AlertTypeThreshold
		case "connectivity":
			alertType = alerts.AlertTypeConnectivity
		case "performance":
			alertType = alerts.AlertTypePerformance
		case "security":
			alertType = alerts.AlertTypeSecurity
		case "configuration":
			alertType = alerts.AlertTypeConfiguration
		default:
			alertType = alerts.AlertTypeThreshold
		}

		// Convert channels
		var channels []alerts.NotificationChannel
		for _, channel := range rule.Channels {
			switch channel {
			case "log":
				channels = append(channels, alerts.ChannelLog)
			case "slack":
				channels = append(channels, alerts.ChannelSlack)
			case "email":
				channels = append(channels, alerts.ChannelEmail)
			}
		}

		rules = append(rules, alerts.AlertRule{
			ID:             rule.ID,
			Name:           rule.Name,
			Description:    rule.Description,
			Enabled:        rule.Enabled,
			Type:           alertType,
			Severity:       severity,
			Conditions:     conditions,
			CooldownPeriod: cooldown,
			MaxAlerts:      rule.MaxAlerts,
			Channels:       channels,
			Tags:           rule.Tags,
		})
	}

	return alerts.AlertConfig{
		Enabled:         config.Enabled,
		Rules:           rules,
		DefaultSeverity: defaultSeverity,
		DefaultCooldown: defaultCooldown,
		MaxActiveAlerts: config.MaxActiveAlerts,
		AlertRetention:  alertRetention,
		Notifications: alerts.NotificationConfig{
			Slack: alerts.SlackConfig{
				Enabled:    config.Notifications.Slack.Enabled,
				WebhookURL: config.Notifications.Slack.WebhookURL,
				Channel:    config.Notifications.Slack.Channel,
				Username:   config.Notifications.Slack.Username,
				IconEmoji:  config.Notifications.Slack.IconEmoji,
				Timeout:    slackTimeout,
			},
			Email: alerts.EmailConfig{
				Enabled:    config.Notifications.Email.Enabled,
				SMTPHost:   config.Notifications.Email.SMTPHost,
				SMTPPort:   config.Notifications.Email.SMTPPort,
				Username:   config.Notifications.Email.Username,
				Password:   config.Notifications.Email.Password,
				From:       config.Notifications.Email.From,
				Recipients: config.Notifications.Email.Recipients,
				UseTLS:     config.Notifications.Email.UseTLS,
				Timeout:    emailTimeout,
			},
		},
		Storage: alerts.StorageConfig{
			Type:      config.Storage.Type,
			Path:      config.Storage.Path,
			MaxSize:   config.Storage.MaxSize,
			Retention: storageRetention,
		},
		Performance: alerts.AlertPerformanceConfig{
			MaxConcurrentNotifications: config.Performance.MaxConcurrentNotifications,
			NotificationTimeout:        notificationTimeout,
			BatchSize:                  config.Performance.BatchSize,
			EvaluationInterval:         evaluationInterval,
		},
	}
}
