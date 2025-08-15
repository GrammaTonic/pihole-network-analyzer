package dhcp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"pihole-analyzer/internal/types"
)

// leaseManager implements the DHCPLeaseManager interface
type leaseManager struct {
	config  *types.DHCPConfig
	storage DHCPStorage
	logger  *slog.Logger
	mu      sync.RWMutex
}

// AllocateIP allocates an IP address for a client
func (lm *leaseManager) AllocateIP(ctx context.Context, clientMAC string, requestedIP string, clientID string) (string, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Debug("Allocating IP address",
		slog.String("client_mac", clientMAC),
		slog.String("requested_ip", requestedIP),
		slog.String("client_id", clientID))
	
	// Check for existing lease for this client
	if existingLease, err := lm.storage.LoadLeaseByMAC(ctx, clientMAC); err == nil && existingLease != nil {
		if existingLease.State == types.LeaseStateActive {
			// Check if lease is still valid
			endTime, err := time.Parse(time.RFC3339, existingLease.EndTime)
			if err == nil && time.Now().Before(endTime) {
				lm.logger.Debug("Returning existing active lease",
					slog.String("ip", existingLease.IP),
					slog.String("client_mac", clientMAC))
				return existingLease.IP, nil
			}
		}
	}
	
	// Check for static reservation
	if reservation, err := lm.storage.LoadReservation(ctx, clientMAC); err == nil && reservation != nil && reservation.Enabled {
		// Check if reserved IP is available
		if existingLease, err := lm.storage.LoadLeaseByIP(ctx, reservation.IP); err != nil || existingLease == nil ||
			existingLease.State != types.LeaseStateActive || existingLease.MAC == clientMAC {
			
			lm.logger.Info("Allocating reserved IP",
				slog.String("ip", reservation.IP),
				slog.String("client_mac", clientMAC))
			
			lease := lm.createLease(reservation.IP, clientMAC, clientID, types.LeaseTypeStatic)
			if err := lm.storage.SaveLease(ctx, lease); err != nil {
				return "", fmt.Errorf("failed to save static lease: %w", err)
			}
			
			return reservation.IP, nil
		}
	}
	
	// Try to allocate requested IP if specified and available
	if requestedIP != "" && lm.isIPInPool(requestedIP) {
		if available, err := lm.isIPAvailable(ctx, requestedIP); err == nil && available {
			lm.logger.Debug("Allocating requested IP",
				slog.String("ip", requestedIP),
				slog.String("client_mac", clientMAC))
			
			lease := lm.createLease(requestedIP, clientMAC, clientID, types.LeaseTypeDynamic)
			if err := lm.storage.SaveLease(ctx, lease); err != nil {
				return "", fmt.Errorf("failed to save requested lease: %w", err)
			}
			
			return requestedIP, nil
		}
	}
	
	// Find next available IP in pool - use internal method to avoid deadlock
	availableIPs, err := lm.getAvailableIPsInternal(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get available IPs: %w", err)
	}
	
	if len(availableIPs) == 0 {
		return "", fmt.Errorf("no available IP addresses in pool")
	}
	
	allocatedIP := availableIPs[0]
	lm.logger.Info("Allocating dynamic IP",
		slog.String("ip", allocatedIP),
		slog.String("client_mac", clientMAC))
	
	lease := lm.createLease(allocatedIP, clientMAC, clientID, types.LeaseTypeDynamic)
	if err := lm.storage.SaveLease(ctx, lease); err != nil {
		return "", fmt.Errorf("failed to save dynamic lease: %w", err)
	}
	
	return allocatedIP, nil
}

// ReleaseIP releases an IP address from a client
func (lm *leaseManager) ReleaseIP(ctx context.Context, ip string, clientMAC string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Info("Releasing IP address",
		slog.String("ip", ip),
		slog.String("client_mac", clientMAC))
	
	lease, err := lm.storage.LoadLeaseByIP(ctx, ip)
	if err != nil {
		return fmt.Errorf("failed to load lease for IP %s: %w", ip, err)
	}
	
	if lease == nil {
		return fmt.Errorf("no lease found for IP %s", ip)
	}
	
	if lease.MAC != clientMAC {
		return fmt.Errorf("lease for IP %s belongs to different client (%s != %s)", ip, lease.MAC, clientMAC)
	}
	
	// Update lease state
	lease.State = types.LeaseStateReleased
	lease.LastRenewal = time.Now().Format(time.RFC3339)
	
	if err := lm.storage.SaveLease(ctx, lease); err != nil {
		return fmt.Errorf("failed to save released lease: %w", err)
	}
	
	return nil
}

// RenewLease renews an existing lease
func (lm *leaseManager) RenewLease(ctx context.Context, ip string, clientMAC string, duration time.Duration) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Debug("Renewing lease",
		slog.String("ip", ip),
		slog.String("client_mac", clientMAC),
		slog.Duration("duration", duration))
	
	lease, err := lm.storage.LoadLeaseByIP(ctx, ip)
	if err != nil {
		return fmt.Errorf("failed to load lease for IP %s: %w", ip, err)
	}
	
	if lease == nil {
		return fmt.Errorf("no lease found for IP %s", ip)
	}
	
	if lease.MAC != clientMAC {
		return fmt.Errorf("lease for IP %s belongs to different client", ip)
	}
	
	// Update lease times
	now := time.Now()
	lease.LastRenewal = now.Format(time.RFC3339)
	lease.EndTime = now.Add(duration).Format(time.RFC3339)
	lease.State = types.LeaseStateActive
	
	if err := lm.storage.SaveLease(ctx, lease); err != nil {
		return fmt.Errorf("failed to save renewed lease: %w", err)
	}
	
	lm.logger.Info("Lease renewed successfully",
		slog.String("ip", ip),
		slog.String("client_mac", clientMAC),
		slog.String("new_end_time", lease.EndTime))
	
	return nil
}

// GetActiveLease returns the active lease for a client
func (lm *leaseManager) GetActiveLease(ctx context.Context, clientMAC string) (*types.DHCPLease, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	lease, err := lm.storage.LoadLeaseByMAC(ctx, clientMAC)
	if err != nil {
		return nil, fmt.Errorf("failed to load lease for MAC %s: %w", clientMAC, err)
	}
	
	if lease != nil && lease.State == types.LeaseStateActive {
		// Check if lease is still valid
		endTime, err := time.Parse(time.RFC3339, lease.EndTime)
		if err == nil && time.Now().Before(endTime) {
			return lease, nil
		}
	}
	
	return nil, nil
}

// GetLeaseByIP returns the lease for a specific IP address
func (lm *leaseManager) GetLeaseByIP(ctx context.Context, ip string) (*types.DHCPLease, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	return lm.storage.LoadLeaseByIP(ctx, ip)
}

// GetAllLeases returns all leases
func (lm *leaseManager) GetAllLeases(ctx context.Context) ([]types.DHCPLease, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	return lm.storage.LoadAllLeases(ctx)
}

// ExpireLeases marks expired leases as expired
func (lm *leaseManager) ExpireLeases(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Debug("Checking for expired leases")
	
	leases, err := lm.storage.LoadAllLeases(ctx)
	if err != nil {
		return fmt.Errorf("failed to load leases: %w", err)
	}
	
	now := time.Now()
	expiredCount := 0
	
	for _, lease := range leases {
		if lease.State == types.LeaseStateActive {
			endTime, err := time.Parse(time.RFC3339, lease.EndTime)
			if err != nil {
				lm.logger.Error("Failed to parse lease end time",
					slog.String("lease_id", lease.ID),
					slog.String("end_time", lease.EndTime),
					slog.String("error", err.Error()))
				continue
			}
			
			if now.After(endTime) {
				lease.State = types.LeaseStateExpired
				if err := lm.storage.SaveLease(ctx, &lease); err != nil {
					lm.logger.Error("Failed to save expired lease",
						slog.String("lease_id", lease.ID),
						slog.String("error", err.Error()))
				} else {
					expiredCount++
				}
			}
		}
	}
	
	if expiredCount > 0 {
		lm.logger.Info("Expired leases", slog.Int("count", expiredCount))
	}
	
	return nil
}

// CleanupExpiredLeases removes old expired leases
func (lm *leaseManager) CleanupExpiredLeases(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Debug("Cleaning up expired leases")
	
	// First expire any leases that should be expired
	if err := lm.ExpireLeases(ctx); err != nil {
		return fmt.Errorf("failed to expire leases: %w", err)
	}
	
	leases, err := lm.storage.LoadAllLeases(ctx)
	if err != nil {
		return fmt.Errorf("failed to load leases: %w", err)
	}
	
	// Remove leases that have been expired for more than 24 hours
	cleanupThreshold := time.Now().Add(-24 * time.Hour)
	cleanupCount := 0
	
	for _, lease := range leases {
		if lease.State == types.LeaseStateExpired || lease.State == types.LeaseStateReleased {
			endTime, err := time.Parse(time.RFC3339, lease.EndTime)
			if err != nil {
				continue
			}
			
			if endTime.Before(cleanupThreshold) {
				if err := lm.storage.DeleteLease(ctx, lease.ID); err != nil {
					lm.logger.Error("Failed to delete expired lease",
						slog.String("lease_id", lease.ID),
						slog.String("error", err.Error()))
				} else {
					cleanupCount++
				}
			}
		}
	}
	
	if cleanupCount > 0 {
		lm.logger.Info("Cleaned up expired leases", slog.Int("count", cleanupCount))
	}
	
	return nil
}

// GetAvailableIPs returns a list of available IP addresses in the pool
func (lm *leaseManager) GetAvailableIPs(ctx context.Context) ([]string, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	return lm.getAvailableIPsInternal(ctx)
}

// getAvailableIPsInternal returns available IPs without acquiring locks (for internal use)
func (lm *leaseManager) getAvailableIPsInternal(ctx context.Context) ([]string, error) {
	// Get all IPs in the pool
	poolIPs, err := lm.generatePoolIPs()
	if err != nil {
		return nil, fmt.Errorf("failed to generate pool IPs: %w", err)
	}
	
	// Get all active leases
	leases, err := lm.storage.LoadAllLeases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load leases: %w", err)
	}
	
	// Create map of allocated IPs
	allocatedIPs := make(map[string]bool)
	for _, lease := range leases {
		if lease.State == types.LeaseStateActive {
			allocatedIPs[lease.IP] = true
		}
	}
	
	// Get all reservations
	reservations, err := lm.storage.LoadAllReservations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load reservations: %w", err)
	}
	
	// Mark reserved IPs as allocated
	for _, reservation := range reservations {
		if reservation.Enabled {
			allocatedIPs[reservation.IP] = true
		}
	}
	
	// Filter out allocated IPs and excluded IPs
	var availableIPs []string
	for _, ip := range poolIPs {
		if !allocatedIPs[ip] && !lm.isIPExcluded(ip) {
			availableIPs = append(availableIPs, ip)
		}
	}
	
	return availableIPs, nil
}

// AddReservation adds a static IP reservation
func (lm *leaseManager) AddReservation(ctx context.Context, reservation *types.DHCPReservation) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Info("Adding IP reservation",
		slog.String("mac", reservation.MAC),
		slog.String("ip", reservation.IP),
		slog.String("hostname", reservation.Hostname))
	
	// Validate reservation
	if !lm.isIPInPool(reservation.IP) {
		return fmt.Errorf("IP %s is not in the DHCP pool", reservation.IP)
	}
	
	// Check if IP is already reserved or leased to another client
	if existingLease, err := lm.storage.LoadLeaseByIP(ctx, reservation.IP); err == nil && existingLease != nil {
		if existingLease.MAC != reservation.MAC && existingLease.State == types.LeaseStateActive {
			return fmt.Errorf("IP %s is already leased to another client", reservation.IP)
		}
	}
	
	return lm.storage.SaveReservation(ctx, reservation)
}

// RemoveReservation removes a static IP reservation
func (lm *leaseManager) RemoveReservation(ctx context.Context, mac string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.logger.Info("Removing IP reservation", slog.String("mac", mac))
	return lm.storage.DeleteReservation(ctx, mac)
}

// GetReservations returns all static IP reservations
func (lm *leaseManager) GetReservations(ctx context.Context) ([]types.DHCPReservation, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	return lm.storage.LoadAllReservations(ctx)
}

// IsReserved checks if an IP address is reserved
func (lm *leaseManager) IsReserved(ctx context.Context, ip string) (bool, *types.DHCPReservation, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	reservations, err := lm.storage.LoadAllReservations(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load reservations: %w", err)
	}
	
	for _, reservation := range reservations {
		if reservation.IP == ip && reservation.Enabled {
			return true, &reservation, nil
		}
	}
	
	return false, nil, nil
}

// Helper methods

func (lm *leaseManager) createLease(ip, mac, clientID string, leaseType types.DHCPLeaseType) *types.DHCPLease {
	now := time.Now()
	
	// Parse lease duration
	leaseDuration, _ := time.ParseDuration(lm.config.LeaseTime)
	if leaseDuration == 0 {
		leaseDuration = 24 * time.Hour // Default to 24 hours
	}
	
	return &types.DHCPLease{
		ID:               fmt.Sprintf("%s_%d", mac, now.Unix()),
		IP:               ip,
		MAC:              mac,
		ClientID:         clientID,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(leaseDuration).Format(time.RFC3339),
		LastRenewal:      now.Format(time.RFC3339),
		State:            types.LeaseStateActive,
		Type:             leaseType,
		Options:          make(map[int]string),
		RequestedOptions: make([]int, 0),
		Metadata:         make(map[string]string),
	}
}

func (lm *leaseManager) isIPInPool(ip string) bool {
	startIP := net.ParseIP(lm.config.Pool.StartIP)
	endIP := net.ParseIP(lm.config.Pool.EndIP)
	checkIP := net.ParseIP(ip)
	
	if startIP == nil || endIP == nil || checkIP == nil {
		return false
	}
	
	return lm.ipInRange(checkIP, startIP, endIP)
}

func (lm *leaseManager) isIPExcluded(ip string) bool {
	for _, excludedIP := range lm.config.Pool.Exclude {
		if ip == excludedIP {
			return true
		}
	}
	return false
}

func (lm *leaseManager) isIPAvailable(ctx context.Context, ip string) (bool, error) {
	lease, err := lm.storage.LoadLeaseByIP(ctx, ip)
	if err != nil {
		return false, err
	}
	
	if lease == nil {
		return true, nil
	}
	
	// Check if lease is active and not expired
	if lease.State == types.LeaseStateActive {
		endTime, err := time.Parse(time.RFC3339, lease.EndTime)
		if err == nil && time.Now().Before(endTime) {
			return false, nil
		}
	}
	
	return true, nil
}

func (lm *leaseManager) generatePoolIPs() ([]string, error) {
	startIP := net.ParseIP(lm.config.Pool.StartIP)
	endIP := net.ParseIP(lm.config.Pool.EndIP)
	
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("invalid IP range")
	}
	
	var ips []string
	for ip := startIP; !ip.Equal(endIP); lm.incrementIP(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, endIP.String()) // Include end IP
	
	return ips, nil
}

func (lm *leaseManager) ipInRange(checkIP, startIP, endIP net.IP) bool {
	if checkIP.To4() == nil || startIP.To4() == nil || endIP.To4() == nil {
		return false
	}
	
	checkInt := lm.ipToInt(checkIP)
	startInt := lm.ipToInt(startIP)
	endInt := lm.ipToInt(endIP)
	
	return checkInt >= startInt && checkInt <= endInt
}

func (lm *leaseManager) ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

func (lm *leaseManager) incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}