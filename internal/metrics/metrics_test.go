package metrics

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if collector.logger != logger {
		t.Fatal("Expected logger to be set correctly")
	}
}

func TestRecordTotalQueries(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Record some queries
	collector.RecordTotalQueries(100)
	collector.RecordTotalQueries(50)

	// Get the metric value
	metric := &dto.Metric{}
	collector.totalQueries.Write(metric)

	expectedValue := 150.0
	actualValue := metric.GetCounter().GetValue()

	if actualValue != expectedValue {
		t.Errorf("Expected total queries to be %f, got %f", expectedValue, actualValue)
	}
}

func TestRecordQueryResponseTime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Record some response times
	collector.RecordQueryResponseTime(100 * time.Millisecond)
	collector.RecordQueryResponseTime(200 * time.Millisecond)

	// Get the metric value
	metric := &dto.Metric{}
	collector.queryResponseTime.Write(metric)

	histogram := metric.GetHistogram()
	if histogram.GetSampleCount() != 2 {
		t.Errorf("Expected 2 samples, got %d", histogram.GetSampleCount())
	}

	// Check that the sum is approximately correct (0.1 + 0.2 = 0.3 seconds)
	expectedSum := 0.3
	actualSum := histogram.GetSampleSum()
	if actualSum < expectedSum-0.01 || actualSum > expectedSum+0.01 {
		t.Errorf("Expected sum to be approximately %f, got %f", expectedSum, actualSum)
	}
}

func TestRecordQueryByType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Record queries by type
	collector.RecordQueryByType("A")
	collector.RecordQueryByType("A")
	collector.RecordQueryByType("AAAA")

	// Test A record queries
	aMetric, err := collector.queriesByType.GetMetricWithLabelValues("A")
	if err != nil {
		t.Fatalf("Failed to get A record metric: %v", err)
	}

	metric := &dto.Metric{}
	aMetric.Write(metric)

	if metric.GetCounter().GetValue() != 2.0 {
		t.Errorf("Expected A record count to be 2, got %f", metric.GetCounter().GetValue())
	}

	// Test AAAA record queries
	aaaaMetric, err := collector.queriesByType.GetMetricWithLabelValues("AAAA")
	if err != nil {
		t.Fatalf("Failed to get AAAA record metric: %v", err)
	}

	aaaaMetric.Write(metric)

	if metric.GetCounter().GetValue() != 1.0 {
		t.Errorf("Expected AAAA record count to be 1, got %f", metric.GetCounter().GetValue())
	}
}

func TestRecordQueryByStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Record queries by status
	collector.RecordQueryByStatus("allowed")
	collector.RecordQueryByStatus("allowed")
	collector.RecordQueryByStatus("blocked")

	// Test allowed queries
	allowedMetric, err := collector.queriesByStatus.GetMetricWithLabelValues("allowed")
	if err != nil {
		t.Fatalf("Failed to get allowed queries metric: %v", err)
	}

	metric := &dto.Metric{}
	allowedMetric.Write(metric)

	if metric.GetCounter().GetValue() != 2.0 {
		t.Errorf("Expected allowed queries count to be 2, got %f", metric.GetCounter().GetValue())
	}

	// Test blocked queries
	blockedMetric, err := collector.queriesByStatus.GetMetricWithLabelValues("blocked")
	if err != nil {
		t.Fatalf("Failed to get blocked queries metric: %v", err)
	}

	blockedMetric.Write(metric)

	if metric.GetCounter().GetValue() != 1.0 {
		t.Errorf("Expected blocked queries count to be 1, got %f", metric.GetCounter().GetValue())
	}
}

func TestSetActiveClients(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.SetActiveClients(25)

	metric := &dto.Metric{}
	collector.activeClients.Write(metric)

	if metric.GetGauge().GetValue() != 25.0 {
		t.Errorf("Expected active clients to be 25, got %f", metric.GetGauge().GetValue())
	}
}

func TestSetUniqueClients(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.SetUniqueClients(15)

	metric := &dto.Metric{}
	collector.uniqueClients.Write(metric)

	if metric.GetGauge().GetValue() != 15.0 {
		t.Errorf("Expected unique clients to be 15, got %f", metric.GetGauge().GetValue())
	}
}

func TestRecordTopClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordTopClient("192.168.1.100", "laptop", 150)

	topClientMetric, err := collector.topClients.GetMetricWithLabelValues("192.168.1.100", "laptop")
	if err != nil {
		t.Fatalf("Failed to get top client metric: %v", err)
	}

	metric := &dto.Metric{}
	topClientMetric.Write(metric)

	if metric.GetGauge().GetValue() != 150.0 {
		t.Errorf("Expected top client queries to be 150, got %f", metric.GetGauge().GetValue())
	}
}

func TestRecordTopDomain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordTopDomain("example.com", 50)
	collector.RecordTopDomain("example.com", 25)

	topDomainMetric, err := collector.topDomains.GetMetricWithLabelValues("example.com")
	if err != nil {
		t.Fatalf("Failed to get top domain metric: %v", err)
	}

	metric := &dto.Metric{}
	topDomainMetric.Write(metric)

	if metric.GetCounter().GetValue() != 75.0 {
		t.Errorf("Expected top domain count to be 75, got %f", metric.GetCounter().GetValue())
	}
}

func TestRecordBlockedAndAllowedDomains(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordBlockedDomains(30)
	collector.RecordAllowedDomains(70)

	// Test blocked domains
	metric := &dto.Metric{}
	collector.blockedDomains.Write(metric)

	if metric.GetCounter().GetValue() != 30.0 {
		t.Errorf("Expected blocked domains to be 30, got %f", metric.GetCounter().GetValue())
	}

	// Test allowed domains
	collector.allowedDomains.Write(metric)

	if metric.GetCounter().GetValue() != 70.0 {
		t.Errorf("Expected allowed domains to be 70, got %f", metric.GetCounter().GetValue())
	}
}

func TestRecordAnalysisProcessTime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordAnalysisProcessTime(5 * time.Second)

	metric := &dto.Metric{}
	collector.analysisProcessTime.Write(metric)

	histogram := metric.GetHistogram()
	if histogram.GetSampleCount() != 1 {
		t.Errorf("Expected 1 sample, got %d", histogram.GetSampleCount())
	}

	if histogram.GetSampleSum() != 5.0 {
		t.Errorf("Expected sum to be 5.0 seconds, got %f", histogram.GetSampleSum())
	}
}

func TestRecordPiholeAPICallTime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordPiholeAPICallTime(250 * time.Millisecond)

	metric := &dto.Metric{}
	collector.piholeAPICallTime.Write(metric)

	histogram := metric.GetHistogram()
	if histogram.GetSampleCount() != 1 {
		t.Errorf("Expected 1 sample, got %d", histogram.GetSampleCount())
	}

	expectedSum := 0.25
	actualSum := histogram.GetSampleSum()
	if actualSum < expectedSum-0.01 || actualSum > expectedSum+0.01 {
		t.Errorf("Expected sum to be approximately %f, got %f", expectedSum, actualSum)
	}
}

func TestRecordError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordError("api_timeout")
	collector.RecordError("api_timeout")
	collector.RecordError("connection_failed")

	// Test API timeout errors
	timeoutMetric, err := collector.errorCount.GetMetricWithLabelValues("api_timeout")
	if err != nil {
		t.Fatalf("Failed to get api_timeout error metric: %v", err)
	}

	metric := &dto.Metric{}
	timeoutMetric.Write(metric)

	if metric.GetCounter().GetValue() != 2.0 {
		t.Errorf("Expected api_timeout errors to be 2, got %f", metric.GetCounter().GetValue())
	}

	// Test connection failed errors
	connMetric, err := collector.errorCount.GetMetricWithLabelValues("connection_failed")
	if err != nil {
		t.Fatalf("Failed to get connection_failed error metric: %v", err)
	}

	connMetric.Write(metric)

	if metric.GetCounter().GetValue() != 1.0 {
		t.Errorf("Expected connection_failed errors to be 1, got %f", metric.GetCounter().GetValue())
	}
}

func TestSetLastAnalysisTime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	testTime := time.Date(2023, 10, 15, 12, 30, 0, 0, time.UTC)
	collector.SetLastAnalysisTime(testTime)

	metric := &dto.Metric{}
	collector.lastAnalysisTime.Write(metric)

	expectedValue := float64(testTime.Unix())
	if metric.GetGauge().GetValue() != expectedValue {
		t.Errorf("Expected last analysis time to be %f, got %f", expectedValue, metric.GetGauge().GetValue())
	}
}

func TestRecordAnalysisDuration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.RecordAnalysisDuration(30 * time.Second)

	metric := &dto.Metric{}
	collector.analysisDuration.Write(metric)

	histogram := metric.GetHistogram()
	if histogram.GetSampleCount() != 1 {
		t.Errorf("Expected 1 sample, got %d", histogram.GetSampleCount())
	}

	if histogram.GetSampleSum() != 30.0 {
		t.Errorf("Expected sum to be 30.0 seconds, got %f", histogram.GetSampleSum())
	}
}

func TestSetDataSourceHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Test healthy state
	collector.SetDataSourceHealth(true)

	metric := &dto.Metric{}
	collector.dataSourceHealth.Write(metric)

	if metric.GetGauge().GetValue() != 1.0 {
		t.Errorf("Expected data source health to be 1.0 (healthy), got %f", metric.GetGauge().GetValue())
	}

	// Test unhealthy state
	collector.SetDataSourceHealth(false)
	collector.dataSourceHealth.Write(metric)

	if metric.GetGauge().GetValue() != 0.0 {
		t.Errorf("Expected data source health to be 0.0 (unhealthy), got %f", metric.GetGauge().GetValue())
	}
}

func TestSetQueriesPerSecond(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	collector.SetQueriesPerSecond(45.5)

	metric := &dto.Metric{}
	collector.queriesPerSecond.Write(metric)

	if metric.GetGauge().GetValue() != 45.5 {
		t.Errorf("Expected queries per second to be 45.5, got %f", metric.GetGauge().GetValue())
	}
}

func TestGetRegistry(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	registry := collector.GetRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be returned, got nil")
	}

	// The registry should be the collector's custom registry
	if registry != collector.registry {
		t.Error("Expected collector's custom registry")
	}
}

func TestConcurrency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Test concurrent access to metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				collector.RecordTotalQueries(1)
				collector.SetActiveClients(float64(id))
				collector.RecordQueryByType("A")
				collector.RecordError("test_error")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that all metrics were recorded correctly
	metric := &dto.Metric{}
	collector.totalQueries.Write(metric)

	if metric.GetCounter().GetValue() != 1000.0 {
		t.Errorf("Expected total queries to be 1000, got %f", metric.GetCounter().GetValue())
	}
}

// Benchmark tests
func BenchmarkRecordTotalQueries(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordTotalQueries(1)
	}
}

func BenchmarkRecordQueryByType(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	queryTypes := []string{"A", "AAAA", "PTR", "CNAME", "MX"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordQueryByType(queryTypes[i%len(queryTypes)])
	}
}

func BenchmarkRecordQueryResponseTime(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordQueryResponseTime(duration)
	}
}

// Helper function to check if a string contains expected log output
func containsLogOutput(t *testing.T, logOutput, expected string) {
	if !strings.Contains(logOutput, expected) {
		t.Errorf("Expected log output to contain '%s', but it didn't. Full output: %s", expected, logOutput)
	}
}
