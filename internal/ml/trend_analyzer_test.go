package ml

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestTrendAnalyzer_Initialize(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	err := analyzer.Initialize(ctx, config)

	if err != nil {
		t.Fatalf("Failed to initialize trend analyzer: %v", err)
	}

	info := analyzer.GetAnalyzerInfo()
	if info.Name == "" {
		t.Error("Analyzer should have a name")
	}

	if len(info.Algorithms) == 0 {
		t.Error("Analyzer should have algorithms listed")
	}
}

func TestTrendAnalyzer_AnalyzeTrends_InsufficientData(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Try with insufficient data
	insufficientData := createTrendTestData(5) // Less than MinDataPoints

	_, err := analyzer.AnalyzeTrends(ctx, insufficientData, 24*time.Hour)
	if err == nil {
		t.Error("Expected error for insufficient data points")
	}
}

func TestTrendAnalyzer_AnalyzeTrends_IncreasingTrend(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with increasing trend
	increasingData := createIncreasingTrendData(50)

	analysis, err := analyzer.AnalyzeTrends(ctx, increasingData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	if analysis.QueryTrend != TrendIncreasing {
		t.Errorf("Expected increasing trend, got %s", analysis.QueryTrend)
	}

	if len(analysis.DomainTrends) == 0 {
		t.Error("Should have domain trends")
	}

	if len(analysis.ClientTrends) == 0 {
		t.Error("Should have client trends")
	}

	if len(analysis.HourlyPatterns) != 24 {
		t.Errorf("Expected 24 hourly patterns, got %d", len(analysis.HourlyPatterns))
	}

	if len(analysis.DailyPatterns) != 7 {
		t.Errorf("Expected 7 daily patterns, got %d", len(analysis.DailyPatterns))
	}
}

func TestTrendAnalyzer_AnalyzeTrends_DecreasingTrend(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with decreasing trend
	decreasingData := createDecreasingTrendData(50)

	analysis, err := analyzer.AnalyzeTrends(ctx, decreasingData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	if analysis.QueryTrend != TrendDecreasing {
		t.Errorf("Expected decreasing trend, got %s", analysis.QueryTrend)
	}
}

func TestTrendAnalyzer_AnalyzeTrends_StableTrend(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with stable trend
	stableData := createStableTrendData(50)

	analysis, err := analyzer.AnalyzeTrends(ctx, stableData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	if analysis.QueryTrend != TrendStable {
		t.Errorf("Expected stable trend, got %s", analysis.QueryTrend)
	}
}

func TestTrendAnalyzer_AnalyzeTrends_VolatileTrend(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with volatile trend - use more data points for better volatility detection
	volatileData := createVolatileTrendData(250)

	analysis, err := analyzer.AnalyzeTrends(ctx, volatileData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	if analysis.QueryTrend != TrendVolatile {
		t.Errorf("Expected volatile trend, got %s", analysis.QueryTrend)
	}
}

func TestTrendAnalyzer_DomainTrends(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with specific domain trends
	domainTrendData := createDomainTrendData(100)

	analysis, err := analyzer.AnalyzeTrends(ctx, domainTrendData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Should have domain trends for top domains
	if len(analysis.DomainTrends) == 0 {
		t.Error("Should have domain trends")
	}

	// Check that domains are sorted by query count
	for i := 1; i < len(analysis.DomainTrends); i++ {
		if analysis.DomainTrends[i].Queries > analysis.DomainTrends[i-1].Queries {
			t.Error("Domain trends should be sorted by query count (descending)")
		}
	}

	// Should not exceed maximum trends
	if len(analysis.DomainTrends) > 20 {
		t.Errorf("Should not exceed 20 domain trends, got %d", len(analysis.DomainTrends))
	}
}

func TestTrendAnalyzer_ClientTrends(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with specific client trends
	clientTrendData := createClientTrendData(100)

	analysis, err := analyzer.AnalyzeTrends(ctx, clientTrendData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Should have client trends for active clients
	if len(analysis.ClientTrends) == 0 {
		t.Error("Should have client trends")
	}

	// Check that clients are sorted by query count
	for i := 1; i < len(analysis.ClientTrends); i++ {
		if analysis.ClientTrends[i].Queries > analysis.ClientTrends[i-1].Queries {
			t.Error("Client trends should be sorted by query count (descending)")
		}
	}

	// Should not exceed maximum trends
	if len(analysis.ClientTrends) > 15 {
		t.Errorf("Should not exceed 15 client trends, got %d", len(analysis.ClientTrends))
	}
}

func TestTrendAnalyzer_HourlyPatterns(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with specific hourly patterns
	hourlyData := createHourlyPatternData(100)

	analysis, err := analyzer.AnalyzeTrends(ctx, hourlyData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Should have patterns for all 24 hours
	if len(analysis.HourlyPatterns) != 24 {
		t.Errorf("Expected 24 hourly patterns, got %d", len(analysis.HourlyPatterns))
	}

	// Patterns should sum to approximately 1.0 (normalized)
	sum := 0.0
	for _, pattern := range analysis.HourlyPatterns {
		sum += pattern
		if pattern < 0 || pattern > 1 {
			t.Errorf("Hourly pattern should be between 0 and 1, got %f", pattern)
		}
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Hourly patterns should sum to ~1.0, got %f", sum)
	}
}

func TestTrendAnalyzer_DailyPatterns(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data spanning multiple days
	dailyData := createDailyPatternData(100)

	analysis, err := analyzer.AnalyzeTrends(ctx, dailyData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Should have patterns for all 7 days
	if len(analysis.DailyPatterns) != 7 {
		t.Errorf("Expected 7 daily patterns, got %d", len(analysis.DailyPatterns))
	}

	// Patterns should sum to approximately 1.0 (normalized)
	sum := 0.0
	for _, pattern := range analysis.DailyPatterns {
		sum += pattern
		if pattern < 0 || pattern > 1 {
			t.Errorf("Daily pattern should be between 0 and 1, got %f", pattern)
		}
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Daily patterns should sum to ~1.0, got %f", sum)
	}
}

func TestTrendAnalyzer_GenerateInsights(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with clear patterns for insights
	insightData := createInsightTestData(200)

	analysis, err := analyzer.AnalyzeTrends(ctx, insightData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Should generate some insights
	if len(analysis.Insights) == 0 {
		t.Error("Should generate at least one insight")
	}

	// Verify insight structure
	for _, insight := range analysis.Insights {
		if insight.Type == "" {
			t.Error("Insight should have a type")
		}
		if insight.Description == "" {
			t.Error("Insight should have a description")
		}
		if insight.Confidence < 0 || insight.Confidence > 1 {
			t.Errorf("Insight confidence should be between 0 and 1, got %f", insight.Confidence)
		}
	}
}

func TestTrendAnalyzer_PredictTrends(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Create data with predictable trend
	predictableData := createPredictableTrendData(100)

	prediction, err := analyzer.PredictTrends(ctx, predictableData, 6*time.Hour)
	if err != nil {
		t.Fatalf("Failed to predict trends: %v", err)
	}

	// Should have predictions for 6 hours
	expectedPoints := 6
	if len(prediction.PredictedQueries) != expectedPoints {
		t.Errorf("Expected %d prediction points, got %d", expectedPoints, len(prediction.PredictedQueries))
	}

	// Confidence should be reasonable
	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Prediction confidence should be between 0 and 1, got %f", prediction.Confidence)
	}

	// Should have methodology
	if prediction.Methodology == "" {
		t.Error("Prediction should have methodology")
	}

	// Verify prediction structure
	for i, forecast := range prediction.PredictedQueries {
		if forecast.PredictedCount < 0 {
			t.Errorf("Predicted count should be non-negative, got %d", forecast.PredictedCount)
		}

		if forecast.ConfidenceInterval.Lower < 0 {
			t.Errorf("Confidence interval lower bound should be non-negative, got %f", forecast.ConfidenceInterval.Lower)
		}

		if forecast.ConfidenceInterval.Upper < forecast.ConfidenceInterval.Lower {
			t.Errorf("Confidence interval upper bound should be >= lower bound")
		}

		// Timestamps should be in the future and ordered
		if i > 0 {
			prevTime := prediction.PredictedQueries[i-1].Timestamp
			if !forecast.Timestamp.After(prevTime) {
				t.Error("Prediction timestamps should be ordered chronologically")
			}
		}
	}
}

func TestTrendAnalyzer_InsufficientDataForPrediction(t *testing.T) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	// Try with insufficient data
	insufficientData := createTrendTestData(3)

	_, err := analyzer.PredictTrends(ctx, insufficientData, 6*time.Hour)
	if err == nil {
		t.Error("Expected error for insufficient data for prediction")
	}
}

// Helper functions for creating specific trend test data

func createIncreasingTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, count*3) // Allow for growth
	baseTime := time.Now().Add(-24 * time.Hour)

	recordID := 1
	for hour := 0; hour < count; hour++ {
		// Increase queries over time
		queriesThisHour := 1 + (hour / 10) // Gradual increase

		for q := 0; q < queriesThisHour; q++ {
			timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(q)*time.Minute*10)

			if len(records) >= cap(records) {
				break
			}

			records = append(records, types.PiholeRecord{
				ID:        recordID,
				DateTime:  timestamp.Format("2006-01-02 15:04:05"),
				Domain:    "google.com",
				Client:    "192.168.1.100",
				QueryType: "A",
				Status:    2,
				Timestamp: timestamp.Format("2006-01-02 15:04:05"),
			})
			recordID++
		}
	}

	return records
}

func createDecreasingTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, count*3)
	baseTime := time.Now().Add(-24 * time.Hour)

	recordID := 1
	for hour := 0; hour < count; hour++ {
		// Decrease queries over time
		queriesThisHour := 5 - (hour / 10) // Gradual decrease
		if queriesThisHour < 1 {
			queriesThisHour = 1
		}

		for q := 0; q < queriesThisHour; q++ {
			timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(q)*time.Minute*10)

			if len(records) >= cap(records) {
				break
			}

			records = append(records, types.PiholeRecord{
				ID:        recordID,
				DateTime:  timestamp.Format("2006-01-02 15:04:05"),
				Domain:    "google.com",
				Client:    "192.168.1.100",
				QueryType: "A",
				Status:    2,
				Timestamp: timestamp.Format("2006-01-02 15:04:05"),
			})
			recordID++
		}
	}

	return records
}

func createStableTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 29) // Evenly spaced

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createVolatileTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, count*5)
	baseTime := time.Now().Add(-24 * time.Hour)

	recordID := 1
	for hour := 0; hour < count; hour++ {
		// Create highly volatile pattern with extreme variations
		var queriesThisHour int
		switch hour % 8 {
		case 0:
			queriesThisHour = 1 // Very low
		case 1:
			queriesThisHour = 15 // Very high
		case 2:
			queriesThisHour = 2 // Low
		case 3:
			queriesThisHour = 20 // Very high
		case 4:
			queriesThisHour = 1 // Very low
		case 5:
			queriesThisHour = 25 // Extremely high
		case 6:
			queriesThisHour = 3 // Low
		case 7:
			queriesThisHour = 18 // High
		}

		for q := 0; q < queriesThisHour; q++ {
			timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(q)*time.Minute*2)

			if len(records) >= cap(records) {
				break
			}

			records = append(records, types.PiholeRecord{
				ID:        recordID,
				DateTime:  timestamp.Format("2006-01-02 15:04:05"),
				Domain:    "google.com",
				Client:    "192.168.1.100",
				QueryType: "A",
				Status:    2,
				Timestamp: timestamp.Format("2006-01-02 15:04:05"),
			})
			recordID++
		}
	}

	return records
}

func createDomainTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := []string{
		"google.com",      // Will be most frequent
		"facebook.com",    // Second most
		"amazon.com",      // Third most
		"rare-domain.com", // Infrequent
	}

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 14)

		// Weight domains to create trends
		var domain string
		if i%10 < 5 { // 50% google.com
			domain = domains[0]
		} else if i%10 < 8 { // 30% facebook.com
			domain = domains[1]
		} else if i%10 < 9 { // 10% amazon.com
			domain = domains[2]
		} else { // 10% rare-domain.com
			domain = domains[3]
		}

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    domain,
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createClientTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	clients := []string{
		"192.168.1.100", // Most active
		"192.168.1.101", // Second most active
		"192.168.1.102", // Less active
		"192.168.1.103", // Least active
	}

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 14)

		// Weight clients to create trends
		var client string
		if i%10 < 4 { // 40% from client 100
			client = clients[0]
		} else if i%10 < 7 { // 30% from client 101
			client = clients[1]
		} else if i%10 < 9 { // 20% from client 102
			client = clients[2]
		} else { // 10% from client 103
			client = clients[3]
		}

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    client,
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createHourlyPatternData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < count; i++ {
		// Create peak during hours 9-17 (business hours)
		hour := i % 24
		var minuteOffset int

		if hour >= 9 && hour <= 17 {
			// More queries during business hours
			minuteOffset = (i / 24) * 30 // 30 minutes apart during peak
		} else {
			// Fewer queries during off hours
			minuteOffset = (i / 24) * 120 // 2 hours apart during off-peak
		}

		timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(minuteOffset)*time.Minute)

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createDailyPatternData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-7 * 24 * time.Hour) // Start 7 days ago

	for i := 0; i < count; i++ {
		// Create different patterns for weekdays vs weekends
		dayIndex := i % 7
		var hourOffset int

		if dayIndex >= 1 && dayIndex <= 5 { // Monday to Friday
			hourOffset = 10 + (i/7)*2 // Business hours
		} else { // Weekend
			hourOffset = 14 + (i/7)*3 // Afternoon hours
		}

		timestamp := baseTime.Add(time.Duration(dayIndex)*24*time.Hour + time.Duration(hourOffset)*time.Hour)

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createInsightTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < count; i++ {
		hour := i % 24

		// Create a clear peak at hour 14 (2 PM)
		var frequency int
		if hour == 14 {
			frequency = 5 // Peak usage
		} else if hour >= 0 && hour <= 6 {
			frequency = 1 // Low usage
		} else {
			frequency = 2 // Normal usage
		}

		timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(i/24)*time.Minute*30)

		for f := 0; f < frequency; f++ {
			idx := i*frequency + f
			if idx >= count {
				break
			}

			records[idx] = types.PiholeRecord{
				ID:        idx + 1,
				DateTime:  timestamp.Add(time.Duration(f) * time.Minute).Format("2006-01-02 15:04:05"),
				Domain:    "google.com",
				Client:    "192.168.1.100",
				QueryType: "A",
				Status:    2,
				Timestamp: timestamp.Add(time.Duration(f) * time.Minute).Format("2006-01-02 15:04:05"),
			}
		}
	}

	return records[:count] // Ensure exact count
}

func createPredictableTrendData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < count; i++ {
		// Create a simple linear increase over time
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 14)

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func BenchmarkTrendAnalyzer_AnalyzeTrends(b *testing.B) {
	config := DefaultMLConfig().TrendAnalysis

	testData := createIncreasingTrendData(200)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := NewTrendAnalyzer(config, nil)
		analyzer.Initialize(ctx, config)
		_, err := analyzer.AnalyzeTrends(ctx, testData, 24*time.Hour)
		if err != nil {
			b.Fatalf("AnalyzeTrends failed: %v", err)
		}
	}
}

func BenchmarkTrendAnalyzer_PredictTrends(b *testing.B) {
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	analyzer.Initialize(ctx, config)

	testData := createPredictableTrendData(200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.PredictTrends(ctx, testData, 6*time.Hour)
		if err != nil {
			b.Fatalf("PredictTrends failed: %v", err)
		}
	}
}
