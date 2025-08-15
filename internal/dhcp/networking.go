package dhcp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"pihole-analyzer/internal/types"
)

// networking implements the DHCPNetworking interface
type networking struct {
	config *types.DHCPConfig
	logger *slog.Logger
	conn   *net.UDPConn
}

// Bind binds to the specified network interface and port
func (n *networking) Bind(address string, port int) error {
	n.logger.Info("Binding to network",
		slog.String("address", address),
		slog.Int("port", port),
		slog.String("interface", n.config.Interface))
	
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}
	
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("failed to bind UDP socket: %w", err)
	}
	
	n.conn = conn
	n.logger.Info("Successfully bound to network",
		slog.String("local_addr", conn.LocalAddr().String()))
	
	return nil
}

// Close closes the network connection
func (n *networking) Close() error {
	if n.conn != nil {
		n.logger.Info("Closing network connection")
		err := n.conn.Close()
		n.conn = nil
		return err
	}
	return nil
}

// ReceivePacket receives a packet from the network
func (n *networking) ReceivePacket(ctx context.Context) ([]byte, error) {
	if n.conn == nil {
		return nil, fmt.Errorf("network not bound")
	}
	
	buffer := make([]byte, 1500) // Standard MTU size
	
	// Set read deadline based on context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Set a reasonable timeout for network operations
		// For now, we'll use a simple 1-second timeout
		deadline := time.Now().Add(1 * time.Second)
		n.conn.SetReadDeadline(deadline)
	}
	
	bytesRead, addr, err := n.conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read UDP packet: %w", err)
	}
	
	n.logger.Debug("Received packet",
		slog.Int("size", bytesRead),
		slog.String("from", addr.String()))
	
	return buffer[:bytesRead], nil
}

// SendPacket sends a packet to the specified destination
func (n *networking) SendPacket(ctx context.Context, data []byte, destIP string, destPort int) error {
	if n.conn == nil {
		return fmt.Errorf("network not bound")
	}
	
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", destIP, destPort))
	if err != nil {
		return fmt.Errorf("failed to resolve destination address: %w", err)
	}
	
	sentBytes, err := n.conn.WriteToUDP(data, addr)
	if err != nil {
		return fmt.Errorf("failed to send UDP packet: %w", err)
	}
	
	n.logger.Debug("Sent packet",
		slog.Int("size", sentBytes),
		slog.String("to", addr.String()))
	
	return nil
}

// GetInterface returns the network interface name
func (n *networking) GetInterface() string {
	return n.config.Interface
}

// GetListenAddress returns the listen address
func (n *networking) GetListenAddress() string {
	return n.config.ListenAddress
}

// GetPort returns the listen port
func (n *networking) GetPort() int {
	return n.config.Port
}

// security implements the DHCPSecurity interface
type security struct {
	config *types.DHCPSecurityConfig
	logger *slog.Logger
}

// IsClientAllowed checks if a client is allowed to get DHCP leases
func (s *security) IsClientAllowed(ctx context.Context, mac string, clientID string) (bool, error) {
	s.logger.Debug("Checking client permissions",
		slog.String("mac", mac),
		slog.String("client_id", clientID))
	
	// Check blocked clients list
	for _, blockedMAC := range s.config.BlockedClients {
		if mac == blockedMAC {
			s.logger.Warn("Client is blocked", slog.String("mac", mac))
			return false, nil
		}
	}
	
	// If allowed clients list is configured, check if client is in it
	if len(s.config.AllowedClients) > 0 {
		for _, allowedMAC := range s.config.AllowedClients {
			if mac == allowedMAC {
				return true, nil
			}
		}
		// Client not in allowed list
		s.logger.Warn("Client not in allowed list", slog.String("mac", mac))
		return false, nil
	}
	
	// If no specific allow list, client is allowed (unless blocked)
	return true, nil
}

// IsClientBlocked checks if a client is explicitly blocked
func (s *security) IsClientBlocked(ctx context.Context, mac string) (bool, error) {
	for _, blockedMAC := range s.config.BlockedClients {
		if mac == blockedMAC {
			return true, nil
		}
	}
	return false, nil
}

// CheckRateLimit checks if a client has exceeded rate limits
func (s *security) CheckRateLimit(ctx context.Context, clientIP string) (bool, error) {
	if !s.config.EnableRateLimit {
		return true, nil // Rate limiting disabled
	}
	
	// Simplified rate limiting - would use proper rate limiting in practice
	s.logger.Debug("Checking rate limit", slog.String("client_ip", clientIP))
	return true, nil // Allow for now
}

// RecordRequest records a request for rate limiting purposes
func (s *security) RecordRequest(ctx context.Context, clientIP string) error {
	if !s.config.EnableRateLimit {
		return nil
	}
	
	s.logger.Debug("Recording request", slog.String("client_ip", clientIP))
	// Would implement request tracking here
	return nil
}

// GenerateFingerprint generates a device fingerprint
func (s *security) GenerateFingerprint(ctx context.Context, request *types.DHCPRequest) (string, error) {
	if !s.config.EnableFingerprinting {
		return "", nil
	}
	
	// Simple fingerprinting based on vendor class and options
	fingerprint := fmt.Sprintf("MAC:%s", request.ClientMAC)
	
	if request.VendorClass != "" {
		fingerprint += fmt.Sprintf(",VC:%s", request.VendorClass)
	}
	
	s.logger.Debug("Generated fingerprint",
		slog.String("mac", request.ClientMAC),
		slog.String("fingerprint", fingerprint))
	
	return fingerprint, nil
}

// ClassifyDevice classifies a device based on its fingerprint
func (s *security) ClassifyDevice(ctx context.Context, fingerprint string) (string, error) {
	if fingerprint == "" {
		return "unknown", nil
	}
	
	// Simple device classification - would be more sophisticated in practice
	if contains(fingerprint, "iPhone") || contains(fingerprint, "iOS") {
		return "mobile_ios", nil
	} else if contains(fingerprint, "Android") {
		return "mobile_android", nil
	} else if contains(fingerprint, "Windows") {
		return "desktop_windows", nil
	} else if contains(fingerprint, "Linux") {
		return "desktop_linux", nil
	} else if contains(fingerprint, "Mac") {
		return "desktop_mac", nil
	}
	
	return "unknown", nil
}

// LogSecurityEvent logs a security-related event
func (s *security) LogSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	if !s.config.LogAllRequests && event.Severity == "low" {
		return nil // Don't log low severity events if detailed logging is disabled
	}
	
	s.logger.Info("Security event",
		slog.String("type", event.Type),
		slog.String("severity", event.Severity),
		slog.String("client_mac", event.ClientMAC),
		slog.String("client_ip", event.ClientIP),
		slog.String("description", event.Description),
		slog.Time("timestamp", event.Timestamp))
	
	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}