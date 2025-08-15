package dhcp

import (
	"context"
	"time"

	"pihole-analyzer/internal/types"
)

// DHCPServer defines the interface for DHCP server functionality
type DHCPServer interface {
	// Server lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Configuration management
	GetConfig() *types.DHCPConfig
	UpdateConfig(config *types.DHCPConfig) error
	ValidateConfig(config *types.DHCPConfig) error

	// Lease management
	GetLeases(ctx context.Context) ([]types.DHCPLease, error)
	GetLease(ctx context.Context, identifier string) (*types.DHCPLease, error)
	CreateReservation(ctx context.Context, reservation *types.DHCPReservation) error
	DeleteReservation(ctx context.Context, mac string) error

	// Server status and statistics
	GetStatus(ctx context.Context) (*types.DHCPServerStatus, error)
	GetStatistics(ctx context.Context) (*types.DHCPStatistics, error)

	// Request handling (internal methods for packet processing)
	HandleDHCPRequest(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error)
}

// DHCPLeaseManager defines the interface for lease management
type DHCPLeaseManager interface {
	// Lease allocation
	AllocateIP(ctx context.Context, clientMAC string, requestedIP string, clientID string) (string, error)
	ReleaseIP(ctx context.Context, ip string, clientMAC string) error
	RenewLease(ctx context.Context, ip string, clientMAC string, duration time.Duration) error

	// Lease queries
	GetActiveLease(ctx context.Context, clientMAC string) (*types.DHCPLease, error)
	GetLeaseByIP(ctx context.Context, ip string) (*types.DHCPLease, error)
	GetAllLeases(ctx context.Context) ([]types.DHCPLease, error)

	// Lease maintenance
	ExpireLeases(ctx context.Context) error
	CleanupExpiredLeases(ctx context.Context) error
	GetAvailableIPs(ctx context.Context) ([]string, error)

	// Reservations
	AddReservation(ctx context.Context, reservation *types.DHCPReservation) error
	RemoveReservation(ctx context.Context, mac string) error
	GetReservations(ctx context.Context) ([]types.DHCPReservation, error)
	IsReserved(ctx context.Context, ip string) (bool, *types.DHCPReservation, error)
}

// DHCPPacketHandler defines the interface for DHCP packet processing
type DHCPPacketHandler interface {
	// Packet processing
	ProcessDiscover(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error)
	ProcessRequest(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error)
	ProcessRelease(ctx context.Context, request *types.DHCPRequest) error
	ProcessDecline(ctx context.Context, request *types.DHCPRequest) error
	ProcessInform(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error)

	// Packet validation
	ValidateRequest(request *types.DHCPRequest) error

	// Option handling
	BuildOptions(ctx context.Context, lease *types.DHCPLease, requestedOptions []int) (map[int]string, error)
	ParseClientOptions(options map[int]string) (*DHCPClientInfo, error)
}

// DHCPStorage defines the interface for lease persistence
type DHCPStorage interface {
	// Lease storage
	SaveLease(ctx context.Context, lease *types.DHCPLease) error
	LoadLease(ctx context.Context, id string) (*types.DHCPLease, error)
	LoadLeaseByIP(ctx context.Context, ip string) (*types.DHCPLease, error)
	LoadLeaseByMAC(ctx context.Context, mac string) (*types.DHCPLease, error)
	LoadAllLeases(ctx context.Context) ([]types.DHCPLease, error)
	DeleteLease(ctx context.Context, id string) error

	// Reservation storage
	SaveReservation(ctx context.Context, reservation *types.DHCPReservation) error
	LoadReservation(ctx context.Context, mac string) (*types.DHCPReservation, error)
	LoadAllReservations(ctx context.Context) ([]types.DHCPReservation, error)
	DeleteReservation(ctx context.Context, mac string) error

	// Statistics storage
	SaveStatistics(ctx context.Context, stats *types.DHCPStatistics) error
	LoadStatistics(ctx context.Context) (*types.DHCPStatistics, error)

	// Storage management
	Initialize(ctx context.Context) error
	Close() error
	Backup(ctx context.Context, path string) error
	Restore(ctx context.Context, path string) error
}

// DHCPNetworking defines the interface for network communication
type DHCPNetworking interface {
	// Network setup
	Bind(address string, port int) error
	Close() error

	// Packet I/O
	ReceivePacket(ctx context.Context) ([]byte, error)
	SendPacket(ctx context.Context, data []byte, destIP string, destPort int) error

	// Network information
	GetInterface() string
	GetListenAddress() string
	GetPort() int
}

// DHCPSecurity defines the interface for security features
type DHCPSecurity interface {
	// Client validation
	IsClientAllowed(ctx context.Context, mac string, clientID string) (bool, error)
	IsClientBlocked(ctx context.Context, mac string) (bool, error)

	// Rate limiting
	CheckRateLimit(ctx context.Context, clientIP string) (bool, error)
	RecordRequest(ctx context.Context, clientIP string) error

	// Device fingerprinting
	GenerateFingerprint(ctx context.Context, request *types.DHCPRequest) (string, error)
	ClassifyDevice(ctx context.Context, fingerprint string) (string, error)

	// Audit logging
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
}

// Supporting types for DHCP operations

// DHCPClientInfo represents parsed client information from DHCP options
type DHCPClientInfo struct {
	MAC         string
	ClientID    string
	Hostname    string
	VendorClass string
	UserClass   string
	Fingerprint string
	DeviceType  string
}

// SecurityEvent represents a security-related event in the DHCP server
type SecurityEvent struct {
	Timestamp   time.Time
	Type        string // "rate_limit", "blocked_client", "suspicious_request", etc.
	ClientMAC   string
	ClientIP    string
	Severity    string // "low", "medium", "high", "critical"
	Description string
	Context     map[string]interface{}
}

// DHCPFactory defines the interface for creating DHCP components
type DHCPFactory interface {
	// Component creation
	CreateServer(config *types.DHCPConfig) (DHCPServer, error)
	CreateLeaseManager(config *types.DHCPConfig, storage DHCPStorage) (DHCPLeaseManager, error)
	CreatePacketHandler(config *types.DHCPConfig, leaseManager DHCPLeaseManager) (DHCPPacketHandler, error)
	CreateStorage(config *types.DHCPStorageConfig) (DHCPStorage, error)
	CreateNetworking(config *types.DHCPConfig) (DHCPNetworking, error)
	CreateSecurity(config *types.DHCPSecurityConfig) (DHCPSecurity, error)
}
