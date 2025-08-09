package ml

import (
	"context"
	"fmt"
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestEngine_Initialize(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)

	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	status := engine.GetStatus()
	if !status.IsInitialized {
		t.Error("Engine should be initialized")
	}

	if status.Status != "initialized" {
		t.Errorf("Expected status 'initialized', got '%s'", status.Status)
	}
}

func TestEngine_Train(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	// Create test data
	testData := createTestData(100)

	err = engine.Train(ctx, testData)
	if err != nil {
		t.Fatalf("Failed to train ML engine: %v", err)
	}

	if !engine.IsTrained() {
		t.Error("Engine should be trained")
	}

	status := engine.GetStatus()
	if status.Status != "trained" {
		t.Errorf("Expected status 'trained', got '%s'", status.Status)
	}
}

func TestEngine_DetectAnomalies(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	// Train with normal data
	normalData := createTestData(100)
	err = engine.Train(ctx, normalData)
	if err != nil {
		t.Fatalf("Failed to train ML engine: %v", err)
	}

	// Test with anomalous data
	anomalousData := createAnomalousTestData(50)
	anomalies, err := engine.DetectAnomalies(ctx, anomalousData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	if len(anomalies) == 0 {
		t.Error("Expected to detect some anomalies")
	}

	// Verify anomaly structure
	for _, anomaly := range anomalies {
		if anomaly.ID == "" {
			t.Error("Anomaly should have an ID")
		}
		if anomaly.Type == "" {
			t.Error("Anomaly should have a type")
		}
		if anomaly.Severity == "" {
			t.Error("Anomaly should have a severity")
		}
		if anomaly.Score < 0 || anomaly.Score > 1 {
			t.Errorf("Anomaly score should be between 0 and 1, got %f", anomaly.Score)
		}
		if anomaly.Confidence < 0 || anomaly.Confidence > 1 {
			t.Errorf("Anomaly confidence should be between 0 and 1, got %f", anomaly.Confidence)
		}
	}
}

func TestEngine_AnalyzeTrends(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	// Create test data with trend
	trendData := createTrendTestData(100)

	analysis, err := engine.AnalyzeTrends(ctx, trendData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}

	if analysis.TotalQueries != len(trendData) {
		t.Errorf("Expected total queries %d, got %d", len(trendData), analysis.TotalQueries)
	}

	if analysis.QueryTrend == "" {
		t.Error("Query trend should not be empty")
	}

	if len(analysis.HourlyPatterns) == 0 {
		t.Error("Hourly patterns should not be empty")
	}

	if len(analysis.DailyPatterns) == 0 {
		t.Error("Daily patterns should not be empty")
	}
}

func TestEngine_PredictTrends(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	// Create test data
	testData := createTestData(100)

	prediction, err := engine.PredictTrends(ctx, testData, 6*time.Hour)
	if err != nil {
		t.Fatalf("Failed to predict trends: %v", err)
	}

	if prediction == nil {
		t.Fatal("Prediction should not be nil")
	}

	if len(prediction.PredictedQueries) == 0 {
		t.Error("Predicted queries should not be empty")
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Prediction confidence should be between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.Methodology == "" {
		t.Error("Prediction methodology should not be empty")
	}
}

func TestEngine_ProcessData(t *testing.T) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize ML engine: %v", err)
	}

	// Train first
	trainingData := createTestData(100)
	err = engine.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train ML engine: %v", err)
	}

	// Process data
	testData := createTestData(50)
	results, err := engine.ProcessData(ctx, testData)
	if err != nil {
		t.Fatalf("Failed to process data: %v", err)
	}

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	// Verify results structure
	if results.ProcessedAt.IsZero() {
		t.Error("ProcessedAt timestamp should be set")
	}

	if results.Summary.HealthScore < 0 || results.Summary.HealthScore > 100 {
		t.Errorf("Health score should be between 0 and 100, got %f", results.Summary.HealthScore)
	}

	if len(results.Summary.Recommendations) == 0 {
		t.Error("Should have at least one recommendation")
	}
}

func TestDefaultMLConfig(t *testing.T) {
	config := DefaultMLConfig()

	if !config.AnomalyDetection.Enabled {
		t.Error("Anomaly detection should be enabled by default")
	}

	if !config.TrendAnalysis.Enabled {
		t.Error("Trend analysis should be enabled by default")
	}

	if config.AnomalyDetection.Sensitivity <= 0 || config.AnomalyDetection.Sensitivity > 1 {
		t.Errorf("Sensitivity should be between 0 and 1, got %f", config.AnomalyDetection.Sensitivity)
	}

	if config.AnomalyDetection.MinConfidence <= 0 || config.AnomalyDetection.MinConfidence > 1 {
		t.Errorf("MinConfidence should be between 0 and 1, got %f", config.AnomalyDetection.MinConfidence)
	}

	if len(config.AnomalyDetection.AnomalyTypes) == 0 {
		t.Error("Should have at least one anomaly type configured")
	}

	if len(config.AnomalyDetection.Thresholds) == 0 {
		t.Error("Should have at least one threshold configured")
	}
}

func TestAnomalyTypes(t *testing.T) {
	expectedTypes := []AnomalyType{
		AnomalyTypeVolumeSpike,
		AnomalyTypeVolumeDropout,
		AnomalyTypeUnusualDomain,
		AnomalyTypeUnusualClient,
		AnomalyTypeQueryPattern,
		AnomalyTypeTimePattern,
		AnomalyTypeResponseTime,
		AnomalyTypeBlockedSpike,
	}

	for _, expectedType := range expectedTypes {
		if string(expectedType) == "" {
			t.Errorf("Anomaly type should not be empty: %v", expectedType)
		}
	}
}

func TestSeverityLevels(t *testing.T) {
	expectedLevels := []SeverityLevel{
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}

	for _, level := range expectedLevels {
		if string(level) == "" {
			t.Errorf("Severity level should not be empty: %v", level)
		}
	}
}

func TestTrendDirections(t *testing.T) {
	expectedDirections := []TrendDirection{
		TrendIncreasing,
		TrendDecreasing,
		TrendStable,
		TrendVolatile,
	}

	for _, direction := range expectedDirections {
		if string(direction) == "" {
			t.Errorf("Trend direction should not be empty: %v", direction)
		}
	}
}

// Helper functions for creating test data
func createTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := []string{
		"google.com",
		"facebook.com",
		"amazon.com",
		"microsoft.com",
		"apple.com",
		"netflix.com",
		"youtube.com",
		"twitter.com",
	}

	clients := []string{
		"192.168.1.100",
		"192.168.1.101",
		"192.168.1.102",
		"192.168.1.103",
	}

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 10)
		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    domains[i%len(domains)],
			Client:    clients[i%len(clients)],
			QueryType: "A",
			Status:    2, // Allowed
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createAnomalousTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now()

	// Create unusual domains and patterns
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second * 10)
		records[i] = types.PiholeRecord{
			ID:        i + 1000,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    fmt.Sprintf("suspicious-domain-%d.com", i), // Unusual domains
			Client:    "192.168.1.200",                            // Single client generating lots of queries
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createTrendTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := []string{"google.com", "facebook.com", "amazon.com"}
	clients := []string{"192.168.1.100", "192.168.1.101"}

	// Create increasing trend
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 15)

		// Create more queries as time progresses (increasing trend)
		multiplier := 1 + (i / 20) // Increasing pattern
		for j := 0; j < multiplier; j++ {
			if len(records) >= count {
				break
			}

			idx := i*multiplier + j
			if idx >= count {
				break
			}

			records[idx] = types.PiholeRecord{
				ID:        idx + 1,
				DateTime:  timestamp.Add(time.Duration(j) * time.Second).Format("2006-01-02 15:04:05"),
				Domain:    domains[idx%len(domains)],
				Client:    clients[idx%len(clients)],
				QueryType: "A",
				Status:    2,
				Timestamp: timestamp.Add(time.Duration(j) * time.Second).Format("2006-01-02 15:04:05"),
			}
		}
	}

	return records
}

func BenchmarkEngine_ProcessData(b *testing.B) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	engine.Initialize(ctx, config)

	// Train with sample data
	trainingData := createTestData(1000)
	engine.Train(ctx, trainingData)

	// Benchmark processing
	testData := createTestData(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ProcessData(ctx, testData)
		if err != nil {
			b.Fatalf("ProcessData failed: %v", err)
		}
	}
}

func BenchmarkEngine_DetectAnomalies(b *testing.B) {
	config := DefaultMLConfig()
	engine := NewEngine(config, nil)

	ctx := context.Background()
	engine.Initialize(ctx, config)

	// Train with sample data
	trainingData := createTestData(1000)
	engine.Train(ctx, trainingData)

	// Benchmark anomaly detection
	testData := createAnomalousTestData(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.DetectAnomalies(ctx, testData)
		if err != nil {
			b.Fatalf("DetectAnomalies failed: %v", err)
		}
	}
}
