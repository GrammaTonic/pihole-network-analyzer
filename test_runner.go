package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

// TestMode enables mock data and offline testing
var TestMode = false
var MockDataInstance *MockData

// InitTestMode sets up the test environment
func InitTestMode() error {
	TestMode = true
	var err error
	MockDataInstance, err = SetupTestEnvironment()
	if err != nil {
		return fmt.Errorf("failed to setup test environment: %v", err)
	}
	return nil
}

// RunTests executes all test scenarios
func RunTests() {
	fmt.Println("DNS Analyzer Test Suite")
	fmt.Println("=======================")

	// Initialize test environment
	err := InitTestMode()
	if err != nil {
		log.Fatalf("Test initialization failed: %v", err)
	}
	defer CleanupTestEnvironment()

	// Test scenarios
	testScenarios := []struct {
		name     string
		testFunc func() error
	}{
		{"CSV Analysis - Default", testCSVAnalysisDefault},
		{"CSV Analysis - No Exclusions", testCSVAnalysisNoExclusions},
		{"CSV Analysis - Online Only", testCSVAnalysisOnlineOnly},
		{"Pi-hole Analysis - Default", testPiholeAnalysisDefault},
		{"Pi-hole Analysis - No Exclusions", testPiholeAnalysisNoExclusions},
		{"Pi-hole Analysis - Online Only", testPiholeAnalysisOnlineOnly},
		{"ARP Table Functionality", testARPFunctionality},
		{"Hostname Resolution", testHostnameResolution},
		{"Exclusion Logic", testExclusionLogic},
	}

	successCount := 0
	for i, scenario := range testScenarios {
		fmt.Printf("\n%d. Running test: %s\n", i+1, scenario.name)
		fmt.Println("   " + strings.Repeat("-", len(scenario.name)))
		
		err := scenario.testFunc()
		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
		} else {
			fmt.Printf("   ✅ PASSED\n")
			successCount++
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 50))
	fmt.Printf("\nTest Results: %d/%d tests passed\n", successCount, len(testScenarios))
	
	if successCount == len(testScenarios) {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Printf("⚠️  %d tests failed\n", len(testScenarios)-successCount)
	}
}

// Test CSV analysis with default settings
func testCSVAnalysisDefault() error {
	csvFile := filepath.Join("test_data", "mock_dns_data.csv")
	
	// Reset flags to default
	*onlineOnlyFlag = false
	*noExcludeFlag = false
	
	clientStats, err := analyzeDNSData(csvFile)
	if err != nil {
		return fmt.Errorf("CSV analysis failed: %v", err)
	}
	
	// Should exclude Docker IPs, so we expect fewer clients
	expectedMinClients := 3 // At least the main network clients
	expectedMaxClients := 8 // But not the Docker ones
	
	if len(clientStats) < expectedMinClients || len(clientStats) > expectedMaxClients {
		return fmt.Errorf("expected %d-%d clients, got %d", expectedMinClients, expectedMaxClients, len(clientStats))
	}
	
	// Check that Docker IPs are excluded
	if _, exists := clientStats["172.20.0.8"]; exists {
		return fmt.Errorf("Docker IP 172.20.0.8 should be excluded but was found")
	}
	
	fmt.Printf("   Found %d clients (Docker IPs correctly excluded)\n", len(clientStats))
	return nil
}

// Test CSV analysis without exclusions
func testCSVAnalysisNoExclusions() error {
	csvFile := filepath.Join("test_data", "mock_dns_data.csv")
	
	// Enable no-exclude flag
	*onlineOnlyFlag = false
	*noExcludeFlag = true
	
	clientStats, err := analyzeDNSData(csvFile)
	if err != nil {
		return fmt.Errorf("CSV analysis failed: %v", err)
	}
	
	// Should include all clients including Docker
	expectedMinClients := 8 // Should include Docker containers
	
	if len(clientStats) < expectedMinClients {
		return fmt.Errorf("expected at least %d clients, got %d", expectedMinClients, len(clientStats))
	}
	
	// Check that Docker IPs are included
	if _, exists := clientStats["172.20.0.8"]; !exists {
		return fmt.Errorf("Docker IP 172.20.0.8 should be included but was not found")
	}
	
	fmt.Printf("   Found %d clients (all IPs included)\n", len(clientStats))
	return nil
}

// Test CSV analysis with online-only filter
func testCSVAnalysisOnlineOnly() error {
	csvFile := filepath.Join("test_data", "mock_dns_data.csv")
	
	// Enable online-only flag
	*onlineOnlyFlag = true
	*noExcludeFlag = false
	
	clientStats, err := analyzeDNSData(csvFile)
	if err != nil {
		return fmt.Errorf("CSV analysis failed: %v", err)
	}
	
	// Mock ARP checking
	err = mockCheckARPStatus(clientStats)
	if err != nil {
		return fmt.Errorf("ARP status check failed: %v", err)
	}
	
	// Count online clients
	onlineCount := 0
	for _, stats := range clientStats {
		if stats.IsOnline {
			onlineCount++
		}
	}
	
	expectedOnlineClients := 4 // Based on mock ARP data (excluding pi.hole which is excluded by default)
	if onlineCount != expectedOnlineClients {
		return fmt.Errorf("expected %d online clients, got %d", expectedOnlineClients, onlineCount)
	}
	
	fmt.Printf("   Found %d online clients\n", onlineCount)
	return nil
}

// Test Pi-hole analysis with default settings
func testPiholeAnalysisDefault() error {
	dbFile := filepath.Join("test_data", "mock_pihole.db")
	
	// Reset flags to default
	*onlineOnlyFlag = false
	*noExcludeFlag = false
	
	clientStats, err := analyzePiholeDatabase(dbFile)
	if err != nil {
		return fmt.Errorf("Pi-hole analysis failed: %v", err)
	}
	
	// Should exclude Docker IPs
	expectedMinClients := 3
	expectedMaxClients := 8
	
	if len(clientStats) < expectedMinClients || len(clientStats) > expectedMaxClients {
		return fmt.Errorf("expected %d-%d clients, got %d", expectedMinClients, expectedMaxClients, len(clientStats))
	}
	
	fmt.Printf("   Found %d clients from Pi-hole database\n", len(clientStats))
	return nil
}

// Test Pi-hole analysis without exclusions
func testPiholeAnalysisNoExclusions() error {
	dbFile := filepath.Join("test_data", "mock_pihole.db")
	
	// Enable no-exclude flag
	*onlineOnlyFlag = false
	*noExcludeFlag = true
	
	clientStats, err := analyzePiholeDatabase(dbFile)
	if err != nil {
		return fmt.Errorf("Pi-hole analysis failed: %v", err)
	}
	
	// Should include all clients including Docker
	expectedMinClients := 7
	
	if len(clientStats) < expectedMinClients {
		return fmt.Errorf("expected at least %d clients, got %d", expectedMinClients, len(clientStats))
	}
	
	fmt.Printf("   Found %d clients (all included)\n", len(clientStats))
	return nil
}

// Test Pi-hole analysis with online-only filter
func testPiholeAnalysisOnlineOnly() error {
	dbFile := filepath.Join("test_data", "mock_pihole.db")
	
	// Enable online-only flag
	*onlineOnlyFlag = true
	*noExcludeFlag = false
	
	clientStats, err := analyzePiholeDatabase(dbFile)
	if err != nil {
		return fmt.Errorf("Pi-hole analysis failed: %v", err)
	}
	
	// Mock ARP checking
	err = mockCheckARPStatus(clientStats)
	if err != nil {
		return fmt.Errorf("ARP status check failed: %v", err)
	}
	
	fmt.Printf("   Found %d clients from Pi-hole database\n", len(clientStats))
	return nil
}

// Test ARP table functionality
func testARPFunctionality() error {
	arpEntries, err := MockARPTable(MockDataInstance)
	if err != nil {
		return fmt.Errorf("mock ARP table failed: %v", err)
	}
	
	expectedEntries := 5 // Based on mock data
	if len(arpEntries) != expectedEntries {
		return fmt.Errorf("expected %d ARP entries, got %d", expectedEntries, len(arpEntries))
	}
	
	// Check specific entries
	if entry, exists := arpEntries["192.168.2.110"]; !exists || !entry.IsOnline {
		return fmt.Errorf("expected 192.168.2.110 to be online in ARP table")
	}
	
	fmt.Printf("   Mock ARP table has %d entries\n", len(arpEntries))
	return nil
}

// Test hostname resolution
func testHostnameResolution() error {
	testIPs := []string{"192.168.2.110", "192.168.2.210", "192.168.2.6"}
	expectedHostnames := []string{"mac.home", "s21-van-marloes.home", "pi.hole"}
	
	for i, ip := range testIPs {
		hostname := MockHostnameResolution(ip, MockDataInstance)
		if hostname != expectedHostnames[i] {
			return fmt.Errorf("expected hostname %s for IP %s, got %s", expectedHostnames[i], ip, hostname)
		}
	}
	
	fmt.Printf("   Hostname resolution working for %d IPs\n", len(testIPs))
	return nil
}

// Test exclusion logic
func testExclusionLogic() error {
	exclusions := getDefaultExclusions("192.168.2.6")
	
	// Test cases
	testCases := []struct {
		ip       string
		hostname string
		shouldExclude bool
		reason    string
	}{
		{"172.20.0.8", "", true, "Docker network"},
		{"192.168.2.6", "", true, "Pi-hole host"},
		{"192.168.2.110", "", false, "Normal client"},
		{"127.0.0.1", "", true, "Loopback"},
		{"192.168.2.100", "pi.hole", true, "Pi-hole hostname"},
	}
	
	for _, tc := range testCases {
		shouldExclude, reason := shouldExcludeClient(tc.ip, tc.hostname, exclusions)
		if shouldExclude != tc.shouldExclude {
			return fmt.Errorf("IP %s: expected exclude=%t, got exclude=%t (reason: %s)", 
				tc.ip, tc.shouldExclude, shouldExclude, reason)
		}
	}
	
	fmt.Printf("   Exclusion logic tested for %d cases\n", len(testCases))
	return nil
}

// mockCheckARPStatus simulates ARP status checking for tests
func mockCheckARPStatus(clientStats map[string]*ClientStats) error {
	arpEntries := MockDataInstance.ARPEntries
	hostnames := MockDataInstance.Hostnames
	
	onlineCount := 0
	for ip, stats := range clientStats {
		if arpEntry, exists := arpEntries[ip]; exists && arpEntry.IsOnline {
			stats.IsOnline = true
			stats.ARPStatus = "online"
			stats.HWAddr = arpEntry.HWAddr
			onlineCount++
		} else {
			stats.IsOnline = false
			stats.ARPStatus = "offline"
		}
		
		// Set hostname if available
		if hostname, exists := hostnames[ip]; exists {
			stats.Hostname = hostname
		}
	}
	
	return nil
}
