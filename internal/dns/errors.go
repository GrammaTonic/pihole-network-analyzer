package dns

import "errors"

// DNS Server errors
var (
	ErrInvalidPort          = errors.New("invalid DNS server port")
	ErrNoProtocolEnabled    = errors.New("no protocol enabled (UDP or TCP)")
	ErrInvalidCacheSize     = errors.New("invalid cache size")
	ErrInvalidConcurrency   = errors.New("invalid max concurrent queries")
	ErrServerNotStarted     = errors.New("DNS server not started")
	ErrServerAlreadyRunning = errors.New("DNS server already running")
	ErrInvalidQuery         = errors.New("invalid DNS query")
	ErrQueryTimeout         = errors.New("DNS query timeout")
	ErrNoUpstreamServers    = errors.New("no upstream DNS servers configured")
	ErrCacheFull            = errors.New("DNS cache is full")
	ErrInvalidDNSMessage    = errors.New("invalid DNS message format")
	ErrUnsupportedQType     = errors.New("unsupported DNS query type")
	ErrServerShutdown       = errors.New("DNS server is shutting down")
)

// DNS Protocol errors
var (
	ErrShortMessage     = errors.New("DNS message too short")
	ErrInvalidHeader    = errors.New("invalid DNS header")
	ErrInvalidQuestion  = errors.New("invalid DNS question")
	ErrInvalidRecord    = errors.New("invalid DNS record")
	ErrTruncatedMessage = errors.New("DNS message truncated")
	ErrInvalidName      = errors.New("invalid DNS name")
	ErrNameTooLong      = errors.New("DNS name too long")
	ErrInvalidLabel     = errors.New("invalid DNS label")
	ErrCompressionLoop  = errors.New("DNS name compression loop detected")
)

// Cache errors
var (
	ErrCacheDisabled     = errors.New("DNS cache is disabled")
	ErrCacheEntryExpired = errors.New("cache entry expired")
	ErrCacheKeyNotFound  = errors.New("cache key not found")
)
