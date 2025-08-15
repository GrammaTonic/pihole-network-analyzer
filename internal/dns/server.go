package dns

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"pihole-analyzer/internal/logger"
)

// Server implements the DNSServer interface
type Server struct {
	config    *Config
	logger    *logger.Logger
	cache     DNSCache
	forwarder DNSForwarder
	parser    DNSParser

	// Server state
	running     atomic.Bool
	udpConn     *net.UDPConn
	tcpListener *net.TCPListener

	// Statistics
	stats   ServerStats
	statsMu sync.RWMutex

	// Shutdown
	shutdownCh chan struct{}
	wg         sync.WaitGroup
}

// NewServer creates a new DNS server
func NewServer(config *Config, logger *logger.Logger) DNSServer {
	return &Server{
		config:     config,
		logger:     logger,
		cache:      NewCache(config.Cache),
		forwarder:  NewForwarder(config.Forwarder),
		parser:     NewParser(),
		shutdownCh: make(chan struct{}),
		stats: ServerStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the DNS server
func (s *Server) Start(ctx context.Context) error {
	if s.running.Load() {
		return ErrServerAlreadyRunning
	}

	s.logger.InfoFields("Starting DNS server", map[string]any{
		"host":          s.config.Host,
		"port":          s.config.Port,
		"udp_enabled":   s.config.UDPEnabled,
		"tcp_enabled":   s.config.TCPEnabled,
		"cache_enabled": s.config.Cache.Enabled,
	})

	// Validate configuration
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Start UDP server if enabled
	if s.config.UDPEnabled {
		if err := s.startUDPServer(); err != nil {
			return fmt.Errorf("failed to start UDP server: %w", err)
		}
		s.logger.Success("UDP DNS server started on %s:%d", s.config.Host, s.config.Port)
	}

	// Start TCP server if enabled
	if s.config.TCPEnabled {
		if err := s.startTCPServer(); err != nil {
			s.stopUDPServer() // Clean up UDP if TCP fails
			return fmt.Errorf("failed to start TCP server: %w", err)
		}
		s.logger.Success("TCP DNS server started on %s:%d", s.config.Host, s.config.Port)
	}

	s.running.Store(true)

	// Start cache cleanup routine
	if s.config.Cache.Enabled {
		s.wg.Add(1)
		go s.cacheCleanupRoutine()
	}

	s.logger.Success("🚀 DNS server started successfully")

	// Wait for shutdown signal
	<-s.shutdownCh

	return nil
}

// Stop stops the DNS server gracefully
func (s *Server) Stop(ctx context.Context) error {
	if !s.running.Load() {
		return ErrServerNotStarted
	}

	s.logger.Info("Stopping DNS server...")

	s.running.Store(false)
	close(s.shutdownCh)

	// Stop servers
	s.stopUDPServer()
	s.stopTCPServer()

	// Wait for all goroutines to finish
	s.wg.Wait()

	s.logger.Success("✅ DNS server stopped gracefully")
	return nil
}

// GetStats returns server statistics
func (s *Server) GetStats() *ServerStats {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	stats := s.stats
	return &stats
}

// HandleQuery processes a DNS query
func (s *Server) HandleQuery(ctx context.Context, query *DNSQuery) (*DNSResponse, error) {
	start := time.Now()

	// Update statistics
	s.updateStats(func(stats *ServerStats) {
		stats.QueriesReceived++
		if query.Protocol == "udp" {
			stats.UDPQueries++
		} else {
			stats.TCPQueries++
		}
	})

	if s.config.LogQueries {
		s.logger.InfoFields("DNS query received", map[string]any{
			"id":       query.ID,
			"domain":   query.Question.Name,
			"type":     query.Question.Type,
			"protocol": query.Protocol,
			"client":   query.Client.String(),
		})
	}

	// Check cache first
	if s.config.Cache.Enabled {
		if entry, found := s.cache.Get(query.Question); found {
			s.updateStats(func(stats *ServerStats) {
				stats.CacheHits++
				stats.QueriesAnswered++
			})

			response := entry.Response
			response.ID = query.ID // Use the query ID
			response.Cached = true
			response.ResponseTime = time.Since(start)

			if s.config.LogQueries {
				s.logger.InfoFields("Cache hit", map[string]any{
					"domain":        query.Question.Name,
					"response_time": response.ResponseTime,
				})
			}

			return response, nil
		}

		s.updateStats(func(stats *ServerStats) {
			stats.CacheMisses++
		})
	}

	// Forward to upstream
	response, err := s.forwarder.Forward(ctx, query)
	if err != nil {
		s.updateStats(func(stats *ServerStats) {
			stats.Errors++
		})

		s.logger.ErrorFields("Failed to forward query", map[string]any{
			"domain": query.Question.Name,
			"error":  err.Error(),
		})

		// Return SERVFAIL response
		return &DNSResponse{
			ID:           query.ID,
			Question:     query.Question,
			ResponseCode: RCodeServFail,
			ResponseTime: time.Since(start),
		}, nil
	}

	// Update response
	response.ID = query.ID
	response.ResponseTime = time.Since(start)

	// Cache the response if caching is enabled
	if s.config.Cache.Enabled && response.ResponseCode == RCodeNoError {
		// Calculate TTL from the response records
		ttl := s.calculateTTL(response)
		if ttl > 0 {
			s.cache.Set(query.Question, response, ttl)
		}
	}

	s.updateStats(func(stats *ServerStats) {
		stats.QueriesForwarded++
		stats.QueriesAnswered++

		// Update average latency
		total := stats.QueriesAnswered
		if total > 1 {
			stats.AverageLatency = time.Duration(
				(int64(stats.AverageLatency)*(total-1) + int64(response.ResponseTime)) / total,
			)
		} else {
			stats.AverageLatency = response.ResponseTime
		}
	})

	if s.config.LogQueries {
		s.logger.InfoFields("Query forwarded", map[string]any{
			"domain":        query.Question.Name,
			"response_code": response.ResponseCode,
			"response_time": response.ResponseTime,
			"answers":       len(response.Answers),
		})
	}

	return response, nil
}

// startUDPServer starts the UDP DNS server
func (s *Server) startUDPServer() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	s.udpConn = conn

	s.wg.Add(1)
	go s.handleUDPQueries()

	return nil
}

// startTCPServer starts the TCP DNS server
func (s *Server) startTCPServer() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.tcpListener = listener.(*net.TCPListener)

	s.wg.Add(1)
	go s.handleTCPConnections()

	return nil
}

// stopUDPServer stops the UDP server
func (s *Server) stopUDPServer() {
	if s.udpConn != nil {
		s.udpConn.Close()
		s.udpConn = nil
	}
}

// stopTCPServer stops the TCP server
func (s *Server) stopTCPServer() {
	if s.tcpListener != nil {
		s.tcpListener.Close()
		s.tcpListener = nil
	}
}

// handleUDPQueries handles incoming UDP DNS queries
func (s *Server) handleUDPQueries() {
	defer s.wg.Done()

	buffer := make([]byte, s.config.BufferSize)

	for s.running.Load() {
		s.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, clientAddr, err := s.udpConn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is expected, continue loop
			}
			if s.running.Load() {
				s.logger.ErrorFields("UDP read error", map[string]any{
					"error": err.Error(),
				})
			}
			continue
		}

		// Handle query in goroutine
		go s.handleUDPQuery(buffer[:n], clientAddr)
	}
}

// handleUDPQuery handles a single UDP DNS query
func (s *Server) handleUDPQuery(data []byte, clientAddr *net.UDPAddr) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ReadTimeout)
	defer cancel()

	// Parse query
	query, err := s.parser.ParseQuery(data)
	if err != nil {
		s.logger.ErrorFields("Failed to parse UDP query", map[string]any{
			"client": clientAddr.String(),
			"error":  err.Error(),
		})
		return
	}

	query.Client = clientAddr
	query.Protocol = "udp"

	// Process query
	response, err := s.HandleQuery(ctx, query)
	if err != nil {
		s.logger.ErrorFields("Failed to process UDP query", map[string]any{
			"client": clientAddr.String(),
			"error":  err.Error(),
		})
		return
	}

	// Serialize response
	responseData, err := s.parser.SerializeResponse(response)
	if err != nil {
		s.logger.ErrorFields("Failed to serialize UDP response", map[string]any{
			"client": clientAddr.String(),
			"error":  err.Error(),
		})
		return
	}

	// Send response
	s.udpConn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
	_, err = s.udpConn.WriteToUDP(responseData, clientAddr)
	if err != nil {
		s.logger.ErrorFields("Failed to send UDP response", map[string]any{
			"client": clientAddr.String(),
			"error":  err.Error(),
		})
	}
}

// handleTCPConnections handles incoming TCP connections
func (s *Server) handleTCPConnections() {
	defer s.wg.Done()

	for s.running.Load() {
		s.tcpListener.SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := s.tcpListener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is expected, continue loop
			}
			if s.running.Load() {
				s.logger.ErrorFields("TCP accept error", map[string]any{
					"error": err.Error(),
				})
			}
			continue
		}

		// Handle connection in goroutine
		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection handles a single TCP connection
func (s *Server) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

	// Read message length (TCP DNS uses 2-byte length prefix)
	lengthBuf := make([]byte, 2)
	_, err := conn.Read(lengthBuf)
	if err != nil {
		s.logger.ErrorFields("Failed to read TCP message length", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
		return
	}

	messageLength := int(lengthBuf[0])<<8 | int(lengthBuf[1])
	if messageLength > s.config.BufferSize {
		s.logger.ErrorFields("TCP message too large", map[string]any{
			"client": conn.RemoteAddr().String(),
			"length": messageLength,
			"max":    s.config.BufferSize,
		})
		return
	}

	// Read message
	messageBuf := make([]byte, messageLength)
	_, err = conn.Read(messageBuf)
	if err != nil {
		s.logger.ErrorFields("Failed to read TCP message", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ReadTimeout)
	defer cancel()

	// Parse query
	query, err := s.parser.ParseQuery(messageBuf)
	if err != nil {
		s.logger.ErrorFields("Failed to parse TCP query", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
		return
	}

	query.Client = conn.RemoteAddr()
	query.Protocol = "tcp"

	// Process query
	response, err := s.HandleQuery(ctx, query)
	if err != nil {
		s.logger.ErrorFields("Failed to process TCP query", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
		return
	}

	// Serialize response
	responseData, err := s.parser.SerializeResponse(response)
	if err != nil {
		s.logger.ErrorFields("Failed to serialize TCP response", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
		return
	}

	// Send response with length prefix
	conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

	lengthPrefix := []byte{byte(len(responseData) >> 8), byte(len(responseData) & 0xFF)}
	_, err = conn.Write(append(lengthPrefix, responseData...))
	if err != nil {
		s.logger.ErrorFields("Failed to send TCP response", map[string]any{
			"client": conn.RemoteAddr().String(),
			"error":  err.Error(),
		})
	}
}

// cacheCleanupRoutine periodically cleans up expired cache entries
func (s *Server) cacheCleanupRoutine() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.Cache.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cache.Cleanup()
		case <-s.shutdownCh:
			return
		}
	}
}

// calculateTTL calculates the TTL for caching from response records
func (s *Server) calculateTTL(response *DNSResponse) time.Duration {
	if len(response.Answers) == 0 {
		return s.config.Cache.DefaultTTL
	}

	// Use the minimum TTL from all answer records
	minTTL := response.Answers[0].TTL
	for _, record := range response.Answers {
		if record.TTL < minTTL {
			minTTL = record.TTL
		}
	}

	ttl := time.Duration(minTTL) * time.Second

	// Apply TTL limits
	if ttl < s.config.Cache.MinTTL {
		ttl = s.config.Cache.MinTTL
	}
	if ttl > s.config.Cache.MaxTTL {
		ttl = s.config.Cache.MaxTTL
	}

	return ttl
}

// updateStats safely updates server statistics
func (s *Server) updateStats(updater func(*ServerStats)) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	updater(&s.stats)
}
