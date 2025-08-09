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

// TrendAnalyzerImpl implements trend analysis using statistical methods
type TrendAnalyzerImpl struct {
	config       TrendAnalysisConfig
	logger       *logger.Logger
	analyzerInfo AnalyzerInfo
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(config TrendAnalysisConfig, parentLogger *slog.Logger) *TrendAnalyzerImpl {
	loggerConfig := &logger.Config{
		Component:     "trend-analyzer",
		Level:         logger.LevelInfo,
		EnableColors:  true,
		EnableEmojis:  true,
		ShowTimestamp: true,
	}

	return &TrendAnalyzerImpl{
		config: config,
		logger: logger.New(loggerConfig),
		analyzerInfo: AnalyzerInfo{
			Name:       "StatisticalTrendAnalyzer",
			Version:    "1.0.0",
			Algorithms: []string{"moving_average", "linear_regression", "seasonal_decomposition"},
			Parameters: make(map[string]interface{}),
		},
	}
}

// Initialize initializes the trend analyzer
func (t *TrendAnalyzerImpl) Initialize(ctx context.Context, config TrendAnalysisConfig) error {
	t.config = config
	t.logger.Info("Initializing trend analyzer with min_data_points=%d, smoothing_factor=%.2f",
		config.MinDataPoints, config.SmoothingFactor)

	t.analyzerInfo.Parameters = map[string]interface{}{
		"analysis_window":  config.AnalysisWindow.String(),
		"forecast_window":  config.ForecastWindow.String(),
		"min_data_points":  config.MinDataPoints,
		"smoothing_factor": config.SmoothingFactor,
	}

	return nil
}

// AnalyzeTrends analyzes trends in DNS query patterns
func (t *TrendAnalyzerImpl) AnalyzeTrends(ctx context.Context, data []types.PiholeRecord, timeWindow time.Duration) (*TrendAnalysis, error) {
	if len(data) < t.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points: need at least %d, got %d", t.config.MinDataPoints, len(data))
	}

	t.logger.Info("Analyzing trends for %d records over %v", len(data), timeWindow)

	analysis := &TrendAnalysis{
		TimeWindow:     timeWindow,
		TotalQueries:   len(data),
		HourlyPatterns: make(map[int]float64),
		DailyPatterns:  make(map[time.Weekday]float64),
		Timestamp:      time.Now(),
	}

	// Analyze overall query trend
	analysis.QueryTrend = t.analyzeQueryTrend(data, timeWindow)

	// Analyze domain trends
	analysis.DomainTrends = t.analyzeDomainTrends(data)

	// Analyze client trends
	analysis.ClientTrends = t.analyzeClientTrends(data)

	// Analyze time patterns
	analysis.HourlyPatterns = t.analyzeHourlyPatterns(data)
	analysis.DailyPatterns = t.analyzeDailyPatterns(data)

	// Generate insights
	analysis.Insights = t.generateInsights(analysis)

	t.analyzerInfo.LastRun = time.Now()

	t.logger.Info("Trend analysis completed with %d insights", len(analysis.Insights))
	return analysis, nil
}

// PredictTrends predicts future query patterns
func (t *TrendAnalyzerImpl) PredictTrends(ctx context.Context, data []types.PiholeRecord, forecastWindow time.Duration) (*TrendPrediction, error) {
	if len(data) < t.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points for prediction: need at least %d, got %d", t.config.MinDataPoints, len(data))
	}

	t.logger.Info("Predicting trends for %v based on %d records", forecastWindow, len(data)) // Group data by time intervals
	timeSeriesData := t.createTimeSeries(data, time.Hour)

	// Apply smoothing
	smoothedData := t.applyExponentialSmoothing(timeSeriesData, t.config.SmoothingFactor)

	// Generate predictions
	forecastPoints := int(forecastWindow.Hours())
	predictions := t.generateForecasts(smoothedData, forecastPoints)

	// Calculate confidence
	confidence := t.calculatePredictionConfidence(timeSeriesData, smoothedData)

	prediction := &TrendPrediction{
		ForecastWindow:   forecastWindow,
		PredictedQueries: predictions,
		Confidence:       confidence,
		Methodology:      "Exponential Smoothing with Linear Trend",
		Timestamp:        time.Now(),
	}

	t.logger.Info("Trend prediction completed with confidence %.2f", prediction.Confidence)
	return prediction, nil
}

// GetAnalyzerInfo returns information about the trend analyzer
func (t *TrendAnalyzerImpl) GetAnalyzerInfo() AnalyzerInfo {
	return t.analyzerInfo
}

// Analysis methods
func (t *TrendAnalyzerImpl) analyzeQueryTrend(data []types.PiholeRecord, timeWindow time.Duration) TrendDirection {
	// Create time series data
	timeSeries := t.createTimeSeries(data, time.Hour)

	if len(timeSeries) < 3 {
		return TrendStable
	}

	// Calculate linear trend using least squares
	slope := t.calculateLinearTrend(timeSeries)
	variance := t.calculateVariance(timeSeries)

	// Determine trend direction based on slope and variance
	slopeThreshold := 0.1 * t.calculateMean(timeSeries)
	varianceThreshold := 0.5 * t.calculateMean(timeSeries)

	if variance > varianceThreshold {
		return TrendVolatile
	} else if slope > slopeThreshold {
		return TrendIncreasing
	} else if slope < -slopeThreshold {
		return TrendDecreasing
	}

	return TrendStable
}

func (t *TrendAnalyzerImpl) analyzeDomainTrends(data []types.PiholeRecord) []DomainTrend {
	// Group data by domain and time
	domainTimeSeries := make(map[string][]timeSeriesPoint)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		domain := strings.ToLower(record.Domain)
		hourKey := timestamp.Truncate(time.Hour)

		if _, exists := domainTimeSeries[domain]; !exists {
			domainTimeSeries[domain] = make([]timeSeriesPoint, 0)
		}

		// Find or create time point
		found := false
		for i, point := range domainTimeSeries[domain] {
			if point.Timestamp.Equal(hourKey) {
				domainTimeSeries[domain][i].Value++
				found = true
				break
			}
		}

		if !found {
			domainTimeSeries[domain] = append(domainTimeSeries[domain], timeSeriesPoint{
				Timestamp: hourKey,
				Value:     1,
			})
		}
	}

	var domainTrends []DomainTrend

	// Analyze trend for each domain with significant activity
	for domain, timeSeries := range domainTimeSeries {
		totalQueries := 0
		for _, point := range timeSeries {
			totalQueries += point.Value
		}

		// Only analyze domains with sufficient activity
		if totalQueries < 5 {
			continue
		}

		// Sort time series by timestamp
		sort.Slice(timeSeries, func(i, j int) bool {
			return timeSeries[i].Timestamp.Before(timeSeries[j].Timestamp)
		})

		// Calculate trend
		slope := t.calculateLinearTrendFromPoints(timeSeries)
		mean := float64(totalQueries) / float64(len(timeSeries))
		changePercentage := (slope / mean) * 100

		var trendDirection TrendDirection
		if math.Abs(changePercentage) < 5 {
			trendDirection = TrendStable
		} else if changePercentage > 0 {
			trendDirection = TrendIncreasing
		} else {
			trendDirection = TrendDecreasing
		}

		domainTrend := DomainTrend{
			Domain:    domain,
			Trend:     trendDirection,
			Change:    changePercentage,
			Queries:   totalQueries,
			IsBlocked: t.isDomainBlocked(domain, data),
		}

		domainTrends = append(domainTrends, domainTrend)
	}

	// Sort by query count (most active first)
	sort.Slice(domainTrends, func(i, j int) bool {
		return domainTrends[i].Queries > domainTrends[j].Queries
	})

	// Return top 20 domains
	maxTrends := 20
	if len(domainTrends) > maxTrends {
		domainTrends = domainTrends[:maxTrends]
	}

	return domainTrends
}

func (t *TrendAnalyzerImpl) analyzeClientTrends(data []types.PiholeRecord) []ClientTrend {
	// Group data by client and time
	clientTimeSeries := make(map[string][]timeSeriesPoint)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		client := record.Client
		hourKey := timestamp.Truncate(time.Hour)

		if _, exists := clientTimeSeries[client]; !exists {
			clientTimeSeries[client] = make([]timeSeriesPoint, 0)
		}

		// Find or create time point
		found := false
		for i, point := range clientTimeSeries[client] {
			if point.Timestamp.Equal(hourKey) {
				clientTimeSeries[client][i].Value++
				found = true
				break
			}
		}

		if !found {
			clientTimeSeries[client] = append(clientTimeSeries[client], timeSeriesPoint{
				Timestamp: hourKey,
				Value:     1,
			})
		}
	}

	var clientTrends []ClientTrend

	// Analyze trend for each client with significant activity
	for client, timeSeries := range clientTimeSeries {
		totalQueries := 0
		for _, point := range timeSeries {
			totalQueries += point.Value
		}

		// Only analyze clients with sufficient activity
		if totalQueries < 10 {
			continue
		}

		// Sort time series by timestamp
		sort.Slice(timeSeries, func(i, j int) bool {
			return timeSeries[i].Timestamp.Before(timeSeries[j].Timestamp)
		})

		// Calculate trend
		slope := t.calculateLinearTrendFromPoints(timeSeries)
		mean := float64(totalQueries) / float64(len(timeSeries))
		changePercentage := (slope / mean) * 100

		var trendDirection TrendDirection
		if math.Abs(changePercentage) < 10 {
			trendDirection = TrendStable
		} else if changePercentage > 0 {
			trendDirection = TrendIncreasing
		} else {
			trendDirection = TrendDecreasing
		}

		clientTrend := ClientTrend{
			ClientIP: client,
			Trend:    trendDirection,
			Change:   changePercentage,
			Queries:  totalQueries,
		}

		clientTrends = append(clientTrends, clientTrend)
	}

	// Sort by query count (most active first)
	sort.Slice(clientTrends, func(i, j int) bool {
		return clientTrends[i].Queries > clientTrends[j].Queries
	})

	// Return top 15 clients
	maxTrends := 15
	if len(clientTrends) > maxTrends {
		clientTrends = clientTrends[:maxTrends]
	}

	return clientTrends
}

func (t *TrendAnalyzerImpl) analyzeHourlyPatterns(data []types.PiholeRecord) map[int]float64 {
	hourCounts := make(map[int]int)
	totalQueries := len(data)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		hour := timestamp.Hour()
		hourCounts[hour]++
	}

	// Normalize to percentages
	hourlyPatterns := make(map[int]float64)
	for hour, count := range hourCounts {
		hourlyPatterns[hour] = float64(count) / float64(totalQueries)
	}

	// Fill missing hours with 0
	for hour := 0; hour < 24; hour++ {
		if _, exists := hourlyPatterns[hour]; !exists {
			hourlyPatterns[hour] = 0.0
		}
	}

	return hourlyPatterns
}

func (t *TrendAnalyzerImpl) analyzeDailyPatterns(data []types.PiholeRecord) map[time.Weekday]float64 {
	dayCounts := make(map[time.Weekday]int)
	totalQueries := len(data)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		day := timestamp.Weekday()
		dayCounts[day]++
	}

	// Normalize to percentages
	dailyPatterns := make(map[time.Weekday]float64)
	for day, count := range dayCounts {
		dailyPatterns[day] = float64(count) / float64(totalQueries)
	}

	// Fill missing days with 0
	for day := time.Sunday; day <= time.Saturday; day++ {
		if _, exists := dailyPatterns[day]; !exists {
			dailyPatterns[day] = 0.0
		}
	}

	return dailyPatterns
}

func (t *TrendAnalyzerImpl) generateInsights(analysis *TrendAnalysis) []TrendInsight {
	var insights []TrendInsight

	// Peak usage insight
	maxHour, maxUsage := t.findPeakHour(analysis.HourlyPatterns)
	if maxUsage > 0.1 { // More than 10% of queries in one hour
		insights = append(insights, TrendInsight{
			Type:        InsightTypePeakUsage,
			Description: fmt.Sprintf("Peak usage occurs at %02d:00 with %.1f%% of daily queries", maxHour, maxUsage*100),
			Impact:      "High",
			Confidence:  0.9,
		})
	}

	// Off-peak usage insight
	minHour, minUsage := t.findOffPeakHour(analysis.HourlyPatterns)
	if minUsage < 0.01 { // Less than 1% of queries in one hour
		insights = append(insights, TrendInsight{
			Type:        InsightTypeOffPeakUsage,
			Description: fmt.Sprintf("Lowest usage occurs at %02d:00 with %.1f%% of daily queries", minHour, minUsage*100),
			Impact:      "Low",
			Confidence:  0.8,
		})
	}

	// Weekend pattern insight
	weekendRatio := t.calculateWeekendRatio(analysis.DailyPatterns)
	if weekendRatio < 0.7 {
		insights = append(insights, TrendInsight{
			Type:        InsightTypeWeekendPattern,
			Description: fmt.Sprintf("Weekend usage is %.1f%% of weekday usage, indicating business-focused network", weekendRatio*100),
			Impact:      "Medium",
			Confidence:  0.8,
		})
	} else if weekendRatio > 1.3 {
		insights = append(insights, TrendInsight{
			Type:        InsightTypeWeekendPattern,
			Description: fmt.Sprintf("Weekend usage is %.1f%% of weekday usage, indicating personal-focused network", weekendRatio*100),
			Impact:      "Medium",
			Confidence:  0.8,
		})
	}

	// Anomalous client insight
	for _, clientTrend := range analysis.ClientTrends {
		if math.Abs(clientTrend.Change) > 50 {
			insights = append(insights, TrendInsight{
				Type:        InsightTypeAnomalousClient,
				Description: fmt.Sprintf("Client %s shows %.1f%% change in query volume", clientTrend.ClientIP, clientTrend.Change),
				Impact:      "High",
				Confidence:  0.7,
			})
		}
	}

	// Emerging threat insight
	for _, domainTrend := range analysis.DomainTrends {
		if domainTrend.Change > 100 && domainTrend.Queries > 20 {
			insights = append(insights, TrendInsight{
				Type:        InsightTypeEmergingThreat,
				Description: fmt.Sprintf("Domain %s shows rapid increase (%.1f%%) - potential emerging threat", domainTrend.Domain, domainTrend.Change),
				Impact:      "High",
				Confidence:  0.6,
			})
		}
	}

	return insights
}

// Prediction methods
func (t *TrendAnalyzerImpl) createTimeSeries(data []types.PiholeRecord, interval time.Duration) []timeSeriesPoint {
	timeMap := make(map[int64]int)

	for _, record := range data {
		timestamp, err := time.Parse("2006-01-02 15:04:05", record.DateTime)
		if err != nil {
			continue
		}

		intervalStart := timestamp.Truncate(interval)
		key := intervalStart.Unix()
		timeMap[key]++
	}

	var timeSeries []timeSeriesPoint
	for timestamp, count := range timeMap {
		timeSeries = append(timeSeries, timeSeriesPoint{
			Timestamp: time.Unix(timestamp, 0),
			Value:     count,
		})
	}

	// Sort by timestamp
	sort.Slice(timeSeries, func(i, j int) bool {
		return timeSeries[i].Timestamp.Before(timeSeries[j].Timestamp)
	})

	return timeSeries
}

func (t *TrendAnalyzerImpl) applyExponentialSmoothing(data []timeSeriesPoint, alpha float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	smoothed := make([]float64, len(data))
	smoothed[0] = float64(data[0].Value)

	for i := 1; i < len(data); i++ {
		smoothed[i] = alpha*float64(data[i].Value) + (1-alpha)*smoothed[i-1]
	}

	return smoothed
}

func (t *TrendAnalyzerImpl) generateForecasts(smoothedData []float64, forecastPoints int) []QueryForecast {
	if len(smoothedData) < 2 {
		return []QueryForecast{}
	}

	// Calculate trend (simple linear trend from last few points)
	trend := t.calculateTrendFromSmoothed(smoothedData)
	lastValue := smoothedData[len(smoothedData)-1]

	var forecasts []QueryForecast
	lastTimestamp := time.Now()

	for i := 1; i <= forecastPoints; i++ {
		predictedValue := lastValue + trend*float64(i)

		// Add some uncertainty bounds
		uncertainty := math.Sqrt(float64(i)) * 5.0 // Increasing uncertainty over time

		forecast := QueryForecast{
			Timestamp:      lastTimestamp.Add(time.Duration(i) * time.Hour),
			PredictedCount: int(math.Max(0, predictedValue)),
			ConfidenceInterval: ConfidenceInterval{
				Lower: math.Max(0, predictedValue-uncertainty),
				Upper: predictedValue + uncertainty,
			},
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts
}

func (t *TrendAnalyzerImpl) calculatePredictionConfidence(original []timeSeriesPoint, smoothed []float64) float64 {
	if len(original) != len(smoothed) || len(original) < 2 {
		return 0.5
	}

	// Calculate Mean Absolute Percentage Error (MAPE)
	var totalError float64
	validPoints := 0

	for i, point := range original {
		if point.Value > 0 {
			error := math.Abs(float64(point.Value)-smoothed[i]) / float64(point.Value)
			totalError += error
			validPoints++
		}
	}

	if validPoints == 0 {
		return 0.5
	}

	mape := totalError / float64(validPoints)
	confidence := math.Max(0.1, 1.0-mape) // Convert MAPE to confidence

	return math.Min(0.95, confidence) // Cap at 95%
}

// Helper types and functions
type timeSeriesPoint struct {
	Timestamp time.Time
	Value     int
}

func (t *TrendAnalyzerImpl) calculateLinearTrend(timeSeries []timeSeriesPoint) float64 {
	if len(timeSeries) < 2 {
		return 0
	}

	n := len(timeSeries)

	// Convert to simple arrays for calculation
	var x, y []float64
	for i, point := range timeSeries {
		x = append(x, float64(i))
		y = append(y, float64(point.Value))
	}

	// Calculate slope using least squares
	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0

	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
	}

	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)

	return slope
}

func (t *TrendAnalyzerImpl) calculateLinearTrendFromPoints(points []timeSeriesPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// Simple slope calculation between first and last points
	firstValue := float64(points[0].Value)
	lastValue := float64(points[len(points)-1].Value)
	timeSpan := points[len(points)-1].Timestamp.Sub(points[0].Timestamp).Hours()

	if timeSpan == 0 {
		return 0
	}

	return (lastValue - firstValue) / timeSpan
}

func (t *TrendAnalyzerImpl) calculateVariance(timeSeries []timeSeriesPoint) float64 {
	if len(timeSeries) < 2 {
		return 0
	}

	mean := t.calculateMean(timeSeries)
	var sumSquaredDiff float64

	for _, point := range timeSeries {
		diff := float64(point.Value) - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(timeSeries)-1)
}

func (t *TrendAnalyzerImpl) calculateMean(timeSeries []timeSeriesPoint) float64 {
	if len(timeSeries) == 0 {
		return 0
	}

	var sum float64
	for _, point := range timeSeries {
		sum += float64(point.Value)
	}

	return sum / float64(len(timeSeries))
}

func (t *TrendAnalyzerImpl) calculateTrendFromSmoothed(smoothed []float64) float64 {
	if len(smoothed) < 2 {
		return 0
	}

	// Use last 5 points to calculate trend
	start := len(smoothed) - 5
	if start < 0 {
		start = 0
	}

	recentData := smoothed[start:]

	// Simple linear trend
	if len(recentData) < 2 {
		return 0
	}

	firstValue := recentData[0]
	lastValue := recentData[len(recentData)-1]

	return (lastValue - firstValue) / float64(len(recentData)-1)
}

func (t *TrendAnalyzerImpl) isDomainBlocked(domain string, data []types.PiholeRecord) bool {
	// Check if domain appears in blocked queries
	for _, record := range data {
		if strings.ToLower(record.Domain) == domain && record.Status == 1 {
			return true
		}
	}
	return false
}

func (t *TrendAnalyzerImpl) findPeakHour(hourlyPatterns map[int]float64) (int, float64) {
	maxHour, maxUsage := 0, 0.0

	for hour, usage := range hourlyPatterns {
		if usage > maxUsage {
			maxHour = hour
			maxUsage = usage
		}
	}

	return maxHour, maxUsage
}

func (t *TrendAnalyzerImpl) findOffPeakHour(hourlyPatterns map[int]float64) (int, float64) {
	minHour, minUsage := 0, 1.0

	for hour, usage := range hourlyPatterns {
		if usage < minUsage {
			minHour = hour
			minUsage = usage
		}
	}

	return minHour, minUsage
}

func (t *TrendAnalyzerImpl) calculateWeekendRatio(dailyPatterns map[time.Weekday]float64) float64 {
	weekdayTotal := dailyPatterns[time.Monday] + dailyPatterns[time.Tuesday] +
		dailyPatterns[time.Wednesday] + dailyPatterns[time.Thursday] +
		dailyPatterns[time.Friday]

	weekendTotal := dailyPatterns[time.Saturday] + dailyPatterns[time.Sunday]

	weekdayAvg := weekdayTotal / 5.0
	weekendAvg := weekendTotal / 2.0

	if weekdayAvg == 0 {
		return 0
	}

	return weekendAvg / weekdayAvg
}
