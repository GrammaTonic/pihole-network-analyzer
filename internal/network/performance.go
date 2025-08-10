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

// DefaultPerformanceAnalyzer implements the PerformanceAnalyzer interface
type DefaultPerformanceAnalyzer struct {
	logger *slog.Logger
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(logger *slog.Logger) PerformanceAnalyzer {
	return &DefaultPerformanceAnalyzer{
		logger: logger,
	}
}

// AnalyzePerformance implements PerformanceAnalyzer.AnalyzePerformance
func (p *DefaultPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats, config types.NetworkPerformanceConfig) (*types.NetworkPerformanceResult, error) {
	p.logger.Info("Starting network performance analysis",
		slog.Int("record_count", len(records)),
		slog.Int("client_count", len(clientStats)),
		slog.Bool("latency_analysis", config.LatencyAnalysis),
		slog.Bool("bandwidth_analysis", config.BandwidthAnalysis),
		slog.Bool("throughput_analysis", config.ThroughputAnalysis))

	startTime := time.Now()

	result := &types.NetworkPerformanceResult{
		OverallScore: 100.0,
	}

	// Analyze latency metrics
	if config.LatencyAnalysis {
		p.logger.Debug("Analyzing latency metrics")
		latencyMetrics, err := p.AnalyzeLatency(records)
		if err != nil {
			p.logger.Error("Failed to analyze latency", slog.String("error", err.Error()))
		} else {
			result.LatencyMetrics = *latencyMetrics
		}
	}

	// Analyze bandwidth metrics
	if config.BandwidthAnalysis {
		p.logger.Debug("Analyzing bandwidth metrics")
		bandwidthMetrics, err := p.AnalyzeBandwidth(records, clientStats)
		if err != nil {
			p.logger.Error("Failed to analyze bandwidth", slog.String("error", err.Error()))
		} else {
			result.BandwidthMetrics = *bandwidthMetrics
		}
	}

	// Analyze throughput metrics
	if config.ThroughputAnalysis {
		p.logger.Debug("Analyzing throughput metrics")
		throughputMetrics, err := p.AnalyzeThroughput(records)
		if err != nil {
			p.logger.Error("Failed to analyze throughput", slog.String("error", err.Error()))
		} else {
			result.ThroughputMetrics = *throughputMetrics
		}
	}

	// Detect packet loss
	if config.PacketLossDetection {
		p.logger.Debug("Detecting packet loss")
		packetLossMetrics, err := p.DetectPacketLoss(records)
		if err != nil {
			p.logger.Error("Failed to detect packet loss", slog.String("error", err.Error()))
		} else {
			result.PacketLossMetrics = *packetLossMetrics
		}
	}

	// Analyze jitter
	if config.JitterAnalysis {
		p.logger.Debug("Analyzing jitter")
		jitterMetrics, err := p.AnalyzeJitter(records)
		if err != nil {
			p.logger.Error("Failed to analyze jitter", slog.String("error", err.Error()))
		} else {
			result.JitterMetrics = *jitterMetrics
		}
	}

	// Assess overall network quality
	qualityAssessment, err := p.AssessNetworkQuality(
		&result.LatencyMetrics, &result.BandwidthMetrics,
		&result.ThroughputMetrics, &result.PacketLossMetrics,
		&result.JitterMetrics, config.QualityThresholds)
	if err != nil {
		p.logger.Error("Failed to assess network quality", slog.String("error", err.Error()))
	} else {
		result.QualityAssessment = *qualityAssessment
		result.OverallScore = p.calculateOverallScore(qualityAssessment)
	}

	duration := time.Since(startTime)
	p.logger.Info("Performance analysis completed",
		slog.String("duration", duration.String()),
		slog.Float64("overall_score", result.OverallScore),
		slog.String("overall_grade", result.QualityAssessment.OverallGrade))

	return result, nil
}

// AnalyzeLatency implements PerformanceAnalyzer.AnalyzeLatency
func (p *DefaultPerformanceAnalyzer) AnalyzeLatency(records []types.PiholeRecord) (*types.LatencyMetrics, error) {
	latencies := make([]float64, 0)
	clientLatencies := make(map[string][]float64)

	// Extract latency data from records
	for _, record := range records {
		// Estimate latency based on record processing time or use ReplyTime if available
		latency := p.estimateLatency(record)
		if latency > 0 {
			latencies = append(latencies, latency)
			clientLatencies[record.Client] = append(clientLatencies[record.Client], latency)
		}
	}

	if len(latencies) == 0 {
		return &types.LatencyMetrics{}, nil
	}

	// Sort latencies for percentile calculations
	sort.Float64s(latencies)

	// Calculate basic statistics
	avgLatency := p.calculateMean(latencies)
	minLatency := latencies[0]
	maxLatency := latencies[len(latencies)-1]

	// Calculate percentiles
	p50 := calculatePercentile(latencies, 50)
	p95 := calculatePercentile(latencies, 95)
	p99 := calculatePercentile(latencies, 99)

	// Calculate per-client averages
	perClient := make(map[string]float64)
	for client, clientLats := range clientLatencies {
		perClient[client] = p.calculateMean(clientLats)
	}

	// Create distribution buckets
	distribution := p.createLatencyDistribution(latencies)

	return &types.LatencyMetrics{
		AvgLatency:   avgLatency,
		MinLatency:   minLatency,
		MaxLatency:   maxLatency,
		P50Latency:   p50,
		P95Latency:   p95,
		P99Latency:   p99,
		PerClient:    perClient,
		Distribution: distribution,
	}, nil
}

// AnalyzeBandwidth implements PerformanceAnalyzer.AnalyzeBandwidth
func (p *DefaultPerformanceAnalyzer) AnalyzeBandwidth(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (*types.BandwidthMetrics, error) {
	// Calculate bandwidth based on query volume and estimated data transfer
	totalBytes := int64(0)
	clientBytes := make(map[string]int64)
	timeSlots := make(map[string]int64) // time slot -> bytes

	// Process records to calculate bandwidth
	for _, record := range records {
		bytes := p.estimateQueryBytes(record)
		totalBytes += bytes
		clientBytes[record.Client] += bytes

		// Group by time slots (5-minute intervals)
		timestamp := parseTimestamp(record.Timestamp)
		timeSlot := timestamp.Truncate(time.Minute * 5).Format(time.RFC3339)
		timeSlots[timeSlot] += bytes
	}

	// Calculate time span
	timeSpan := p.calculateTimeSpanFromRecords(records)
	if timeSpan.Seconds() == 0 {
		timeSpan = time.Hour // Default
	}

	// Calculate bandwidth metrics
	totalBandwidthMbps := p.bytesToMbps(totalBytes, timeSpan)
	avgBandwidthMbps := totalBandwidthMbps

	// Find peak bandwidth
	peakBandwidthMbps := 0.0
	timeDistribution := make([]types.BandwidthTimeSlot, 0)
	
	for timeSlot, bytes := range timeSlots {
		slotBandwidth := p.bytesToMbps(bytes, time.Minute*5)
		if slotBandwidth > peakBandwidthMbps {
			peakBandwidthMbps = slotBandwidth
		}
		
		timeDistribution = append(timeDistribution, types.BandwidthTimeSlot{
			TimeSlot:  timeSlot,
			Bandwidth: slotBandwidth,
		})
	}

	// Sort time distribution
	sort.Slice(timeDistribution, func(i, j int) bool {
		return timeDistribution[i].TimeSlot < timeDistribution[j].TimeSlot
	})

	// Calculate per-client bandwidth
	perClient := make(map[string]float64)
	for client, bytes := range clientBytes {
		perClient[client] = p.bytesToMbps(bytes, timeSpan)
	}

	return &types.BandwidthMetrics{
		TotalBandwidth:   totalBandwidthMbps,
		AvgBandwidth:     avgBandwidthMbps,
		PeakBandwidth:    peakBandwidthMbps,
		PerClient:        perClient,
		TimeDistribution: timeDistribution,
	}, nil
}

// AnalyzeThroughput implements PerformanceAnalyzer.AnalyzeThroughput
func (p *DefaultPerformanceAnalyzer) AnalyzeThroughput(records []types.PiholeRecord) (*types.ThroughputMetrics, error) {
	if len(records) == 0 {
		return &types.ThroughputMetrics{}, nil
	}

	// Calculate time span
	timeSpan := p.calculateTimeSpanFromRecords(records)
	if timeSpan.Seconds() == 0 {
		timeSpan = time.Hour
	}

	// Calculate queries per second
	qps := float64(len(records)) / timeSpan.Seconds()

	// Calculate peak QPS by analyzing 1-minute windows
	minuteWindows := make(map[string]int)
	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		minute := timestamp.Truncate(time.Minute).Format(time.RFC3339)
		minuteWindows[minute]++
	}

	peakQPS := 0.0
	for _, count := range minuteWindows {
		minuteQPS := float64(count) / 60.0 // queries per second in this minute
		if minuteQPS > peakQPS {
			peakQPS = minuteQPS
		}
	}

	// Calculate response rate (assuming all DNS queries get responses)
	responseRate := 100.0 // DNS queries typically always get responses

	// Estimate average processing time
	processingTimes := make([]float64, 0)
	for _, record := range records {
		if processingTime := p.estimateProcessingTime(record); processingTime > 0 {
			processingTimes = append(processingTimes, processingTime)
		}
	}

	avgProcessingTime := 0.0
	if len(processingTimes) > 0 {
		avgProcessingTime = p.calculateMean(processingTimes)
	}

	return &types.ThroughputMetrics{
		QueriesPerSecond: qps,
		PeakQPS:          peakQPS,
		AvgQPS:           qps,
		ResponseRate:     responseRate,
		ProcessingTime:   avgProcessingTime,
	}, nil
}

// DetectPacketLoss implements PerformanceAnalyzer.DetectPacketLoss
func (p *DefaultPerformanceAnalyzer) DetectPacketLoss(records []types.PiholeRecord) (*types.PacketLossMetrics, error) {
	// For DNS queries, packet loss is difficult to detect directly
	// We can infer it from failed queries or timeouts

	totalQueries := int64(len(records))
	lostPackets := int64(0)
	clientLoss := make(map[string]float64)
	burstEvents := make([]types.LossBurst, 0)

	// Analyze query status codes to infer packet loss
	statusCounts := make(map[int]int)
	clientStatusCounts := make(map[string]map[int]int)

	for _, record := range records {
		statusCounts[record.Status]++
		
		if _, exists := clientStatusCounts[record.Client]; !exists {
			clientStatusCounts[record.Client] = make(map[int]int)
		}
		clientStatusCounts[record.Client][record.Status]++

		// Status codes indicating potential packet loss or failures
		// (specific codes depend on Pi-hole implementation)
		if record.Status == 3 || record.Status == 4 { // Example: NXDOMAIN or timeout
			lostPackets++
		}
	}

	lossPercentage := 0.0
	if totalQueries > 0 {
		lossPercentage = float64(lostPackets) / float64(totalQueries) * 100
	}

	// Calculate per-client loss
	for client, statuses := range clientStatusCounts {
		clientTotal := 0
		clientLost := 0
		for status, count := range statuses {
			clientTotal += count
			if status == 3 || status == 4 {
				clientLost += count
			}
		}
		if clientTotal > 0 {
			clientLoss[client] = float64(clientLost) / float64(clientTotal) * 100
		}
	}

	// Detect burst loss events (simplified)
	// In a real implementation, this would analyze time windows for concentrated losses
	if lossPercentage > 5 { // If overall loss > 5%, consider it a burst
		burstEvents = append(burstEvents, types.LossBurst{
			StartTime:   time.Now().Add(-time.Hour).Format(time.RFC3339),
			Duration:    "1h",
			LostPackets: lostPackets,
			LossRate:    lossPercentage,
		})
	}

	return &types.PacketLossMetrics{
		LossPercentage: lossPercentage,
		TotalLost:      lostPackets,
		TotalSent:      totalQueries,
		PerClient:      clientLoss,
		BurstLoss:      burstEvents,
	}, nil
}

// AnalyzeJitter implements PerformanceAnalyzer.AnalyzeJitter
func (p *DefaultPerformanceAnalyzer) AnalyzeJitter(records []types.PiholeRecord) (*types.JitterMetrics, error) {
	// Calculate jitter based on response time variations
	clientJitters := make(map[string][]float64)
	allJitters := make([]float64, 0)

	// Group by client and calculate response time variations
	clientTimes := make(map[string][]float64)
	for _, record := range records {
		responseTime := p.estimateLatency(record)
		if responseTime > 0 {
			clientTimes[record.Client] = append(clientTimes[record.Client], responseTime)
		}
	}

	// Calculate jitter for each client
	for client, times := range clientTimes {
		if len(times) < 2 {
			continue
		}

		// Calculate jitter as variation in response times
		jitters := make([]float64, 0, len(times)-1)
		for i := 1; i < len(times); i++ {
			jitter := math.Abs(times[i] - times[i-1])
			jitters = append(jitters, jitter)
			allJitters = append(allJitters, jitter)
		}
		clientJitters[client] = jitters
	}

	if len(allJitters) == 0 {
		return &types.JitterMetrics{}, nil
	}

	// Calculate overall statistics
	avgJitter := p.calculateMean(allJitters)
	maxJitter := 0.0
	for _, jitter := range allJitters {
		if jitter > maxJitter {
			maxJitter = jitter
		}
	}

	jitterStdDev := calculateStandardDeviation(allJitters)

	// Calculate per-client jitter
	perClient := make(map[string]float64)
	for client, jitters := range clientJitters {
		if len(jitters) > 0 {
			perClient[client] = p.calculateMean(jitters)
		}
	}

	return &types.JitterMetrics{
		AvgJitter:    avgJitter,
		MaxJitter:    maxJitter,
		JitterStdDev: jitterStdDev,
		PerClient:    perClient,
	}, nil
}

// AssessNetworkQuality implements PerformanceAnalyzer.AssessNetworkQuality
func (p *DefaultPerformanceAnalyzer) AssessNetworkQuality(latency *types.LatencyMetrics, bandwidth *types.BandwidthMetrics, throughput *types.ThroughputMetrics, packetLoss *types.PacketLossMetrics, jitter *types.JitterMetrics, thresholds types.QualityThresholds) (*types.QualityAssessment, error) {
	assessment := &types.QualityAssessment{
		Recommendations: make([]string, 0),
		Issues:          make([]types.QualityIssue, 0),
	}

	scores := make([]float64, 0)

	// Assess latency
	latencyScore := 100.0
	if latency.AvgLatency > thresholds.MaxLatency {
		latencyScore = math.Max(0, 100-((latency.AvgLatency-thresholds.MaxLatency)/thresholds.MaxLatency)*100)
		assessment.Issues = append(assessment.Issues, types.QualityIssue{
			Type:        "latency",
			Severity:    p.getSeverity(latencyScore),
			Description: fmt.Sprintf("Average latency (%.2fms) exceeds threshold (%.2fms)", latency.AvgLatency, thresholds.MaxLatency),
			Impact:      "Network responsiveness may be degraded",
			Resolution:  "Check network infrastructure and reduce network congestion",
		})
		assessment.Recommendations = append(assessment.Recommendations, "Investigate high latency sources")
	}
	assessment.LatencyGrade = p.scoreToGrade(latencyScore)
	scores = append(scores, latencyScore)

	// Assess bandwidth
	bandwidthScore := 100.0
	if bandwidth.AvgBandwidth < thresholds.MinBandwidth {
		bandwidthScore = math.Max(0, (bandwidth.AvgBandwidth/thresholds.MinBandwidth)*100)
		assessment.Issues = append(assessment.Issues, types.QualityIssue{
			Type:        "bandwidth",
			Severity:    p.getSeverity(bandwidthScore),
			Description: fmt.Sprintf("Average bandwidth (%.2f Mbps) below minimum threshold (%.2f Mbps)", bandwidth.AvgBandwidth, thresholds.MinBandwidth),
			Impact:      "Network capacity may be insufficient",
			Resolution:  "Consider upgrading network bandwidth or optimizing traffic",
		})
		assessment.Recommendations = append(assessment.Recommendations, "Monitor bandwidth utilization patterns")
	}
	assessment.BandwidthGrade = p.scoreToGrade(bandwidthScore)
	scores = append(scores, bandwidthScore)

	// Assess reliability (packet loss)
	reliabilityScore := 100.0
	if packetLoss.LossPercentage > thresholds.MaxPacketLoss {
		reliabilityScore = math.Max(0, 100-((packetLoss.LossPercentage-thresholds.MaxPacketLoss)/thresholds.MaxPacketLoss)*100)
		assessment.Issues = append(assessment.Issues, types.QualityIssue{
			Type:        "packet_loss",
			Severity:    p.getSeverity(reliabilityScore),
			Description: fmt.Sprintf("Packet loss (%.2f%%) exceeds threshold (%.2f%%)", packetLoss.LossPercentage, thresholds.MaxPacketLoss),
			Impact:      "Connection reliability may be compromised",
			Resolution:  "Check network equipment and connections for errors",
		})
		assessment.Recommendations = append(assessment.Recommendations, "Investigate packet loss causes")
	}
	assessment.ReliabilityGrade = p.scoreToGrade(reliabilityScore)
	scores = append(scores, reliabilityScore)

	// Calculate overall grade
	overallScore := p.calculateMean(scores)
	assessment.OverallGrade = p.scoreToGrade(overallScore)

	// Add general recommendations
	if len(assessment.Issues) == 0 {
		assessment.Recommendations = append(assessment.Recommendations, "Network performance is within acceptable parameters")
	} else {
		assessment.Recommendations = append(assessment.Recommendations, "Regular monitoring recommended to track performance trends")
	}

	return assessment, nil
}

// Helper methods

// estimateLatency estimates latency from a DNS record
func (p *DefaultPerformanceAnalyzer) estimateLatency(record types.PiholeRecord) float64 {
	// Use ReplyTime if available, otherwise estimate based on timestamp processing
	if record.ReplyTime > 0 {
		return record.ReplyTime
	}
	
	// Estimate based on domain complexity and query type
	baseLatency := 10.0 // Base DNS query latency in ms
	
	// Add latency based on domain length (longer domains may take longer to resolve)
	domainLatency := float64(len(record.Domain)) * 0.1
	
	// Add latency based on query type
	typeLatency := 0.0
	switch record.QueryType {
	case "A":
		typeLatency = 0.0
	case "AAAA":
		typeLatency = 1.0
	case "MX":
		typeLatency = 2.0
	case "TXT":
		typeLatency = 3.0
	default:
		typeLatency = 1.0
	}
	
	return baseLatency + domainLatency + typeLatency
}

// estimateQueryBytes estimates bytes transferred for a DNS query
func (p *DefaultPerformanceAnalyzer) estimateQueryBytes(record types.PiholeRecord) int64 {
	// DNS header: 12 bytes
	// Question: domain name + 4 bytes (type + class)
	// Answer: varies by type
	
	baseSize := int64(12 + len(record.Domain) + 4)
	
	// Add estimated answer size
	switch record.QueryType {
	case "A":
		baseSize += 16 // A record response
	case "AAAA":
		baseSize += 28 // AAAA record response
	case "CNAME":
		baseSize += int64(len(record.Domain)) + 10
	case "MX":
		baseSize += int64(len(record.Domain)) + 14
	case "TXT":
		baseSize += 50 // Estimated TXT record size
	default:
		baseSize += 20 // Default response size
	}
	
	return baseSize
}

// estimateProcessingTime estimates query processing time
func (p *DefaultPerformanceAnalyzer) estimateProcessingTime(record types.PiholeRecord) float64 {
	// Base processing time
	processingTime := 0.5 // 0.5ms base
	
	// Add time based on domain complexity
	processingTime += float64(len(record.Domain)) * 0.01
	
	// Add time based on query type
	switch record.QueryType {
	case "A":
		processingTime += 0.1
	case "AAAA":
		processingTime += 0.2
	case "MX":
		processingTime += 0.3
	case "TXT":
		processingTime += 0.5
	default:
		processingTime += 0.2
	}
	
	return processingTime
}

// calculateTimeSpanFromRecords calculates time span from records
func (p *DefaultPerformanceAnalyzer) calculateTimeSpanFromRecords(records []types.PiholeRecord) time.Duration {
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
		return time.Hour // Minimum duration
	}

	return duration
}

// bytesToMbps converts bytes to Mbps
func (p *DefaultPerformanceAnalyzer) bytesToMbps(bytes int64, duration time.Duration) float64 {
	if duration.Seconds() == 0 {
		return 0
	}
	
	bytesPerSecond := float64(bytes) / duration.Seconds()
	bitsPerSecond := bytesPerSecond * 8
	mbps := bitsPerSecond / (1024 * 1024)
	
	return mbps
}

// calculateMean calculates the mean of a slice of float64 values
func (p *DefaultPerformanceAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	
	return sum / float64(len(values))
}

// createLatencyDistribution creates latency distribution buckets
func (p *DefaultPerformanceAnalyzer) createLatencyDistribution(latencies []float64) []types.LatencyBucket {
	buckets := []types.LatencyBucket{
		{RangeStart: 0, RangeEnd: 10, Count: 0},
		{RangeStart: 10, RangeEnd: 50, Count: 0},
		{RangeStart: 50, RangeEnd: 100, Count: 0},
		{RangeStart: 100, RangeEnd: 200, Count: 0},
		{RangeStart: 200, RangeEnd: 500, Count: 0},
		{RangeStart: 500, RangeEnd: math.Inf(1), Count: 0},
	}

	total := int64(len(latencies))
	
	for _, latency := range latencies {
		for i := range buckets {
			if latency >= buckets[i].RangeStart && latency < buckets[i].RangeEnd {
				buckets[i].Count++
				break
			}
		}
	}

	// Calculate percentages
	for i := range buckets {
		if total > 0 {
			buckets[i].Percentage = float64(buckets[i].Count) / float64(total) * 100
		}
	}

	return buckets
}

// calculateOverallScore calculates overall performance score
func (p *DefaultPerformanceAnalyzer) calculateOverallScore(assessment *types.QualityAssessment) float64 {
	grades := []string{assessment.LatencyGrade, assessment.BandwidthGrade, assessment.ReliabilityGrade}
	total := 0.0
	count := 0

	for _, grade := range grades {
		if grade != "" {
			total += p.gradeToScore(grade)
			count++
		}
	}

	if count == 0 {
		return 100.0
	}

	return total / float64(count)
}

// gradeToScore converts letter grade to numeric score
func (p *DefaultPerformanceAnalyzer) gradeToScore(grade string) float64 {
	switch grade {
	case "A":
		return 95.0
	case "B":
		return 85.0
	case "C":
		return 75.0
	case "D":
		return 65.0
	case "F":
		return 50.0
	default:
		return 75.0
	}
}

// scoreToGrade converts numeric score to letter grade
func (p *DefaultPerformanceAnalyzer) scoreToGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// getSeverity determines severity based on score
func (p *DefaultPerformanceAnalyzer) getSeverity(score float64) string {
	switch {
	case score < 50:
		return "CRITICAL"
	case score < 70:
		return "HIGH"
	case score < 85:
		return "MEDIUM"
	default:
		return "LOW"
	}
}