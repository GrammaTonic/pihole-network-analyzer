package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// Test basic color functions
func TestColorFunctions(t *testing.T) {
	// Enable test mode to bypass terminal detection
	EnableTestMode()
	defer DisableColors()

	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{"Red", Red, "test", "\033[31mtest\033[0m"},
		{"Green", Green, "test", "\033[32mtest\033[0m"},
		{"Yellow", Yellow, "test", "\033[33mtest\033[0m"},
		{"Blue", Blue, "test", "\033[34mtest\033[0m"},
		{"Purple", Purple, "test", "\033[35mtest\033[0m"},
		{"Cyan", Cyan, "test", "\033[36mtest\033[0m"},
		{"Gray", Gray, "test", "\033[90mtest\033[0m"},
		{"BoldRed", BoldRed, "test", "\033[1;31mtest\033[0m"},
		{"BoldGreen", BoldGreen, "test", "\033[1;32mtest\033[0m"},
		{"BoldYellow", BoldYellow, "test", "\033[1;33mtest\033[0m"},
		{"BoldBlue", BoldBlue, "test", "\033[1;34mtest\033[0m"},
		{"BoldPurple", BoldPurple, "test", "\033[1;35mtest\033[0m"},
		{"BoldCyan", BoldCyan, "test", "\033[1;36mtest\033[0m"},
		{"BoldWhite", BoldWhite, "test", "\033[1;37mtest\033[0m"},
		{"Bold", Bold, "test", "\033[1mtest\033[0m"},
		{"Underline", Underline, "test", "\033[4mtest\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

// Test status-specific color functions
func TestStatusColorFunctions(t *testing.T) {
	EnableTestMode()
	defer DisableColors()

	tests := []struct {
		name     string
		function func(string) string
		input    string
		contains string
	}{
		{"Success", Success, "OK", "\033[1;32m"},   // BoldGreen
		{"Warning", Warning, "WARN", "\033[1;33m"}, // BoldYellow
		{"Error", Error, "ERR", "\033[1;31m"},      // BoldRed
		{"Info", Info, "INFO", "\033[1;36m"},       // BoldCyan
		{"Header", Header, "HEADER", "\033[1;37m"}, // BoldWhite
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
			if !strings.HasSuffix(result, "\033[0m") {
				t.Errorf("%s() = %q, should end with reset code", tt.name, result)
			}
		})
	}
}

// Test online status function
func TestOnlineStatus(t *testing.T) {
	EnableTestMode()
	EnableEmojis()
	defer DisableColors()

	tests := []struct {
		name      string
		isOnline  bool
		arpStatus string
		contains  []string
	}{
		{"Online with emoji", true, "reachable", []string{"✅", "Online", "\033[1;32m"}},
		{"Offline with emoji", false, "timeout", []string{"❌", "Offline", "\033[1;31m"}},
		{"Unknown status", true, "unknown", []string{"❓", "Unknown", "\033[90m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OnlineStatus(tt.isOnline, tt.arpStatus)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("OnlineStatus(%t, %q) = %q, should contain %q",
						tt.isOnline, tt.arpStatus, result, expected)
				}
			}
		})
	}
}

// Test online status without emojis
func TestOnlineStatusNoEmoji(t *testing.T) {
	EnableTestMode()
	DisableEmojis()
	defer func() {
		DisableColors()
		EnableEmojis() // Reset for other tests
	}()

	tests := []struct {
		name        string
		isOnline    bool
		arpStatus   string
		contains    []string
		notContains []string
	}{
		{"Online no emoji", true, "reachable", []string{"✓", "Online"}, []string{"✅"}},
		{"Offline no emoji", false, "timeout", []string{"✗", "Offline"}, []string{"❌"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OnlineStatus(tt.isOnline, tt.arpStatus)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("OnlineStatus(%t, %q) = %q, should contain %q",
						tt.isOnline, tt.arpStatus, result, expected)
				}
			}
			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("OnlineStatus(%t, %q) = %q, should not contain %q",
						tt.isOnline, tt.arpStatus, result, notExpected)
				}
			}
		})
	}

	// Re-enable emojis for other tests
	EnableEmojis()
}

// Test percentage coloring
func TestColoredPercentage(t *testing.T) {
	EnableTestMode()
	defer DisableColors()

	tests := []struct {
		name     string
		value    float64
		contains string
		color    string
	}{
		{"High percentage", 45.5, "45.50%", "\033[1;31m"},   // BoldRed
		{"Medium percentage", 20.0, "20.00%", "\033[1;33m"}, // BoldYellow
		{"Low-medium percentage", 8.5, "8.50%", "\033[33m"}, // Yellow
		{"Very low percentage", 2.1, "2.10%", "\033[90m"},   // Gray
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColoredPercentage(tt.value)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("ColoredPercentage(%f) = %q, should contain %q",
					tt.value, result, tt.contains)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("ColoredPercentage(%f) = %q, should contain color %q",
					tt.value, result, tt.color)
			}
		})
	}
}

// Test query count coloring
func TestColoredQueryCount(t *testing.T) {
	EnableTestMode()

	tests := []struct {
		name     string
		count    int
		contains string
		color    string
	}{
		{"Very high count", 15000, "15000", "\033[1;31m"}, // BoldRed
		{"High count", 7500, "7500", "\033[1;33m"},        // BoldYellow
		{"Medium count", 2500, "2500", "\033[33m"},        // Yellow
		{"Low count", 500, "500", "\033[36m"},             // Cyan
		{"Very low count", 50, "50", "\033[90m"},          // Gray
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColoredQueryCount(tt.count)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("ColoredQueryCount(%d) = %q, should contain %q",
					tt.count, result, tt.contains)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("ColoredQueryCount(%d) = %q, should contain color %q",
					tt.count, result, tt.color)
			}
		})
	}
}

// Test domain count coloring
func TestColoredDomainCount(t *testing.T) {
	EnableTestMode()

	tests := []struct {
		name     string
		count    int
		contains string
		color    string
	}{
		{"Very diverse", 750, "750", "\033[1;35m"},   // BoldPurple
		{"Diverse", 350, "350", "\033[35m"},          // Purple
		{"Medium diversity", 100, "100", "\033[34m"}, // Blue
		{"Low diversity", 25, "25", "\033[90m"},      // Gray
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColoredDomainCount(tt.count)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("ColoredDomainCount(%d) = %q, should contain %q",
					tt.count, result, tt.contains)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("ColoredDomainCount(%d) = %q, should contain color %q",
					tt.count, result, tt.color)
			}
		})
	}
}

// Test IP address highlighting
func TestHighlightIP(t *testing.T) {
	EnableTestMode()

	tests := []struct {
		name  string
		ip    string
		color string
		desc  string
	}{
		{"Private 192.168", "192.168.1.100", "\033[1;34m", "BoldBlue"},
		{"Private 10.x", "10.0.0.5", "\033[1;34m", "BoldBlue"},
		{"Private 172.x", "172.16.0.10", "\033[1;34m", "BoldBlue"},
		{"Public IP", "8.8.8.8", "\033[1;33m", "BoldYellow"},
		{"Public IP 2", "1.1.1.1", "\033[1;33m", "BoldYellow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightIP(tt.ip)
			if !strings.Contains(result, tt.ip) {
				t.Errorf("HighlightIP(%q) = %q, should contain IP", tt.ip, result)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("HighlightIP(%q) = %q, should contain %s color %q",
					tt.ip, result, tt.desc, tt.color)
			}
		})
	}
}

// Test domain highlighting
func TestHighlightDomain(t *testing.T) {
	EnableTestMode()

	tests := []struct {
		name   string
		domain string
		color  string
		desc   string
	}{
		{"Google service", "google.com", "\033[1;32m", "BoldGreen"},
		{"Microsoft service", "api.microsoft.com", "\033[1;32m", "BoldGreen"},
		{"GitHub development", "github.com", "\033[1;36m", "BoldCyan"},
		{"StackOverflow development", "stackoverflow.com", "\033[1;36m", "BoldCyan"},
		{"Ads domain", "ads.example.com", "\033[1;31m", "BoldRed"},
		{"Tracking domain", "tracking.analytics.com", "\033[1;31m", "BoldRed"},
		{"Doubleclick ads", "doubleclick.net", "\033[1;31m", "BoldRed"},
		{"Telemetry domain", "telemetry.service.com", "\033[1;31m", "BoldRed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightDomain(tt.domain)
			if !strings.Contains(result, tt.domain) {
				t.Errorf("HighlightDomain(%q) = %q, should contain domain", tt.domain, result)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("HighlightDomain(%q) = %q, should contain %s color %q",
					tt.domain, result, tt.desc, tt.color)
			}
		})
	}
}

// Test domain highlighting - no color cases
func TestHighlightDomainNoColor(t *testing.T) {
	EnableTestMode()

	regularDomains := []string{
		"example.com",
		"netflix.com",
		"spotify.com",
		"amazon.com",
		"facebook.com",
		"twitter.com",
	}

	for _, domain := range regularDomains {
		t.Run("No color for "+domain, func(t *testing.T) {
			result := HighlightDomain(domain)
			if result != domain {
				t.Errorf("HighlightDomain(%q) = %q, should return unchanged domain",
					domain, result)
			}
		})
	}
}

// Test section headers
func TestSectionHeader(t *testing.T) {
	EnableTestMode()

	title := "TEST SECTION"
	result := SectionHeader(title)

	// Should contain the title
	if !strings.Contains(result, title) {
		t.Errorf("SectionHeader(%q) should contain title", title)
	}

	// Should contain color codes when colors are enabled
	if !strings.Contains(result, "\033[1;36m") { // BoldCyan
		t.Errorf("SectionHeader(%q) should contain BoldCyan color code", title)
	}

	// Should contain border characters
	if !strings.Contains(result, "=") {
		t.Errorf("SectionHeader(%q) should contain border characters", title)
	}
}

// Test subsection headers
func TestSubSectionHeader(t *testing.T) {
	EnableTestMode()

	title := "Subsection Test"
	result := SubSectionHeader(title)

	// Should contain the title
	if !strings.Contains(result, title) {
		t.Errorf("SubSectionHeader(%q) should contain title", title)
	}

	// Should contain color codes when colors are enabled
	if !strings.Contains(result, "\033[1;36m") { // BoldCyan
		t.Errorf("SubSectionHeader(%q) should contain BoldCyan color code", title)
	}

	// Should contain border characters
	if !strings.Contains(result, "-") {
		t.Errorf("SubSectionHeader(%q) should contain border characters", title)
	}
}

// Test processing indicator
func TestProcessingIndicator(t *testing.T) {
	EnableTestMode()
	EnableEmojis()

	message := "Processing test data"
	result := ProcessingIndicator(message)

	// Should contain the message
	if !strings.Contains(result, message) {
		t.Errorf("ProcessingIndicator(%q) should contain message", message)
	}

	// Should contain emoji when enabled
	if !strings.Contains(result, "🔄") {
		t.Errorf("ProcessingIndicator(%q) should contain emoji", message)
	}

	// Should contain color code
	if !strings.Contains(result, "\033[1;36m") { // BoldCyan (Info)
		t.Errorf("ProcessingIndicator(%q) should contain BoldCyan color code", message)
	}
}

// Test color stripping functions
func TestStripColorCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No colors", "plain text", "plain text"},
		{"Red text", "\033[31mred text\033[0m", "red text"},
		{"Bold green", "\033[1;32mbold green\033[0m", "bold green"},
		{"Multiple colors", "\033[31mred\033[0m and \033[32mgreen\033[0m", "red and green"},
		{"Complex formatting", "\033[1;36mBold Cyan\033[0m with \033[90mgray\033[0m", "Bold Cyan with gray"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripColorCodes(tt.input)
			if result != tt.expected {
				t.Errorf("stripColorCodes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test display length calculation
func TestGetDisplayLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Plain text", "hello world", 11},
		{"Red text", "\033[31mhello\033[0m", 5},
		{"Bold text", "\033[1mworld\033[0m", 5},
		{"Complex colors", "\033[1;32mGreen\033[0m \033[31mRed\033[0m", 9}, // "Green Red"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDisplayLength(tt.input)
			if result != tt.expected {
				t.Errorf("getDisplayLength(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// Test table column formatting
func TestFormatTableColumn(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		width  int
		minLen int
	}{
		{"Plain text", "hello", 10, 10},
		{"Colored text", "\033[31mhello\033[0m", 10, 10},
		{"Long text", "this is a very long text", 10, 24}, // Should not truncate
		{"Short colored", "\033[1;32mOK\033[0m", 15, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTableColumn(tt.text, tt.width)
			if len(result) < tt.minLen {
				t.Errorf("formatTableColumn(%q, %d) = %q (len=%d), should be at least %d chars",
					tt.text, tt.width, result, len(result), tt.minLen)
			}
			// Should contain original text
			displayOnly := stripColorCodes(result)
			originalDisplay := stripColorCodes(tt.text)
			if !strings.Contains(displayOnly, originalDisplay) {
				t.Errorf("formatTableColumn(%q, %d) should contain original text", tt.text, tt.width)
			}
		})
	}
}

// Test color configuration
func TestColorConfiguration(t *testing.T) {
	// Test enable/disable colors
	EnableTestMode()
	if colorConfig.ForceDisabled {
		t.Error("EnableTestMode() should set ForceDisabled to false")
	}
	if !colorConfig.Enabled {
		t.Error("EnableTestMode() should set Enabled to true")
	}

	DisableColors()
	if !colorConfig.ForceDisabled {
		t.Error("DisableColors() should set ForceDisabled to true")
	}

	// Test enable/disable emojis
	EnableEmojis()
	if !colorConfig.UseEmoji {
		t.Error("EnableEmojis() should set UseEmoji to true")
	}

	DisableEmojis()
	if colorConfig.UseEmoji {
		t.Error("DisableEmojis() should set UseEmoji to false")
	}

	// Reset to defaults for other tests
	EnableTestMode()
	EnableEmojis()
}

// Test color detection with disabled colors
func TestColorDetectionDisabled(t *testing.T) {
	DisableColors()

	// All color functions should return plain text
	result := Red("test")
	if result != "test" {
		t.Errorf("Red() with disabled colors should return %q, got %q", "test", result)
	}

	result = BoldGreen("test")
	if result != "test" {
		t.Errorf("BoldGreen() with disabled colors should return %q, got %q", "test", result)
	}

	// Re-enable for other tests
	EnableTestMode()
}

// Test Windows color detection (simulated)
func TestWindowsColorDetection(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Save original env
		origWT := os.Getenv("WT_SESSION")
		origTerm := os.Getenv("TERM_PROGRAM")

		// Test without Windows Terminal
		os.Unsetenv("WT_SESSION")
		os.Unsetenv("TERM_PROGRAM")

		result := colorEnabled()
		if result {
			t.Error("colorEnabled() should return false on Windows without WT_SESSION or TERM_PROGRAM")
		}

		// Test with Windows Terminal
		os.Setenv("WT_SESSION", "test")
		result = colorEnabled()
		if !result && !colorConfig.ForceDisabled {
			t.Error("colorEnabled() should return true on Windows with WT_SESSION")
		}

		// Restore original env
		if origWT != "" {
			os.Setenv("WT_SESSION", origWT)
		}
		if origTerm != "" {
			os.Setenv("TERM_PROGRAM", origTerm)
		}
	}
}

// Benchmark color functions
func BenchmarkColorFunctions(b *testing.B) {
	EnableTestMode()
	text := "benchmark test text"

	b.Run("Red", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Red(text)
		}
	})

	b.Run("BoldGreen", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BoldGreen(text)
		}
	})

	b.Run("HighlightDomain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			HighlightDomain("google.com")
		}
	})

	b.Run("ColoredQueryCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ColoredQueryCount(5000)
		}
	})
}

// Benchmark table formatting
func BenchmarkTableFormatting(b *testing.B) {
	text := "\033[31mColored Text\033[0m"

	b.Run("formatTableColumn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatTableColumn(text, 20)
		}
	})

	b.Run("stripColorCodes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stripColorCodes(text)
		}
	})

	b.Run("getDisplayLength", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getDisplayLength(text)
		}
	})
}
