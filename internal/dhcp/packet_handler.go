package dhcp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"pihole-analyzer/internal/types"
)

// packetHandler implements the DHCPPacketHandler interface
type packetHandler struct {
	config       *types.DHCPConfig
	leaseManager DHCPLeaseManager
	logger       *slog.Logger
}

// ProcessDiscover processes a DHCP DISCOVER message
func (ph *packetHandler) ProcessDiscover(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error) {
	ph.logger.Debug("Processing DHCP DISCOVER",
		slog.String("client_mac", request.ClientMAC),
		slog.String("requested_ip", request.RequestedIP))
	
	// Try to allocate an IP address
	allocatedIP, err := ph.leaseManager.AllocateIP(ctx, request.ClientMAC, request.RequestedIP, request.ClientID)
	if err != nil {
		ph.logger.Error("Failed to allocate IP", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}
	
	// Build DHCP OFFER response
	response := &types.DHCPResponse{
		MessageType:   2, // DHCP OFFER
		TransactionID: request.TransactionID,
		ClientMAC:     request.ClientMAC,
		YourIP:        allocatedIP,
		ServerIP:      ph.config.ListenAddress,
		Options:       make(map[int]string),
		LeaseTime:     uint32(ph.getLeaseTimeSeconds()),
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	
	// Add standard DHCP options
	ph.addStandardOptions(response)
	
	ph.logger.Info("DHCP OFFER sent",
		slog.String("client_mac", request.ClientMAC),
		slog.String("offered_ip", allocatedIP))
	
	return response, nil
}

// ProcessRequest processes a DHCP REQUEST message
func (ph *packetHandler) ProcessRequest(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error) {
	ph.logger.Debug("Processing DHCP REQUEST",
		slog.String("client_mac", request.ClientMAC),
		slog.String("requested_ip", request.RequestedIP))
	
	// Validate the request
	if err := ph.ValidateRequest(request); err != nil {
		ph.logger.Warn("Invalid DHCP REQUEST", slog.String("error", err.Error()))
		return ph.buildNAK(request, "Invalid request")
	}
	
	// Check if we have an active lease for this client
	existingLease, err := ph.leaseManager.GetActiveLease(ctx, request.ClientMAC)
	if err != nil {
		ph.logger.Error("Failed to get active lease", slog.String("error", err.Error()))
		return ph.buildNAK(request, "Internal error")
	}
	
	var assignedIP string
	
	if existingLease != nil && existingLease.IP == request.RequestedIP {
		// Renew existing lease
		leaseDuration := ph.getLeaseDuration()
		if err := ph.leaseManager.RenewLease(ctx, existingLease.IP, request.ClientMAC, leaseDuration); err != nil {
			ph.logger.Error("Failed to renew lease", slog.String("error", err.Error()))
			return ph.buildNAK(request, "Failed to renew lease")
		}
		assignedIP = existingLease.IP
	} else {
		// Try to allocate the requested IP
		allocatedIP, err := ph.leaseManager.AllocateIP(ctx, request.ClientMAC, request.RequestedIP, request.ClientID)
		if err != nil {
			ph.logger.Error("Failed to allocate IP", slog.String("error", err.Error()))
			return ph.buildNAK(request, "Cannot allocate requested IP")
		}
		assignedIP = allocatedIP
	}
	
	// Build DHCP ACK response
	response := &types.DHCPResponse{
		MessageType:   5, // DHCP ACK
		TransactionID: request.TransactionID,
		ClientMAC:     request.ClientMAC,
		YourIP:        assignedIP,
		ServerIP:      ph.config.ListenAddress,
		Options:       make(map[int]string),
		LeaseTime:     uint32(ph.getLeaseTimeSeconds()),
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	
	// Add standard DHCP options
	ph.addStandardOptions(response)
	
	ph.logger.Info("DHCP ACK sent",
		slog.String("client_mac", request.ClientMAC),
		slog.String("assigned_ip", assignedIP))
	
	return response, nil
}

// ProcessRelease processes a DHCP RELEASE message
func (ph *packetHandler) ProcessRelease(ctx context.Context, request *types.DHCPRequest) error {
	ph.logger.Debug("Processing DHCP RELEASE",
		slog.String("client_mac", request.ClientMAC),
		slog.String("client_ip", request.ClientIP))
	
	if err := ph.leaseManager.ReleaseIP(ctx, request.ClientIP, request.ClientMAC); err != nil {
		ph.logger.Error("Failed to release IP", slog.String("error", err.Error()))
		return fmt.Errorf("failed to release IP: %w", err)
	}
	
	ph.logger.Info("IP released",
		slog.String("client_mac", request.ClientMAC),
		slog.String("released_ip", request.ClientIP))
	
	return nil
}

// ProcessDecline processes a DHCP DECLINE message
func (ph *packetHandler) ProcessDecline(ctx context.Context, request *types.DHCPRequest) error {
	ph.logger.Debug("Processing DHCP DECLINE",
		slog.String("client_mac", request.ClientMAC),
		slog.String("declined_ip", request.RequestedIP))
	
	// Mark IP as declined (could implement conflict detection here)
	ph.logger.Warn("IP declined by client",
		slog.String("client_mac", request.ClientMAC),
		slog.String("declined_ip", request.RequestedIP))
	
	// In a full implementation, we would mark this IP as problematic
	// and potentially remove it from the available pool temporarily
	
	return nil
}

// ProcessInform processes a DHCP INFORM message
func (ph *packetHandler) ProcessInform(ctx context.Context, request *types.DHCPRequest) (*types.DHCPResponse, error) {
	ph.logger.Debug("Processing DHCP INFORM",
		slog.String("client_mac", request.ClientMAC),
		slog.String("client_ip", request.ClientIP))
	
	// Build DHCP ACK response with configuration options only
	response := &types.DHCPResponse{
		MessageType:   5, // DHCP ACK
		TransactionID: request.TransactionID,
		ClientMAC:     request.ClientMAC,
		YourIP:        "", // No IP assignment for INFORM
		ServerIP:      ph.config.ListenAddress,
		Options:       make(map[int]string),
		LeaseTime:     0, // No lease for INFORM
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	
	// Add configuration options only
	ph.addStandardOptions(response)
	
	ph.logger.Info("DHCP INFORM ACK sent", slog.String("client_mac", request.ClientMAC))
	
	return response, nil
}

// ValidateRequest validates a DHCP request
func (ph *packetHandler) ValidateRequest(request *types.DHCPRequest) error {
	if request.ClientMAC == "" {
		return fmt.Errorf("client MAC address is required")
	}
	
	if request.TransactionID == 0 {
		return fmt.Errorf("transaction ID is required")
	}
	
	return nil
}

// BuildOptions builds DHCP options for a lease
func (ph *packetHandler) BuildOptions(ctx context.Context, lease *types.DHCPLease, requestedOptions []int) (map[int]string, error) {
	options := make(map[int]string)
	
	// Add standard options
	if ph.config.Pool.Gateway != "" {
		options[3] = ph.config.Pool.Gateway // Router
	}
	
	if len(ph.config.Pool.DNSServers) > 0 {
		// DNS Servers (simplified - would handle multiple servers properly)
		options[6] = ph.config.Pool.DNSServers[0]
	}
	
	if ph.config.Options.DomainName != "" {
		options[15] = ph.config.Options.DomainName // Domain Name
	}
	
	// Add lease time
	options[51] = fmt.Sprintf("%d", ph.getLeaseTimeSeconds())
	
	// Add server identifier
	options[54] = ph.config.ListenAddress
	
	// Add custom options from configuration
	for optionCode, optionValue := range ph.config.Options.CustomOptions {
		options[optionCode] = optionValue
	}
	
	return options, nil
}

// ParseClientOptions parses client options from a DHCP request
func (ph *packetHandler) ParseClientOptions(options map[int]string) (*DHCPClientInfo, error) {
	clientInfo := &DHCPClientInfo{
		Fingerprint: "", // Would be generated based on options
	}
	
	// Parse hostname (option 12)
	if hostname, ok := options[12]; ok {
		clientInfo.Hostname = hostname
	}
	
	// Parse vendor class (option 60)
	if vendorClass, ok := options[60]; ok {
		clientInfo.VendorClass = vendorClass
	}
	
	// Parse client identifier (option 61)
	if clientID, ok := options[61]; ok {
		clientInfo.ClientID = clientID
	}
	
	// Generate device fingerprint based on options
	clientInfo.Fingerprint = ph.generateFingerprint(options)
	
	return clientInfo, nil
}

// Helper methods

func (ph *packetHandler) buildNAK(request *types.DHCPRequest, reason string) (*types.DHCPResponse, error) {
	ph.logger.Warn("Sending DHCP NAK",
		slog.String("client_mac", request.ClientMAC),
		slog.String("reason", reason))
	
	return &types.DHCPResponse{
		MessageType:   6, // DHCP NAK
		TransactionID: request.TransactionID,
		ClientMAC:     request.ClientMAC,
		YourIP:        "",
		ServerIP:      ph.config.ListenAddress,
		Options:       map[int]string{56: reason}, // Message option
		LeaseTime:     0,
		Timestamp:     time.Now().Format(time.RFC3339),
	}, nil
}

func (ph *packetHandler) addStandardOptions(response *types.DHCPResponse) {
	// Subnet mask (option 1)
	if ph.config.Pool.Subnet != "" {
		// Extract subnet mask from CIDR notation
		response.Options[1] = "255.255.255.0" // Simplified
	}
	
	// Router (option 3)
	if ph.config.Pool.Gateway != "" {
		response.Options[3] = ph.config.Pool.Gateway
	}
	
	// DNS servers (option 6)
	if len(ph.config.Pool.DNSServers) > 0 {
		response.Options[6] = ph.config.Pool.DNSServers[0] // Simplified
	}
	
	// Domain name (option 15)
	if ph.config.Options.DomainName != "" {
		response.Options[15] = ph.config.Options.DomainName
	}
	
	// Lease time (option 51)
	response.Options[51] = fmt.Sprintf("%d", response.LeaseTime)
	
	// Server identifier (option 54)
	response.Options[54] = ph.config.ListenAddress
	
	// Add custom options
	for optionCode, optionValue := range ph.config.Options.CustomOptions {
		response.Options[optionCode] = optionValue
	}
}

func (ph *packetHandler) getLeaseTimeSeconds() int {
	duration, err := time.ParseDuration(ph.config.LeaseTime)
	if err != nil {
		return 86400 // Default to 24 hours
	}
	return int(duration.Seconds())
}

func (ph *packetHandler) getLeaseDuration() time.Duration {
	duration, err := time.ParseDuration(ph.config.LeaseTime)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return duration
}

func (ph *packetHandler) generateFingerprint(options map[int]string) string {
	// Simplified fingerprinting - would be more sophisticated in practice
	fingerprint := ""
	
	// Use parameter request list (option 55) and vendor class (option 60)
	if paramList, ok := options[55]; ok {
		fingerprint += "PRL:" + paramList + ";"
	}
	
	if vendorClass, ok := options[60]; ok {
		fingerprint += "VC:" + vendorClass + ";"
	}
	
	return fingerprint
}