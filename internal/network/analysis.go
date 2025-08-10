package network

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// EnhancedNetworkAnalyzer implements the NetworkAnalyzer interface
type EnhancedNetworkAnalyzer struct {
	config      types.NetworkAnalysisConfig
	logger      *slog.Logger
	dpi         DeepPacketInspector
	traffic     TrafficPatternAnalyzer
	security    SecurityAnalyzer
	performance PerformanceAnalyzer
	visualizer  NetworkVisualizer
	initialized bool
}

// NewEnhancedNetworkAnalyzer creates a new enhanced network analyzer
func NewEnhancedNetworkAnalyzer(loggerInstance *slog.Logger) *EnhancedNetworkAnalyzer {
	if loggerInstance == nil {
		customLogger := logger.New(&logger.Config{Component: "network-analyzer"})
		loggerInstance = customLogger.GetSlogger() // Get underlying slog.Logger
	}

	return &EnhancedNetworkAnalyzer{
		logger: loggerInstance,
	}
}

// Initialize implements NetworkAnalyzer.Initialize
func (e *EnhancedNetworkAnalyzer) Initialize(ctx context.Context, config types.NetworkAnalysisConfig) error {
	e.logger.Info("Initializing enhanced network analyzer", 
		slog.Bool("enabled", config.Enabled),
		slog.Bool("dpi_enabled", config.DeepPacketInspection.Enabled),
		slog.Bool("traffic_patterns_enabled", config.TrafficPatterns.Enabled),
		slog.Bool("security_enabled", config.SecurityAnalysis.Enabled),
		slog.Bool("performance_enabled", config.Performance.Enabled))

	e.config = config

	// Initialize component analyzers
	if config.DeepPacketInspection.Enabled {
		e.dpi = NewDeepPacketInspector(e.logger)
	}

	if config.TrafficPatterns.Enabled {
		e.traffic = NewTrafficPatternAnalyzer(e.logger)
	}

	if config.SecurityAnalysis.Enabled {
		e.security = NewSecurityAnalyzer(e.logger)
	}

	if config.Performance.Enabled {
		e.performance = NewPerformanceAnalyzer(e.logger)
	}

	e.visualizer = NewNetworkVisualizer(e.logger)
	e.initialized = true

	e.logger.Info("Enhanced network analyzer initialized successfully")
	return nil
}

// AnalyzeTraffic implements NetworkAnalyzer.AnalyzeTraffic
func (e *EnhancedNetworkAnalyzer) AnalyzeTraffic(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (*types.NetworkAnalysisResult, error) {
	if !e.initialized {
		return nil, fmt.Errorf("analyzer not initialized")
	}

	e.logger.Info("Starting comprehensive network traffic analysis",
		slog.Int("record_count", len(records)),
		slog.Int("client_count", len(clientStats)))

	startTime := time.Now()
	analysisID := fmt.Sprintf("analysis_%d", startTime.Unix())

	result := &types.NetworkAnalysisResult{
		Timestamp:   startTime.Format(time.RFC3339),
		AnalysisID:  analysisID,
		Duration:    "",
		Summary: &types.NetworkAnalysisSummary{
			TotalClients:  len(clientStats),
			TotalQueries:  int64(len(records)),
			KeyInsights:   make([]string, 0),
			HealthScore:   100.0,
			OverallHealth: "GOOD",
			ThreatLevel:   "LOW",
		},
	}

	// Deep Packet Inspection
	if e.config.DeepPacketInspection.Enabled && e.dpi != nil {
		e.logger.Info("Performing deep packet inspection")
		packetResult, err := e.dpi.InspectPackets(ctx, records, e.config.DeepPacketInspection)
		if err != nil {
			e.logger.Error("Deep packet inspection failed", slog.String("error", err.Error()))
		} else {
			result.PacketAnalysis = packetResult
			e.logger.Info("Deep packet inspection completed",
				slog.Int64("analyzed_packets", packetResult.AnalyzedPackets),
				slog.Int("anomalies", len(packetResult.Anomalies)))
		}
	}

	// Traffic Pattern Analysis
	if e.config.TrafficPatterns.Enabled && e.traffic != nil {
		e.logger.Info("Analyzing traffic patterns")
		trafficResult, err := e.traffic.AnalyzePatterns(ctx, records, clientStats, e.config.TrafficPatterns)
		if err != nil {
			e.logger.Error("Traffic pattern analysis failed", slog.String("error", err.Error()))
		} else {
			result.TrafficPatterns = trafficResult
			e.logger.Info("Traffic pattern analysis completed",
				slog.Int("patterns_detected", len(trafficResult.DetectedPatterns)),
				slog.Int("anomalies", len(trafficResult.Anomalies)))
		}
	}

	// Security Analysis
	if e.config.SecurityAnalysis.Enabled && e.security != nil {
		e.logger.Info("Performing security analysis")
		securityResult, err := e.security.AnalyzeSecurity(ctx, records, clientStats, e.config.SecurityAnalysis)
		if err != nil {
			e.logger.Error("Security analysis failed", slog.String("error", err.Error()))
		} else {
			result.SecurityAnalysis = securityResult
			result.Summary.ThreatLevel = securityResult.ThreatLevel
			e.logger.Info("Security analysis completed",
				slog.String("threat_level", securityResult.ThreatLevel),
				slog.Int("threats_detected", len(securityResult.DetectedThreats)))
		}
	}

	// Performance Analysis
	if e.config.Performance.Enabled && e.performance != nil {
		e.logger.Info("Analyzing network performance")
		performanceResult, err := e.performance.AnalyzePerformance(ctx, records, clientStats, e.config.Performance)
		if err != nil {
			e.logger.Error("Performance analysis failed", slog.String("error", err.Error()))
		} else {
			result.Performance = performanceResult
			result.Summary.HealthScore = performanceResult.OverallScore
			e.logger.Info("Performance analysis completed",
				slog.Float64("overall_score", performanceResult.OverallScore))
		}
	}

	// Calculate analysis duration
	duration := time.Since(startTime)
	result.Duration = duration.String()

	// Update summary with insights
	e.generateSummaryInsights(result)

	e.logger.Info("Network traffic analysis completed",
		slog.String("analysis_id", analysisID),
		slog.String("duration", duration.String()),
		slog.String("health", result.Summary.OverallHealth),
		slog.Float64("health_score", result.Summary.HealthScore))

	return result, nil
}

// GetCapabilities implements NetworkAnalyzer.GetCapabilities
func (e *EnhancedNetworkAnalyzer) GetCapabilities() []string {
	capabilities := []string{"basic_analysis"}

	if e.config.DeepPacketInspection.Enabled {
		capabilities = append(capabilities, "deep_packet_inspection", "protocol_analysis", "packet_anomaly_detection")
	}

	if e.config.TrafficPatterns.Enabled {
		capabilities = append(capabilities, "traffic_patterns", "bandwidth_analysis", "temporal_patterns", "client_behavior")
	}

	if e.config.SecurityAnalysis.Enabled {
		capabilities = append(capabilities, "security_analysis", "threat_detection", "dns_anomalies", "port_scan_detection")
	}

	if e.config.Performance.Enabled {
		capabilities = append(capabilities, "performance_analysis", "latency_analysis", "throughput_analysis", "quality_assessment")
	}

	return capabilities
}

// IsHealthy implements NetworkAnalyzer.IsHealthy
func (e *EnhancedNetworkAnalyzer) IsHealthy() bool {
	return e.initialized
}

// generateSummaryInsights generates insights for the analysis summary
func (e *EnhancedNetworkAnalyzer) generateSummaryInsights(result *types.NetworkAnalysisResult) {
	insights := []string{}

	// Packet analysis insights
	if result.PacketAnalysis != nil {
		if len(result.PacketAnalysis.Anomalies) > 0 {
			insights = append(insights, fmt.Sprintf("Detected %d packet anomalies requiring attention", len(result.PacketAnalysis.Anomalies)))
		}
		if result.PacketAnalysis.AnalyzedPackets > 1000000 {
			insights = append(insights, "High traffic volume detected - network experiencing heavy load")
		}
	}

	// Traffic pattern insights
	if result.TrafficPatterns != nil {
		if len(result.TrafficPatterns.Anomalies) > 0 {
			insights = append(insights, fmt.Sprintf("Found %d traffic pattern anomalies", len(result.TrafficPatterns.Anomalies)))
		}
		if len(result.TrafficPatterns.DetectedPatterns) > 5 {
			insights = append(insights, "Multiple traffic patterns detected - network shows complex usage patterns")
		}
	}

	// Security insights
	if result.SecurityAnalysis != nil {
		threatCount := len(result.SecurityAnalysis.DetectedThreats)
		if threatCount > 0 {
			insights = append(insights, fmt.Sprintf("Security analysis found %d potential threats", threatCount))
		}
		if result.SecurityAnalysis.ThreatLevel == "HIGH" || result.SecurityAnalysis.ThreatLevel == "CRITICAL" {
			insights = append(insights, "Elevated threat level detected - immediate attention recommended")
		}
	}

	// Performance insights
	if result.Performance != nil {
		if result.Performance.OverallScore < 70 {
			insights = append(insights, "Network performance below optimal - consider investigation")
		}
		if result.Performance.LatencyMetrics.AvgLatency > 100 {
			insights = append(insights, "High latency detected - network responsiveness may be impacted")
		}
	}

	// Overall health assessment
	if result.Summary.HealthScore >= 90 {
		result.Summary.OverallHealth = "EXCELLENT"
	} else if result.Summary.HealthScore >= 80 {
		result.Summary.OverallHealth = "GOOD"
	} else if result.Summary.HealthScore >= 70 {
		result.Summary.OverallHealth = "FAIR"
	} else if result.Summary.HealthScore >= 60 {
		result.Summary.OverallHealth = "POOR"
	} else {
		result.Summary.OverallHealth = "CRITICAL"
	}

	// Count active clients
	activeCount := 0
	for _, anomaly := range result.TrafficPatterns.Anomalies {
		activeCount += len(anomaly.Affected)
	}
	result.Summary.ActiveClients = activeCount

	// Set anomaly count
	anomalyCount := 0
	if result.PacketAnalysis != nil {
		anomalyCount += len(result.PacketAnalysis.Anomalies)
	}
	if result.TrafficPatterns != nil {
		anomalyCount += len(result.TrafficPatterns.Anomalies)
	}
	if result.SecurityAnalysis != nil {
		anomalyCount += len(result.SecurityAnalysis.DetectedThreats)
	}
	result.Summary.AnomaliesDetected = anomalyCount

	if len(insights) == 0 {
		insights = append(insights, "Network analysis completed successfully - no significant issues detected")
	}

	result.Summary.KeyInsights = insights
}

// Factory implementation

// DefaultAnalyzerFactory implements AnalyzerFactory
type DefaultAnalyzerFactory struct {
	logger *slog.Logger
}

// NewAnalyzerFactory creates a new analyzer factory
func NewAnalyzerFactory(loggerInstance *slog.Logger) AnalyzerFactory {
	if loggerInstance == nil {
		customLogger := logger.New(&logger.Config{Component: "analyzer-factory"})
		loggerInstance = customLogger.GetSlogger() // Get underlying slog.Logger
	}
	return &DefaultAnalyzerFactory{logger: loggerInstance}
}

// CreateNetworkAnalyzer implements AnalyzerFactory.CreateNetworkAnalyzer
func (f *DefaultAnalyzerFactory) CreateNetworkAnalyzer(config types.NetworkAnalysisConfig) (NetworkAnalyzer, error) {
	analyzer := NewEnhancedNetworkAnalyzer(f.logger)
	err := analyzer.Initialize(context.Background(), config)
	return analyzer, err
}

// CreateDPIAnalyzer implements AnalyzerFactory.CreateDPIAnalyzer
func (f *DefaultAnalyzerFactory) CreateDPIAnalyzer(config types.DPIConfig) (DeepPacketInspector, error) {
	return NewDeepPacketInspector(f.logger), nil
}

// CreateTrafficAnalyzer implements AnalyzerFactory.CreateTrafficAnalyzer
func (f *DefaultAnalyzerFactory) CreateTrafficAnalyzer(config types.TrafficPatternsConfig) (TrafficPatternAnalyzer, error) {
	return NewTrafficPatternAnalyzer(f.logger), nil
}

// CreateSecurityAnalyzer implements AnalyzerFactory.CreateSecurityAnalyzer
func (f *DefaultAnalyzerFactory) CreateSecurityAnalyzer(config types.SecurityAnalysisConfig) (SecurityAnalyzer, error) {
	return NewSecurityAnalyzer(f.logger), nil
}

// CreatePerformanceAnalyzer implements AnalyzerFactory.CreatePerformanceAnalyzer
func (f *DefaultAnalyzerFactory) CreatePerformanceAnalyzer(config types.NetworkPerformanceConfig) (PerformanceAnalyzer, error) {
	return NewPerformanceAnalyzer(f.logger), nil
}

// CreateVisualizer implements AnalyzerFactory.CreateVisualizer
func (f *DefaultAnalyzerFactory) CreateVisualizer() (NetworkVisualizer, error) {
	return NewNetworkVisualizer(f.logger), nil
}

// Utility functions for network analysis

// parseTimestamp safely parses a timestamp string
func parseTimestamp(timestamp string) time.Time {
	if timestamp == "" {
		return time.Now()
	}

	// Try different timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"1136239445", // Unix timestamp
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t
		}
	}

	// Try parsing as Unix timestamp
	if unix, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		return time.Unix(unix, 0)
	}

	return time.Now()
}

// calculatePercentile calculates the nth percentile of a slice of float64 values
func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := percentile/100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// normalizeScore ensures a score is between 0 and 1
func normalizeScore(score float64) float64 {
	return math.Max(0, math.Min(1, score))
}

// calculateZScore calculates the z-score for a value given mean and standard deviation
func calculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// isIPv4 checks if a string is a valid IPv4 address
func isIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}

// isPrivateIP checks if an IP address is in a private range
func isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.", "172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"127.",
		"169.254.",
	}

	for _, prefix := range privateRanges {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}

	return false
}

// generateAnalysisID generates a unique analysis ID
func generateAnalysisID() string {
	return fmt.Sprintf("na_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// timeToHour converts a time to the hour of day (0-23)
func timeToHour(t time.Time) int {
	return t.Hour()
}

// timeToDay converts a time to day of week (0-6, Sunday=0)
func timeToDay(t time.Time) int {
	return int(t.Weekday())
}

// calculateStandardDeviation calculates standard deviation of float64 values
func calculateStandardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values) - 1)

	return math.Sqrt(variance)
}

// mapKeys returns the keys of a map[string]int64
func mapKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// maxInt64 returns the maximum of two int64 values
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// minFloat64 returns the minimum of two float64 values
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat64 returns the maximum of two float64 values
func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}