package dhcp

import (
	"fmt"
	"log/slog"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// NewServer creates a new DHCP server instance with the given configuration
func NewServer(config *types.DHCPConfig, l *slog.Logger) (DHCPServer, error) {
	if l == nil {
		loggerInstance := logger.New(&logger.Config{Component: "dhcp"})
		l = loggerInstance.GetSlogger()
	}
	
	factory := NewFactory(l)
	return factory.CreateServer(config)
}

// NewLeaseManager creates a new DHCP lease manager instance
func NewLeaseManager(config *types.DHCPConfig, storage DHCPStorage, l *slog.Logger) (DHCPLeaseManager, error) {
	if l == nil {
		loggerInstance := logger.New(&logger.Config{Component: "dhcp-lease-manager"})
		l = loggerInstance.GetSlogger()
	}
	
	factory := NewFactory(l)
	return factory.CreateLeaseManager(config, storage)
}

// NewStorage creates a new DHCP storage instance based on configuration
func NewStorage(config *types.DHCPStorageConfig, l *slog.Logger) (DHCPStorage, error) {
	if l == nil {
		loggerInstance := logger.New(&logger.Config{Component: "dhcp-storage"})
		l = loggerInstance.GetSlogger()
	}
	
	factory := NewFactory(l)
	return factory.CreateStorage(config)
}

// DefaultDHCPConfig returns a default DHCP configuration
func DefaultDHCPConfig() *types.DHCPConfig {
	return &types.DHCPConfig{
		Enabled:       false,
		Interface:     "eth0",
		ListenAddress: "0.0.0.0",
		Port:          67,
		LeaseTime:     "24h",
		MaxLeaseTime:  "72h",
		RenewalTime:   "12h",
		RebindTime:    "21h",
		Pool: types.DHCPPoolConfig{
			StartIP:    "192.168.1.100",
			EndIP:      "192.168.1.200",
			Subnet:     "192.168.1.0/24",
			Gateway:    "192.168.1.1",
			DNSServers: []string{"192.168.1.1", "8.8.8.8"},
			Exclude:    []string{},
		},
		Options: types.DHCPOptionsConfig{
			Router:           "192.168.1.1",
			DomainName:       "local",
			DomainNameServer: []string{"192.168.1.1", "8.8.8.8"},
			MTU:              1500,
			CustomOptions:    make(map[int]string),
		},
		Reservations: []types.DHCPReservation{},
		Storage: types.DHCPStorageConfig{
			Type:         "memory",
			Path:         "",
			BackupPath:   "",
			SyncInterval: "5m",
			MaxLeases:    10000,
		},
		Performance: types.DHCPPerfConfig{
			MaxConnections:       1000,
			ReadTimeout:          "30s",
			WriteTimeout:         "30s",
			BufferSize:           65536,
			WorkerPoolSize:       10,
			LeaseCleanupInterval: "1h",
		},
		Security: types.DHCPSecurityConfig{
			EnableRateLimit:      false,
			MaxRequestsPerIP:     100,
			AllowedClients:       []string{},
			BlockedClients:       []string{},
			RequireClientID:      false,
			LogAllRequests:       false,
			EnableFingerprinting: true,
		},
	}
}

// ValidateDHCPConfig validates a DHCP configuration
func ValidateDHCPConfig(config *types.DHCPConfig) error {
	f := NewFactory(nil)
	factoryImpl, ok := f.(*factory)
	if !ok {
		return fmt.Errorf("invalid factory type")
	}
	return factoryImpl.validateServerConfig(config)
}