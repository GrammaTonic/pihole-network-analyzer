package dhcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pihole-analyzer/internal/types"
)

// server implements the DHCPServer interface
type server struct {
	config        *types.DHCPConfig
	storage       DHCPStorage
	leaseManager  DHCPLeaseManager
	packetHandler DHCPPacketHandler
	networking    DHCPNetworking
	security      DHCPSecurity
	logger        *slog.Logger

	// Server state
	running    bool
	startTime  time.Time
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	statistics *types.DHCPStatistics
}

// Start starts the DHCP server
func (s *server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("DHCP server is already running")
	}

	s.logger.Info("Starting DHCP server",
		slog.String("interface", s.config.Interface),
		slog.String("listen_address", s.config.ListenAddress),
		slog.Int("port", s.config.Port))

	// Initialize storage
	if err := s.storage.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Bind to network interface
	if err := s.networking.Bind(s.config.ListenAddress, s.config.Port); err != nil {
		return fmt.Errorf("failed to bind to network: %w", err)
	}

	// Initialize statistics
	s.statistics = &types.DHCPStatistics{
		RequestsByType: make(map[string]int64),
		RequestsByHour: make(map[string]int64),
		TopClients:     make([]types.DHCPClientStat, 0),
		RecentActivity: make([]types.DHCPActivity, 0),
	}

	// Create context for server operations
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Start server goroutines
	go s.packetProcessor()
	go s.leaseCleanupWorker()
	go s.statisticsUpdater()

	s.running = true
	s.startTime = time.Now()

	s.logger.Info("DHCP server started successfully")
	return nil
}

// Stop stops the DHCP server
func (s *server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("DHCP server is not running")
	}

	s.logger.Info("Stopping DHCP server")

	// Cancel server context
	if s.cancel != nil {
		s.cancel()
	}

	// Close networking
	if err := s.networking.Close(); err != nil {
		s.logger.Error("Error closing networking", slog.String("error", err.Error()))
	}

	// Close storage
	if err := s.storage.Close(); err != nil {
		s.logger.Error("Error closing storage", slog.String("error", err.Error()))
	}

	s.running = false
	s.logger.Info("DHCP server stopped")
	return nil
}

// IsRunning returns whether the DHCP server is running
func (s *server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetConfig returns the current DHCP configuration
func (s *server) GetConfig() *types.DHCPConfig {
	return s.config
}

// UpdateConfig updates the DHCP server configuration
func (s *server) UpdateConfig(config *types.DHCPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	s.config = config
	s.logger.Info("DHCP configuration updated")
	return nil
}

// ValidateConfig validates a DHCP configuration
func (s *server) ValidateConfig(config *types.DHCPConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate pool configuration
	if config.Pool.StartIP == "" || config.Pool.EndIP == "" {
		return fmt.Errorf("IP pool start and end addresses must be specified")
	}

	// Additional validation would go here (IP range validation, etc.)
	return nil
}

// GetLeases returns all DHCP leases
func (s *server) GetLeases(ctx context.Context) ([]types.DHCPLease, error) {
	return s.leaseManager.GetAllLeases(ctx)
}

// GetLease returns a specific DHCP lease
func (s *server) GetLease(ctx context.Context, identifier string) (*types.DHCPLease, error) {
	// Try to find by MAC address first, then by IP
	if lease, err := s.storage.LoadLeaseByMAC(ctx, identifier); err == nil && lease != nil {
		return lease, nil
	}

	return s.storage.LoadLeaseByIP(ctx, identifier)
}

// CreateReservation creates a new static IP reservation
func (s *server) CreateReservation(ctx context.Context, reservation *types.DHCPReservation) error {
	s.logger.Info("Creating DHCP reservation",
		slog.String("mac", reservation.MAC),
		slog.String("ip", reservation.IP),
		slog.String("hostname", reservation.Hostname))

	return s.leaseManager.AddReservation(ctx, reservation)
}

// DeleteReservation removes a static IP reservation
func (s *server) DeleteReservation(ctx context.Context, mac string) error {
	s.logger.Info("Deleting DHCP reservation", slog.String("mac", mac))
	return s.leaseManager.RemoveReservation(ctx, mac)
}

// GetStatus returns the current server status
func (s *server) GetStatus(ctx context.Context) (*types.DHCPServerStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	poolInfo, err := s.getPoolInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool info: %w", err)
	}

	status := &types.DHCPServerStatus{
		Running:       s.running,
		StartTime:     s.getStartTimeString(),
		ConfigValid:   s.ValidateConfig(s.config) == nil,
		Interface:     s.config.Interface,
		ListenAddress: s.config.ListenAddress,
		PoolInfo:      *poolInfo,
		Statistics:    *s.statistics,
		RecentErrors:  make([]types.DHCPError, 0), // Would be populated from actual error tracking
		Version:       "1.0.0",
	}

	return status, nil
}

// GetStatistics returns DHCP server statistics
func (s *server) GetStatistics(ctx context.Context) (*types.DHCPStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Update current statistics
	stats := *s.statistics

	// Get active leases count
	leases, err := s.leaseManager.GetAllLeases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get leases: %w", err)
	}

	activeCount := 0
	for _, lease := range leases {
		if lease.State == types.LeaseStateActive {
			activeCount++
		}
	}

	stats.ActiveLeases = activeCount
	stats.Uptime = time.Since(s.startTime).String()

	return &stats, nil
}

// HandleDHCPRequest processes a DHCP request (internal method)
func (s *server) HandleDHCPRequest(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error) {
	s.logger.Debug("Handling DHCP request",
		slog.Int("message_type", request.MessageType),
		slog.String("client_mac", request.ClientMAC))

	// Update statistics
	s.updateRequestStatistics(request)

	// Security checks
	if allowed, err := s.security.IsClientAllowed(ctx, request.ClientMAC, request.ClientID); err != nil {
		return nil, fmt.Errorf("security check failed: %w", err)
	} else if !allowed {
		s.logger.Warn("Client not allowed", slog.String("mac", request.ClientMAC))
		return nil, fmt.Errorf("client not allowed")
	}

	// Route to appropriate handler based on message type
	switch request.MessageType {
	case 1: // DHCP Discover
		return s.packetHandler.ProcessDiscover(ctx, request)
	case 3: // DHCP Request
		return s.packetHandler.ProcessRequest(ctx, request)
	case 7: // DHCP Release
		err := s.packetHandler.ProcessRelease(ctx, request)
		return nil, err
	case 4: // DHCP Decline
		err := s.packetHandler.ProcessDecline(ctx, request)
		return nil, err
	case 8: // DHCP Inform
		return s.packetHandler.ProcessInform(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported DHCP message type: %d", request.MessageType)
	}
}

// packetProcessor runs the main packet processing loop
func (s *server) packetProcessor() {
	s.logger.Info("Starting DHCP packet processor")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Stopping packet processor")
			return
		default:
			// Receive packet from network
			data, err := s.networking.ReceivePacket(s.ctx)
			if err != nil {
				s.logger.Error("Failed to receive packet", slog.String("error", err.Error()))
				continue
			}

			// Parse DHCP request (simplified - would need actual DHCP packet parsing)
			request, err := s.parsePacket(data)
			if err != nil {
				s.logger.Error("Failed to parse packet", slog.String("error", err.Error()))
				continue
			}

			// Handle request
			response, err := s.HandleDHCPRequest(s.ctx, request)
			if err != nil {
				s.logger.Error("Failed to handle DHCP request", slog.String("error", err.Error()))
				continue
			}

			// Send response if we have one
			if response != nil {
				responseData, err := s.buildPacket(response)
				if err != nil {
					s.logger.Error("Failed to build response packet", slog.String("error", err.Error()))
					continue
				}

				if err := s.networking.SendPacket(s.ctx, responseData, response.ClientMAC, 68); err != nil {
					s.logger.Error("Failed to send response", slog.String("error", err.Error()))
				}
			}
		}
	}
}

// leaseCleanupWorker periodically cleans up expired leases
func (s *server) leaseCleanupWorker() {
	s.logger.Info("Starting lease cleanup worker")

	ticker := time.NewTicker(5 * time.Minute) // Default cleanup interval
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Stopping lease cleanup worker")
			return
		case <-ticker.C:
			if err := s.leaseManager.CleanupExpiredLeases(s.ctx); err != nil {
				s.logger.Error("Failed to cleanup expired leases", slog.String("error", err.Error()))
			}
		}
	}
}

// statisticsUpdater periodically updates server statistics
func (s *server) statisticsUpdater() {
	s.logger.Info("Starting statistics updater")

	ticker := time.NewTicker(1 * time.Minute) // Update every minute
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Stopping statistics updater")
			return
		case <-ticker.C:
			if err := s.updateStatistics(); err != nil {
				s.logger.Error("Failed to update statistics", slog.String("error", err.Error()))
			}
		}
	}
}

// Helper methods

func (s *server) getPoolInfo(ctx context.Context) (*types.DHCPPoolInfo, error) {
	availableIPs, err := s.leaseManager.GetAvailableIPs(ctx)
	if err != nil {
		return nil, err
	}

	leases, err := s.leaseManager.GetAllLeases(ctx)
	if err != nil {
		return nil, err
	}

	allocatedCount := 0
	for _, lease := range leases {
		if lease.State == types.LeaseStateActive {
			allocatedCount++
		}
	}

	totalIPs := len(availableIPs) + allocatedCount
	utilizationRate := 0.0
	if totalIPs > 0 {
		utilizationRate = float64(allocatedCount) / float64(totalIPs) * 100
	}

	reservations, _ := s.leaseManager.GetReservations(ctx)
	reservedCount := len(reservations)

	return &types.DHCPPoolInfo{
		StartIP:         s.config.Pool.StartIP,
		EndIP:           s.config.Pool.EndIP,
		TotalIPs:        totalIPs,
		AllocatedIPs:    allocatedCount,
		AvailableIPs:    len(availableIPs),
		ReservedIPs:     reservedCount,
		UtilizationRate: utilizationRate,
	}, nil
}

func (s *server) updateRequestStatistics(request *types.DHCPRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.statistics.TotalRequests++

	// Update by message type
	messageType := fmt.Sprintf("type_%d", request.MessageType)
	s.statistics.RequestsByType[messageType]++

	// Update by hour
	hour := time.Now().Format("2006-01-02_15")
	s.statistics.RequestsByHour[hour]++
}

func (s *server) updateStatistics() error {
	// Save current statistics to storage
	return s.storage.SaveStatistics(s.ctx, s.statistics)
}

// Simplified packet parsing/building (would need proper DHCP protocol implementation)
func (s *server) parsePacket(data []byte) (*types.DHCPRequest, error) {
	// This is a placeholder - actual implementation would parse DHCP packets
	return &types.DHCPRequest{
		MessageType:   1, // DHCP Discover
		TransactionID: 12345,
		ClientMAC:     "00:11:22:33:44:55",
		Options:       make(map[int]string),
		Timestamp:     time.Now().Format(time.RFC3339),
	}, nil
}

func (s *server) buildPacket(response *types.DHCPResponse) ([]byte, error) {
	// This is a placeholder - actual implementation would build DHCP packets
	return []byte("DHCP Response"), nil
}

func (s *server) getStartTimeString() string {
	if s.startTime.IsZero() {
		return ""
	}
	return s.startTime.Format(time.RFC3339)
}
