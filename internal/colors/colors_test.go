package colors

import (
	"strings"
	"testing"
)

// TestColorRed tests the red color functionality
func TestColorRed(t *testing.T) {
	EnableTestMode() // Enable test mode for reliable testing
	defer DisableTestMode()

	text := "test"
	result := Red(text)
	if !strings.Contains(result, text) {
		t.Errorf("Red() should contain original text, got %s", result)
	}
}

// TestColorGreen tests the green color functionality
func TestColorGreen(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	text := "test"
	result := Green(text)
	if !strings.Contains(result, text) {
		t.Errorf("Green() should contain original text, got %s", result)
	}
}

// TestColorYellow tests the yellow color functionality
func TestColorYellow(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	text := "test"
	result := Yellow(text)
	if !strings.Contains(result, text) {
		t.Errorf("Yellow() should contain original text, got %s", result)
	}
}

// TestColorBlue tests the blue color functionality
func TestColorBlue(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	text := "test"
	result := Blue(text)
	if !strings.Contains(result, text) {
		t.Errorf("Blue() should contain original text, got %s", result)
	}
}

// TestHighlightDomain tests domain highlighting
func TestHighlightDomain(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	domain := "example.com"
	result := HighlightDomain(domain)
	if !strings.Contains(result, domain) {
		t.Errorf("HighlightDomain() should contain original domain, got %s", result)
	}
}

// TestStatusOnline tests online status formatting
func TestStatusOnline(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	result := OnlineStatus(true, "active")
	if result == "" {
		t.Error("OnlineStatus(true) should return non-empty string")
	}
}

// TestStatusOffline tests offline status formatting
func TestStatusOffline(t *testing.T) {
	EnableTestMode()
	defer DisableTestMode()

	result := OnlineStatus(false, "inactive")
	if result == "" {
		t.Error("OnlineStatus(false) should return non-empty string")
	}
}

// TestStripColors tests color stripping functionality
func TestStripColors(t *testing.T) {
	coloredText := "\033[31mred text\033[0m"
	result := stripColorCodes(coloredText)
	expected := "red text"
	if result != expected {
		t.Errorf("stripColorCodes() = %s, want %s", result, expected)
	}
}

// TestGetDisplayLength tests display length calculation
func TestGetDisplayLength(t *testing.T) {
	text := "hello"
	result := GetDisplayLength(text)
	if result != 5 {
		t.Errorf("GetDisplayLength() = %d, want 5", result)
	}
}

// TestFormatTableColumn tests table column formatting
func TestFormatTableColumn(t *testing.T) {
	text := "test"
	width := 10
	result := FormatTableColumn(text, width)
	if GetDisplayLength(result) < width {
		t.Errorf("FormatTableColumn() should pad to width %d", width)
	}
}

// TestColoredColored tests colored functions exist
func TestColoredQueryCount(t *testing.T) {
	result := ColoredQueryCount(100)
	if result == "" {
		t.Error("ColoredQueryCount() should return non-empty string")
	}
}

// TestColoredPercentage tests percentage coloring
func TestColoredPercentage(t *testing.T) {
	result := ColoredPercentage(50.0)
	if result == "" {
		t.Error("ColoredPercentage() should return non-empty string")
	}
}

// BenchmarkColorRed benchmarks red color performance
func BenchmarkColorRed(b *testing.B) {
	text := "benchmark text"
	for i := 0; i < b.N; i++ {
		Red(text)
	}
}

// BenchmarkColorGreen benchmarks green color performance
func BenchmarkColorGreen(b *testing.B) {
	text := "benchmark text"
	for i := 0; i < b.N; i++ {
		Green(text)
	}
}

// BenchmarkColorHighlightDomain benchmarks domain highlighting performance
func BenchmarkColorHighlightDomain(b *testing.B) {
	domain := "example.com"
	for i := 0; i < b.N; i++ {
		HighlightDomain(domain)
	}
}
