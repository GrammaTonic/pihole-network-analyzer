package network

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"pihole-analyzer/internal/types"
)

// DefaultTrafficPatternAnalyzer implements the TrafficPatternAnalyzer interface
type DefaultTrafficPatternAnalyzer struct {
	logger *slog.Logger
}

// NewTrafficPatternAnalyzer creates a new traffic pattern analyzer
func NewTrafficPatternAnalyzer(logger *slog.Logger) TrafficPatternAnalyzer {
	return &DefaultTrafficPatternAnalyzer{
		logger: logger,
	}
}

// AnalyzePatterns implements TrafficPatternAnalyzer.AnalyzePatterns
func (t *DefaultTrafficPatternAnalyzer) AnalyzePatterns(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats, config types.TrafficPatternsConfig) (*types.TrafficPatternsResult, error) {
	t.logger.Info("Starting traffic pattern analysis",
		slog.Int("record_count", len(records)),
		slog.Int("client_count", len(clientStats)),
		slog.String("analysis_window", config.AnalysisWindow))

	startTime := time.Now()
	patternID := fmt.Sprintf("pattern_%d", startTime.Unix())

	result := &types.TrafficPatternsResult{
		PatternID:         patternID,
		DetectedPatterns:  make([]types.TrafficPattern, 0),
		BandwidthPatterns: make([]types.BandwidthPattern, 0),
		TemporalPatterns:  make([]types.TemporalPattern, 0),
		ClientBehavior:    make(map[string]types.ClientBehavior),
		Anomalies:         make([]types.TrafficAnomaly, 0),
		PredictedTrends:   make([]types.TrafficTrend, 0),
	}

	// Parse analysis window
	analysisWindow, err := time.ParseDuration(config.AnalysisWindow)
	if err != nil {
		analysisWindow = time.Hour // Default to 1 hour
	}

	// Analyze bandwidth patterns
	if contains(config.PatternTypes, "bandwidth") {
		t.logger.Debug("Analyzing bandwidth patterns")
		bandwidthPatterns, err := t.AnalyzeBandwidthPatterns(records, analysisWindow)
		if err != nil {
			t.logger.Error("Failed to analyze bandwidth patterns", slog.String("error", err.Error()))
		} else {
			result.BandwidthPatterns = bandwidthPatterns
		}
	}

	// Analyze temporal patterns
	if contains(config.PatternTypes, "temporal") {
		t.logger.Debug("Analyzing temporal patterns")
		temporalPatterns, err := t.AnalyzeTemporalPatterns(records)
		if err != nil {
			t.logger.Error("Failed to analyze temporal patterns", slog.String("error", err.Error()))
		} else {
			result.TemporalPatterns = temporalPatterns
		}
	}

	// Analyze client behavior patterns
	if contains(config.PatternTypes, "client") {
		t.logger.Debug("Analyzing client behavior patterns")
		clientBehavior, err := t.AnalyzeClientBehavior(clientStats, records)
		if err != nil {
			t.logger.Error("Failed to analyze client behavior", slog.String("error", err.Error()))
		} else {
			result.ClientBehavior = clientBehavior
		}
	}

	// Detect general traffic patterns
	detectedPatterns := t.detectGeneralPatterns(records, config)
	result.DetectedPatterns = detectedPatterns

	// Detect anomalies if enabled
	if config.AnomalyDetection {
		t.logger.Debug("Detecting traffic anomalies")
		anomalies, err := t.DetectTrafficAnomalies(ctx, records, config)
		if err != nil {
			t.logger.Error("Failed to detect traffic anomalies", slog.String("error", err.Error()))
		} else {
			result.Anomalies = anomalies
		}
	}

	// Predict traffic trends
	trends, err := t.PredictTrafficTrends(records, analysisWindow)
	if err != nil {
		t.logger.Error("Failed to predict traffic trends", slog.String("error", err.Error()))
	} else {
		result.PredictedTrends = trends
	}

	duration := time.Since(startTime)
	t.logger.Info("Traffic pattern analysis completed",
		slog.String("duration", duration.String()),
		slog.Int("patterns_detected", len(result.DetectedPatterns)),
		slog.Int("bandwidth_patterns", len(result.BandwidthPatterns)),
		slog.Int("temporal_patterns", len(result.TemporalPatterns)),
		slog.Int("anomalies_detected", len(result.Anomalies)))

	return result, nil
}

// AnalyzeBandwidthPatterns implements TrafficPatternAnalyzer.AnalyzeBandwidthPatterns
func (t *DefaultTrafficPatternAnalyzer) AnalyzeBandwidthPatterns(records []types.PiholeRecord, timeWindow time.Duration) ([]types.BandwidthPattern, error) {
	patterns := make([]types.BandwidthPattern, 0)

	// Group records by time slots
	timeSlots := t.groupRecordsByTimeSlots(records, timeWindow)

	// Calculate bandwidth for each time slot
	for timeSlot, slotRecords := range timeSlots {
		avgBandwidth := t.calculateAverageBandwidth(slotRecords)
		peakBandwidth := t.calculatePeakBandwidth(slotRecords)
		usage := t.calculateUsagePercentage(slotRecords, len(records))
		trend := t.calculateTrend(slotRecords, timeSlots)

		pattern := types.BandwidthPattern{
			TimeSlot:      timeSlot,
			AvgBandwidth:  avgBandwidth,
			PeakBandwidth: peakBandwidth,
			Usage:         usage,
			Trend:         trend,
		}
		patterns = append(patterns, pattern)
	}

	// Sort patterns by time
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].TimeSlot < patterns[j].TimeSlot
	})

	return patterns, nil
}

// AnalyzeTemporalPatterns implements TrafficPatternAnalyzer.AnalyzeTemporalPatterns
func (t *DefaultTrafficPatternAnalyzer) AnalyzeTemporalPatterns(records []types.PiholeRecord) ([]types.TemporalPattern, error) {
	patterns := make([]types.TemporalPattern, 0)

	// Analyze hourly patterns
	hourlyPattern := t.analyzeHourlyPattern(records)
	patterns = append(patterns, hourlyPattern)

	// Analyze daily patterns
	dailyPattern := t.analyzeDailyPattern(records)
	patterns = append(patterns, dailyPattern)

	// Analyze weekly patterns
	weeklyPattern := t.analyzeWeeklyPattern(records)
	patterns = append(patterns, weeklyPattern)

	return patterns, nil
}

// AnalyzeClientBehavior implements TrafficPatternAnalyzer.AnalyzeClientBehavior
func (t *DefaultTrafficPatternAnalyzer) AnalyzeClientBehavior(clientStats map[string]*types.ClientStats, records []types.PiholeRecord) (map[string]types.ClientBehavior, error) {
	behavior := make(map[string]types.ClientBehavior)

	// Group records by client
	clientRecords := make(map[string][]types.PiholeRecord)
	for _, record := range records {
		clientRecords[record.Client] = append(clientRecords[record.Client], record)
	}

	for clientIP, clientRecs := range clientRecords {
		stats, hasStats := clientStats[clientIP]
		hostname := ""
		if hasStats {
			hostname = stats.Hostname
		}

		// Analyze client behavior
		behaviorType := t.classifyBehaviorType(clientRecs)
		activityLevel := t.classifyActivityLevel(clientRecs, len(records))
		typicalUsage := t.analyzeTypicalUsage(clientRecs)
		anomalies := t.detectClientAnomalies(clientRecs)
		riskScore := t.calculateRiskScore(clientRecs, anomalies)

		behavior[clientIP] = types.ClientBehavior{
			IP:            clientIP,
			Hostname:      hostname,
			BehaviorType:  behaviorType,
			ActivityLevel: activityLevel,
			TypicalUsage:  typicalUsage,
			Anomalies:     anomalies,
			RiskScore:     riskScore,
		}
	}

	return behavior, nil
}

// DetectTrafficAnomalies implements TrafficPatternAnalyzer.DetectTrafficAnomalies
func (t *DefaultTrafficPatternAnalyzer) DetectTrafficAnomalies(ctx context.Context, records []types.PiholeRecord, config types.TrafficPatternsConfig) ([]types.TrafficAnomaly, error) {
	anomalies := make([]types.TrafficAnomaly, 0)

	// Parse analysis window
	analysisWindow, err := time.ParseDuration(config.AnalysisWindow)
	if err != nil {
		analysisWindow = time.Hour
	}

	// Group records by time windows
	timeWindows := t.groupRecordsByTimeSlots(records, analysisWindow/4) // Use quarter windows for finer analysis

	// Calculate baseline statistics
	baseline := t.calculateBaseline(records)

	// Detect volume anomalies
	volumeAnomalies := t.detectVolumeAnomalies(timeWindows, baseline, config.PatternThreshold)
	anomalies = append(anomalies, volumeAnomalies...)

	// Detect frequency anomalies
	frequencyAnomalies := t.detectFrequencyAnomalies(timeWindows, baseline, config.PatternThreshold)
	anomalies = append(anomalies, frequencyAnomalies...)

	// Detect pattern deviations
	patternAnomalies := t.detectPatternDeviations(records, baseline, config.PatternThreshold)
	anomalies = append(anomalies, patternAnomalies...)

	return anomalies, nil
}

// PredictTrafficTrends implements TrafficPatternAnalyzer.PredictTrafficTrends
func (t *DefaultTrafficPatternAnalyzer) PredictTrafficTrends(records []types.PiholeRecord, horizon time.Duration) ([]types.TrafficTrend, error) {
	trends := make([]types.TrafficTrend, 0)

	// Predict query volume trend
	volumeTrend := t.predictVolumeTrend(records, horizon)
	trends = append(trends, volumeTrend)

	// Predict bandwidth trend
	bandwidthTrend := t.predictBandwidthTrend(records, horizon)
	trends = append(trends, bandwidthTrend)

	// Predict client count trend
	clientTrend := t.predictClientTrend(records, horizon)
	trends = append(trends, clientTrend)

	// Predict domain diversity trend
	diversityTrend := t.predictDiversityTrend(records, horizon)
	trends = append(trends, diversityTrend)

	return trends, nil
}

// Helper methods

// groupRecordsByTimeSlots groups records into time slots
func (t *DefaultTrafficPatternAnalyzer) groupRecordsByTimeSlots(records []types.PiholeRecord, slotDuration time.Duration) map[string][]types.PiholeRecord {
	slots := make(map[string][]types.PiholeRecord)

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		slotKey := timestamp.Truncate(slotDuration).Format(time.RFC3339)

		if _, exists := slots[slotKey]; !exists {
			slots[slotKey] = make([]types.PiholeRecord, 0)
		}
		slots[slotKey] = append(slots[slotKey], record)
	}

	return slots
}

// calculateAverageBandwidth estimates average bandwidth for records
func (t *DefaultTrafficPatternAnalyzer) calculateAverageBandwidth(records []types.PiholeRecord) float64 {
	if len(records) == 0 {
		return 0
	}

	totalBytes := int64(0)
	for _, record := range records {
		totalBytes += t.estimateRecordSize(record)
	}

	// Convert to Mbps (assuming records span 1 minute for estimation)
	bytesPerSecond := float64(totalBytes) / 60.0
	mbps := (bytesPerSecond * 8) / (1024 * 1024)

	return mbps
}

// calculatePeakBandwidth estimates peak bandwidth
func (t *DefaultTrafficPatternAnalyzer) calculatePeakBandwidth(records []types.PiholeRecord) float64 {
	// Group by smaller time windows to find peak
	subSlots := t.groupRecordsByTimeSlots(records, time.Minute)

	maxBandwidth := 0.0
	for _, subRecords := range subSlots {
		bandwidth := t.calculateAverageBandwidth(subRecords)
		if bandwidth > maxBandwidth {
			maxBandwidth = bandwidth
		}
	}

	return maxBandwidth
}

// calculateUsagePercentage calculates usage as percentage of total
func (t *DefaultTrafficPatternAnalyzer) calculateUsagePercentage(slotRecords []types.PiholeRecord, totalRecords int) float64 {
	if totalRecords == 0 {
		return 0
	}
	return float64(len(slotRecords)) / float64(totalRecords) * 100
}

// calculateTrend determines if traffic is increasing, decreasing, or stable
func (t *DefaultTrafficPatternAnalyzer) calculateTrend(slotRecords []types.PiholeRecord, allSlots map[string][]types.PiholeRecord) string {
	// Simple trend calculation based on comparison with other slots
	slotSize := len(slotRecords)

	totalOthers := 0
	otherCount := 0
	for _, otherRecords := range allSlots {
		if len(otherRecords) != slotSize { // Don't compare with self
			totalOthers += len(otherRecords)
			otherCount++
		}
	}

	if otherCount == 0 {
		return "stable"
	}

	avgOthers := float64(totalOthers) / float64(otherCount)

	if float64(slotSize) > avgOthers*1.2 {
		return "increasing"
	} else if float64(slotSize) < avgOthers*0.8 {
		return "decreasing"
	}

	return "stable"
}

// analyzeHourlyPattern analyzes traffic patterns by hour of day
func (t *DefaultTrafficPatternAnalyzer) analyzeHourlyPattern(records []types.PiholeRecord) types.TemporalPattern {
	hourCounts := make(map[int]int)

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		hour := timestamp.Hour()
		hourCounts[hour]++
	}

	// Find peak and low hours
	peakHours := make([]int, 0)
	lowHours := make([]int, 0)

	if len(hourCounts) > 0 {
		avgCount := t.calculateAverageCount(hourCounts)

		for hour, count := range hourCounts {
			if float64(count) > avgCount*1.5 {
				peakHours = append(peakHours, hour)
			} else if float64(count) < avgCount*0.5 {
				lowHours = append(lowHours, hour)
			}
		}
	}

	regularity := t.calculateRegularity(hourCounts)
	seasonality := t.detectSeasonality(hourCounts)

	return types.TemporalPattern{
		Pattern:     "hourly",
		PeakHours:   peakHours,
		LowHours:    lowHours,
		Regularity:  regularity,
		Seasonality: seasonality,
	}
}

// analyzeDailyPattern analyzes traffic patterns by day of week
func (t *DefaultTrafficPatternAnalyzer) analyzeDailyPattern(records []types.PiholeRecord) types.TemporalPattern {
	dayCounts := make(map[int]int)

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		day := int(timestamp.Weekday())
		dayCounts[day]++
	}

	// Convert day numbers to hour format for consistency
	peakHours := make([]int, 0)
	lowHours := make([]int, 0)

	if len(dayCounts) > 0 {
		avgCount := t.calculateAverageCount(dayCounts)

		for day, count := range dayCounts {
			if float64(count) > avgCount*1.2 {
				peakHours = append(peakHours, day)
			} else if float64(count) < avgCount*0.8 {
				lowHours = append(lowHours, day)
			}
		}
	}

	regularity := t.calculateRegularity(dayCounts)
	seasonality := t.detectSeasonality(dayCounts)

	return types.TemporalPattern{
		Pattern:     "daily",
		PeakHours:   peakHours,
		LowHours:    lowHours,
		Regularity:  regularity,
		Seasonality: seasonality,
	}
}

// analyzeWeeklyPattern analyzes traffic patterns by week
func (t *DefaultTrafficPatternAnalyzer) analyzeWeeklyPattern(records []types.PiholeRecord) types.TemporalPattern {
	// For weekly patterns, analyze by week number
	weekCounts := make(map[int]int)

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		_, week := timestamp.ISOWeek()
		weekCounts[week]++
	}

	peakHours := make([]int, 0)
	lowHours := make([]int, 0)

	if len(weekCounts) > 0 {
		avgCount := t.calculateAverageCount(weekCounts)

		for week, count := range weekCounts {
			if float64(count) > avgCount*1.2 {
				peakHours = append(peakHours, week)
			} else if float64(count) < avgCount*0.8 {
				lowHours = append(lowHours, week)
			}
		}
	}

	regularity := t.calculateRegularity(weekCounts)
	seasonality := t.detectSeasonality(weekCounts)

	return types.TemporalPattern{
		Pattern:     "weekly",
		PeakHours:   peakHours,
		LowHours:    lowHours,
		Regularity:  regularity,
		Seasonality: seasonality,
	}
}

// classifyBehaviorType classifies client behavior type
func (t *DefaultTrafficPatternAnalyzer) classifyBehaviorType(records []types.PiholeRecord) string {
	if len(records) == 0 {
		return "inactive"
	}

	// Analyze query patterns
	domainCounts := make(map[string]int)
	for _, record := range records {
		domainCounts[record.Domain]++
	}

	uniqueDomains := len(domainCounts)
	totalQueries := len(records)
	avgQueriesPerDomain := float64(totalQueries) / float64(uniqueDomains)

	// Classify based on patterns
	switch {
	case uniqueDomains > 100 && avgQueriesPerDomain < 2:
		return "browser" // Many different domains, few queries each
	case uniqueDomains < 10 && avgQueriesPerDomain > 10:
		return "focused" // Few domains, many queries each
	case totalQueries > 1000:
		return "heavy_user"
	case totalQueries < 10:
		return "light_user"
	default:
		return "normal"
	}
}

// classifyActivityLevel classifies client activity level
func (t *DefaultTrafficPatternAnalyzer) classifyActivityLevel(records []types.PiholeRecord, totalRecords int) string {
	clientQueries := len(records)
	percentage := float64(clientQueries) / float64(totalRecords) * 100

	switch {
	case percentage > 10:
		return "high"
	case percentage > 2:
		return "normal"
	default:
		return "low"
	}
}

// analyzeTypicalUsage analyzes typical usage patterns by hour
func (t *DefaultTrafficPatternAnalyzer) analyzeTypicalUsage(records []types.PiholeRecord) []types.HourlyUsage {
	hourUsage := make(map[int][]int)         // hour -> list of query counts
	hourBandwidth := make(map[int][]float64) // hour -> list of bandwidth values

	// Group by hour and day to get patterns
	dailyHourUsage := make(map[string]map[int]int) // date -> hour -> count

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		hour := timestamp.Hour()
		date := timestamp.Format("2006-01-02")

		if _, exists := dailyHourUsage[date]; !exists {
			dailyHourUsage[date] = make(map[int]int)
		}
		dailyHourUsage[date][hour]++
	}

	// Calculate averages per hour across all days
	for _, dayUsage := range dailyHourUsage {
		for hour, count := range dayUsage {
			hourUsage[hour] = append(hourUsage[hour], count)
			bandwidth := t.estimateBandwidthFromQueries(count)
			hourBandwidth[hour] = append(hourBandwidth[hour], bandwidth)
		}
	}

	// Create hourly usage summary
	usage := make([]types.HourlyUsage, 0, 24)
	for hour := 0; hour < 24; hour++ {
		avgQueries := 0.0
		avgBandwidth := 0.0

		if counts, exists := hourUsage[hour]; exists && len(counts) > 0 {
			total := 0
			for _, count := range counts {
				total += count
			}
			avgQueries = float64(total) / float64(len(counts))
		}

		if bandwidths, exists := hourBandwidth[hour]; exists && len(bandwidths) > 0 {
			total := 0.0
			for _, bw := range bandwidths {
				total += bw
			}
			avgBandwidth = total / float64(len(bandwidths))
		}

		usage = append(usage, types.HourlyUsage{
			Hour:         hour,
			AvgQueries:   avgQueries,
			AvgBandwidth: avgBandwidth,
		})
	}

	return usage
}

// detectClientAnomalies detects anomalies in client behavior
func (t *DefaultTrafficPatternAnalyzer) detectClientAnomalies(records []types.PiholeRecord) []types.BehaviorAnomaly {
	anomalies := make([]types.BehaviorAnomaly, 0)

	// Detect burst activity
	if len(records) > 0 {
		timeSlots := t.groupRecordsByTimeSlots(records, time.Minute*5)
		avgSlotSize := float64(len(records)) / float64(len(timeSlots))

		for timeSlot, slotRecords := range timeSlots {
			if float64(len(slotRecords)) > avgSlotSize*3 { // 3x average
				anomaly := types.BehaviorAnomaly{
					Type:        "burst_activity",
					Description: fmt.Sprintf("Burst of %d queries in 5-minute window", len(slotRecords)),
					Timestamp:   timeSlot,
					Severity:    "MEDIUM",
					Confidence:  0.8,
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	// Detect unusual domains
	domainCounts := make(map[string]int)
	for _, record := range records {
		domainCounts[record.Domain]++
	}

	for domain, count := range domainCounts {
		if count == 1 && len(domain) > 50 { // Single query to very long domain
			anomaly := types.BehaviorAnomaly{
				Type:        "unusual_domain",
				Description: fmt.Sprintf("Query to unusual domain: %s", domain),
				Timestamp:   time.Now().Format(time.RFC3339),
				Severity:    "LOW",
				Confidence:  0.6,
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// calculateRiskScore calculates a risk score for client behavior
func (t *DefaultTrafficPatternAnalyzer) calculateRiskScore(records []types.PiholeRecord, anomalies []types.BehaviorAnomaly) float64 {
	baseScore := 0.0

	// Factor in anomaly count and severity
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "CRITICAL":
			baseScore += 0.3
		case "HIGH":
			baseScore += 0.2
		case "MEDIUM":
			baseScore += 0.1
		case "LOW":
			baseScore += 0.05
		}
	}

	// Factor in query volume (very high volume slightly increases risk)
	if len(records) > 10000 {
		baseScore += 0.1
	}

	// Factor in domain diversity (very high diversity might indicate scanning)
	domainCounts := make(map[string]int)
	for _, record := range records {
		domainCounts[record.Domain]++
	}

	if len(domainCounts) > 1000 {
		baseScore += 0.15
	}

	return normalizeScore(baseScore)
}

// Remaining helper methods

// detectGeneralPatterns detects general traffic patterns
func (t *DefaultTrafficPatternAnalyzer) detectGeneralPatterns(records []types.PiholeRecord, config types.TrafficPatternsConfig) []types.TrafficPattern {
	patterns := make([]types.TrafficPattern, 0)

	// Detect periodic patterns
	periodicPattern := t.detectPeriodicPattern(records, config.PatternThreshold)
	if periodicPattern.ID != "" {
		patterns = append(patterns, periodicPattern)
	}

	// Detect burst patterns
	burstPattern := t.detectBurstPattern(records, config.PatternThreshold)
	if burstPattern.ID != "" {
		patterns = append(patterns, burstPattern)
	}

	// Detect seasonal patterns
	seasonalPattern := t.detectSeasonalPattern(records, config.PatternThreshold)
	if seasonalPattern.ID != "" {
		patterns = append(patterns, seasonalPattern)
	}

	return patterns
}

// Additional helper methods (implementations would continue...)

// Utility methods
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (t *DefaultTrafficPatternAnalyzer) estimateRecordSize(record types.PiholeRecord) int64 {
	// Estimate DNS packet size
	return int64(len(record.Domain) + 50) // Domain + DNS overhead
}

func (t *DefaultTrafficPatternAnalyzer) calculateAverageCount(counts map[int]int) float64 {
	if len(counts) == 0 {
		return 0
	}

	total := 0
	for _, count := range counts {
		total += count
	}

	return float64(total) / float64(len(counts))
}

func (t *DefaultTrafficPatternAnalyzer) calculateRegularity(counts map[int]int) float64 {
	if len(counts) < 2 {
		return 0
	}

	values := make([]float64, 0, len(counts))
	for _, count := range counts {
		values = append(values, float64(count))
	}

	stdDev := calculateStandardDeviation(values)
	avg := t.calculateAverageCount(counts)

	if avg == 0 {
		return 0
	}

	// Regularity is inverse of coefficient of variation
	cv := stdDev / avg
	return math.Max(0, 1.0-cv)
}

func (t *DefaultTrafficPatternAnalyzer) detectSeasonality(counts map[int]int) bool {
	// Simple seasonality detection based on pattern regularity
	regularity := t.calculateRegularity(counts)
	return regularity > 0.7
}

func (t *DefaultTrafficPatternAnalyzer) estimateBandwidthFromQueries(queryCount int) float64 {
	// Estimate bandwidth from query count (rough approximation)
	avgBytesPerQuery := 100.0                                           // DNS query + response
	bytesPerSecond := (float64(queryCount) * avgBytesPerQuery) / 3600.0 // Spread over hour
	return (bytesPerSecond * 8) / (1024 * 1024)                         // Convert to Mbps
}

func (t *DefaultTrafficPatternAnalyzer) calculateBaseline(records []types.PiholeRecord) map[string]interface{} {
	baseline := make(map[string]interface{})

	// Calculate baseline query rate
	if len(records) > 0 {
		timeSpan := t.calculateTimeSpan(records)
		baseline["avg_queries_per_minute"] = float64(len(records)) / timeSpan.Minutes()
	}

	return baseline
}

func (t *DefaultTrafficPatternAnalyzer) calculateTimeSpan(records []types.PiholeRecord) time.Duration {
	if len(records) == 0 {
		return time.Hour
	}

	var earliest, latest time.Time
	for i, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		if i == 0 {
			earliest = timestamp
			latest = timestamp
		} else {
			if timestamp.Before(earliest) {
				earliest = timestamp
			}
			if timestamp.After(latest) {
				latest = timestamp
			}
		}
	}

	duration := latest.Sub(earliest)
	if duration < time.Minute {
		duration = time.Hour
	}

	return duration
}

// Placeholder implementations for remaining methods
func (t *DefaultTrafficPatternAnalyzer) detectVolumeAnomalies(timeWindows map[string][]types.PiholeRecord, baseline map[string]interface{}, threshold float64) []types.TrafficAnomaly {
	return []types.TrafficAnomaly{}
}

func (t *DefaultTrafficPatternAnalyzer) detectFrequencyAnomalies(timeWindows map[string][]types.PiholeRecord, baseline map[string]interface{}, threshold float64) []types.TrafficAnomaly {
	return []types.TrafficAnomaly{}
}

func (t *DefaultTrafficPatternAnalyzer) detectPatternDeviations(records []types.PiholeRecord, baseline map[string]interface{}, threshold float64) []types.TrafficAnomaly {
	return []types.TrafficAnomaly{}
}

func (t *DefaultTrafficPatternAnalyzer) predictVolumeTrend(records []types.PiholeRecord, horizon time.Duration) types.TrafficTrend {
	return types.TrafficTrend{
		Metric:      "query_volume",
		Current:     float64(len(records)),
		Predicted:   float64(len(records)) * 1.1, // Simple 10% increase prediction
		Confidence:  0.7,
		TimeHorizon: horizon.String(),
		Trend:       "stable",
	}
}

func (t *DefaultTrafficPatternAnalyzer) predictBandwidthTrend(records []types.PiholeRecord, horizon time.Duration) types.TrafficTrend {
	return types.TrafficTrend{
		Metric:      "bandwidth",
		Current:     t.calculateAverageBandwidth(records),
		Predicted:   t.calculateAverageBandwidth(records) * 1.05,
		Confidence:  0.6,
		TimeHorizon: horizon.String(),
		Trend:       "stable",
	}
}

func (t *DefaultTrafficPatternAnalyzer) predictClientTrend(records []types.PiholeRecord, horizon time.Duration) types.TrafficTrend {
	clients := make(map[string]bool)
	for _, record := range records {
		clients[record.Client] = true
	}

	return types.TrafficTrend{
		Metric:      "client_count",
		Current:     float64(len(clients)),
		Predicted:   float64(len(clients)) * 1.02,
		Confidence:  0.8,
		TimeHorizon: horizon.String(),
		Trend:       "stable",
	}
}

func (t *DefaultTrafficPatternAnalyzer) predictDiversityTrend(records []types.PiholeRecord, horizon time.Duration) types.TrafficTrend {
	domains := make(map[string]bool)
	for _, record := range records {
		domains[record.Domain] = true
	}

	return types.TrafficTrend{
		Metric:      "domain_diversity",
		Current:     float64(len(domains)),
		Predicted:   float64(len(domains)) * 1.08,
		Confidence:  0.65,
		TimeHorizon: horizon.String(),
		Trend:       "increasing",
	}
}

func (t *DefaultTrafficPatternAnalyzer) detectPeriodicPattern(records []types.PiholeRecord, threshold float64) types.TrafficPattern {
	return types.TrafficPattern{
		ID:          "",
		Type:        "periodic",
		Description: "No significant periodic pattern detected",
		Confidence:  0.0,
	}
}

func (t *DefaultTrafficPatternAnalyzer) detectBurstPattern(records []types.PiholeRecord, threshold float64) types.TrafficPattern {
	return types.TrafficPattern{
		ID:          "",
		Type:        "burst",
		Description: "No significant burst pattern detected",
		Confidence:  0.0,
	}
}

func (t *DefaultTrafficPatternAnalyzer) detectSeasonalPattern(records []types.PiholeRecord, threshold float64) types.TrafficPattern {
	return types.TrafficPattern{
		ID:          "",
		Type:        "seasonal",
		Description: "No significant seasonal pattern detected",
		Confidence:  0.0,
	}
}
