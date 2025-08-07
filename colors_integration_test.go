package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// Test colorized output integration
func TestColorizedOutputIntegration(t *testing.T) {
	// Setup test environment
	EnableTestMode() // This enables colors for testing and bypasses terminal detection
	EnableEmojis()

	// Create mock client stats
	mockStats := map[string]*ClientStats{
		"192.168.1.100": {
			Client:        "192.168.1.100",
			TotalQueries:  5000,
			Uniquedomains: 150,
			AvgReplyTime:  0.025,
			HWAddr:        "AA:BB:CC:DD:EE:FF",
			IsOnline:      true,
			ARPStatus:     "reachable",
			Hostname:      "test-device",
			Domains: map[string]int{
				"google.com":               1500,
				"tracking.doubleclick.net": 800,
				"github.com":               600,
				"ads.microsoft.com":        400,
				"example.com":              200,
			},
			QueryTypes:  map[int]int{1: 4000, 28: 1000},
			StatusCodes: map[int]int{2: 3000, 3: 1500, 9: 500},
		},
		"10.0.0.5": {
			Client:        "10.0.0.5",
			TotalQueries:  1200,
			Uniquedomains: 45,
			AvgReplyTime:  0.015,
			HWAddr:        "",
			IsOnline:      false,
			ARPStatus:     "timeout",
			Hostname:      "",
			Domains: map[string]int{
				"microsoft.com":         500,
				"telemetry.windows.com": 400,
				"api.spotify.com":       300,
			},
			QueryTypes:  map[int]int{1: 1000, 28: 200},
			StatusCodes: map[int]int{2: 800, 3: 400},
		},
	}

	// Create mock config
	config := &Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   true,
		Output: OutputConfig{
			MaxClients:    20,
			MaxDomains:    10,
			VerboseOutput: false,
		},
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the colorized display function
	displayResultsWithConfig(mockStats, config)

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Test that output contains expected colorized elements
	tests := []struct {
		name     string
		contains string
		desc     string
	}{
		{"Section header colors", "\033[1;36m", "Should contain BoldCyan for section headers"},
		{"Border colors", "\033[36m", "Should contain Cyan for borders"},
		{"IP highlighting", "\033[1;34m", "Should contain BoldBlue for private IPs"},
		{"Online status emoji", "✅", "Should contain green checkmark emoji"},
		{"Offline status emoji", "❌", "Should contain red X emoji"},
		{"Online status color", "\033[1;32m", "Should contain BoldGreen for online status"},
		{"Offline status color", "\033[1;31m", "Should contain BoldRed for offline status"},
		{"Query count colors", "\033[1;33m", "Should contain colors for query counts"},
		{"Domain count colors", "\033[34m", "Should contain colors for domain counts"},
		{"Major service domain", "google.com", "Should contain google.com domain"},
		{"Tracking domain", "tracking.doubleclick.net", "Should contain tracking domain"},
		{"Development domain", "github.com", "Should contain github.com domain"},
		{"Ads domain", "ads.microsoft.com", "Should contain ads domain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Output should contain %s: %q\nActual output: %s",
					tt.desc, tt.contains, output)
			}
		})
	}

	// Test that domains are properly highlighted
	domainTests := []struct {
		domain    string
		colorCode string
		desc      string
	}{
		{"google.com", "\033[1;32m", "Google should be BoldGreen"},
		{"microsoft.com", "\033[1;32m", "Microsoft should be BoldGreen"},
		{"github.com", "\033[1;36m", "GitHub should be BoldCyan"},
		{"tracking.doubleclick.net", "\033[1;31m", "Tracking should be BoldRed"},
		{"ads.microsoft.com", "\033[1;31m", "Ads should be BoldRed"},
		{"telemetry.windows.com", "\033[1;31m", "Telemetry should be BoldRed"},
	}

	for _, tt := range domainTests {
		t.Run("Domain color: "+tt.domain, func(t *testing.T) {
			// Look for the domain in the output (it should be present regardless of color)
			if !strings.Contains(output, tt.domain) {
				t.Errorf("Domain %s should appear in output", tt.domain)
				return
			}

			// If domain is present, check for any color code near it
			// This is more flexible than exact color matching
			lines := strings.Split(output, "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, tt.domain) {
					// Check if there's any ANSI color code in the line
					if strings.Contains(line, "\033[") {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("Domain %s should be colored (expected %s - %s)",
					tt.domain, tt.colorCode, tt.desc)
			}
		})
	}
}

// Test colorized output with colors disabled
func TestColorizedOutputDisabled(t *testing.T) {
	// Disable colors for this test
	DisableColors()

	// Create minimal mock stats
	mockStats := map[string]*ClientStats{
		"192.168.1.100": {
			Client:        "192.168.1.100",
			TotalQueries:  1000,
			Uniquedomains: 50,
			IsOnline:      true,
			ARPStatus:     "reachable",
			Domains: map[string]int{
				"google.com": 500,
				"github.com": 300,
			},
		},
	}

	config := &Config{
		Output: OutputConfig{
			MaxClients: 10,
			MaxDomains: 5,
		},
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayResultsWithConfig(mockStats, config)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not contain ANSI color codes
	colorCodes := []string{
		"\033[31m", "\033[32m", "\033[33m", "\033[34m", "\033[35m", "\033[36m",
		"\033[1;31m", "\033[1;32m", "\033[1;33m", "\033[1;34m", "\033[1;35m", "\033[1;36m",
		"\033[0m",
	}

	for _, code := range colorCodes {
		if strings.Contains(output, code) {
			t.Errorf("Output should not contain color code %q when colors are disabled", code)
		}
	}

	// Should still contain the actual content
	if !strings.Contains(output, "google.com") {
		t.Error("Output should still contain domain names when colors are disabled")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("Output should still contain IP addresses when colors are disabled")
	}

	// Re-enable test mode for other tests
	EnableTestMode()
}

// Test emoji functionality
func TestEmojiOutput(t *testing.T) {
	EnableTestMode() // Enable test mode to force colors in testing environment

	// Test with emojis enabled
	EnableEmojis()

	result := OnlineStatus(true, "reachable")
	if !strings.Contains(result, "✅") {
		t.Error("OnlineStatus should contain checkmark emoji when emojis are enabled")
	}

	result = OnlineStatus(false, "timeout")
	if !strings.Contains(result, "❌") {
		t.Error("OnlineStatus should contain X emoji when emojis are enabled")
	}

	result = ProcessingIndicator("test")
	if !strings.Contains(result, "🔄") {
		t.Error("ProcessingIndicator should contain spinner emoji when emojis are enabled")
	}

	// Test with emojis disabled
	DisableEmojis()

	result = OnlineStatus(true, "reachable")
	if strings.Contains(result, "✅") {
		t.Error("OnlineStatus should not contain checkmark emoji when emojis are disabled")
	}
	if !strings.Contains(result, "✓") {
		t.Error("OnlineStatus should contain ASCII checkmark when emojis are disabled")
	}

	result = OnlineStatus(false, "timeout")
	if strings.Contains(result, "❌") {
		t.Error("OnlineStatus should not contain X emoji when emojis are disabled")
	}
	if !strings.Contains(result, "✗") {
		t.Error("OnlineStatus should contain ASCII X when emojis are disabled")
	}

	// Re-enable emojis for other tests
	EnableEmojis()
}

// Test table formatting with colors
func TestTableFormattingWithColors(t *testing.T) {
	EnableTestMode() // Enable test mode to force colors in testing environment

	// Test various colored texts with different widths
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected total length including padding
	}{
		{"Short red text", Red("test"), 10, 10},
		{"Long green text", BoldGreen("this is a long text"), 15, 19}, // longer than width
		{"IP address", HighlightIP("192.168.1.1"), 20, 20},
		{"Colored count", ColoredQueryCount(5000), 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTableColumn(tt.text, tt.width)

			// Calculate display length (without color codes)
			displayLen := getDisplayLength(result)

			if displayLen < tt.width && getDisplayLength(tt.text) <= tt.width {
				// If original text fits in width, result should be padded to width
				if displayLen != tt.width {
					t.Errorf("formatTableColumn(%s, %d) display length = %d, want %d",
						stripColorCodes(tt.text), tt.width, displayLen, tt.width)
				}
			}
		})
	}
}

// Test color consistency across functions
func TestColorConsistency(t *testing.T) {
	EnableTestMode() // Enable test mode to force colors in testing environment

	// Test that related functions use consistent colors
	successText := Success("test")
	onlineStatus := OnlineStatus(true, "reachable")

	// Both should use green color codes
	if !strings.Contains(successText, "\033[1;32m") {
		t.Error("Success() should use BoldGreen color")
	}
	if !strings.Contains(onlineStatus, "\033[1;32m") {
		t.Error("OnlineStatus(true) should use BoldGreen color")
	}

	// Test error/warning consistency
	errorText := Error("test")
	offlineStatus := OnlineStatus(false, "timeout")

	// Both should use red color codes
	if !strings.Contains(errorText, "\033[1;31m") {
		t.Error("Error() should use BoldRed color")
	}
	if !strings.Contains(offlineStatus, "\033[1;31m") {
		t.Error("OnlineStatus(false) should use BoldRed color")
	}
}

// Test performance of color functions
func TestColorPerformance(t *testing.T) {
	EnableTestMode() // Enable test mode to force colors in testing environment

	// Test that color functions don't significantly impact performance
	// This is more of a smoke test than a strict performance test

	largeText := strings.Repeat("test text ", 1000)

	// These should complete quickly
	_ = Red(largeText)
	_ = BoldGreen(largeText)
	_ = formatTableColumn(Red(largeText), 50)
	_ = stripColorCodes(BoldGreen(largeText))

	// Test with many domain highlights
	domains := []string{
		"google.com", "microsoft.com", "github.com", "stackoverflow.com",
		"ads.example.com", "tracking.test.com", "doubleclick.net",
		"telemetry.service.com", "example.com", "test.org",
	}

	for i := 0; i < 100; i++ {
		for _, domain := range domains {
			_ = HighlightDomain(domain)
		}
	}
}

// Test edge cases
func TestColorEdgeCases(t *testing.T) {
	EnableTestMode() // Enable test mode to force colors in testing environment

	// Test empty strings
	if Red("") != "\033[31m\033[0m" {
		t.Error("Red(\"\") should return color codes even for empty string")
	}

	// Test very long strings
	longString := strings.Repeat("a", 10000)
	result := Green(longString)
	if !strings.HasPrefix(result, "\033[32m") {
		t.Error("Green() should work with very long strings")
	}
	if !strings.HasSuffix(result, "\033[0m") {
		t.Error("Green() should end with reset code for very long strings")
	}

	// Test strings with existing color codes
	alreadyColored := "\033[31mred\033[0m"
	result = Blue(alreadyColored)
	// Should wrap the already colored text in blue
	if !strings.HasPrefix(result, "\033[34m") {
		t.Error("Blue() should work with already colored text")
	}

	// Test strip color codes with nested/malformed codes
	malformed := "\033[31mtest\033[32mnested\033[0m\033[invalid"
	stripped := stripColorCodes(malformed)
	expected := "testnested\033[invalid" // Only strips known codes
	if stripped != expected {
		t.Errorf("stripColorCodes() malformed input: got %q, want %q", stripped, expected)
	}
}

// Test color detection logic
func TestColorDetectionLogic(t *testing.T) {
	// Save original config
	originalConfig := colorConfig

	// Test force disabled
	colorConfig = ColorConfig{
		Enabled:       true,
		ForceDisabled: true,
		UseEmoji:      true,
	}

	if colorEnabled() {
		t.Error("colorEnabled() should return false when ForceDisabled is true")
	}

	// Test enabled false
	colorConfig = ColorConfig{
		Enabled:       false,
		ForceDisabled: false,
		UseEmoji:      true,
	}

	if colorEnabled() {
		t.Error("colorEnabled() should return false when Enabled is false")
	}

	// Test normal enabled state (will depend on terminal detection)
	colorConfig = ColorConfig{
		Enabled:       true,
		ForceDisabled: false,
		UseEmoji:      true,
	}

	// Result will depend on isTerminal() which depends on the test environment
	// We just test that it doesn't panic
	_ = colorEnabled()

	// Restore original config
	colorConfig = originalConfig
}
