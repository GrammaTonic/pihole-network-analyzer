package dns

import (
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
)

func TestFactory_CreateServer(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:     logger.LogLevel("INFO"),
		Component: "dns-test",
	}
	testLogger := logger.New(loggerConfig)

	factory := NewFactory(testLogger)

	config := &Config{
		Enabled:    true,
		Host:       "127.0.0.1",
		Port:       5353,
		TCPEnabled: true,
		UDPEnabled: true,

		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,

		Cache: CacheConfig{
			Enabled:         true,
			MaxSize:         1000,
			DefaultTTL:      300 * time.Second,
			MaxTTL:          24 * time.Hour,
			MinTTL:          10 * time.Second,
			CleanupInterval: 5 * time.Minute,
			EvictionPolicy:  "lru",
		},

		Forwarder: ForwarderConfig{
			Enabled:   true,
			Upstreams: []string{"8.8.8.8:53", "8.8.4.4:53"},
			Timeout:   5 * time.Second,
			Retries:   2,
		},

		LogQueries:           true,
		LogLevel:             1,
		MaxConcurrentQueries: 100,
		BufferSize:           4096,
	}

	server, err := factory.CreateServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Error("Expected non-nil server")
	}

	// Test server stats
	stats := server.GetStats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestFactory_CreateCache(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:     logger.LogLevel("INFO"),
		Component: "dns-test",
	}
	testLogger := logger.New(loggerConfig)

	factory := NewFactory(testLogger)

	config := &CacheConfig{
		Enabled:         true,
		MaxSize:         1000,
		DefaultTTL:      300 * time.Second,
		MaxTTL:          24 * time.Hour,
		MinTTL:          10 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}

	cache, err := factory.CreateCache(config)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	if cache == nil {
		t.Error("Expected non-nil cache")
	}

	// Test cache functionality
	question := DNSQuestion{
		Name:  "test.com",
		Type:  TypeA,
		Class: ClassIN,
	}

	_, found := cache.Get(question)
	if found {
		t.Error("Expected cache miss for new entry")
	}

	response := &DNSResponse{
		ID:           1234,
		Question:     question,
		ResponseCode: RCodeNoError,
	}

	cache.Set(question, response, 300*time.Second)

	_, found = cache.Get(question)
	if !found {
		t.Error("Expected cache hit after setting entry")
	}
}

func TestFactory_CreateForwarder(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:     logger.LogLevel("INFO"),
		Component: "dns-test",
	}
	testLogger := logger.New(loggerConfig)

	factory := NewFactory(testLogger)

	config := &ForwarderConfig{
		Enabled:        true,
		Upstreams:      []string{"8.8.8.8:53", "8.8.4.4:53"},
		Timeout:        5 * time.Second,
		Retries:        2,
		HealthCheck:    false, // Disable for testing
		HealthInterval: 30 * time.Second,
		LoadBalancing:  "round_robin",
		EDNS0Enabled:   false,
		UDPSize:        4096,
	}

	forwarder, err := factory.CreateForwarder(config)
	if err != nil {
		t.Fatalf("Failed to create forwarder: %v", err)
	}

	if forwarder == nil {
		t.Error("Expected non-nil forwarder")
	}

	// Test forwarder functionality
	upstreams := forwarder.GetUpstreams()
	if len(upstreams) != 2 {
		t.Errorf("Expected 2 upstreams, got %d", len(upstreams))
	}

	expectedUpstreams := []string{"8.8.8.8:53", "8.8.4.4:53"}
	for i, upstream := range upstreams {
		if upstream != expectedUpstreams[i] {
			t.Errorf("Expected upstream %s, got %s", expectedUpstreams[i], upstream)
		}
	}
}

func TestFactory_CreateParser(t *testing.T) {
	loggerConfig := &logger.Config{
		Level:     logger.LogLevel("INFO"),
		Component: "dns-test",
	}
	testLogger := logger.New(loggerConfig)

	factory := NewFactory(testLogger)

	parser := factory.CreateParser()
	if parser == nil {
		t.Error("Expected non-nil parser")
	}

	// Test parser functionality
	query := &DNSQuery{
		ID: 0x1234,
		Question: DNSQuestion{
			Name:  "example.com",
			Type:  TypeA,
			Class: ClassIN,
		},
	}

	data, err := parser.SerializeQuery(query)
	if err != nil {
		t.Fatalf("Failed to serialize query: %v", err)
	}

	parsedQuery, err := parser.ParseQuery(data)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	if parsedQuery.ID != query.ID {
		t.Errorf("Expected ID %d, got %d", query.ID, parsedQuery.ID)
	}

	if parsedQuery.Question.Name != query.Question.Name {
		t.Errorf("Expected name %s, got %s", query.Question.Name, parsedQuery.Question.Name)
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 5353,
				TCPEnabled:           true,
				UDPEnabled:           true,
				MaxConcurrentQueries: 100,
				Cache: CacheConfig{
					MaxSize: 1000,
				},
			},
			expectError: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 0,
				TCPEnabled:           true,
				UDPEnabled:           true,
				MaxConcurrentQueries: 100,
				Cache: CacheConfig{
					MaxSize: 1000,
				},
			},
			expectError: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 65536,
				TCPEnabled:           true,
				UDPEnabled:           true,
				MaxConcurrentQueries: 100,
				Cache: CacheConfig{
					MaxSize: 1000,
				},
			},
			expectError: true,
		},
		{
			name: "no protocols enabled",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 5353,
				TCPEnabled:           false,
				UDPEnabled:           false,
				MaxConcurrentQueries: 100,
				Cache: CacheConfig{
					MaxSize: 1000,
				},
			},
			expectError: true,
		},
		{
			name: "invalid cache size",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 5353,
				TCPEnabled:           true,
				UDPEnabled:           true,
				MaxConcurrentQueries: 100,
				Cache: CacheConfig{
					MaxSize: -1,
				},
			},
			expectError: true,
		},
		{
			name: "invalid concurrency",
			config: &Config{
				Enabled:              true,
				Host:                 "127.0.0.1",
				Port:                 5353,
				TCPEnabled:           true,
				UDPEnabled:           true,
				MaxConcurrentQueries: 0,
				Cache: CacheConfig{
					MaxSize: 1000,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}
