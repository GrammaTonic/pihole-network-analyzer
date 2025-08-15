package dns

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
)

func TestDNSServer_Integration(t *testing.T) {
	// Skip integration test in CI environments where external DNS may not work
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip in CI environments
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Create logger
	loggerConfig := &logger.Config{
		Level:     logger.LogLevel("ERROR"), // Reduce noise in tests
		Component: "dns-integration-test",
	}
	testLogger := logger.New(loggerConfig)

	// Create DNS server config
	config := &Config{
		Enabled:    true,
		Host:       "127.0.0.1",
		Port:       5355, // Use a specific test port
		TCPEnabled: true,
		UDPEnabled: true,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,

		Cache: CacheConfig{
			Enabled:         true,
			MaxSize:         100,
			DefaultTTL:      30 * time.Second,
			MaxTTL:          1 * time.Hour,
			MinTTL:          5 * time.Second,
			CleanupInterval: 1 * time.Minute,
			EvictionPolicy:  "lru",
		},

		Forwarder: ForwarderConfig{
			Enabled:        true,
			Upstreams:      []string{"8.8.8.8:53"}, // Single upstream for testing
			Timeout:        5 * time.Second,
			Retries:        1,
			HealthCheck:    false, // Disable health check for testing
			HealthInterval: 30 * time.Second,
			LoadBalancing:  "round_robin",
			EDNS0Enabled:   false,
			UDPSize:        512,
		},

		LogQueries:           false, // Reduce noise
		LogLevel:             0,
		MaxConcurrentQueries: 10,
		BufferSize:           512,
	}

	// Create server
	server := NewServer(config, testLogger)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		err := server.Start(ctx)
		if err != nil {
			serverErr <- err
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test UDP query
	t.Run("UDP Query", func(t *testing.T) {
		testUDPQuery(t, config.Host, config.Port)
	})

	// Stop server
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		t.Errorf("Error stopping server: %v", err)
	}

	// Check for server errors
	select {
	case err := <-serverErr:
		if err != context.Canceled {
			t.Errorf("Server error: %v", err)
		}
	default:
		// No error, which is good
	}
}

func testUDPQuery(t *testing.T, host string, port int) {
	// Create UDP connection
	conn, err := net.Dial("udp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		t.Skipf("Could not create UDP connection: %v", err)
	}
	defer conn.Close()

	// Create a simple DNS query for google.com A record
	parser := NewParser()
	query := &DNSQuery{
		ID: 0x1234,
		Question: DNSQuestion{
			Name:  "google.com",
			Type:  TypeA,
			Class: ClassIN,
		},
	}

	queryData, err := parser.SerializeQuery(query)
	if err != nil {
		t.Fatalf("Failed to serialize query: %v", err)
	}

	// Set timeout
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send query
	_, err = conn.Write(queryData)
	if err != nil {
		t.Skipf("Failed to send query: %v", err)
	}

	// Read response
	responseData := make([]byte, 512)
	n, err := conn.Read(responseData)
	if err != nil {
		t.Skipf("Failed to read response: %v", err)
	}

	// Parse response
	response, err := parser.ParseResponse(responseData[:n])
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response
	if response.ID != query.ID {
		t.Errorf("Expected response ID %d, got %d", query.ID, response.ID)
	}

	if response.Question.Name != query.Question.Name {
		t.Errorf("Expected question name %s, got %s", query.Question.Name, response.Question.Name)
	}

	// For this test, we just verify we got a proper DNS response structure
	// The actual content depends on external DNS resolution
	t.Logf("Received DNS response with %d answers", len(response.Answers))
}

func TestDNSCache_Integration(t *testing.T) {
	// Create cache
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         10,
		DefaultTTL:      1 * time.Second, // Short TTL for testing
		MaxTTL:          1 * time.Hour,
		MinTTL:          1 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		EvictionPolicy:  "lru",
	}

	cache := NewCache(config)

	// Test cache operations
	question := DNSQuestion{
		Name:  "test.local",
		Type:  TypeA,
		Class: ClassIN,
	}

	response := &DNSResponse{
		ID:       1234,
		Question: question,
		Answers: []DNSRecord{
			{
				Name:  "test.local",
				Type:  TypeA,
				Class: ClassIN,
				TTL:   1,
				Data:  []byte{127, 0, 0, 1},
			},
		},
		ResponseCode: RCodeNoError,
	}

	// Test cache miss
	_, found := cache.Get(question)
	if found {
		t.Error("Expected cache miss")
	}

	// Test cache set
	cache.Set(question, response, 1*time.Second)

	// Test cache hit
	entry, found := cache.Get(question)
	if !found {
		t.Error("Expected cache hit")
	}

	if entry.Response.ID != response.ID {
		t.Errorf("Expected response ID %d, got %d", response.ID, entry.Response.ID)
	}

	// Test expiration
	time.Sleep(1500 * time.Millisecond) // Wait for expiration

	_, found = cache.Get(question)
	if found {
		t.Error("Expected cache miss after expiration")
	}

	// Test cleanup
	cache.Set(question, response, 1*time.Hour) // Long TTL
	_, found = cache.Get(question)
	if !found {
		t.Error("Expected cache hit before cleanup")
	}

	cache.Cleanup()
	// Entry should still be there since TTL is long
	_, found = cache.Get(question)
	if !found {
		t.Error("Expected cache hit after cleanup (not expired)")
	}

	// Test stats
	stats := cache.GetStats()
	if stats.Size != 1 {
		t.Errorf("Expected cache size 1, got %d", stats.Size)
	}

	if stats.Hits < 1 {
		t.Errorf("Expected at least 1 hit, got %d", stats.Hits)
	}
}

func TestDNSForwarder_Integration(t *testing.T) {
	// Skip if no internet connection
	if testing.Short() {
		t.Skip("Skipping forwarder integration test in short mode")
	}

	// Skip in CI environments
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping forwarder integration test in CI environment")
	}

	config := ForwarderConfig{
		Enabled:       true,
		Upstreams:     []string{"8.8.8.8:53"},
		Timeout:       5 * time.Second,
		Retries:       1,
		HealthCheck:   false,
		LoadBalancing: "round_robin",
		EDNS0Enabled:  false,
		UDPSize:       512,
	}

	forwarder := NewForwarder(config)

	query := &DNSQuery{
		ID: 0x5678,
		Question: DNSQuestion{
			Name:  "google.com",
			Type:  TypeA,
			Class: ClassIN,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := forwarder.Forward(ctx, query)
	if err != nil {
		t.Skipf("Failed to forward query (may be network issue): %v", err)
	}

	if response.ID != query.ID {
		t.Errorf("Expected response ID %d, got %d", query.ID, response.ID)
	}

	if response.Question.Name != query.Question.Name {
		t.Errorf("Expected question name %s, got %s", query.Question.Name, response.Question.Name)
	}

	t.Logf("Forwarded query successfully, got %d answers", len(response.Answers))
}
