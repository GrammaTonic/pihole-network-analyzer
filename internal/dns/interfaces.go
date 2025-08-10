package dns

import (
	"context"
	"net"
	"time"
)

// DNSQuery represents a DNS query
type DNSQuery struct {
	ID       uint16
	Question DNSQuestion
	Client   net.Addr
	Protocol string // "udp" or "tcp"
}

// DNSQuestion represents the question section of a DNS query
type DNSQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

// DNSResponse represents a DNS response
type DNSResponse struct {
	ID           uint16
	Question     DNSQuestion
	Answers      []DNSRecord
	Authorities  []DNSRecord
	Additional   []DNSRecord
	ResponseCode uint8
	Cached       bool
	ResponseTime time.Duration
}

// DNSRecord represents a DNS resource record
type DNSRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

// CacheEntry represents a cached DNS response
type CacheEntry struct {
	Response   *DNSResponse
	ExpiresAt  time.Time
	AccessTime time.Time
	HitCount   int64
}

// DNSServer defines the main DNS server interface
type DNSServer interface {
	// Start starts the DNS server
	Start(ctx context.Context) error
	
	// Stop stops the DNS server gracefully
	Stop(ctx context.Context) error
	
	// GetStats returns server statistics
	GetStats() *ServerStats
	
	// HandleQuery processes a DNS query
	HandleQuery(ctx context.Context, query *DNSQuery) (*DNSResponse, error)
}

// DNSCache defines the DNS caching interface
type DNSCache interface {
	// Get retrieves a cached response
	Get(question DNSQuestion) (*CacheEntry, bool)
	
	// Set stores a response in cache
	Set(question DNSQuestion, response *DNSResponse, ttl time.Duration)
	
	// Delete removes an entry from cache
	Delete(question DNSQuestion)
	
	// Clear removes all entries from cache
	Clear()
	
	// GetStats returns cache statistics
	GetStats() *CacheStats
	
	// Cleanup removes expired entries
	Cleanup()
}

// DNSForwarder defines the interface for forwarding DNS queries
type DNSForwarder interface {
	// Forward sends a query to upstream DNS servers
	Forward(ctx context.Context, query *DNSQuery) (*DNSResponse, error)
	
	// GetUpstreams returns the list of upstream servers
	GetUpstreams() []string
	
	// SetUpstreams sets the upstream DNS servers
	SetUpstreams(upstreams []string)
}

// DNSParser defines the interface for parsing DNS messages
type DNSParser interface {
	// ParseQuery parses a DNS query from raw bytes
	ParseQuery(data []byte) (*DNSQuery, error)
	
	// SerializeResponse serializes a DNS response to raw bytes
	SerializeResponse(response *DNSResponse) ([]byte, error)
	
	// ParseResponse parses a DNS response from raw bytes
	ParseResponse(data []byte) (*DNSResponse, error)
	
	// SerializeQuery serializes a DNS query to raw bytes
	SerializeQuery(query *DNSQuery) ([]byte, error)
}

// ServerStats contains DNS server statistics
type ServerStats struct {
	StartTime        time.Time
	QueriesReceived  int64
	QueriesAnswered  int64
	QueriesForwarded int64
	CacheHits        int64
	CacheMisses      int64
	Errors           int64
	AverageLatency   time.Duration
	UDPQueries       int64
	TCPQueries       int64
}

// CacheStats contains DNS cache statistics
type CacheStats struct {
	Size        int
	MaxSize     int
	HitRate     float64
	Hits        int64
	Misses      int64
	Evictions   int64
	LastCleanup time.Time
}

// DNSServerFactory creates DNS server components
type DNSServerFactory interface {
	// CreateServer creates a DNS server instance
	CreateServer(config *Config) (DNSServer, error)
	
	// CreateCache creates a DNS cache instance
	CreateCache(config *CacheConfig) (DNSCache, error)
	
	// CreateForwarder creates a DNS forwarder instance
	CreateForwarder(config *ForwarderConfig) (DNSForwarder, error)
	
	// CreateParser creates a DNS parser instance
	CreateParser() DNSParser
}