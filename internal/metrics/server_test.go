package metrics

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "9090",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.collector != collector {
		t.Error("Expected collector to be set correctly")
	}

	if server.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	expectedAddr := "localhost:9090"
	if server.server.Addr != expectedAddr {
		t.Errorf("Expected server address to be %s, got %s", expectedAddr, server.server.Addr)
	}
}

func TestServerStartInBackground(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19090", // Use a different port to avoid conflicts
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)

	// Start server in background
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test that the server is running by making a request to health endpoint
	resp, err := http.Get("http://localhost:19090/health")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := `{"status":"healthy","component":"metrics-server"}`
	if string(body) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(body))
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServerEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19091", // Use a different port
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name               string
		endpoint           string
		expectedStatusCode int
		expectedContent    string
	}{
		{
			name:               "Root endpoint",
			endpoint:           "/",
			expectedStatusCode: http.StatusOK,
			expectedContent:    "Pi-hole Network Analyzer Metrics Server",
		},
		{
			name:               "Health endpoint",
			endpoint:           "/health",
			expectedStatusCode: http.StatusOK,
			expectedContent:    `{"status":"healthy","component":"metrics-server"}`,
		},
		{
			name:               "Metrics endpoint",
			endpoint:           "/metrics",
			expectedStatusCode: http.StatusOK,
			expectedContent:    "pihole_analyzer_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:19091%s", tt.endpoint)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if !strings.Contains(string(body), tt.expectedContent) {
				t.Errorf("Expected response to contain '%s', got: %s", tt.expectedContent, string(body))
			}
		})
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServerMetricsEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	// Add some test data to metrics
	collector.RecordTotalQueries(100)
	collector.SetActiveClients(5)
	collector.RecordQueryByType("A")
	collector.RecordQueryByStatus("allowed")

	config := ServerConfig{
		Port:    "19092", // Use a different port
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test metrics endpoint
	resp, err := http.Get("http://localhost:19092/metrics")
	if err != nil {
		t.Fatalf("Failed to connect to metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	metrics := string(body)

	// Check that our metrics are present
	expectedMetrics := []string{
		"pihole_analyzer_total_queries",
		"pihole_analyzer_active_clients",
		"pihole_analyzer_queries_by_type_total",
		"pihole_analyzer_queries_by_status_total",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(metrics, metric) {
			t.Errorf("Expected metrics response to contain %s", metric)
		}
	}

	// Check for specific metric values
	if !strings.Contains(metrics, `pihole_analyzer_total_queries 100`) {
		t.Error("Expected total queries metric to show value 100")
	}

	if !strings.Contains(metrics, `pihole_analyzer_active_clients 5`) {
		t.Error("Expected active clients metric to show value 5")
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServerGetCollector(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19093",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)

	returnedCollector := server.GetCollector()
	if returnedCollector != collector {
		t.Error("Expected GetCollector to return the same collector instance")
	}
}

func TestServerStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19094",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://localhost:19094/health")
	if err != nil {
		t.Fatalf("Server should be running: %v", err)
	}
	resp.Body.Close()

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Give the server a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Verify server is stopped
	_, err = http.Get("http://localhost:19094/health")
	if err == nil {
		t.Error("Expected server to be stopped, but it's still responding")
	}
}

func TestServerStopTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19095",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the server with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should either succeed quickly or timeout
	err := server.Stop(ctx)
	// We don't test for a specific error since shutdown might complete quickly
	_ = err

	// Give it a proper stop
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	server.Stop(ctx2)
}

func TestServerConfigVariations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	collector := New(logger)

	tests := []struct {
		name   string
		config ServerConfig
	}{
		{
			name: "Default config",
			config: ServerConfig{
				Port:    "9090",
				Host:    "localhost",
				Enabled: true,
			},
		},
		{
			name: "Different port",
			config: ServerConfig{
				Port:    "8080",
				Host:    "localhost",
				Enabled: true,
			},
		},
		{
			name: "Different host",
			config: ServerConfig{
				Port:    "9090",
				Host:    "127.0.0.1",
				Enabled: true,
			},
		},
		{
			name: "Disabled server",
			config: ServerConfig{
				Port:    "9090",
				Host:    "localhost",
				Enabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.config, collector, logger)

			if server == nil {
				t.Fatal("Expected server to be created, got nil")
			}

			expectedAddr := tt.config.Host + ":" + tt.config.Port
			if server.server.Addr != expectedAddr {
				t.Errorf("Expected server address to be %s, got %s", expectedAddr, server.server.Addr)
			}
		})
	}
}

// Test server with concurrent requests
func TestServerConcurrency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19096",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make concurrent requests
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get("http://localhost:19096/health")
			if err != nil {
				done <- err
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				done <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				return
			}
			done <- nil
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

// Benchmark tests
func BenchmarkServerHealthEndpoint(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	config := ServerConfig{
		Port:    "19097",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:19097/health")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func BenchmarkServerMetricsEndpoint(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	collector := New(logger)

	// Add some test data
	collector.RecordTotalQueries(1000)
	collector.SetActiveClients(10)

	config := ServerConfig{
		Port:    "19098",
		Host:    "localhost",
		Enabled: true,
	}

	server := NewServer(config, collector, logger)
	server.StartInBackground()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:19098/metrics")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}
