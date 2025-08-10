package dns

import (
	"time"
)

// Config represents the DNS server configuration
type Config struct {
	// Server settings
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	TCPEnabled bool   `json:"tcp_enabled"`
	UDPEnabled bool   `json:"udp_enabled"`
	
	// Timeouts
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	
	// Cache configuration
	Cache CacheConfig `json:"cache"`
	
	// Forwarder configuration
	Forwarder ForwarderConfig `json:"forwarder"`
	
	// Logging
	LogQueries bool `json:"log_queries"`
	LogLevel   int  `json:"log_level"`
	
	// Performance
	MaxConcurrentQueries int `json:"max_concurrent_queries"`
	BufferSize          int `json:"buffer_size"`
}

// CacheConfig represents DNS cache configuration
type CacheConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxSize        int           `json:"max_size"`
	DefaultTTL     time.Duration `json:"default_ttl"`
	MaxTTL         time.Duration `json:"max_ttl"`
	MinTTL         time.Duration `json:"min_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	
	// Cache strategy: "lru", "lfu", "ttl"
	EvictionPolicy string `json:"eviction_policy"`
	
	// Memory limits
	MaxMemoryMB int `json:"max_memory_mb"`
}

// ForwarderConfig represents DNS forwarder configuration
type ForwarderConfig struct {
	Enabled        bool          `json:"enabled"`
	Upstreams      []string      `json:"upstreams"`
	Timeout        time.Duration `json:"timeout"`
	Retries        int           `json:"retries"`
	HealthCheck    bool          `json:"health_check"`
	HealthInterval time.Duration `json:"health_interval"`
	
	// Load balancing: "round_robin", "random", "fastest"
	LoadBalancing string `json:"load_balancing"`
	
	// EDNS0 support
	EDNS0Enabled bool `json:"edns0_enabled"`
	UDPSize      int  `json:"udp_size"`
}

// DefaultConfig returns a default DNS server configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:    false,
		Host:       "0.0.0.0",
		Port:       5353, // Alternative DNS port to avoid conflicts
		TCPEnabled: true,
		UDPEnabled: true,
		
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		
		Cache: CacheConfig{
			Enabled:         true,
			MaxSize:         10000,
			DefaultTTL:      300 * time.Second, // 5 minutes
			MaxTTL:          24 * time.Hour,    // 24 hours
			MinTTL:          10 * time.Second,
			CleanupInterval: 5 * time.Minute,
			EvictionPolicy:  "lru",
			MaxMemoryMB:     100,
		},
		
		Forwarder: ForwarderConfig{
			Enabled: true,
			Upstreams: []string{
				"8.8.8.8:53",    // Google DNS
				"8.8.4.4:53",    // Google DNS
				"1.1.1.1:53",    // Cloudflare DNS
				"1.0.0.1:53",    // Cloudflare DNS
			},
			Timeout:        5 * time.Second,
			Retries:        2,
			HealthCheck:    true,
			HealthInterval: 30 * time.Second,
			LoadBalancing:  "round_robin",
			EDNS0Enabled:   true,
			UDPSize:        4096,
		},
		
		LogQueries:           true,
		LogLevel:             1, // Info level
		MaxConcurrentQueries: 1000,
		BufferSize:           4096,
	}
}

// DNS Record Types
const (
	TypeA     uint16 = 1
	TypeNS    uint16 = 2
	TypeCNAME uint16 = 5
	TypeSOA   uint16 = 6
	TypePTR   uint16 = 12
	TypeMX    uint16 = 15
	TypeTXT   uint16 = 16
	TypeAAAA  uint16 = 28
	TypeSRV   uint16 = 33
)

// DNS Classes
const (
	ClassIN uint16 = 1 // Internet
)

// DNS Response Codes
const (
	RCodeNoError  uint8 = 0 // No error
	RCodeFormErr  uint8 = 1 // Format error
	RCodeServFail uint8 = 2 // Server failure
	RCodeNXDomain uint8 = 3 // Non-existent domain
	RCodeNotImp   uint8 = 4 // Not implemented
	RCodeRefused  uint8 = 5 // Query refused
)

// DNS Header flags
const (
	FlagQR uint16 = 1 << 15 // Query/Response
	FlagAA uint16 = 1 << 10 // Authoritative Answer
	FlagTC uint16 = 1 << 9  // Truncated
	FlagRD uint16 = 1 << 8  // Recursion Desired
	FlagRA uint16 = 1 << 7  // Recursion Available
)

// Validate validates the DNS configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidPort
	}
	
	if !c.UDPEnabled && !c.TCPEnabled {
		return ErrNoProtocolEnabled
	}
	
	if c.Cache.MaxSize < 0 {
		return ErrInvalidCacheSize
	}
	
	if c.MaxConcurrentQueries < 1 {
		return ErrInvalidConcurrency
	}
	
	return nil
}