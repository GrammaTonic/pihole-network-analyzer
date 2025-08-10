package dhcp

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

func TestDHCPFactory(t *testing.T) {
	factory := NewFactory(nil)
	
	t.Run("CreateServer", func(t *testing.T) {
		config := DefaultDHCPConfig()
		config.Interface = "test0"
		config.ListenAddress = "127.0.0.1"
		config.Enabled = true // Enable the server for testing
		
		server, err := factory.CreateServer(config)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		
		if server == nil {
			t.Fatal("Server should not be nil")
		}
		
		if !server.GetConfig().Enabled {
			t.Error("Server should be enabled")
		}
	})
	
	t.Run("CreateLeaseManager", func(t *testing.T) {
		config := DefaultDHCPConfig()
		storage, err := factory.CreateStorage(&config.Storage)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		
		leaseManager, err := factory.CreateLeaseManager(config, storage)
		if err != nil {
			t.Fatalf("Failed to create lease manager: %v", err)
		}
		
		if leaseManager == nil {
			t.Fatal("Lease manager should not be nil")
		}
	})
	
	t.Run("CreateStorage", func(t *testing.T) {
		config := &types.DHCPStorageConfig{
			Type: "memory",
			Path: "",
		}
		
		storage, err := factory.CreateStorage(config)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		
		if storage == nil {
			t.Fatal("Storage should not be nil")
		}
	})
	
	t.Run("InvalidConfig", func(t *testing.T) {
		config := &types.DHCPConfig{
			// Missing required fields
		}
		
		_, err := factory.CreateServer(config)
		if err == nil {
			t.Error("Should fail with invalid configuration")
		}
	})
}

func TestMemoryStorage(t *testing.T) {
	loggerInstance := logger.New(&logger.Config{Component: "test-storage"})
	storage := &memoryStorage{
		config: &types.DHCPStorageConfig{Type: "memory"},
		logger: loggerInstance.GetSlogger(),
	}
	ctx := context.Background()
	
	// Initialize storage
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()
	
	t.Run("SaveAndLoadLease", func(t *testing.T) {
		lease := &types.DHCPLease{
			ID:        "test-lease-1",
			IP:        "192.168.1.100",
			MAC:       "00:11:22:33:44:55",
			StartTime: time.Now().Format(time.RFC3339),
			EndTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			State:     types.LeaseStateActive,
			Type:      types.LeaseTypeDynamic,
		}
		
		// Save lease
		if err := storage.SaveLease(ctx, lease); err != nil {
			t.Fatalf("Failed to save lease: %v", err)
		}
		
		// Load lease by ID
		loadedLease, err := storage.LoadLease(ctx, lease.ID)
		if err != nil {
			t.Fatalf("Failed to load lease: %v", err)
		}
		
		if loadedLease.IP != lease.IP {
			t.Errorf("Expected IP %s, got %s", lease.IP, loadedLease.IP)
		}
		
		// Load lease by IP
		leaseByIP, err := storage.LoadLeaseByIP(ctx, lease.IP)
		if err != nil {
			t.Fatalf("Failed to load lease by IP: %v", err)
		}
		
		if leaseByIP.MAC != lease.MAC {
			t.Errorf("Expected MAC %s, got %s", lease.MAC, leaseByIP.MAC)
		}
		
		// Load lease by MAC
		leaseByMAC, err := storage.LoadLeaseByMAC(ctx, lease.MAC)
		if err != nil {
			t.Fatalf("Failed to load lease by MAC: %v", err)
		}
		
		if leaseByMAC.IP != lease.IP {
			t.Errorf("Expected IP %s, got %s", lease.IP, leaseByMAC.IP)
		}
	})
	
	t.Run("SaveAndLoadReservation", func(t *testing.T) {
		reservation := &types.DHCPReservation{
			MAC:         "00:11:22:33:44:66",
			IP:          "192.168.1.50",
			Hostname:    "test-device",
			Description: "Test reservation",
			Enabled:     true,
		}
		
		// Save reservation
		if err := storage.SaveReservation(ctx, reservation); err != nil {
			t.Fatalf("Failed to save reservation: %v", err)
		}
		
		// Load reservation
		loadedReservation, err := storage.LoadReservation(ctx, reservation.MAC)
		if err != nil {
			t.Fatalf("Failed to load reservation: %v", err)
		}
		
		if loadedReservation.IP != reservation.IP {
			t.Errorf("Expected IP %s, got %s", reservation.IP, loadedReservation.IP)
		}
		
		if loadedReservation.Hostname != reservation.Hostname {
			t.Errorf("Expected hostname %s, got %s", reservation.Hostname, loadedReservation.Hostname)
		}
	})
	
	t.Run("LoadAllLeases", func(t *testing.T) {
		// Clear storage first
		storage.leases = make(map[string]*types.DHCPLease)
		
		// Add multiple leases
		lease1 := &types.DHCPLease{
			ID:  "lease-1",
			IP:  "192.168.1.101",
			MAC: "00:11:22:33:44:01",
		}
		lease2 := &types.DHCPLease{
			ID:  "lease-2",
			IP:  "192.168.1.102",
			MAC: "00:11:22:33:44:02",
		}
		
		storage.SaveLease(ctx, lease1)
		storage.SaveLease(ctx, lease2)
		
		leases, err := storage.LoadAllLeases(ctx)
		if err != nil {
			t.Fatalf("Failed to load all leases: %v", err)
		}
		
		if len(leases) != 2 {
			t.Errorf("Expected 2 leases, got %d", len(leases))
		}
	})
	
	t.Run("DeleteLease", func(t *testing.T) {
		lease := &types.DHCPLease{
			ID:  "delete-test",
			IP:  "192.168.1.103",
			MAC: "00:11:22:33:44:03",
		}
		
		// Save lease
		storage.SaveLease(ctx, lease)
		
		// Verify it exists
		_, err := storage.LoadLease(ctx, lease.ID)
		if err != nil {
			t.Fatalf("Lease should exist before deletion: %v", err)
		}
		
		// Delete lease
		if err := storage.DeleteLease(ctx, lease.ID); err != nil {
			t.Fatalf("Failed to delete lease: %v", err)
		}
		
		// Verify it's gone
		_, err = storage.LoadLease(ctx, lease.ID)
		if err == nil {
			t.Error("Lease should not exist after deletion")
		}
	})
}

func TestLeaseManager(t *testing.T) {
	config := DefaultDHCPConfig()
	config.Pool.StartIP = "192.168.1.100"
	config.Pool.EndIP = "192.168.1.110"
	
	loggerInstance := logger.New(&logger.Config{Component: "test-lease-manager"})
	storage := &memoryStorage{
		config: &config.Storage,
		logger: loggerInstance.GetSlogger(),
	}
	ctx := context.Background()
	storage.Initialize(ctx)
	defer storage.Close()
	
	lm := &leaseManager{
		config:  config,
		storage: storage,
		logger:  loggerInstance.GetSlogger(),
	}
	
	t.Run("AllocateIP", func(t *testing.T) {
		clientMAC := "00:11:22:33:44:77"
		
		// Allocate IP
		ip, err := lm.AllocateIP(ctx, clientMAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate IP: %v", err)
		}
		
		if ip == "" {
			t.Fatal("Allocated IP should not be empty")
		}
		
		// Verify lease was created
		lease, err := storage.LoadLeaseByMAC(ctx, clientMAC)
		if err != nil {
			t.Fatalf("Failed to load lease: %v", err)
		}
		
		if lease.IP != ip {
			t.Errorf("Expected IP %s, got %s", ip, lease.IP)
		}
		
		if lease.State != types.LeaseStateActive {
			t.Errorf("Expected state %s, got %s", types.LeaseStateActive, lease.State)
		}
	})
	
	t.Run("AllocateSpecificIP", func(t *testing.T) {
		clientMAC := "00:11:22:33:44:88"
		requestedIP := "192.168.1.105"
		
		// First check if the IP is available by getting available IPs
		availableIPs, err := lm.GetAvailableIPs(ctx)
		if err != nil {
			t.Fatalf("Failed to get available IPs: %v", err)
		}
		
		// Find an available IP from the list for testing
		if len(availableIPs) > 0 {
			requestedIP = availableIPs[0] // Use the first available IP
		}
		
		// Allocate specific IP
		ip, err := lm.AllocateIP(ctx, clientMAC, requestedIP, "")
		if err != nil {
			t.Fatalf("Failed to allocate specific IP: %v", err)
		}
		
		if ip != requestedIP {
			t.Errorf("Expected IP %s, got %s", requestedIP, ip)
		}
	})
	
	t.Run("ReleaseIP", func(t *testing.T) {
		clientMAC := "00:11:22:33:44:99"
		
		// Allocate IP first
		ip, err := lm.AllocateIP(ctx, clientMAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate IP: %v", err)
		}
		
		// Release IP
		if err := lm.ReleaseIP(ctx, ip, clientMAC); err != nil {
			t.Fatalf("Failed to release IP: %v", err)
		}
		
		// Verify lease state changed
		lease, err := storage.LoadLeaseByIP(ctx, ip)
		if err != nil {
			t.Fatalf("Failed to load lease: %v", err)
		}
		
		if lease.State != types.LeaseStateReleased {
			t.Errorf("Expected state %s, got %s", types.LeaseStateReleased, lease.State)
		}
	})
	
	t.Run("RenewLease", func(t *testing.T) {
		clientMAC := "00:11:22:33:44:AA"
		
		// Allocate IP first
		ip, err := lm.AllocateIP(ctx, clientMAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate IP: %v", err)
		}
		
		// Get original lease
		originalLease, err := storage.LoadLeaseByMAC(ctx, clientMAC) // Use MAC instead of IP to get the right lease
		if err != nil {
			t.Fatalf("Failed to load original lease: %v", err)
		}
		
		// Wait a bit to ensure time difference
		time.Sleep(10 * time.Millisecond)
		
		// Renew lease
		duration := 48 * time.Hour
		if err := lm.RenewLease(ctx, ip, clientMAC, duration); err != nil {
			t.Fatalf("Failed to renew lease: %v", err)
		}
		
		// Verify lease was renewed
		renewedLease, err := storage.LoadLeaseByMAC(ctx, clientMAC) // Use MAC instead of IP
		if err != nil {
			t.Fatalf("Failed to load renewed lease: %v", err)
		}
		
		if renewedLease.LastRenewal == originalLease.LastRenewal {
			t.Error("Last renewal time should have been updated")
		}
		
		if renewedLease.EndTime == originalLease.EndTime {
			t.Error("End time should have been updated")
		}
	})
	
	t.Run("GetAvailableIPs", func(t *testing.T) {
		// Get available IPs
		availableIPs, err := lm.GetAvailableIPs(ctx)
		if err != nil {
			t.Fatalf("Failed to get available IPs: %v", err)
		}
		
		if len(availableIPs) == 0 {
			t.Error("Should have available IPs")
		}
		
		// Verify IPs are in range
		for _, ip := range availableIPs {
			if !lm.isIPInPool(ip) {
				t.Errorf("IP %s should be in pool", ip)
			}
		}
	})
	
	t.Run("AddReservation", func(t *testing.T) {
		reservation := &types.DHCPReservation{
			MAC:         "00:11:22:33:44:BB",
			IP:          "192.168.1.107",
			Hostname:    "reserved-device",
			Description: "Test reservation",
			Enabled:     true,
		}
		
		// Add reservation
		if err := lm.AddReservation(ctx, reservation); err != nil {
			t.Fatalf("Failed to add reservation: %v", err)
		}
		
		// Verify reservation exists
		reservations, err := lm.GetReservations(ctx)
		if err != nil {
			t.Fatalf("Failed to get reservations: %v", err)
		}
		
		found := false
		for _, r := range reservations {
			if r.MAC == reservation.MAC && r.IP == reservation.IP {
				found = true
				break
			}
		}
		
		if !found {
			t.Error("Reservation should exist")
		}
		
		// Test allocation with reservation
		ip, err := lm.AllocateIP(ctx, reservation.MAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate reserved IP: %v", err)
		}
		
		if ip != reservation.IP {
			t.Errorf("Should allocate reserved IP %s, got %s", reservation.IP, ip)
		}
	})
}

func TestPacketHandler(t *testing.T) {
	config := DefaultDHCPConfig()
	
	loggerInstance := logger.New(&logger.Config{Component: "test-packet-handler"})
	storage := &memoryStorage{
		config: &config.Storage,
		logger: loggerInstance.GetSlogger(),
	}
	ctx := context.Background()
	storage.Initialize(ctx)
	defer storage.Close()
	
	lm := &leaseManager{
		config:  config,
		storage: storage,
		logger:  loggerInstance.GetSlogger(),
	}
	
	ph := &packetHandler{
		config:       config,
		leaseManager: lm,
		logger:       loggerInstance.GetSlogger(),
	}
	
	t.Run("ProcessDiscover", func(t *testing.T) {
		request := &types.DHCPRequest{
			MessageType:   1, // DHCP Discover
			TransactionID: 12345,
			ClientMAC:     "00:11:22:33:44:CC",
			Options:       make(map[int]string),
			Timestamp:     time.Now().Format(time.RFC3339),
		}
		
		response, err := ph.ProcessDiscover(ctx, request)
		if err != nil {
			t.Fatalf("Failed to process discover: %v", err)
		}
		
		if response == nil {
			t.Fatal("Response should not be nil")
		}
		
		if response.MessageType != 2 {
			t.Errorf("Expected message type 2 (OFFER), got %d", response.MessageType)
		}
		
		if response.YourIP == "" {
			t.Error("Offered IP should not be empty")
		}
		
		if response.TransactionID != request.TransactionID {
			t.Errorf("Transaction ID mismatch: expected %d, got %d", 
				request.TransactionID, response.TransactionID)
		}
	})
	
	t.Run("ProcessRequest", func(t *testing.T) {
		// First allocate an IP
		clientMAC := "00:11:22:33:44:DD"
		allocatedIP, err := lm.AllocateIP(ctx, clientMAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate IP: %v", err)
		}
		
		request := &types.DHCPRequest{
			MessageType:   3, // DHCP Request
			TransactionID: 12346,
			ClientMAC:     clientMAC,
			RequestedIP:   allocatedIP,
			Options:       make(map[int]string),
			Timestamp:     time.Now().Format(time.RFC3339),
		}
		
		response, err := ph.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Failed to process request: %v", err)
		}
		
		if response == nil {
			t.Fatal("Response should not be nil")
		}
		
		if response.MessageType != 5 {
			t.Errorf("Expected message type 5 (ACK), got %d", response.MessageType)
		}
		
		if response.YourIP != allocatedIP {
			t.Errorf("Expected assigned IP %s, got %s", allocatedIP, response.YourIP)
		}
	})
	
	t.Run("ProcessRelease", func(t *testing.T) {
		// First allocate an IP
		clientMAC := "00:11:22:33:44:EE"
		allocatedIP, err := lm.AllocateIP(ctx, clientMAC, "", "")
		if err != nil {
			t.Fatalf("Failed to allocate IP: %v", err)
		}
		
		request := &types.DHCPRequest{
			MessageType: 7, // DHCP Release
			ClientMAC:   clientMAC,
			ClientIP:    allocatedIP,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		
		err = ph.ProcessRelease(ctx, request)
		if err != nil {
			t.Fatalf("Failed to process release: %v", err)
		}
		
		// Verify lease was released
		lease, err := storage.LoadLeaseByIP(ctx, allocatedIP)
		if err != nil {
			t.Fatalf("Failed to load lease: %v", err)
		}
		
		if lease.State != types.LeaseStateReleased {
			t.Errorf("Expected state %s, got %s", types.LeaseStateReleased, lease.State)
		}
	})
	
	t.Run("ValidateRequest", func(t *testing.T) {
		// Valid request
		validRequest := &types.DHCPRequest{
			MessageType:   1,
			TransactionID: 12345,
			ClientMAC:     "00:11:22:33:44:55",
		}
		
		if err := ph.ValidateRequest(validRequest); err != nil {
			t.Errorf("Valid request should pass validation: %v", err)
		}
		
		// Invalid request - missing MAC
		invalidRequest := &types.DHCPRequest{
			MessageType:   1,
			TransactionID: 12345,
			ClientMAC:     "",
		}
		
		if err := ph.ValidateRequest(invalidRequest); err == nil {
			t.Error("Invalid request should fail validation")
		}
		
		// Invalid request - missing transaction ID
		invalidRequest2 := &types.DHCPRequest{
			MessageType:   1,
			TransactionID: 0,
			ClientMAC:     "00:11:22:33:44:55",
		}
		
		if err := ph.ValidateRequest(invalidRequest2); err == nil {
			t.Error("Invalid request should fail validation")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultDHCPConfig()
	
	if config == nil {
		t.Fatal("Default config should not be nil")
	}
	
	if config.Interface == "" {
		t.Error("Default interface should not be empty")
	}
	
	if config.Port <= 0 {
		t.Error("Default port should be positive")
	}
	
	if config.Pool.StartIP == "" {
		t.Error("Default start IP should not be empty")
	}
	
	if config.Pool.EndIP == "" {
		t.Error("Default end IP should not be empty")
	}
	
	if config.LeaseTime == "" {
		t.Error("Default lease time should not be empty")
	}
	
	// Test config validation
	if err := ValidateDHCPConfig(config); err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}
}