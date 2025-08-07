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
		{"Colorized Output - Colors Enabled", testColorizedOutputEnabled},
		{"Colorized Output - Colors Disabled", testColorizedOutputDisabled},
		{"Domain Highlighting", testDomainHighlighting},
		{"Table Formatting", testTableFormatting},
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
		ip            string
		hostname      string
		shouldExclude bool
		reason        string
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

	fmt.Printf("   Mock ARP status applied to %d clients\n", len(clientStats))
	return nil
}

// Test colorized output with colors enabled
func testColorizedOutputEnabled() error {
	// Enable colors and emojis
	EnableColors()
	EnableEmojis()

	// Test basic color functions
	redText := Red("test")
	if !strings.Contains(redText, "\033[31m") || !strings.Contains(redText, "\033[0m") {
		return fmt.Errorf("Red() function not working properly")
	}

	greenText := BoldGreen("success")
	if !strings.Contains(greenText, "\033[1;32m") {
		return fmt.Errorf("BoldGreen() function not working properly")
	}

	// Test status functions
	onlineStatus := OnlineStatus(true, "reachable")
	if !strings.Contains(onlineStatus, "✅") || !strings.Contains(onlineStatus, "Online") {
		return fmt.Errorf("OnlineStatus() for online client not working properly")
	}

	offlineStatus := OnlineStatus(false, "timeout")
	if !strings.Contains(offlineStatus, "❌") || !strings.Contains(offlineStatus, "Offline") {
		return fmt.Errorf("OnlineStatus() for offline client not working properly")
	}

	// Test percentage coloring
	highPerc := ColoredPercentage(35.5)
	if !strings.Contains(highPerc, "\033[1;31m") { // Should be BoldRed
		return fmt.Errorf("ColoredPercentage() high value not colored properly")
	}

	lowPerc := ColoredPercentage(2.1)
	if !strings.Contains(lowPerc, "\033[90m") { // Should be Gray
		return fmt.Errorf("ColoredPercentage() low value not colored properly")
	}

	fmt.Printf("   Color functions working correctly\n")
	return nil
}

// Test colorized output with colors disabled
func testColorizedOutputDisabled() error {
	// Disable colors
	DisableColors()

	// Test that color functions return plain text
	redText := Red("test")
	if strings.Contains(redText, "\033[") {
		return fmt.Errorf("Red() should return plain text when colors disabled, got: %q", redText)
	}
	if redText != "test" {
		return fmt.Errorf("Red() should return original text when colors disabled")
	}

	// Test status functions without colors
	onlineStatus := OnlineStatus(true, "reachable")
	if strings.Contains(onlineStatus, "\033[") {
		return fmt.Errorf("OnlineStatus() should not contain color codes when disabled")
	}

	// Test that emojis still work when colors are disabled
	if !strings.Contains(onlineStatus, "✅") {
		return fmt.Errorf("OnlineStatus() should still show emojis when colors disabled")
	}

	// Re-enable colors for other tests
	EnableColors()

	fmt.Printf("   Color disabling working correctly\n")
	return nil
}

// Test domain highlighting functionality
func testDomainHighlighting() error {
	EnableColors()

	testCases := []struct {
		domain    string
		colorCode string
		category  string
	}{
		{"google.com", "\033[1;32m", "major service"},
		{"microsoft.com", "\033[1;32m", "major service"},
		{"github.com", "\033[1;36m", "development"},
		{"stackoverflow.com", "\033[1;36m", "development"},
		{"ads.example.com", "\033[1;31m", "ads/tracking"},
		{"tracking.test.com", "\033[1;31m", "ads/tracking"},
		{"doubleclick.net", "\033[1;31m", "ads/tracking"},
		{"telemetry.service.com", "\033[1;31m", "ads/tracking"},
		{"example.com", "", "no highlighting"},
		{"netflix.com", "", "no highlighting"},
	}

	for _, tc := range testCases {
		result := HighlightDomain(tc.domain)
		
		if tc.colorCode == "" {
			// Should return unchanged domain
			if result != tc.domain {
				return fmt.Errorf("HighlightDomain(%q) should return unchanged for %s, got: %q", 
					tc.domain, tc.category, result)
			}
		} else {
			// Should contain the color code
			if !strings.Contains(result, tc.colorCode) {
				return fmt.Errorf("HighlightDomain(%q) should contain %q for %s, got: %q", 
					tc.domain, tc.colorCode, tc.category, result)
			}
			// Should contain the domain name
			if !strings.Contains(result, tc.domain) {
				return fmt.Errorf("HighlightDomain(%q) should contain domain name, got: %q", 
					tc.domain, result)
			}
		}
	}

	fmt.Printf("   Domain highlighting working for %d test cases\n", len(testCases))
	return nil
}

// Test table formatting with colors
func testTableFormatting() error {
	EnableColors()

	// Test color code stripping
	coloredText := BoldRed("colored text")
	stripped := stripColorCodes(coloredText)
	if stripped != "colored text" {
		return fmt.Errorf("stripColorCodes() failed: expected 'colored text', got %q", stripped)
	}

	// Test display length calculation
	displayLen := getDisplayLength(coloredText)
	if displayLen != 12 { // "colored text" is 12 characters
		return fmt.Errorf("getDisplayLength() failed: expected 12, got %d", displayLen)
	}

	// Test table column formatting
	formatted := formatTableColumn(coloredText, 20)
	formattedDisplayLen := getDisplayLength(formatted)
	if formattedDisplayLen != 20 {
		return fmt.Errorf("formatTableColumn() failed: expected display length 20, got %d", 
			formattedDisplayLen)
	}

	// Test that colored text is preserved in formatting
	if !strings.Contains(formatted, "\033[1;31m") {
		return fmt.Errorf("formatTableColumn() should preserve color codes")
	}

	// Test right-aligned formatting
	rightFormatted := formatTableColumnRight("test", 10)
	rightDisplayLen := getDisplayLength(rightFormatted)
	if rightDisplayLen != 10 {
		return fmt.Errorf("formatTableColumnRight() failed: expected display length 10, got %d", 
			rightDisplayLen)
	}

	// Test with various IP addresses
	testIPs := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1", "8.8.8.8"}
	for _, ip := range testIPs {
		highlighted := HighlightIP(ip)
		if !strings.Contains(highlighted, ip) {
			return fmt.Errorf("HighlightIP(%q) should contain the IP address", ip)
		}
		
		// Private IPs should be blue, public should be yellow
		if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
			if !strings.Contains(highlighted, "\033[1;34m") { // BoldBlue
				return fmt.Errorf("HighlightIP(%q) should be blue for private IP", ip)
			}
		} else {
			if !strings.Contains(highlighted, "\033[1;33m") { // BoldYellow
				return fmt.Errorf("HighlightIP(%q) should be yellow for public IP", ip)
			}
		}
	}

	fmt.Printf("   Table formatting working correctly\n")
	return nil
}
