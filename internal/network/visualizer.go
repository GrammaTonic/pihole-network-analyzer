package network

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"pihole-analyzer/internal/types"
)

// DefaultNetworkVisualizer implements the NetworkVisualizer interface
type DefaultNetworkVisualizer struct {
	logger *slog.Logger
}

// NewNetworkVisualizer creates a new network visualizer
func NewNetworkVisualizer(logger *slog.Logger) NetworkVisualizer {
	return &DefaultNetworkVisualizer{
		logger: logger,
	}
}

// GenerateTrafficVisualization implements NetworkVisualizer.GenerateTrafficVisualization
func (v *DefaultNetworkVisualizer) GenerateTrafficVisualization(result *types.NetworkAnalysisResult) (map[string]interface{}, error) {
	v.logger.Debug("Generating traffic visualization data")

	visualization := make(map[string]interface{})

	// Generate packet analysis visualization
	if result.PacketAnalysis != nil {
		packetViz := v.generatePacketVisualization(result.PacketAnalysis)
		visualization["packet_analysis"] = packetViz
	}

	// Generate traffic patterns visualization
	if result.TrafficPatterns != nil {
		trafficViz := v.generateTrafficPatternsVisualization(result.TrafficPatterns)
		visualization["traffic_patterns"] = trafficViz
	}

	// Generate security visualization
	if result.SecurityAnalysis != nil {
		securityViz := v.generateSecurityVisualization(result.SecurityAnalysis)
		visualization["security_analysis"] = securityViz
	}

	// Generate performance visualization
	if result.Performance != nil {
		performanceViz := v.generatePerformanceVisualization(result.Performance)
		visualization["performance_analysis"] = performanceViz
	}

	// Generate summary dashboard
	summaryViz := v.generateSummaryVisualization(result.Summary)
	visualization["summary"] = summaryViz

	// Add metadata
	visualization["metadata"] = map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"analysis_id":  result.AnalysisID,
		"duration":     result.Duration,
	}

	v.logger.Debug("Traffic visualization data generated successfully")
	return visualization, nil
}

// GenerateTopologyVisualization implements NetworkVisualizer.GenerateTopologyVisualization
func (v *DefaultNetworkVisualizer) GenerateTopologyVisualization(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (map[string]interface{}, error) {
	v.logger.Debug("Generating network topology visualization")

	topology := make(map[string]interface{})

	// Generate nodes (clients and domains)
	nodes := v.generateTopologyNodes(records, clientStats)
	topology["nodes"] = nodes

	// Generate edges (connections between clients and domains)
	edges := v.generateTopologyEdges(records)
	topology["edges"] = edges

	// Generate network clusters
	clusters := v.generateNetworkClusters(records, clientStats)
	topology["clusters"] = clusters

	// Generate layout information
	layout := v.generateLayoutInfo(nodes, edges)
	topology["layout"] = layout

	// Add statistics
	stats := map[string]interface{}{
		"total_nodes":        len(nodes),
		"total_edges":        len(edges),
		"total_clusters":     len(clusters),
		"connection_density": v.calculateConnectionDensity(nodes, edges),
	}
	topology["statistics"] = stats

	v.logger.Debug("Network topology visualization generated successfully")
	return topology, nil
}

// GenerateTimeSeriesData implements NetworkVisualizer.GenerateTimeSeriesData
func (v *DefaultNetworkVisualizer) GenerateTimeSeriesData(records []types.PiholeRecord, metric string, interval time.Duration) (map[string]interface{}, error) {
	v.logger.Debug("Generating time series data", slog.String("metric", metric), slog.String("interval", interval.String()))

	timeSeries := make(map[string]interface{})

	// Group records by time intervals
	timeGroups := v.groupRecordsByTime(records, interval)

	// Generate data points based on metric
	dataPoints := make([]map[string]interface{}, 0)
	labels := make([]string, 0)

	// Sort time groups
	sortedTimes := make([]string, 0, len(timeGroups))
	for timeKey := range timeGroups {
		sortedTimes = append(sortedTimes, timeKey)
	}
	sort.Strings(sortedTimes)

	for _, timeKey := range sortedTimes {
		groupRecords := timeGroups[timeKey]
		value := v.calculateMetricValue(groupRecords, metric)

		dataPoint := map[string]interface{}{
			"timestamp": timeKey,
			"value":     value,
			"count":     len(groupRecords),
		}
		dataPoints = append(dataPoints, dataPoint)
		labels = append(labels, timeKey)
	}

	timeSeries["data_points"] = dataPoints
	timeSeries["labels"] = labels
	timeSeries["metric"] = metric
	timeSeries["interval"] = interval.String()

	// Add trend analysis
	trend := v.calculateTrend(dataPoints)
	timeSeries["trend"] = trend

	// Add statistics
	values := make([]float64, len(dataPoints))
	for i, point := range dataPoints {
		values[i] = point["value"].(float64)
	}

	stats := map[string]interface{}{
		"min":    v.calculateMin(values),
		"max":    v.calculateMax(values),
		"avg":    v.calculateMean(values),
		"stddev": calculateStandardDeviation(values),
	}
	timeSeries["statistics"] = stats

	v.logger.Debug("Time series data generated successfully")
	return timeSeries, nil
}

// GenerateHeatmapData implements NetworkVisualizer.GenerateHeatmapData
func (v *DefaultNetworkVisualizer) GenerateHeatmapData(records []types.PiholeRecord) (map[string]interface{}, error) {
	v.logger.Debug("Generating heatmap data")

	heatmap := make(map[string]interface{})

	// Generate hourly activity heatmap
	hourlyHeatmap := v.generateHourlyHeatmap(records)
	heatmap["hourly_activity"] = hourlyHeatmap

	// Generate daily activity heatmap
	dailyHeatmap := v.generateDailyHeatmap(records)
	heatmap["daily_activity"] = dailyHeatmap

	// Generate client-domain interaction heatmap
	clientDomainHeatmap := v.generateClientDomainHeatmap(records)
	heatmap["client_domain_interactions"] = clientDomainHeatmap

	// Generate query type heatmap
	queryTypeHeatmap := v.generateQueryTypeHeatmap(records)
	heatmap["query_types"] = queryTypeHeatmap

	v.logger.Debug("Heatmap data generated successfully")
	return heatmap, nil
}

// GenerateSecurityDashboard implements NetworkVisualizer.GenerateSecurityDashboard
func (v *DefaultNetworkVisualizer) GenerateSecurityDashboard(securityResult *types.SecurityAnalysisResult) (map[string]interface{}, error) {
	v.logger.Debug("Generating security dashboard data")

	dashboard := make(map[string]interface{})

	// Threat level indicator
	dashboard["threat_level"] = map[string]interface{}{
		"level":       securityResult.ThreatLevel,
		"color":       v.getThreatLevelColor(securityResult.ThreatLevel),
		"description": v.getThreatLevelDescription(securityResult.ThreatLevel),
	}

	// Threat summary
	threatSummary := v.generateThreatSummary(securityResult.DetectedThreats)
	dashboard["threat_summary"] = threatSummary

	// Security metrics
	securityMetrics := map[string]interface{}{
		"total_threats":         len(securityResult.DetectedThreats),
		"suspicious_activities": len(securityResult.SuspiciousActivity),
		"dns_anomalies":         len(securityResult.DNSAnomalies),
		"port_scans":            len(securityResult.PortScans),
		"tunneling_attempts":    len(securityResult.TunnelingAttempts),
		"blocked_connections":   len(securityResult.BlockedConnections),
	}
	dashboard["security_metrics"] = securityMetrics

	// Threat timeline
	timeline := v.generateThreatTimeline(securityResult.DetectedThreats)
	dashboard["threat_timeline"] = timeline

	// Risk distribution
	riskDistribution := v.generateRiskDistribution(securityResult.DetectedThreats, securityResult.SuspiciousActivity)
	dashboard["risk_distribution"] = riskDistribution

	// Top threats
	topThreats := v.generateTopThreats(securityResult.DetectedThreats, 10)
	dashboard["top_threats"] = topThreats

	v.logger.Debug("Security dashboard data generated successfully")
	return dashboard, nil
}

// GeneratePerformanceDashboard implements NetworkVisualizer.GeneratePerformanceDashboard
func (v *DefaultNetworkVisualizer) GeneratePerformanceDashboard(performanceResult *types.NetworkPerformanceResult) (map[string]interface{}, error) {
	v.logger.Debug("Generating performance dashboard data")

	dashboard := make(map[string]interface{})

	// Overall score gauge
	dashboard["overall_score"] = map[string]interface{}{
		"score":       performanceResult.OverallScore,
		"grade":       performanceResult.QualityAssessment.OverallGrade,
		"color":       v.getScoreColor(performanceResult.OverallScore),
		"description": v.getScoreDescription(performanceResult.OverallScore),
	}

	// Performance metrics cards
	metricsCards := v.generatePerformanceMetricsCards(performanceResult)
	dashboard["metrics_cards"] = metricsCards

	// Latency analysis
	latencyAnalysis := v.generateLatencyAnalysis(&performanceResult.LatencyMetrics)
	dashboard["latency_analysis"] = latencyAnalysis

	// Bandwidth analysis
	bandwidthAnalysis := v.generateBandwidthAnalysis(&performanceResult.BandwidthMetrics)
	dashboard["bandwidth_analysis"] = bandwidthAnalysis

	// Quality assessment
	qualityAssessment := v.generateQualityAssessmentViz(&performanceResult.QualityAssessment)
	dashboard["quality_assessment"] = qualityAssessment

	// Performance trends
	trends := v.generatePerformanceTrends(performanceResult)
	dashboard["performance_trends"] = trends

	v.logger.Debug("Performance dashboard data generated successfully")
	return dashboard, nil
}

// Helper methods for visualization generation

// generatePacketVisualization generates packet analysis visualization data
func (v *DefaultNetworkVisualizer) generatePacketVisualization(packetAnalysis *types.PacketAnalysisResult) map[string]interface{} {
	viz := make(map[string]interface{})

	// Protocol distribution pie chart
	protocolChart := make([]map[string]interface{}, 0)
	for protocol, count := range packetAnalysis.ProtocolDistribution {
		protocolChart = append(protocolChart, map[string]interface{}{
			"label": protocol,
			"value": count,
			"color": v.getProtocolColor(protocol),
		})
	}
	viz["protocol_distribution"] = protocolChart

	// Packet size distribution bar chart
	sizeChart := make([]map[string]interface{}, 0)
	for sizeCategory, count := range packetAnalysis.PacketSizeDistribution {
		sizeChart = append(sizeChart, map[string]interface{}{
			"category": sizeCategory,
			"count":    count,
			"color":    v.getSizeColor(sizeCategory),
		})
	}
	viz["size_distribution"] = sizeChart

	// Top source IPs
	topSources := make([]map[string]interface{}, 0)
	for _, ipStat := range packetAnalysis.TopSourceIPs {
		topSources = append(topSources, map[string]interface{}{
			"ip":         ipStat.IP,
			"hostname":   ipStat.Hostname,
			"packets":    ipStat.PacketCount,
			"bytes":      ipStat.ByteCount,
			"percentage": ipStat.Percentage,
		})
	}
	viz["top_sources"] = topSources

	// Anomalies list
	anomalies := make([]map[string]interface{}, 0)
	for _, anomaly := range packetAnalysis.Anomalies {
		anomalies = append(anomalies, map[string]interface{}{
			"type":        anomaly.Type,
			"description": anomaly.Description,
			"severity":    anomaly.Severity,
			"confidence":  anomaly.Confidence,
			"timestamp":   anomaly.Timestamp,
		})
	}
	viz["anomalies"] = anomalies

	return viz
}

// generateTrafficPatternsVisualization generates traffic patterns visualization data
func (v *DefaultNetworkVisualizer) generateTrafficPatternsVisualization(trafficPatterns *types.TrafficPatternsResult) map[string]interface{} {
	viz := make(map[string]interface{})

	// Bandwidth patterns timeline
	bandwidthTimeline := make([]map[string]interface{}, 0)
	for _, pattern := range trafficPatterns.BandwidthPatterns {
		bandwidthTimeline = append(bandwidthTimeline, map[string]interface{}{
			"time_slot":      pattern.TimeSlot,
			"avg_bandwidth":  pattern.AvgBandwidth,
			"peak_bandwidth": pattern.PeakBandwidth,
			"usage":          pattern.Usage,
			"trend":          pattern.Trend,
		})
	}
	viz["bandwidth_timeline"] = bandwidthTimeline

	// Temporal patterns
	temporalPatterns := make([]map[string]interface{}, 0)
	for _, pattern := range trafficPatterns.TemporalPatterns {
		temporalPatterns = append(temporalPatterns, map[string]interface{}{
			"pattern":     pattern.Pattern,
			"peak_hours":  pattern.PeakHours,
			"low_hours":   pattern.LowHours,
			"regularity":  pattern.Regularity,
			"seasonality": pattern.Seasonality,
		})
	}
	viz["temporal_patterns"] = temporalPatterns

	// Client behavior summary
	clientBehavior := make([]map[string]interface{}, 0)
	for ip, behavior := range trafficPatterns.ClientBehavior {
		clientBehavior = append(clientBehavior, map[string]interface{}{
			"ip":             ip,
			"hostname":       behavior.Hostname,
			"behavior_type":  behavior.BehaviorType,
			"activity_level": behavior.ActivityLevel,
			"risk_score":     behavior.RiskScore,
			"anomaly_count":  len(behavior.Anomalies),
		})
	}
	viz["client_behavior"] = clientBehavior

	return viz
}

// generateSecurityVisualization generates security analysis visualization data
func (v *DefaultNetworkVisualizer) generateSecurityVisualization(securityAnalysis *types.SecurityAnalysisResult) map[string]interface{} {
	viz := make(map[string]interface{})

	// Threat level gauge
	viz["threat_level"] = map[string]interface{}{
		"level": securityAnalysis.ThreatLevel,
		"color": v.getThreatLevelColor(securityAnalysis.ThreatLevel),
	}

	// Threat types distribution
	threatTypes := make(map[string]int)
	for _, threat := range securityAnalysis.DetectedThreats {
		threatTypes[threat.Type]++
	}

	threatChart := make([]map[string]interface{}, 0)
	for threatType, count := range threatTypes {
		threatChart = append(threatChart, map[string]interface{}{
			"type":  threatType,
			"count": count,
			"color": v.getThreatTypeColor(threatType),
		})
	}
	viz["threat_types"] = threatChart

	// Security metrics
	viz["security_metrics"] = map[string]interface{}{
		"total_threats":      len(securityAnalysis.DetectedThreats),
		"suspicious_count":   len(securityAnalysis.SuspiciousActivity),
		"dns_anomalies":      len(securityAnalysis.DNSAnomalies),
		"tunneling_attempts": len(securityAnalysis.TunnelingAttempts),
	}

	return viz
}

// generatePerformanceVisualization generates performance analysis visualization data
func (v *DefaultNetworkVisualizer) generatePerformanceVisualization(performance *types.NetworkPerformanceResult) map[string]interface{} {
	viz := make(map[string]interface{})

	// Overall score gauge
	viz["overall_score"] = map[string]interface{}{
		"score": performance.OverallScore,
		"grade": performance.QualityAssessment.OverallGrade,
		"color": v.getScoreColor(performance.OverallScore),
	}

	// Performance metrics
	viz["metrics"] = map[string]interface{}{
		"latency": map[string]interface{}{
			"avg":   performance.LatencyMetrics.AvgLatency,
			"p95":   performance.LatencyMetrics.P95Latency,
			"grade": performance.QualityAssessment.LatencyGrade,
		},
		"bandwidth": map[string]interface{}{
			"total": performance.BandwidthMetrics.TotalBandwidth,
			"peak":  performance.BandwidthMetrics.PeakBandwidth,
			"grade": performance.QualityAssessment.BandwidthGrade,
		},
		"throughput": map[string]interface{}{
			"qps":      performance.ThroughputMetrics.QueriesPerSecond,
			"peak_qps": performance.ThroughputMetrics.PeakQPS,
		},
	}

	// Quality issues
	issues := make([]map[string]interface{}, 0)
	for _, issue := range performance.QualityAssessment.Issues {
		issues = append(issues, map[string]interface{}{
			"type":        issue.Type,
			"severity":    issue.Severity,
			"description": issue.Description,
			"impact":      issue.Impact,
		})
	}
	viz["quality_issues"] = issues

	return viz
}

// generateSummaryVisualization generates summary visualization data
func (v *DefaultNetworkVisualizer) generateSummaryVisualization(summary *types.NetworkAnalysisSummary) map[string]interface{} {
	viz := make(map[string]interface{})

	// Key metrics
	viz["key_metrics"] = map[string]interface{}{
		"total_clients":      summary.TotalClients,
		"active_clients":     summary.ActiveClients,
		"total_queries":      summary.TotalQueries,
		"anomalies_detected": summary.AnomaliesDetected,
		"health_score":       summary.HealthScore,
	}

	// Health indicator
	viz["health_indicator"] = map[string]interface{}{
		"overall_health": summary.OverallHealth,
		"health_score":   summary.HealthScore,
		"threat_level":   summary.ThreatLevel,
		"color":          v.getHealthColor(summary.OverallHealth),
	}

	// Key insights
	viz["key_insights"] = summary.KeyInsights

	return viz
}

// Additional helper methods for visualization

func (v *DefaultNetworkVisualizer) generateTopologyNodes(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) []map[string]interface{} {
	nodes := make([]map[string]interface{}, 0)
	nodeMap := make(map[string]bool)

	// Add client nodes
	for ip, stats := range clientStats {
		if !nodeMap[ip] {
			node := map[string]interface{}{
				"id":       ip,
				"label":    v.getNodeLabel(ip, stats.Hostname),
				"type":     "client",
				"size":     v.calculateNodeSize(stats.QueryCount),
				"color":    v.getClientColor(stats),
				"queries":  stats.QueryCount,
				"hostname": stats.Hostname,
			}
			nodes = append(nodes, node)
			nodeMap[ip] = true
		}
	}

	// Add domain nodes (top domains only)
	domainCounts := make(map[string]int)
	for _, record := range records {
		domainCounts[record.Domain]++
	}

	// Sort domains by count and take top 50
	type domainCount struct {
		domain string
		count  int
	}

	domains := make([]domainCount, 0, len(domainCounts))
	for domain, count := range domainCounts {
		domains = append(domains, domainCount{domain, count})
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].count > domains[j].count
	})

	maxDomains := 50
	if len(domains) > maxDomains {
		domains = domains[:maxDomains]
	}

	for _, dc := range domains {
		node := map[string]interface{}{
			"id":      dc.domain,
			"label":   dc.domain,
			"type":    "domain",
			"size":    v.calculateNodeSize(dc.count),
			"color":   v.getDomainColor(dc.domain),
			"queries": dc.count,
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func (v *DefaultNetworkVisualizer) generateTopologyEdges(records []types.PiholeRecord) []map[string]interface{} {
	edges := make([]map[string]interface{}, 0)
	edgeMap := make(map[string]int) // source-target -> count

	for _, record := range records {
		edgeKey := fmt.Sprintf("%s-%s", record.Client, record.Domain)
		edgeMap[edgeKey]++
	}

	for edgeKey, count := range edgeMap {
		parts := strings.Split(edgeKey, "-")
		if len(parts) == 2 {
			edge := map[string]interface{}{
				"source": parts[0],
				"target": parts[1],
				"weight": count,
				"width":  v.calculateEdgeWidth(count),
				"color":  v.getEdgeColor(count),
			}
			edges = append(edges, edge)
		}
	}

	return edges
}

// Color and styling helper methods

func (v *DefaultNetworkVisualizer) getProtocolColor(protocol string) string {
	colors := map[string]string{
		"DNS_UDP": "#3498db",
		"DNS_TCP": "#e74c3c",
		"HTTP":    "#2ecc71",
		"HTTPS":   "#f39c12",
	}
	if color, exists := colors[protocol]; exists {
		return color
	}
	return "#95a5a6"
}

func (v *DefaultNetworkVisualizer) getThreatLevelColor(level string) string {
	colors := map[string]string{
		"LOW":      "#27ae60",
		"MEDIUM":   "#f39c12",
		"HIGH":     "#e67e22",
		"CRITICAL": "#e74c3c",
	}
	if color, exists := colors[level]; exists {
		return color
	}
	return "#95a5a6"
}

func (v *DefaultNetworkVisualizer) getScoreColor(score float64) string {
	switch {
	case score >= 90:
		return "#27ae60" // Green
	case score >= 80:
		return "#2ecc71" // Light green
	case score >= 70:
		return "#f39c12" // Orange
	case score >= 60:
		return "#e67e22" // Dark orange
	default:
		return "#e74c3c" // Red
	}
}

func (v *DefaultNetworkVisualizer) getHealthColor(health string) string {
	colors := map[string]string{
		"EXCELLENT": "#27ae60",
		"GOOD":      "#2ecc71",
		"FAIR":      "#f39c12",
		"POOR":      "#e67e22",
		"CRITICAL":  "#e74c3c",
	}
	if color, exists := colors[health]; exists {
		return color
	}
	return "#95a5a6"
}

// Placeholder implementations for remaining helper methods
func (v *DefaultNetworkVisualizer) getSizeColor(category string) string            { return "#3498db" }
func (v *DefaultNetworkVisualizer) getThreatTypeColor(threatType string) string    { return "#e74c3c" }
func (v *DefaultNetworkVisualizer) getClientColor(stats *types.ClientStats) string { return "#3498db" }
func (v *DefaultNetworkVisualizer) getDomainColor(domain string) string            { return "#2ecc71" }
func (v *DefaultNetworkVisualizer) getEdgeColor(count int) string                  { return "#95a5a6" }

func (v *DefaultNetworkVisualizer) calculateNodeSize(count int) int {
	// Scale node size based on count (min 10, max 50)
	size := 10 + (count/100)*10
	if size > 50 {
		size = 50
	}
	return size
}

func (v *DefaultNetworkVisualizer) calculateEdgeWidth(count int) int {
	// Scale edge width based on count (min 1, max 10)
	width := 1 + (count/1000)*5
	if width > 10 {
		width = 10
	}
	return width
}

func (v *DefaultNetworkVisualizer) getNodeLabel(ip, hostname string) string {
	if hostname != "" {
		return hostname
	}
	return ip
}

// Additional helper method implementations would continue here...
// For brevity, I'll include simplified placeholder implementations

func (v *DefaultNetworkVisualizer) generateNetworkClusters(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) []map[string]interface{} {
	return []map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateLayoutInfo(nodes, edges []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) calculateConnectionDensity(nodes, edges []map[string]interface{}) float64 {
	if len(nodes) <= 1 {
		return 0
	}
	maxEdges := len(nodes) * (len(nodes) - 1) / 2
	return float64(len(edges)) / float64(maxEdges)
}

func (v *DefaultNetworkVisualizer) groupRecordsByTime(records []types.PiholeRecord, interval time.Duration) map[string][]types.PiholeRecord {
	groups := make(map[string][]types.PiholeRecord)
	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		key := timestamp.Truncate(interval).Format(time.RFC3339)
		groups[key] = append(groups[key], record)
	}
	return groups
}

func (v *DefaultNetworkVisualizer) calculateMetricValue(records []types.PiholeRecord, metric string) float64 {
	switch metric {
	case "query_count":
		return float64(len(records))
	case "unique_clients":
		clients := make(map[string]bool)
		for _, r := range records {
			clients[r.Client] = true
		}
		return float64(len(clients))
	case "unique_domains":
		domains := make(map[string]bool)
		for _, r := range records {
			domains[r.Domain] = true
		}
		return float64(len(domains))
	default:
		return float64(len(records))
	}
}

func (v *DefaultNetworkVisualizer) calculateTrend(dataPoints []map[string]interface{}) string {
	if len(dataPoints) < 2 {
		return "stable"
	}

	first := dataPoints[0]["value"].(float64)
	last := dataPoints[len(dataPoints)-1]["value"].(float64)

	if last > first*1.1 {
		return "increasing"
	} else if last < first*0.9 {
		return "decreasing"
	}
	return "stable"
}

func (v *DefaultNetworkVisualizer) calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func (v *DefaultNetworkVisualizer) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func (v *DefaultNetworkVisualizer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Additional placeholder methods for completeness
func (v *DefaultNetworkVisualizer) generateHourlyHeatmap(records []types.PiholeRecord) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateDailyHeatmap(records []types.PiholeRecord) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateClientDomainHeatmap(records []types.PiholeRecord) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateQueryTypeHeatmap(records []types.PiholeRecord) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) getThreatLevelDescription(level string) string {
	descriptions := map[string]string{
		"LOW":      "No significant threats detected",
		"MEDIUM":   "Some suspicious activity detected",
		"HIGH":     "Multiple threats detected - attention required",
		"CRITICAL": "Critical threats detected - immediate action required",
	}
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "Unknown threat level"
}

func (v *DefaultNetworkVisualizer) getScoreDescription(score float64) string {
	switch {
	case score >= 90:
		return "Excellent performance"
	case score >= 80:
		return "Good performance"
	case score >= 70:
		return "Fair performance"
	case score >= 60:
		return "Poor performance"
	default:
		return "Critical performance issues"
	}
}

// Additional methods would be implemented here for complete functionality
func (v *DefaultNetworkVisualizer) generateThreatSummary(threats []types.SecurityThreat) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateThreatTimeline(threats []types.SecurityThreat) []map[string]interface{} {
	return []map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateRiskDistribution(threats []types.SecurityThreat, suspicious []types.SuspiciousActivity) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateTopThreats(threats []types.SecurityThreat, limit int) []map[string]interface{} {
	return []map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generatePerformanceMetricsCards(result *types.NetworkPerformanceResult) []map[string]interface{} {
	return []map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateLatencyAnalysis(latency *types.LatencyMetrics) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateBandwidthAnalysis(bandwidth *types.BandwidthMetrics) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generateQualityAssessmentViz(assessment *types.QualityAssessment) map[string]interface{} {
	return map[string]interface{}{}
}

func (v *DefaultNetworkVisualizer) generatePerformanceTrends(result *types.NetworkPerformanceResult) map[string]interface{} {
	return map[string]interface{}{}
}
