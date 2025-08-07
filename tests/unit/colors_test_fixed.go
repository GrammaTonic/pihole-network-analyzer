package main

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"pihole-network-analyzer/internal/colors"
)

// Test basic color functions
func TestColorFunctions(t *testing.T) {
	// Enable test mode to bypass terminal detection
	colors.EnableTestMode()
	defer colors.DisableColors()

	tests := []struct {
		name     string
		function func(string) string
		input    string
		contains string
	}{
		{"Red", colors.Red, "test", "\033[31m"},
		{"Green", colors.Green, "test", "\033[32m"},
		{"Yellow", colors.Yellow, "test", "\033[33m"},
		{"Blue", colors.Blue, "test", "\033[34m"},
		{"Purple", colors.Purple, "test", "\033[35m"},
		{"Cyan", colors.Cyan, "test", "\033[36m"},
		{"White", colors.White, "test", "\033[37m"},
		{"Gray", colors.Gray, "test", "\033[90m"},
		{"BoldRed", colors.BoldRed, "test", "\033[1;31m"},
		{"BoldGreen", colors.BoldGreen, "test", "\033[1;32m"},
		{"BoldYellow", colors.BoldYellow, "test", "\033[1;33m"},
		{"BoldBlue", colors.BoldBlue, "test", "\033[1;34m"},
		{"BoldPurple", colors.BoldPurple, "test", "\033[1;35m"},
		{"BoldCyan", colors.BoldCyan, "test", "\033[1;36m"},
		{"BoldWhite", colors.BoldWhite, "test", "\033[1;37m"},
		{"Bold", colors.Bold, "test", "\033[1m"},
		{"Underline", colors.Underline, "test", "\033[4m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("%s() = %q, should contain %q", tt.name, result, tt.contains)
			}
			if !strings.Contains(result, tt.input) {
				t.Errorf("%s() = %q, should contain input %q", tt.name, result, tt.input)
			}
		})
	}
}

// Test status-specific color functions
func TestStatusColorFunctions(t *testing.T) {
	colors.EnableTestMode()
	defer colors.DisableColors()

	tests := []struct {
		name     string
		function func(string) string
		input    string
		contains string
	}{
		{"Success", colors.Success, "OK", "\033[1;32m"},   // BoldGreen
		{"Warning", colors.Warning, "WARN", "\033[1;33m"}, // BoldYellow
		{"Error", colors.Error, "ERR", "\033[1;31m"},      // BoldRed
		{"Info", colors.Info, "INFO", "\033[1;36m"},       // BoldCyan
		{"Header", colors.Header, "HEADER", "\033[1;37m"}, // BoldWhite
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("%s() = %q, should contain %q", tt.name, result, tt.contains)
			}
			if !strings.Contains(result, tt.input) {
				t.Errorf("%s() = %q, should contain input %q", tt.name, result, tt.input)
			}
		})
	}
}

// Test emoji status functions
func TestEmojiStatusFunctions(t *testing.T) {
	colors.EnableTestMode()
	colors.EnableEmojis()
	defer colors.DisableColors()

	tests := []struct {
		name      string
		isOnline  bool
		arpStatus string
		contains  string
	}{
		{"Online with ARP", true, "active", "🟢"},
		{"Online without ARP", true, "", "🟡"},
		{"Offline", false, "", "🔴"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.OnlineStatus(tt.isOnline, tt.arpStatus)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("OnlineStatus(%t, %q) = %q, should contain %q", tt.isOnline, tt.arpStatus, result, tt.contains)
			}
		})
	}
}

// Test emoji status functions without emojis
func TestStatusFunctionsNoEmojis(t *testing.T) {
	colors.EnableTestMode()
	colors.DisableEmojis()
	defer func() {
		colors.DisableColors()
		colors.EnableEmojis() // Reset for other tests
	}()

	tests := []struct {
		name      string
		isOnline  bool
		arpStatus string
		contains  string
	}{
		{"Online with ARP", true, "active", "ONLINE"},
		{"Online without ARP", true, "", "ONLINE"},
		{"Offline", false, "", "OFFLINE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.OnlineStatus(tt.isOnline, tt.arpStatus)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("OnlineStatus(%t, %q) = %q, should contain %q", tt.isOnline, tt.arpStatus, result, tt.contains)
			}
		})
	}
}

// Test percentage coloring functionality
func TestColoredPercentage(t *testing.T) {
	// Enable emoji support for this test
	colors.EnableEmojis()

	// Enable test mode so we can reliably test color output
	colors.EnableTestMode()
	defer colors.DisableColors()

	tests := []struct {
		name     string
		value    float64
		contains []string // Multiple strings that should be present
	}{
		{"Zero percent", 0.0, []string{"0.00", "%"}},
		{"Low percent", 15.5, []string{"15.50", "%"}},
		{"Medium percent", 45.0, []string{"45.00", "%"}},
		{"High percent", 75.25, []string{"75.25", "%"}},
		{"Full percent", 100.0, []string{"100.00", "%"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.ColoredPercentage(tt.value)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("ColoredPercentage(%f) = %q, should contain %q", tt.value, result, expected)
				}
			}
		})
	}
}

// Test query count coloring
func TestColoredQueryCount(t *testing.T) {
	colors.EnableTestMode()

	tests := []struct {
		name  string
		count int
		check func(string) bool
	}{
		{"Zero queries", 0, func(s string) bool { return strings.Contains(s, "0") }},
		{"Low queries", 50, func(s string) bool { return strings.Contains(s, "50") }},
		{"Medium queries", 500, func(s string) bool { return strings.Contains(s, "500") }},
		{"High queries", 5000, func(s string) bool { return strings.Contains(s, "5,000") }},
		{"Very high queries", 50000, func(s string) bool { return strings.Contains(s, "50,000") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.ColoredQueryCount(tt.count)
			if !tt.check(result) {
				t.Errorf("ColoredQueryCount(%d) = %q, check failed", tt.count, result)
			}
		})
	}
}

// Test domain count coloring
func TestColoredDomainCount(t *testing.T) {
	colors.EnableTestMode()

	tests := []struct {
		name  string
		count int
		check func(string) bool
	}{
		{"Zero domains", 0, func(s string) bool { return strings.Contains(s, "0") }},
		{"Few domains", 5, func(s string) bool { return strings.Contains(s, "5") }},
		{"Many domains", 50, func(s string) bool { return strings.Contains(s, "50") }},
		{"Lots of domains", 500, func(s string) bool { return strings.Contains(s, "500") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.ColoredDomainCount(tt.count)
			if !tt.check(result) {
				t.Errorf("ColoredDomainCount(%d) = %q, check failed", tt.count, result)
			}
		})
	}
}

// Test IP highlighting
func TestHighlightIP(t *testing.T) {
	colors.EnableTestMode()

	tests := []struct {
		name string
		ip   string
	}{
		{"IPv4 local", "192.168.1.1"},
		{"IPv4 public", "8.8.8.8"},
		{"IPv6 local", "::1"},
		{"IPv6 full", "2001:4860:4860::8888"},
		{"Invalid IP", "not.an.ip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.HighlightIP(tt.ip)
			if !strings.Contains(result, tt.ip) {
				t.Errorf("HighlightIP(%q) = %q, should contain %q", tt.ip, result, tt.ip)
			}
		})
	}
}

// Test domain highlighting
func TestHighlightDomain(t *testing.T) {
	colors.EnableTestMode()

	tests := []struct {
		name   string
		domain string
	}{
		{"Simple domain", "example.com"},
		{"Subdomain", "www.example.com"},
		{"Complex domain", "api.v2.service.example.co.uk"},
		{"Single word", "localhost"},
		{"With protocol", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.HighlightDomain(tt.domain)
			// Domain should be present in the output
			if !strings.Contains(result, tt.domain) {
				t.Errorf("HighlightDomain(%q) = %q, should contain %q", tt.domain, result, tt.domain)
			}
		})
	}
}

// Test domain highlighting with wildcard matching
func TestHighlightDomainWildcard(t *testing.T) {
	colors.EnableTestMode()

	domains := []string{"example.com", "test.example.com", "api.example.com"}

	for _, domain := range domains {
		t.Run(domain, func(t *testing.T) {
			result := colors.HighlightDomain(domain)
			if !strings.Contains(result, domain) {
				t.Errorf("HighlightDomain(%q) = %q, should contain domain", domain, result)
			}
		})
	}
}

// Test section headers
func TestSectionHeader(t *testing.T) {
	colors.EnableTestMode()
	title := "Test Section"

	result := colors.SectionHeader(title)

	// Should contain the title
	if !strings.Contains(result, title) {
		t.Errorf("SectionHeader(%q) = %q, should contain title", title, result)
	}

	// Should contain some formatting characters (like equals signs or dashes)
	if !strings.ContainsAny(result, "=-") {
		t.Errorf("SectionHeader(%q) = %q, should contain formatting characters", title, result)
	}
}

// Test subsection headers
func TestSubSectionHeader(t *testing.T) {
	colors.EnableTestMode()
	title := "Test Subsection"

	result := colors.SubSectionHeader(title)

	// Should contain the title
	if !strings.Contains(result, title) {
		t.Errorf("SubSectionHeader(%q) = %q, should contain title", title, result)
	}

	// Should contain some formatting characters
	if !strings.ContainsAny(result, "=-") {
		t.Errorf("SubSectionHeader(%q) = %q, should contain formatting characters", title, result)
	}
}

// Test processing indicator
func TestProcessingIndicator(t *testing.T) {
	colors.EnableTestMode()
	colors.EnableEmojis()
	message := "Processing data"

	result := colors.ProcessingIndicator(message)

	// Should contain the message
	if !strings.Contains(result, message) {
		t.Errorf("ProcessingIndicator(%q) = %q, should contain message", message, result)
	}
}

// Test stripColorCodes function (internal function testing)
func TestStripColorCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No colors", "plain text", "plain text"},
		{"Simple color", "\033[31mred text\033[0m", "red text"},
		{"Multiple colors", "\033[31mred\033[32mgreen\033[0m", "redgreen"},
		{"Complex colors", "\033[1;31mbold red\033[0m normal", "bold red normal"},
		{"Only color codes", "\033[31m\033[0m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test this through a public function that uses stripColorCodes
			// Since it's not exported, we'll test the display length function
			result := colors.GetDisplayLength(tt.input)
			expected := len(tt.expected)
			if result != expected {
				t.Errorf("GetDisplayLength(%q) = %d, want %d", tt.input, result, expected)
			}
		})
	}
}

// Test getDisplayLength function
func TestGetDisplayLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Plain text", "hello", 5},
		{"With colors", "\033[31mhello\033[0m", 5},
		{"Empty string", "", 0},
		{"Only colors", "\033[31m\033[0m", 0},
		{"Mixed content", "\033[31mred\033[0m and \033[32mgreen\033[0m", 13}, // "red and green"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.GetDisplayLength(tt.input)
			if result != tt.expected {
				t.Errorf("GetDisplayLength(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// Test formatTableColumn function
func TestFormatTableColumn(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // Expected display width
	}{
		{"Short text", "hi", 10, 10},
		{"Exact width", "exact", 5, 5},
		{"Long text", "this is too long", 5, 5},
		{"With colors", "\033[31mred\033[0m", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colors.FormatTableColumn(tt.text, tt.width)
			// Check that the display width is correct by stripping colors and measuring
			displayOnly := colors.GetDisplayLength(result)
			originalDisplay := colors.GetDisplayLength(tt.text)

			// For truncated text, display should be <= width
			// For padded text, display should be exactly width (if original was shorter)
			if originalDisplay > tt.width {
				if displayOnly > tt.width {
					t.Errorf("FormatTableColumn(%q, %d) display width = %d, should be <= %d", tt.text, tt.width, displayOnly, tt.width)
				}
			} else {
				if displayOnly != tt.expected {
					t.Errorf("FormatTableColumn(%q, %d) display width = %d, want %d", tt.text, tt.width, displayOnly, tt.expected)
				}
			}
		})
	}
}

// Test color configuration functions
func TestColorConfiguration(t *testing.T) {
	colors.EnableTestMode()
	if colors.IsColorEnabled() {
		t.Error("Colors should be disabled in test mode initially")
	}
	if !colors.IsTestMode() {
		t.Error("Test mode should be enabled")
	}

	colors.DisableColors()
	if !colors.IsForceDisabled() {
		t.Error("Colors should be force disabled after DisableColors()")
	}

	colors.EnableEmojis()
	if !colors.IsEmojiEnabled() {
		t.Error("Emojis should be enabled after EnableEmojis()")
	}

	colors.DisableEmojis()
	if colors.IsEmojiEnabled() {
		t.Error("Emojis should be disabled after DisableEmojis()")
	}
}

// Test Windows compatibility
func TestWindowsCompatibility(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// On Windows, colors might be handled differently
	colors.EnableTestMode()
	result := colors.Red("test")

	// Should still contain the text even if colors are disabled on Windows
	if !strings.Contains(result, "test") {
		t.Errorf("Red() should contain input text even on Windows, got: %q", result)
	}
}

// Test environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Save original values
	originalTerm := os.Getenv("TERM")
	originalNoColor := os.Getenv("NO_COLOR")

	defer func() {
		os.Setenv("TERM", originalTerm)
		os.Setenv("NO_COLOR", originalNoColor)
	}()

	// Test NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")
	colors.EnableTestMode()

	// Colors should be disabled when NO_COLOR is set
	result := colors.Red("test")
	// In test mode, we can still test the logic
	if !strings.Contains(result, "test") {
		t.Errorf("Function should still return text when NO_COLOR is set, got: %q", result)
	}
}
