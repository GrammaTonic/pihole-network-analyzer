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
	"pihole-analyzer/internal/web"
)

// TestWebBasic_ServerStart tests that the web server starts and responds
func TestWebBasic_ServerStart(t *testing.T) {
	// Create test logger
	testLogger := logger.New(&logger.Config{
		Component: "web-basic-test",
		Level:     "info",
	})

	// Create mock data source
	mockProvider := NewMockWebDataSource()

	// Use a specific test port to avoid conflicts
	testPort := 18081
	config := &web.Config{
		Port:            testPort,
		Host:            "127.0.0.1",
		EnableWeb:       true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnableWebSocket: true,
		WebSocketConfig: web.DefaultWebSocketConfig(),
	}

	// Create web server
	server, err := web.NewServer(config, mockProvider, testLogger)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	// Start server in background
	go func() {
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			testLogger.Error("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)

	// Test health endpoint
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to connect to health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health endpoint returned status %d", resp.StatusCode)
	}

	// Stop server
	server.Stop()

	testLogger.Success("Basic server test completed successfully")
}

// TestWebBasic_Dashboard tests the dashboard page without browser automation
func TestWebBasic_Dashboard(t *testing.T) {
	// Create test logger
	testLogger := logger.New(&logger.Config{
		Component: "web-basic-test",
		Level:     "info",
	})

	// Create mock data source
	mockProvider := NewMockWebDataSource()

	// Use a specific test port to avoid conflicts
	testPort := 18082
	config := &web.Config{
		Port:            testPort,
		Host:            "127.0.0.1",
		EnableWeb:       true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnableWebSocket: true,
		WebSocketConfig: web.DefaultWebSocketConfig(),
	}

	// Create web server
	server, err := web.NewServer(config, mockProvider, testLogger)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	// Start server in background
	go func() {
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			testLogger.Error("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)

	// Test dashboard endpoint
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Failed to connect to dashboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Dashboard returned status %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content, got %s", contentType)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Check for key dashboard elements
	expectedContent := []string{
		"Pi-hole Network Analyzer",
		"Total Queries",
		"Unique Clients",
		"Client Statistics",
		"2450", // Total queries from mock data
		"8",    // Unique clients from mock data
	}

	for _, expected := range expectedContent {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Dashboard missing expected content: %s", expected)
		}
	}

	// Stop server
	server.Stop()

	testLogger.Success("Dashboard test completed successfully")
}

// TestWebBasic_APIEndpoints tests all API endpoints
func TestWebBasic_APIEndpoints(t *testing.T) {
	// Create test logger
	testLogger := logger.New(&logger.Config{
		Component: "web-basic-test",
		Level:     "info",
	})

	// Create mock data source
	mockProvider := NewMockWebDataSource()

	// Use a specific test port to avoid conflicts
	testPort := 18083
	config := &web.Config{
		Port:            testPort,
		Host:            "127.0.0.1",
		EnableWeb:       true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnableWebSocket: true,
		WebSocketConfig: web.DefaultWebSocketConfig(),
	}

	// Create web server
	server, err := web.NewServer(config, mockProvider, testLogger)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	// Start server in background
	go func() {
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			testLogger.Error("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)

	// Test API endpoints
	testCases := []struct {
		endpoint   string
		name       string
		expectJSON bool
		expectHTTP bool
	}{
		{"/api/status", "Status API", true, false},
		{"/api/analysis", "Analysis API", true, false},
		{"/api/clients", "Clients API", true, false},
		{"/health", "Health Check", true, false},
		{"/enhanced", "Enhanced Dashboard", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(baseURL + tc.endpoint)
			if err != nil {
				t.Fatalf("Failed to request %s: %v", tc.endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", tc.endpoint, resp.StatusCode)
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if tc.expectJSON && !strings.Contains(contentType, "application/json") {
				t.Errorf("Expected JSON content type for %s, got %s", tc.endpoint, contentType)
			}
			if tc.expectHTTP && !strings.Contains(contentType, "text/html") {
				t.Errorf("Expected HTML content type for %s, got %s", tc.endpoint, contentType)
			}

			// Read body to ensure it's not empty
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Failed to read response body for %s: %v", tc.endpoint, err)
			}

			if len(body) == 0 {
				t.Errorf("Empty response body for %s", tc.endpoint)
			}
		})
	}

	// Stop server
	server.Stop()

	testLogger.Success("API endpoints test completed successfully")
}

// TestWebBasic_ConcurrentRequests tests handling concurrent requests
func TestWebBasic_ConcurrentRequests(t *testing.T) {
	// Create test logger
	testLogger := logger.New(&logger.Config{
		Component: "web-basic-test",
		Level:     "info",
	})

	// Create mock data source
	mockProvider := NewMockWebDataSource()

	// Use a specific test port to avoid conflicts
	testPort := 18084
	config := &web.Config{
		Port:            testPort,
		Host:            "127.0.0.1",
		EnableWeb:       true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnableWebSocket: true,
		WebSocketConfig: web.DefaultWebSocketConfig(),
	}

	// Create web server
	server, err := web.NewServer(config, mockProvider, testLogger)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	// Start server in background
	go func() {
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			testLogger.Error("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)

	// Perform concurrent requests to test load handling
	concurrentRequests := 5
	done := make(chan bool, concurrentRequests)
	errors := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(index int) {
			defer func() { done <- true }()

			resp, err := http.Get(baseURL + "/api/status")
			if err != nil {
				errors <- fmt.Errorf("concurrent request %d failed: %v", index, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("concurrent request %d got status %d", index, resp.StatusCode)
				return
			}
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	// Stop server
	server.Stop()

	testLogger.Success("Concurrent requests test completed successfully")
}
