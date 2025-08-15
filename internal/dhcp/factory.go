package dhcp

import (
	"fmt"
	"log/slog"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// factory implements the DHCPFactory interface
type factory struct {
	logger *slog.Logger
}

// NewFactory creates a new DHCP component factory
func NewFactory(l *slog.Logger) DHCPFactory {
	if l == nil {
		loggerInstance := logger.New(&logger.Config{Component: "dhcp-factory"})
		l = loggerInstance.GetSlogger()
	}
	
	return &factory{
		logger: l,
	}
}

// CreateServer creates a new DHCP server instance
func (f *factory) CreateServer(config *types.DHCPConfig) (DHCPServer, error) {
	f.logger.Debug("Creating DHCP server", slog.Any("config", config))
	
	if err := f.validateServerConfig(config); err != nil {
		return nil, fmt.Errorf("invalid server configuration: %w", err)
	}
	
	// Create storage
	storage, err := f.CreateStorage(&config.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	
	// Create lease manager
	leaseManager, err := f.CreateLeaseManager(config, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease manager: %w", err)
	}
	
	// Create packet handler
	packetHandler, err := f.CreatePacketHandler(config, leaseManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create packet handler: %w", err)
	}
	
	// Create networking
	networking, err := f.CreateNetworking(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create networking: %w", err)
	}
	
	// Create security
	security, err := f.CreateSecurity(&config.Security)
	if err != nil {
		return nil, fmt.Errorf("failed to create security: %w", err)
	}
	
	server := &server{
		config:        config,
		storage:       storage,
		leaseManager:  leaseManager,
		packetHandler: packetHandler,
		networking:    networking,
		security:      security,
		logger:        f.logger.With(slog.String("component", "dhcp-server")),
		statistics: &types.DHCPStatistics{
			RequestsByType: make(map[string]int64),
			RequestsByHour: make(map[string]int64),
			TopClients:     make([]types.DHCPClientStat, 0),
			RecentActivity: make([]types.DHCPActivity, 0),
		},
	}
	
	return server, nil
}

// CreateLeaseManager creates a new DHCP lease manager
func (f *factory) CreateLeaseManager(config *types.DHCPConfig, storage DHCPStorage) (DHCPLeaseManager, error) {
	f.logger.Debug("Creating DHCP lease manager")
	
	return &leaseManager{
		config:  config,
		storage: storage,
		logger:  f.logger.With(slog.String("component", "dhcp-lease-manager")),
	}, nil
}

// CreatePacketHandler creates a new DHCP packet handler
func (f *factory) CreatePacketHandler(config *types.DHCPConfig, leaseManager DHCPLeaseManager) (DHCPPacketHandler, error) {
	f.logger.Debug("Creating DHCP packet handler")
	
	return &packetHandler{
		config:       config,
		leaseManager: leaseManager,
		logger:       f.logger.With(slog.String("component", "dhcp-packet-handler")),
	}, nil
}

// CreateStorage creates a new DHCP storage implementation
func (f *factory) CreateStorage(config *types.DHCPStorageConfig) (DHCPStorage, error) {
	f.logger.Debug("Creating DHCP storage", slog.String("type", config.Type))
	
	switch config.Type {
	case "memory":
		return &memoryStorage{
			config: config,
			logger: f.logger.With(slog.String("component", "dhcp-memory-storage")),
		}, nil
	case "file":
		return &fileStorage{
			config: config,
			logger: f.logger.With(slog.String("component", "dhcp-file-storage")),
		}, nil
	case "database":
		return &databaseStorage{
			config: config,
			logger: f.logger.With(slog.String("component", "dhcp-database-storage")),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// CreateNetworking creates a new DHCP networking implementation
func (f *factory) CreateNetworking(config *types.DHCPConfig) (DHCPNetworking, error) {
	f.logger.Debug("Creating DHCP networking", 
		slog.String("interface", config.Interface),
		slog.String("listen_address", config.ListenAddress),
		slog.Int("port", config.Port))
	
	return &networking{
		config: config,
		logger: f.logger.With(slog.String("component", "dhcp-networking")),
	}, nil
}

// CreateSecurity creates a new DHCP security implementation
func (f *factory) CreateSecurity(config *types.DHCPSecurityConfig) (DHCPSecurity, error) {
	f.logger.Debug("Creating DHCP security", slog.Any("config", config))
	
	return &security{
		config: config,
		logger: f.logger.With(slog.String("component", "dhcp-security")),
	}, nil
}

// validateServerConfig validates the DHCP server configuration
func (f *factory) validateServerConfig(config *types.DHCPConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if config.Interface == "" {
		return fmt.Errorf("interface must be specified")
	}
	
	if config.ListenAddress == "" {
		return fmt.Errorf("listen address must be specified")
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	
	if config.Pool.StartIP == "" {
		return fmt.Errorf("pool start IP must be specified")
	}
	
	if config.Pool.EndIP == "" {
		return fmt.Errorf("pool end IP must be specified")
	}
	
	if config.Pool.Subnet == "" {
		return fmt.Errorf("pool subnet must be specified")
	}
	
	if config.LeaseTime == "" {
		return fmt.Errorf("lease time must be specified")
	}
	
	return nil
}