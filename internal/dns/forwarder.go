package dns

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

// Forwarder implements the DNSForwarder interface
type Forwarder struct {
	mu        sync.RWMutex
	upstreams []string
	config    ForwarderConfig
	parser    DNSParser
	lastUsed  int
	healthMap map[string]bool
}

// NewForwarder creates a new DNS forwarder
func NewForwarder(config ForwarderConfig) DNSForwarder {
	f := &Forwarder{
		upstreams: config.Upstreams,
		config:    config,
		parser:    NewParser(),
		healthMap: make(map[string]bool),
	}

	// Initialize all upstreams as healthy
	for _, upstream := range config.Upstreams {
		f.healthMap[upstream] = true
	}

	// Start health checker if enabled
	if config.HealthCheck {
		go f.healthChecker()
	}

	return f
}

// Forward sends a query to upstream DNS servers
func (f *Forwarder) Forward(ctx context.Context, query *DNSQuery) (*DNSResponse, error) {
	if !f.config.Enabled {
		return nil, ErrNoUpstreamServers
	}

	// Serialize the query
	queryData, err := f.parser.SerializeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	// Try each upstream with retries
	upstreams := f.getHealthyUpstreams()
	if len(upstreams) == 0 {
		return nil, ErrNoUpstreamServers
	}

	var lastErr error
	for attempt := 0; attempt < f.config.Retries+1; attempt++ {
		upstream := f.selectUpstream(upstreams)

		response, err := f.queryUpstream(ctx, upstream, queryData)
		if err != nil {
			lastErr = err
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("all upstream queries failed: %w", lastErr)
}

// GetUpstreams returns the list of upstream servers
func (f *Forwarder) GetUpstreams() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	upstreams := make([]string, len(f.upstreams))
	copy(upstreams, f.upstreams)
	return upstreams
}

// SetUpstreams sets the upstream DNS servers
func (f *Forwarder) SetUpstreams(upstreams []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.upstreams = upstreams

	// Reset health map
	f.healthMap = make(map[string]bool)
	for _, upstream := range upstreams {
		f.healthMap[upstream] = true
	}
}

// getHealthyUpstreams returns only healthy upstream servers
func (f *Forwarder) getHealthyUpstreams() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var healthy []string
	for _, upstream := range f.upstreams {
		if f.healthMap[upstream] {
			healthy = append(healthy, upstream)
		}
	}

	// If no healthy upstreams, return all (failsafe)
	if len(healthy) == 0 {
		return f.upstreams
	}

	return healthy
}

// selectUpstream selects an upstream server based on load balancing strategy
func (f *Forwarder) selectUpstream(upstreams []string) string {
	if len(upstreams) == 0 {
		return ""
	}

	switch f.config.LoadBalancing {
	case "random":
		return upstreams[rand.Intn(len(upstreams))]
	case "fastest":
		// For now, use round robin. In a real implementation,
		// we would track response times and select the fastest
		fallthrough
	case "round_robin":
		fallthrough
	default:
		f.mu.Lock()
		defer f.mu.Unlock()

		f.lastUsed = (f.lastUsed + 1) % len(upstreams)
		return upstreams[f.lastUsed]
	}
}

// queryUpstream sends a query to a specific upstream server
func (f *Forwarder) queryUpstream(ctx context.Context, upstream string, queryData []byte) (*DNSResponse, error) {
	// Create connection with timeout
	conn, err := net.DialTimeout("udp", upstream, f.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to upstream %s: %w", upstream, err)
	}
	defer conn.Close()

	// Set deadline for the entire operation
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(f.config.Timeout)
	}
	conn.SetDeadline(deadline)

	// Send query
	_, err = conn.Write(queryData)
	if err != nil {
		return nil, fmt.Errorf("failed to send query to %s: %w", upstream, err)
	}

	// Read response
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", upstream, err)
	}

	// Parse response
	response, err := f.parser.ParseResponse(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse response from %s: %w", upstream, err)
	}

	return response, nil
}

// healthChecker periodically checks the health of upstream servers
func (f *Forwarder) healthChecker() {
	ticker := time.NewTicker(f.config.HealthInterval)
	defer ticker.Stop()

	for range ticker.C {
		f.checkUpstreamHealth()
	}
}

// checkUpstreamHealth checks the health of all upstream servers
func (f *Forwarder) checkUpstreamHealth() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, upstream := range f.upstreams {
		healthy := f.isUpstreamHealthy(upstream)
		f.healthMap[upstream] = healthy
	}
}

// isUpstreamHealthy checks if a specific upstream server is healthy
func (f *Forwarder) isUpstreamHealthy(upstream string) bool {
	// Create a simple health check query (A record for ".")
	query := &DNSQuery{
		ID: uint16(rand.Intn(65536)),
		Question: DNSQuestion{
			Name:  ".",
			Type:  TypeNS,
			Class: ClassIN,
		},
	}

	queryData, err := f.parser.SerializeQuery(query)
	if err != nil {
		return false
	}

	// Set a shorter timeout for health checks
	conn, err := net.DialTimeout("udp", upstream, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	_, err = conn.Write(queryData)
	if err != nil {
		return false
	}

	buffer := make([]byte, 512)
	_, err = conn.Read(buffer)
	return err == nil
}
