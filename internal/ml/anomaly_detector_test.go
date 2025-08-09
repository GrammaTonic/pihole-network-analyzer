package ml

import (
	"context"
	"fmt"
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestStatisticalAnomalyDetector_Initialize(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	err := detector.Initialize(ctx, config)

	if err != nil {
		t.Fatalf("Failed to initialize anomaly detector: %v", err)
	}
}

func TestStatisticalAnomalyDetector_Train(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	err := detector.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize anomaly detector: %v", err)
	}

	// Create training data
	trainingData := createNormalTrainingData(200)

	err = detector.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train anomaly detector: %v", err)
	}

	if !detector.IsTrained() {
		t.Error("Detector should be trained")
	}

	modelInfo := detector.GetModelInfo()
	if modelInfo.TrainingSize != len(trainingData) {
		t.Errorf("Expected training size %d, got %d", len(trainingData), modelInfo.TrainingSize)
	}
}

func TestStatisticalAnomalyDetector_DetectVolumeSpike(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Train with normal data (low volume)
	normalData := createNormalTrainingData(100)
	detector.Train(ctx, normalData)

	// Create data with volume spike
	spikeData := createVolumeSpikeData(500) // Much higher volume

	anomalies, err := detector.DetectAnomalies(ctx, spikeData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	// Should detect volume spike anomalies
	volumeSpikes := 0
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyTypeVolumeSpike {
			volumeSpikes++
		}
	}

	if volumeSpikes == 0 {
		t.Error("Expected to detect volume spike anomalies")
	}
}

func TestStatisticalAnomalyDetector_DetectUnusualDomain(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Train with common domains
	normalData := createNormalTrainingData(100)
	detector.Train(ctx, normalData)

	// Create data with unusual domains
	unusualData := createUnusualDomainData(50)

	anomalies, err := detector.DetectAnomalies(ctx, unusualData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	// Should detect unusual domain anomalies
	unusualDomains := 0
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyTypeUnusualDomain {
			unusualDomains++
		}
	}

	if unusualDomains == 0 {
		t.Error("Expected to detect unusual domain anomalies")
	}
}

func TestStatisticalAnomalyDetector_DetectUnusualClient(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Train with normal client behavior
	normalData := createNormalTrainingData(100)
	detector.Train(ctx, normalData)

	// Create data with unusual client behavior
	unusualClientData := createUnusualClientData(100)

	anomalies, err := detector.DetectAnomalies(ctx, unusualClientData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	// Should detect unusual client anomalies
	unusualClients := 0
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyTypeUnusualClient {
			unusualClients++
		}
	}

	if unusualClients == 0 {
		t.Error("Expected to detect unusual client anomalies")
	}
}

func TestStatisticalAnomalyDetector_ConfidenceFiltering(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	config.MinConfidence = 0.8 // High confidence threshold

	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Train with normal data
	normalData := createNormalTrainingData(100)
	detector.Train(ctx, normalData)

	// Create mildly anomalous data
	mildlyAnomalousData := createMildlyAnomalousData(50)

	anomalies, err := detector.DetectAnomalies(ctx, mildlyAnomalousData)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	// All returned anomalies should meet confidence threshold
	for _, anomaly := range anomalies {
		if anomaly.Confidence < config.MinConfidence {
			t.Errorf("Anomaly confidence %f below threshold %f", anomaly.Confidence, config.MinConfidence)
		}
	}
}

func TestStatisticalAnomalyDetector_InsufficientData(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Try to train with insufficient data
	insufficientData := createNormalTrainingData(5) // Too few records

	err := detector.Train(ctx, insufficientData)
	if err == nil {
		t.Error("Expected error for insufficient training data")
	}
}

func TestCalculateStats(t *testing.T) {
	values := []int{1, 2, 3, 4, 5}
	mean, stddev := calculateStats(values)

	expectedMean := 3.0
	if mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mean)
	}

	// Standard deviation should be approximately 1.58
	if stddev < 1.5 || stddev > 1.7 {
		t.Errorf("Expected stddev around 1.58, got %f", stddev)
	}
}

func TestCalculateFloatStats(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mean, stddev := calculateFloatStats(values)

	expectedMean := 3.0
	if mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, mean)
	}

	// Standard deviation should be approximately 1.58
	if stddev < 1.5 || stddev > 1.7 {
		t.Errorf("Expected stddev around 1.58, got %f", stddev)
	}
}

func TestSeverityCalculation(t *testing.T) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	testCases := []struct {
		value    float64
		mean     float64
		stddev   float64
		expected SeverityLevel
	}{
		{100, 50, 10, SeverityCritical}, // 5 standard deviations
		{80, 50, 10, SeverityHigh},      // 3 standard deviations
		{70, 50, 10, SeverityMedium},    // 2 standard deviations
		{60, 50, 10, SeverityLow},       // 1 standard deviation
	}

	for _, tc := range testCases {
		severity := detector.calculateSeverity(tc.value, tc.mean, tc.stddev)
		if severity != tc.expected {
			t.Errorf("Expected severity %s for value %f (mean=%f, stddev=%f), got %s",
				tc.expected, tc.value, tc.mean, tc.stddev, severity)
		}
	}
}

// Helper functions for creating specific test data

func createNormalTrainingData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now().Add(-24 * time.Hour)

	// Common domains for training
	domains := []string{
		"google.com",
		"facebook.com",
		"amazon.com",
		"microsoft.com",
		"apple.com",
	}

	// Normal client IPs
	clients := []string{
		"192.168.1.100",
		"192.168.1.101",
		"192.168.1.102",
	}

	for i := 0; i < count; i++ {
		// Distribute queries evenly over 24 hours
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 14) // ~14 minutes apart

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

func createVolumeSpikeData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now()

	// Create high volume in short time window (spike)
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second * 2) // 2 seconds apart = high volume

		records[i] = types.PiholeRecord{
			ID:        i + 1000,
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

func createUnusualDomainData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now()

	// Create queries to unusual/suspicious domains
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 2)

		records[i] = types.PiholeRecord{
			ID:        i + 2000,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    fmt.Sprintf("malicious-domain-%d.badsite.com", i),
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createUnusualClientData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now()

	// Create many queries from a single unusual client
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second * 30)

		records[i] = types.PiholeRecord{
			ID:        i + 3000,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "google.com",
			Client:    "192.168.1.200", // Unusual client not seen in training
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func createMildlyAnomalousData(count int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, count)
	baseTime := time.Now()

	// Create mildly anomalous data that should have low confidence
	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute * 3)

		records[i] = types.PiholeRecord{
			ID:        i + 4000,
			DateTime:  timestamp.Format("2006-01-02 15:04:05"),
			Domain:    "moderately-unusual.com", // Slightly unusual but not clearly malicious
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    2,
			Timestamp: timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return records
}

func BenchmarkAnomalyDetector_Train(b *testing.B) {
	config := DefaultMLConfig().AnomalyDetection

	trainingData := createNormalTrainingData(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector := NewStatisticalAnomalyDetector(config, nil)
		detector.Initialize(ctx, config)
		err := detector.Train(ctx, trainingData)
		if err != nil {
			b.Fatalf("Training failed: %v", err)
		}
	}
}

func BenchmarkAnomalyDetector_DetectAnomalies(b *testing.B) {
	config := DefaultMLConfig().AnomalyDetection
	detector := NewStatisticalAnomalyDetector(config, nil)

	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Train once
	trainingData := createNormalTrainingData(1000)
	detector.Train(ctx, trainingData)

	// Benchmark detection
	testData := createVolumeSpikeData(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := detector.DetectAnomalies(ctx, testData)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}
	}
}
