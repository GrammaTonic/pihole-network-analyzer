package dhcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"pihole-analyzer/internal/types"
)

// memoryStorage implements DHCPStorage using in-memory storage
type memoryStorage struct {
	config       *types.DHCPStorageConfig
	logger       *slog.Logger
	leases       map[string]*types.DHCPLease
	reservations map[string]*types.DHCPReservation
	statistics   *types.DHCPStatistics
	mu           sync.RWMutex
}

// Initialize initializes the memory storage
func (ms *memoryStorage) Initialize(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	ms.leases = make(map[string]*types.DHCPLease)
	ms.reservations = make(map[string]*types.DHCPReservation)
	ms.statistics = &types.DHCPStatistics{
		RequestsByType: make(map[string]int64),
		RequestsByHour: make(map[string]int64),
		TopClients:     make([]types.DHCPClientStat, 0),
		RecentActivity: make([]types.DHCPActivity, 0),
	}
	
	ms.logger.Info("Memory storage initialized")
	return nil
}

// Close closes the memory storage
func (ms *memoryStorage) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	ms.leases = nil
	ms.reservations = nil
	ms.statistics = nil
	
	ms.logger.Info("Memory storage closed")
	return nil
}

// SaveLease saves a lease to memory
func (ms *memoryStorage) SaveLease(ctx context.Context, lease *types.DHCPLease) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if lease == nil {
		return fmt.Errorf("lease cannot be nil")
	}
	
	ms.leases[lease.ID] = lease
	ms.logger.Debug("Lease saved to memory",
		slog.String("lease_id", lease.ID),
		slog.String("ip", lease.IP),
		slog.String("mac", lease.MAC))
	
	return nil
}

// LoadLease loads a lease from memory by ID
func (ms *memoryStorage) LoadLease(ctx context.Context, id string) (*types.DHCPLease, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	lease, exists := ms.leases[id]
	if !exists {
		return nil, fmt.Errorf("lease not found: %s", id)
	}
	
	// Return a copy to prevent external modification
	leaseCopy := *lease
	return &leaseCopy, nil
}

// LoadLeaseByIP loads a lease from memory by IP address
func (ms *memoryStorage) LoadLeaseByIP(ctx context.Context, ip string) (*types.DHCPLease, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	for _, lease := range ms.leases {
		if lease.IP == ip {
			leaseCopy := *lease
			return &leaseCopy, nil
		}
	}
	
	return nil, fmt.Errorf("lease not found for IP: %s", ip)
}

// LoadLeaseByMAC loads a lease from memory by MAC address
func (ms *memoryStorage) LoadLeaseByMAC(ctx context.Context, mac string) (*types.DHCPLease, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	for _, lease := range ms.leases {
		if lease.MAC == mac && lease.State == types.LeaseStateActive {
			leaseCopy := *lease
			return &leaseCopy, nil
		}
	}
	
	return nil, fmt.Errorf("active lease not found for MAC: %s", mac)
}

// LoadAllLeases loads all leases from memory
func (ms *memoryStorage) LoadAllLeases(ctx context.Context) ([]types.DHCPLease, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	leases := make([]types.DHCPLease, 0, len(ms.leases))
	for _, lease := range ms.leases {
		leases = append(leases, *lease)
	}
	
	return leases, nil
}

// DeleteLease deletes a lease from memory
func (ms *memoryStorage) DeleteLease(ctx context.Context, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.leases[id]; !exists {
		return fmt.Errorf("lease not found: %s", id)
	}
	
	delete(ms.leases, id)
	ms.logger.Debug("Lease deleted from memory", slog.String("lease_id", id))
	
	return nil
}

// SaveReservation saves a reservation to memory
func (ms *memoryStorage) SaveReservation(ctx context.Context, reservation *types.DHCPReservation) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if reservation == nil {
		return fmt.Errorf("reservation cannot be nil")
	}
	
	ms.reservations[reservation.MAC] = reservation
	ms.logger.Debug("Reservation saved to memory",
		slog.String("mac", reservation.MAC),
		slog.String("ip", reservation.IP))
	
	return nil
}

// LoadReservation loads a reservation from memory by MAC address
func (ms *memoryStorage) LoadReservation(ctx context.Context, mac string) (*types.DHCPReservation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	reservation, exists := ms.reservations[mac]
	if !exists {
		return nil, fmt.Errorf("reservation not found: %s", mac)
	}
	
	reservationCopy := *reservation
	return &reservationCopy, nil
}

// LoadAllReservations loads all reservations from memory
func (ms *memoryStorage) LoadAllReservations(ctx context.Context) ([]types.DHCPReservation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	reservations := make([]types.DHCPReservation, 0, len(ms.reservations))
	for _, reservation := range ms.reservations {
		reservations = append(reservations, *reservation)
	}
	
	return reservations, nil
}

// DeleteReservation deletes a reservation from memory
func (ms *memoryStorage) DeleteReservation(ctx context.Context, mac string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.reservations[mac]; !exists {
		return fmt.Errorf("reservation not found: %s", mac)
	}
	
	delete(ms.reservations, mac)
	ms.logger.Debug("Reservation deleted from memory", slog.String("mac", mac))
	
	return nil
}

// SaveStatistics saves statistics to memory
func (ms *memoryStorage) SaveStatistics(ctx context.Context, stats *types.DHCPStatistics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if stats == nil {
		return fmt.Errorf("statistics cannot be nil")
	}
	
	ms.statistics = stats
	ms.logger.Debug("Statistics saved to memory")
	
	return nil
}

// LoadStatistics loads statistics from memory
func (ms *memoryStorage) LoadStatistics(ctx context.Context) (*types.DHCPStatistics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.statistics == nil {
		// Return empty statistics if none exist
		return &types.DHCPStatistics{
			RequestsByType: make(map[string]int64),
			RequestsByHour: make(map[string]int64),
			TopClients:     make([]types.DHCPClientStat, 0),
			RecentActivity: make([]types.DHCPActivity, 0),
		}, nil
	}
	
	statsCopy := *ms.statistics
	return &statsCopy, nil
}

// Backup creates a backup of the memory storage (not applicable for memory)
func (ms *memoryStorage) Backup(ctx context.Context, path string) error {
	return fmt.Errorf("backup not supported for memory storage")
}

// Restore restores from a backup (not applicable for memory)
func (ms *memoryStorage) Restore(ctx context.Context, path string) error {
	return fmt.Errorf("restore not supported for memory storage")
}