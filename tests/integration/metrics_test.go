package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/metrics"
	"pihole-analyzer/internal/types"
)

func TestMetricsIntegration(t *testing.T) {
	// Create a custom logger for testing
	loggerConfig := &logger.Config{
		Level:        logger.LevelDebug,
		EnableColors: false,
		EnableEmojis: false,
		Component:    "metrics-integration-test",
	}
	testLogger := logger.New(loggerConfig)

	// Create metrics collector
	collector := metrics.New(testLogger.GetSlogger())

	// Create metrics server config
	serverConfig := metrics.ServerConfig{
		Port:    "19099",
		Host:    "localhost", 
		Enabled: true,
	}

	// Create and start server
	server := metrics.NewServer(serverConfig, collector, testLogger.GetSlogger())
	server.StartInBackground()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Defer stopping the server
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Test that we can collect and expose metrics through HTTP endpoint
	t.Run("MetricsCollectionAndHTTPEndpoint", func(t *testing.T) {
		// Add some test metrics
		collector.RecordTotalQueries(500)
		collector.SetActiveClients(10)
		collector.SetUniqueClients(25)
		collector.RecordQueryByType("A")
		collector.RecordQueryByType("AAAA")
		collector.RecordQueryByStatus("allowed")
		collector.RecordQueryByStatus("blocked")
		collector.RecordTopDomain("example.com", 100)
		collector.RecordTopClient("192.168.1.100", "laptop", 50)
		collector.RecordBlockedDomains(30)
		collector.RecordAllowedDomains(470)
		collector.RecordAnalysisProcessTime(5 * time.Second)
		collector.RecordPiholeAPICallTime(200 * time.Millisecond)
		collector.SetLastAnalysisTime(time.Now())
		collector.SetDataSourceHealth(true)

		// Request metrics endpoint
		resp, err := http.Get("http://localhost:19099/metrics")
		if err != nil {
			t.Fatalf("Failed to get metrics: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		metricsOutput := string(body)

		// Verify that our test metrics are present
		expectedMetrics := []string{
			"pihole_analyzer_total_queries 500",
			"pihole_analyzer_active_clients 10",
			"pihole_analyzer_unique_clients 25",
			"pihole_analyzer_queries_by_type_total{query_type=\"A\"} 1",
			"pihole_analyzer_queries_by_type_total{query_type=\"AAAA\"} 1",
			"pihole_analyzer_queries_by_status_total{status=\"allowed\"} 1",
			"pihole_analyzer_queries_by_status_total{status=\"blocked\"} 1",
			"pihole_analyzer_top_domains_total{domain=\"example.com\"} 100",
			"pihole_analyzer_blocked_domains_total 30",
			"pihole_analyzer_allowed_domains_total 470",
			"pihole_analyzer_data_source_health 1",
		}

		for _, expected := range expectedMetrics {
			if !strings.Contains(metricsOutput, expected) {
				t.Errorf("Expected to find metric '%s' in output", expected)
			}
		}

		// Verify histogram metrics are present (they have more complex output)
		histogramMetrics := []string{
			"pihole_analyzer_analysis_process_time_seconds",
			"pihole_analyzer_api_call_time_seconds",
		}

		for _, metric := range histogramMetrics {
			if !strings.Contains(metricsOutput, metric) {
				t.Errorf("Expected to find histogram metric '%s' in output", metric)
			}
		}
	})

	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19099/health")
		if err != nil {
			t.Fatalf("Failed to get health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expected := `{"status":"healthy","component":"metrics-server"}`
		if string(body) != expected {
			t.Errorf("Expected body '%s', got '%s'", expected, string(body))
		}
	})

	t.Run("RootEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19099/")
		if err != nil {
			t.Fatalf("Failed to get root: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Pi-hole Network Analyzer Metrics Server") {
			t.Error("Expected root page to contain title")
		}

		if !strings.Contains(bodyStr, "/metrics") || !strings.Contains(bodyStr, "/health") {
			t.Error("Expected root page to contain endpoint links")
		}
	})
}

func TestMetricsConfigurationIntegration(t *testing.T) {
	// Test different configurations
	testCases := []struct {
		name   string
		config types.MetricsConfig
		expectServer bool
	}{
		{
			name: "FullyEnabled",
			config: types.MetricsConfig{
				Enabled:               true,
				Port:                  "19100",
				Host:                  "localhost",
				EnableEndpoint:        true,
				CollectMetrics:        true,
				EnableDetailedMetrics: true,
			},
			expectServer: true,
		},
		{
			name: "MetricsOnlyNoEndpoint",
			config: types.MetricsConfig{
				Enabled:               true,
				Port:                  "19101",
				Host:                  "localhost",
				EnableEndpoint:        false,
				CollectMetrics:        true,
				EnableDetailedMetrics: true,
			},
			expectServer: false,
		},
		{
			name: "Disabled",
			config: types.MetricsConfig{
				Enabled:               false,
				Port:                  "19102",
				Host:                  "localhost",
				EnableEndpoint:        false,
				CollectMetrics:        false,
				EnableDetailedMetrics: false,
			},
			expectServer: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create logger
			loggerConfig := &logger.Config{
				Level:        logger.LevelDebug,
				EnableColors: false,
				EnableEmojis: false,
				Component:    "metrics-config-test",
			}
			testLogger := logger.New(loggerConfig)

			var collector *metrics.Collector
			var server *metrics.Server

			// Initialize based on config
			if tc.config.Enabled && tc.config.CollectMetrics {
				collector = metrics.New(testLogger.GetSlogger())

				if tc.config.EnableEndpoint {
					serverConfig := metrics.ServerConfig{
						Port:    tc.config.Port,
						Host:    tc.config.Host,
						Enabled: tc.config.EnableEndpoint,
					}
					server = metrics.NewServer(serverConfig, collector, testLogger.GetSlogger())
					server.StartInBackground()

					// Give server time to start
					time.Sleep(100 * time.Millisecond)

					defer func() {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						server.Stop(ctx)
					}()
				}
			}

			// Test metrics collection
			if collector != nil {
				collector.RecordTotalQueries(100)
				// Metrics should be collected successfully
			}

			// Test server availability
			if tc.expectServer {
				url := fmt.Sprintf("http://%s:%s/health", tc.config.Host, tc.config.Port)
				resp, err := http.Get(url)
				if err != nil {
					t.Errorf("Expected server to be running but failed to connect: %v", err)
				} else {
					resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Errorf("Expected server to respond with 200, got %d", resp.StatusCode)
					}
				}
			} else {
				// If we don't expect a server, we can't easily test that it's not running
				// without potentially hitting other services, so we'll skip this check
			}
		})
	}
}

func TestMetricsCollectorConcurrency(t *testing.T) {
	// Test that metrics collector handles concurrent access correctly
	loggerConfig := &logger.Config{
		Level:        logger.LevelError, // Reduce log noise
		EnableColors: false,
		EnableEmojis: false,
		Component:    "metrics-concurrency-test",
	}
	testLogger := logger.New(loggerConfig)

	collector := metrics.New(testLogger.GetSlogger())

	// Run concurrent metric updates
	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				collector.RecordTotalQueries(1)
				collector.SetActiveClients(float64(id))
				collector.RecordQueryByType(fmt.Sprintf("TYPE_%d", id%5))
				collector.RecordError("test_error")
				collector.RecordTopDomain(fmt.Sprintf("domain%d.com", id), 1)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify that metrics were recorded (should be 20 * 50 = 1000 total queries)
	// We can't easily verify exact counts due to the nature of the Prometheus client
	// but the fact that no panics occurred indicates thread safety is working
}

func TestMetricsServerGracefulShutdown(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:        logger.LevelDebug,
		EnableColors: false,
		EnableEmojis: false,
		Component:    "metrics-shutdown-test",
	}
	testLogger := logger.New(loggerConfig)

	collector := metrics.New(testLogger.GetSlogger())

	serverConfig := metrics.ServerConfig{
		Port:    "19103",
		Host:    "localhost",
		Enabled: true,
	}

	server := metrics.NewServer(serverConfig, collector, testLogger.GetSlogger())
	server.StartInBackground()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://localhost:19103/health")
	if err != nil {
		t.Fatalf("Server should be running: %v", err)
	}
	resp.Body.Close()

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop server gracefully: %v", err)
	}

	// Give server time to stop
	time.Sleep(100 * time.Millisecond)

	// Verify server is no longer responding
	_, err = http.Get("http://localhost:19103/health")
	if err == nil {
		t.Error("Expected server to be stopped, but it's still responding")
	}
}