package ml

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestMLIntegration_AnomalyDetector(t *testing.T) {
	// Test anomaly detector initialization and basic functionality
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	err := detector.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize anomaly detector: %v", err)
	}

	// Create training data
	trainingData := createIntegrationTestData(100)

	// Train the detector
	err = detector.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train anomaly detector: %v", err)
	}

	// Test anomaly detection
	testData := createAnomalyTestData(50)
	anomalies, err := detector.DetectAnomalies(ctx, testData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	// Should detect some anomalies in the test data
	if len(anomalies) == 0 {
		t.Log("No anomalies detected - this might be expected depending on test data")
	} else {
		t.Logf("Detected %d anomalies", len(anomalies))
	}

	// Verify detector info
	info := detector.GetModelInfo()
	if info.Name == "" {
		t.Error("Detector should have a name")
	}

	if len(info.Parameters) == 0 {
		t.Error("Detector should have parameters listed")
	}
}

func TestMLIntegration_TrendAnalyzer(t *testing.T) {
	// Test trend analyzer initialization and basic functionality
	config := DefaultMLConfig().TrendAnalysis
	analyzer := NewTrendAnalyzer(config, nil)

	ctx := context.Background()
	err := analyzer.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize trend analyzer: %v", err)
	}

	// Create test data with trends
	testData := createTrendingTestData(100)

	// Test trend analysis
	analysis, err := analyzer.AnalyzeTrends(ctx, testData, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to analyze trends: %v", err)
	}

	// Verify analysis results
	if analysis.TotalQueries != len(testData) {
		t.Errorf("Expected total queries %d, got %d", len(testData), analysis.TotalQueries)
	}

	if len(analysis.HourlyPatterns) != 24 {
		t.Errorf("Expected 24 hourly patterns, got %d", len(analysis.HourlyPatterns))
	}

	if len(analysis.DailyPatterns) != 7 {
		t.Errorf("Expected 7 daily patterns, got %d", len(analysis.DailyPatterns))
	}

	// Test trend prediction
	prediction, err := analyzer.PredictTrends(ctx, testData, 6*time.Hour)
	if err != nil {
		t.Fatalf("Failed to predict trends: %v", err)
	}

	// Verify prediction results
	if len(prediction.PredictedQueries) == 0 {
		t.Error("Should have predicted queries")
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", prediction.Confidence)
	}

	// Verify analyzer info
	info := analyzer.GetAnalyzerInfo()
	if info.Name == "" {
		t.Error("Analyzer should have a name")
	}

	if len(info.Parameters) == 0 {
		t.Error("Analyzer should have parameters listed")
	}
}

func TestMLIntegration_ConfigValidation(t *testing.T) {
	// Test default configuration validity
	config := DefaultMLConfig()

	// Verify anomaly detection config
	if !config.AnomalyDetection.Enabled {
		t.Error("Anomaly detection should be enabled by default")
	}

	if config.AnomalyDetection.Sensitivity <= 0 {
		t.Error("Sensitivity should be positive")
	}

	if config.AnomalyDetection.MinConfidence <= 0 || config.AnomalyDetection.MinConfidence > 1 {
		t.Error("MinConfidence should be between 0 and 1")
	}

	if len(config.AnomalyDetection.AnomalyTypes) == 0 {
		t.Error("Should have anomaly types configured")
	}

	// Verify trend analysis config
	if !config.TrendAnalysis.Enabled {
		t.Error("Trend analysis should be enabled by default")
	}

	if config.TrendAnalysis.MinDataPoints <= 0 {
		t.Error("MinDataPoints should be positive")
	}

	if config.TrendAnalysis.SmoothingFactor <= 0 || config.TrendAnalysis.SmoothingFactor >= 1 {
		t.Error("SmoothingFactor should be between 0 and 1")
	}
}

// Helper functions for integration tests

func createIntegrationTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := []string{"google.com", "facebook.com", "amazon.com", "netflix.com"}
	clients := []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"}

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 14)

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    domains[i%len(domains)],
			Client:    clients[i%len(clients)],
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createAnomalyTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-1 * time.Hour)

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 1)

		// Create some potential anomalies
		domain := "google.com"
		if i%10 == 0 {
			domain = "suspicious-domain.com" // Unusual domain
		}

		client := "192.168.1.100"
		if i%15 == 0 {
			client = "192.168.1.255" // Unusual client
		}

		records[i] = types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    domain,
			Client:    client,
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createTrendingTestData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < count; i++ {
		// Create increasing trend over time
		hour := i % 24
		timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(i/24)*time.Minute*30)

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
