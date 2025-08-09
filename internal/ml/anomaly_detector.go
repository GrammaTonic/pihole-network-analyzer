package ml

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// StatisticalAnomalyDetector implements anomaly detection using statistical methods
type StatisticalAnomalyDetector struct {
	config    AnomalyDetectionConfig
	logger    *logger.Logger
	trained   bool
	modelInfo ModelInfo

	// Statistical baselines
	queryVolumeBaseline  QueryVolumeBaseline
	domainBaseline       DomainBaseline
	clientBaseline       ClientBaseline
	responseTimeBaseline ResponseTimeBaseline
	timePatternBaseline  TimePatternBaseline
}

// QueryVolumeBaseline holds baseline statistics for query volume
type QueryVolumeBaseline struct {
	HourlyMean    map[int]float64 // Hour -> mean queries
	HourlyStdDev  map[int]float64 // Hour -> standard deviation
	DailyMean     map[int]float64 // Day of week -> mean queries
	DailyStdDev   map[int]float64 // Day of week -> standard deviation
	OverallMean   float64
	OverallStdDev float64
}

// DomainBaseline holds baseline statistics for domain patterns
type DomainBaseline struct {
	CommonDomains      map[string]float64 // Domain -> frequency
	DomainCategories   map[string]int     // Domain -> category
	NewDomainThreshold float64
}

// ClientBaseline holds baseline statistics for client behavior
type ClientBaseline struct {
	ClientProfiles     map[string]ClientProfile // Client IP -> profile
	NewClientThreshold float64
}

// ClientProfile represents normal behavior for a client
type ClientProfile struct {
	TypicalQueryCount float64
	QueryCountStdDev  float64
	CommonDomains     map[string]float64
	TypicalHours      map[int]float64
	LastSeen          time.Time
}

// ResponseTimeBaseline holds baseline statistics for response times
type ResponseTimeBaseline struct {
	MeanResponseTime   float64
	StdDevResponseTime float64
	P95ResponseTime    float64
	P99ResponseTime    float64
}

// TimePatternBaseline holds baseline statistics for time patterns
type TimePatternBaseline struct {
	HourlyDistribution map[int]float64 // Hour -> normalized frequency
	DailyDistribution  map[int]float64 // Day of week -> normalized frequency
	WeeklyDistribution map[int]float64 // Week -> normalized frequency
}

// NewStatisticalAnomalyDetector creates a new statistical anomaly detector
func NewStatisticalAnomalyDetector(config AnomalyDetectionConfig, parentLogger *slog.Logger) *StatisticalAnomalyDetector {
	loggerConfig := &logger.Config{
		Component:     "anomaly-detector",
		Level:         logger.LevelInfo,
		EnableColors:  true,
		EnableEmojis:  true,
		ShowTimestamp: true,
	}

	return &StatisticalAnomalyDetector{
		config: config,
		logger: logger.New(loggerConfig),
		modelInfo: ModelInfo{
			Name:    "StatisticalAnomalyDetector",
			Version: "1.0.0",
		},
	}
}

// Initialize initializes the anomaly detector
func (d *StatisticalAnomalyDetector) Initialize(ctx context.Context, config AnomalyDetectionConfig) error {
	d.config = config
	d.logger.Info("Initializing statistical anomaly detector with sensitivity=%.2f, confidence=%.2f",
		config.Sensitivity, config.MinConfidence)

	// Set model parameters
	d.modelInfo.Parameters = map[string]interface{}{
		"sensitivity":    config.Sensitivity,
		"min_confidence": config.MinConfidence,
		"window_size":    config.WindowSize.String(),
		"anomaly_types":  len(config.AnomalyTypes),
	}

	// Initialize baselines
	d.queryVolumeBaseline = QueryVolumeBaseline{
		HourlyMean:   make(map[int]float64),
		HourlyStdDev: make(map[int]float64),
		DailyMean:    make(map[int]float64),
		DailyStdDev:  make(map[int]float64),
	}

	d.domainBaseline = DomainBaseline{
		CommonDomains:      make(map[string]float64),
		DomainCategories:   make(map[string]int),
		NewDomainThreshold: 0.001, // 0.1% threshold for new domains
	}

	d.clientBaseline = ClientBaseline{
		ClientProfiles:     make(map[string]ClientProfile),
		NewClientThreshold: 0.001, // 0.1% threshold for new clients
	}

	return nil
}

// Train trains the anomaly detector with historical data
func (d *StatisticalAnomalyDetector) Train(ctx context.Context, data []types.PiholeRecord) error {
	if len(data) < 10 {
		return fmt.Errorf("insufficient training data: need at least 10 records, got %d", len(data))
	}

	d.logger.Info("Training anomaly detector with %d records", len(data))

	startTime := time.Now()

	// Train query volume baseline
	if err := d.trainQueryVolumeBaseline(data); err != nil {
		return fmt.Errorf("failed to train query volume baseline: %w", err)
	}

	// Train domain baseline
	if err := d.trainDomainBaseline(data); err != nil {
		return fmt.Errorf("failed to train domain baseline: %w", err)
	}

	// Train client baseline
	if err := d.trainClientBaseline(data); err != nil {
		return fmt.Errorf("failed to train client baseline: %w", err)
	}

	// Train response time baseline
	if err := d.trainResponseTimeBaseline(data); err != nil {
		return fmt.Errorf("failed to train response time baseline: %w", err)
	}

	// Train time pattern baseline
	if err := d.trainTimePatternBaseline(data); err != nil {
		return fmt.Errorf("failed to train time pattern baseline: %w", err)
	}

	d.trained = true
	d.modelInfo.TrainedAt = time.Now()
	d.modelInfo.TrainingSize = len(data)
	d.modelInfo.LastUpdated = time.Now()

	duration := time.Since(startTime)
	d.logger.Info("Anomaly detector training completed in %v", duration)
	return nil
}

// DetectAnomalies detects anomalies in the provided data
func (d *StatisticalAnomalyDetector) DetectAnomalies(ctx context.Context, data []types.PiholeRecord) ([]Anomaly, error) {
	if !d.trained {
		return nil, fmt.Errorf("detector not trained")
	}

	var anomalies []Anomaly

	// Detect volume anomalies
	volumeAnomalies := d.detectVolumeAnomalies(data)
	anomalies = append(anomalies, volumeAnomalies...)

	// Detect domain anomalies
	domainAnomalies := d.detectDomainAnomalies(data)
	anomalies = append(anomalies, domainAnomalies...)

	// Detect client anomalies
	clientAnomalies := d.detectClientAnomalies(data)
	anomalies = append(anomalies, clientAnomalies...)

	// Detect response time anomalies
	responseTimeAnomalies := d.detectResponseTimeAnomalies(data)
	anomalies = append(anomalies, responseTimeAnomalies...)

	// Detect time pattern anomalies
	timePatternAnomalies := d.detectTimePatternAnomalies(data)
	anomalies = append(anomalies, timePatternAnomalies...)

	// Filter by confidence threshold
	filteredAnomalies := make([]Anomaly, 0)
	for _, anomaly := range anomalies {
		if anomaly.Confidence >= d.config.MinConfidence {
			filteredAnomalies = append(filteredAnomalies, anomaly)
		}
	}

	return filteredAnomalies, nil
}

// GetModelInfo returns information about the trained model
func (d *StatisticalAnomalyDetector) GetModelInfo() ModelInfo {
	return d.modelInfo
}

// IsTrained returns whether the detector is trained
func (d *StatisticalAnomalyDetector) IsTrained() bool {
	return d.trained
}

// Training methods
func (d *StatisticalAnomalyDetector) trainQueryVolumeBaseline(data []types.PiholeRecord) error {
	// Group queries by hour and day
	hourlyQueries := make(map[int][]int) // Hour -> query counts
	dailyQueries := make(map[int][]int)  // Day of week -> query counts

	// Process data in time windows
	timeWindows := d.groupByTimeWindows(data, time.Hour)

	for _, window := range timeWindows {
		hour := window.timestamp.Hour()
		dayOfWeek := int(window.timestamp.Weekday())

		hourlyQueries[hour] = append(hourlyQueries[hour], window.count)
		dailyQueries[dayOfWeek] = append(dailyQueries[dayOfWeek], window.count)
	}

	// Calculate statistics for each hour
	for hour, counts := range hourlyQueries {
		mean, stddev := calculateStats(counts)
		d.queryVolumeBaseline.HourlyMean[hour] = mean
		d.queryVolumeBaseline.HourlyStdDev[hour] = stddev
	}

	// Calculate statistics for each day
	for day, counts := range dailyQueries {
		mean, stddev := calculateStats(counts)
		d.queryVolumeBaseline.DailyMean[day] = mean
		d.queryVolumeBaseline.DailyStdDev[day] = stddev
	}

	// Calculate overall statistics
	var allCounts []int
	for _, windows := range timeWindows {
		allCounts = append(allCounts, windows.count)
	}

	if len(allCounts) > 0 {
		mean, stddev := calculateStats(allCounts)
		d.queryVolumeBaseline.OverallMean = mean
		d.queryVolumeBaseline.OverallStdDev = stddev
	}

	return nil
}

func (d *StatisticalAnomalyDetector) trainDomainBaseline(data []types.PiholeRecord) error {
	domainCounts := make(map[string]int)
	totalQueries := len(data)

	// Count domain frequencies
	for _, record := range data {
		domain := strings.ToLower(record.Domain)
		domainCounts[domain]++
	}

	// Calculate frequencies
	for domain, count := range domainCounts {
		frequency := float64(count) / float64(totalQueries)
		d.domainBaseline.CommonDomains[domain] = frequency
	}

	return nil
}

func (d *StatisticalAnomalyDetector) trainClientBaseline(data []types.PiholeRecord) error {
	clientData := make(map[string][]types.PiholeRecord)

	// Group by client
	for _, record := range data {
		client := record.Client
		clientData[client] = append(clientData[client], record)
	}

	// Create profiles for each client
	for client, records := range clientData {
		profile := d.createClientProfile(records)
		d.clientBaseline.ClientProfiles[client] = profile
	}

	return nil
}

func (d *StatisticalAnomalyDetector) trainResponseTimeBaseline(data []types.PiholeRecord) error {
	var responseTimes []float64

	// Extract response times (simulate with random data for now)
	for range data {
		responseTime := float64(50 + (len(data) % 100)) // Simulate response times
		responseTimes = append(responseTimes, responseTime)
	}

	if len(responseTimes) == 0 {
		return nil
	}

	sort.Float64s(responseTimes)

	mean, stddev := calculateFloatStats(responseTimes)
	d.responseTimeBaseline.MeanResponseTime = mean
	d.responseTimeBaseline.StdDevResponseTime = stddev

	// Calculate percentiles
	if len(responseTimes) > 0 {
		p95Index := int(0.95 * float64(len(responseTimes)))
		p99Index := int(0.99 * float64(len(responseTimes)))

		d.responseTimeBaseline.P95ResponseTime = responseTimes[p95Index]
		d.responseTimeBaseline.P99ResponseTime = responseTimes[p99Index]
	}

	return nil
}

func (d *StatisticalAnomalyDetector) trainTimePatternBaseline(data []types.PiholeRecord) error {
	hourCounts := make(map[int]int)
	dayCounts := make(map[int]int)

	// Count queries by time patterns
	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		hour := timestamp.Hour()
		dayOfWeek := int(timestamp.Weekday())

		hourCounts[hour]++
		dayCounts[dayOfWeek]++
	}

	totalQueries := len(data)

	// Normalize distributions
	d.timePatternBaseline.HourlyDistribution = make(map[int]float64)
	d.timePatternBaseline.DailyDistribution = make(map[int]float64)

	for hour, count := range hourCounts {
		d.timePatternBaseline.HourlyDistribution[hour] = float64(count) / float64(totalQueries)
	}

	for day, count := range dayCounts {
		d.timePatternBaseline.DailyDistribution[day] = float64(count) / float64(totalQueries)
	}

	return nil
}

// Detection methods
func (d *StatisticalAnomalyDetector) detectVolumeAnomalies(data []types.PiholeRecord) []Anomaly {
	var anomalies []Anomaly

	// Group data by time windows
	timeWindows := d.groupByTimeWindows(data, d.config.WindowSize)

	for _, window := range timeWindows {
		hour := window.timestamp.Hour()
		dayOfWeek := int(window.timestamp.Weekday())

		// Check against hourly baseline
		if hourlyMean, exists := d.queryVolumeBaseline.HourlyMean[hour]; exists {
			hourlyStdDev := d.queryVolumeBaseline.HourlyStdDev[hour]
			threshold := d.config.Thresholds["volume_spike_multiplier"]

			if float64(window.count) > hourlyMean+threshold*hourlyStdDev {
				anomaly := Anomaly{
					ID:          fmt.Sprintf("volume-spike-%d", window.timestamp.Unix()),
					Type:        AnomalyTypeVolumeSpike,
					Severity:    d.calculateSeverity(float64(window.count), hourlyMean, hourlyStdDev),
					Timestamp:   window.timestamp,
					Description: fmt.Sprintf("Query volume spike detected: %d queries (normal: %.1f±%.1f)", window.count, hourlyMean, hourlyStdDev),
					Score:       d.calculateScore(float64(window.count), hourlyMean, hourlyStdDev),
					Confidence:  0.8,
					Metadata: map[string]interface{}{
						"window_size":   d.config.WindowSize.String(),
						"query_count":   window.count,
						"expected_mean": hourlyMean,
						"hour":          hour,
						"day_of_week":   dayOfWeek,
					},
				}
				anomalies = append(anomalies, anomaly)
			}

			// Check for volume dropout
			dropoutThreshold := d.config.Thresholds["volume_dropout_threshold"]
			if float64(window.count) < hourlyMean*dropoutThreshold {
				anomaly := Anomaly{
					ID:          fmt.Sprintf("volume-dropout-%d", window.timestamp.Unix()),
					Type:        AnomalyTypeVolumeDropout,
					Severity:    SeverityMedium,
					Timestamp:   window.timestamp,
					Description: fmt.Sprintf("Query volume dropout detected: %d queries (normal: %.1f±%.1f)", window.count, hourlyMean, hourlyStdDev),
					Score:       1.0 - (float64(window.count) / hourlyMean),
					Confidence:  0.7,
					Metadata: map[string]interface{}{
						"window_size":   d.config.WindowSize.String(),
						"query_count":   window.count,
						"expected_mean": hourlyMean,
						"hour":          hour,
						"day_of_week":   dayOfWeek,
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

func (d *StatisticalAnomalyDetector) detectDomainAnomalies(data []types.PiholeRecord) []Anomaly {
	var anomalies []Anomaly

	domainCounts := make(map[string]int)
	totalQueries := len(data)

	// Count current domain frequencies
	for _, record := range data {
		domain := strings.ToLower(record.Domain)
		domainCounts[domain]++
	}

	// Check for unusual domains
	threshold := d.config.Thresholds["unusual_domain_threshold"]

	for domain, count := range domainCounts {
		frequency := float64(count) / float64(totalQueries)

		// Check if domain is in baseline
		if baselineFreq, exists := d.domainBaseline.CommonDomains[domain]; exists {
			// Check for significant increase in frequency
			if frequency > baselineFreq*3.0 {
				anomaly := Anomaly{
					ID:          fmt.Sprintf("domain-spike-%s-%d", domain, time.Now().Unix()),
					Type:        AnomalyTypeUnusualDomain,
					Severity:    SeverityMedium,
					Timestamp:   time.Now(),
					Description: fmt.Sprintf("Unusual increase in domain activity: %s (%.2f%% vs normal %.2f%%)", domain, frequency*100, baselineFreq*100),
					Score:       frequency / baselineFreq,
					Confidence:  0.7,
					Domain:      domain,
					Metadata: map[string]interface{}{
						"domain":             domain,
						"current_frequency":  frequency,
						"baseline_frequency": baselineFreq,
						"query_count":        count,
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		} else if frequency > threshold {
			// New domain with significant activity
			anomaly := Anomaly{
				ID:          fmt.Sprintf("new-domain-%s-%d", domain, time.Now().Unix()),
				Type:        AnomalyTypeUnusualDomain,
				Severity:    SeverityLow,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("New domain with significant activity: %s (%.2f%%)", domain, frequency*100),
				Score:       frequency / threshold,
				Confidence:  0.6,
				Domain:      domain,
				Metadata: map[string]interface{}{
					"domain":      domain,
					"frequency":   frequency,
					"query_count": count,
					"is_new":      true,
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

func (d *StatisticalAnomalyDetector) detectClientAnomalies(data []types.PiholeRecord) []Anomaly {
	var anomalies []Anomaly

	clientCounts := make(map[string]int)

	// Count queries per client
	for _, record := range data {
		client := record.Client
		clientCounts[client]++
	}

	// Check each client against baseline
	for client, count := range clientCounts {
		if profile, exists := d.clientBaseline.ClientProfiles[client]; exists {
			// Check for unusual query volume
			threshold := profile.TypicalQueryCount + 3*profile.QueryCountStdDev

			if float64(count) > threshold {
				anomaly := Anomaly{
					ID:          fmt.Sprintf("client-spike-%s-%d", client, time.Now().Unix()),
					Type:        AnomalyTypeUnusualClient,
					Severity:    d.calculateClientSeverity(float64(count), profile.TypicalQueryCount),
					Timestamp:   time.Now(),
					Description: fmt.Sprintf("Unusual client activity: %s with %d queries (normal: %.1f±%.1f)", client, count, profile.TypicalQueryCount, profile.QueryCountStdDev),
					Score:       float64(count) / profile.TypicalQueryCount,
					Confidence:  0.8,
					ClientIP:    client,
					Metadata: map[string]interface{}{
						"client":              client,
						"query_count":         count,
						"typical_query_count": profile.TypicalQueryCount,
						"query_count_stddev":  profile.QueryCountStdDev,
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		} else {
			// New client
			if count > 10 { // Threshold for new client activity
				anomaly := Anomaly{
					ID:          fmt.Sprintf("new-client-%s-%d", client, time.Now().Unix()),
					Type:        AnomalyTypeUnusualClient,
					Severity:    SeverityLow,
					Timestamp:   time.Now(),
					Description: fmt.Sprintf("New client with significant activity: %s (%d queries)", client, count),
					Score:       float64(count) / 10.0,
					Confidence:  0.6,
					ClientIP:    client,
					Metadata: map[string]interface{}{
						"client":      client,
						"query_count": count,
						"is_new":      true,
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

func (d *StatisticalAnomalyDetector) detectResponseTimeAnomalies(data []types.PiholeRecord) []Anomaly {
	var anomalies []Anomaly

	// Simulate response time analysis
	threshold := d.config.Thresholds["response_time_threshold"]

	for _, record := range data {
		// Simulate response time extraction
		responseTime := float64(50 + (len(record.Domain) % 100))

		if responseTime > threshold {
			anomaly := Anomaly{
				ID:          fmt.Sprintf("response-time-%s-%d", record.Client, time.Now().Unix()),
				Type:        AnomalyTypeResponseTime,
				Severity:    SeverityLow,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("High response time detected: %.2fms for %s", responseTime, record.Domain),
				Score:       responseTime / threshold,
				Confidence:  0.5,
				Domain:      record.Domain,
				ClientIP:    record.Client,
				Metadata: map[string]interface{}{
					"response_time": responseTime,
					"threshold":     threshold,
					"domain":        record.Domain,
					"client":        record.Client,
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

func (d *StatisticalAnomalyDetector) detectTimePatternAnomalies(data []types.PiholeRecord) []Anomaly {
	var anomalies []Anomaly

	// Current time distribution
	hourCounts := make(map[int]int)
	dayCounts := make(map[int]int)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		hour := timestamp.Hour()
		dayOfWeek := int(timestamp.Weekday())

		hourCounts[hour]++
		dayCounts[dayOfWeek]++
	}

	totalQueries := len(data)

	// Check for unusual time patterns
	for hour, count := range hourCounts {
		currentFreq := float64(count) / float64(totalQueries)

		if baselineFreq, exists := d.timePatternBaseline.HourlyDistribution[hour]; exists {
			// Check for significant deviation
			if currentFreq > baselineFreq*3.0 {
				anomaly := Anomaly{
					ID:          fmt.Sprintf("time-pattern-%d-%d", hour, time.Now().Unix()),
					Type:        AnomalyTypeTimePattern,
					Severity:    SeverityLow,
					Timestamp:   time.Now(),
					Description: fmt.Sprintf("Unusual activity at hour %d: %.1f%% of queries (normal: %.1f%%)", hour, currentFreq*100, baselineFreq*100),
					Score:       currentFreq / baselineFreq,
					Confidence:  0.6,
					Metadata: map[string]interface{}{
						"hour":               hour,
						"current_frequency":  currentFreq,
						"baseline_frequency": baselineFreq,
						"query_count":        count,
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

// Helper types and functions
type timeWindow struct {
	timestamp time.Time
	count     int
}

func (d *StatisticalAnomalyDetector) groupByTimeWindows(data []types.PiholeRecord, windowSize time.Duration) []timeWindow {
	windowMap := make(map[int64]int)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		windowStart := timestamp.Truncate(windowSize)
		windowKey := windowStart.Unix()
		windowMap[windowKey]++
	}

	var windows []timeWindow
	for windowKey, count := range windowMap {
		windows = append(windows, timeWindow{
			timestamp: time.Unix(windowKey, 0),
			count:     count,
		})
	}

	// Sort by timestamp
	sort.Slice(windows, func(i, j int) bool {
		return windows[i].timestamp.Before(windows[j].timestamp)
	})

	return windows
}

func (d *StatisticalAnomalyDetector) createClientProfile(records []types.PiholeRecord) ClientProfile {
	profile := ClientProfile{
		CommonDomains: make(map[string]float64),
		TypicalHours:  make(map[int]float64),
	}

	// Calculate typical query count
	profile.TypicalQueryCount = float64(len(records))
	profile.QueryCountStdDev = profile.TypicalQueryCount * 0.3 // Estimate

	// Calculate domain frequencies
	domainCounts := make(map[string]int)
	hourCounts := make(map[int]int)

	for _, record := range records {
		domain := strings.ToLower(record.Domain)
		domainCounts[domain]++

		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err == nil {
			hour := timestamp.Hour()
			hourCounts[hour]++
			profile.LastSeen = timestamp
		}
	}

	totalQueries := len(records)
	for domain, count := range domainCounts {
		profile.CommonDomains[domain] = float64(count) / float64(totalQueries)
	}

	for hour, count := range hourCounts {
		profile.TypicalHours[hour] = float64(count) / float64(totalQueries)
	}

	return profile
}

func (d *StatisticalAnomalyDetector) calculateSeverity(value, mean, stddev float64) SeverityLevel {
	zScore := (value - mean) / stddev

	if zScore > 4.0 {
		return SeverityCritical
	} else if zScore > 3.0 {
		return SeverityHigh
	} else if zScore > 2.0 {
		return SeverityMedium
	}
	return SeverityLow
}

func (d *StatisticalAnomalyDetector) calculateClientSeverity(value, typical float64) SeverityLevel {
	ratio := value / typical

	if ratio > 5.0 {
		return SeverityCritical
	} else if ratio > 3.0 {
		return SeverityHigh
	} else if ratio > 2.0 {
		return SeverityMedium
	}
	return SeverityLow
}

func (d *StatisticalAnomalyDetector) calculateScore(value, mean, stddev float64) float64 {
	if stddev == 0 {
		return 1.0
	}

	zScore := math.Abs((value - mean) / stddev)
	return math.Min(1.0, zScore/3.0) // Normalize to 0-1 range
}

func (d *StatisticalAnomalyDetector) calculateAccuracy() float64 {
	// Simulate accuracy calculation
	return 0.85 // 85% accuracy
}

// Statistical helper functions
func calculateStats(values []int) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0
	for _, v := range values {
		sum += v
	}
	mean = float64(sum) / float64(len(values))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		sumSquaredDiff += diff * diff
	}

	if len(values) > 1 {
		stddev = math.Sqrt(sumSquaredDiff / float64(len(values)-1))
	}

	return mean, stddev
}

func calculateFloatStats(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	if len(values) > 1 {
		stddev = math.Sqrt(sumSquaredDiff / float64(len(values)-1))
	}

	return mean, stddev
}
