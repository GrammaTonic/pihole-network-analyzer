package colors

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Color codes for terminal output
const (
	// Reset
	ColorReset = "\033[0m"

	// Regular colors
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// Bold colors
	ColorBoldRed    = "\033[1;31m"
	ColorBoldGreen  = "\033[1;32m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
	ColorBoldPurple = "\033[1;35m"
	ColorBoldCyan   = "\033[1;36m"
	ColorBoldWhite  = "\033[1;37m"

	// Background colors
	ColorBgRed    = "\033[41m"
	ColorBgGreen  = "\033[42m"
	ColorBgYellow = "\033[43m"
	ColorBgBlue   = "\033[44m"

	// Special formatting
	ColorBold      = "\033[1m"
	ColorUnderline = "\033[4m"
	ColorReverse   = "\033[7m"
)

// ColorConfig holds color preferences
type ColorConfig struct {
	Enabled       bool
	ForceDisabled bool
	UseEmoji      bool
	TestMode      bool // For unit testing - bypasses terminal detection
}

// Global color configuration
var colorConfig = ColorConfig{
	Enabled:  true,
	UseEmoji: true,
}

// colorEnabled checks if colors should be used
func colorEnabled() bool {
	if colorConfig.ForceDisabled {
		return false
	}

	if !colorConfig.Enabled {
		return false
	}

	// In test mode, bypass terminal and OS checks
	if colorConfig.TestMode {
		return true
	}

	// Disable colors on Windows unless explicitly enabled
	if runtime.GOOS == "windows" {
		// Check for Windows Terminal or other color-capable terminals
		if os.Getenv("WT_SESSION") == "" && os.Getenv("TERM_PROGRAM") == "" {
			return false
		}
	}

	// Check if output is being piped or redirected
	if !isTerminal() {
		return false
	}

	return true
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Colorize applies color to text if colors are enabled
func Colorize(color, text string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + ColorReset
}

// Color helper functions
func Red(text string) string    { return Colorize(ColorRed, text) }
func Green(text string) string  { return Colorize(ColorGreen, text) }
func Yellow(text string) string { return Colorize(ColorYellow, text) }
func Blue(text string) string   { return Colorize(ColorBlue, text) }
func Purple(text string) string { return Colorize(ColorPurple, text) }
func Cyan(text string) string   { return Colorize(ColorCyan, text) }
func Gray(text string) string   { return Colorize(ColorGray, text) }

// Bold color helper functions
func BoldRed(text string) string    { return Colorize(ColorBoldRed, text) }
func BoldGreen(text string) string  { return Colorize(ColorBoldGreen, text) }
func BoldYellow(text string) string { return Colorize(ColorBoldYellow, text) }
func BoldBlue(text string) string   { return Colorize(ColorBoldBlue, text) }
func BoldPurple(text string) string { return Colorize(ColorBoldPurple, text) }
func BoldCyan(text string) string   { return Colorize(ColorBoldCyan, text) }
func BoldWhite(text string) string  { return Colorize(ColorBoldWhite, text) }

// Formatting helper functions
func Bold(text string) string      { return Colorize(ColorBold, text) }
func Underline(text string) string { return Colorize(ColorUnderline, text) }

// Status-specific color functions
func Success(text string) string { return BoldGreen(text) }
func Warning(text string) string { return BoldYellow(text) }
func Error(text string) string   { return BoldRed(text) }
func Info(text string) string    { return BoldCyan(text) }
func Header(text string) string  { return BoldWhite(text) }

// Online status with colors and emojis
func OnlineStatus(isOnline bool, arpStatus string) string {
	if arpStatus == "unknown" {
		return Gray("❓ Unknown")
	}

	if isOnline {
		if colorConfig.UseEmoji {
			return Success("✅ Online")
		}
		return Success("✓ Online")
	} else {
		if colorConfig.UseEmoji {
			return Error("❌ Offline")
		}
		return Error("✗ Offline")
	}
}

// Percentage coloring based on value
func ColoredPercentage(value float64) string {
	text := fmt.Sprintf("%.2f%%", value)

	switch {
	case value >= 30.0:
		return BoldRed(text) // High usage
	case value >= 15.0:
		return BoldYellow(text) // Medium usage
	case value >= 5.0:
		return Yellow(text) // Low-medium usage
	default:
		return Gray(text) // Very low usage
	}
}

// Query count coloring
func ColoredQueryCount(count int) string {
	text := fmt.Sprintf("%d", count)

	switch {
	case count >= 10000:
		return BoldRed(text) // Very high
	case count >= 5000:
		return BoldYellow(text) // High
	case count >= 1000:
		return Yellow(text) // Medium
	case count >= 100:
		return Cyan(text) // Low
	default:
		return Gray(text) // Very low
	}
}

// Domain count coloring
func ColoredDomainCount(count int) string {
	text := fmt.Sprintf("%d", count)

	switch {
	case count >= 500:
		return BoldPurple(text) // Very diverse
	case count >= 200:
		return Purple(text) // Diverse
	case count >= 50:
		return Blue(text) // Medium diversity
	default:
		return Gray(text) // Low diversity
	}
}

// Section headers with decorative borders
func SectionHeader(title string) string {
	if !colorEnabled() {
		border := strings.Repeat("=", 80)
		return fmt.Sprintf("\n%s\n%s\n%s", border, title, border)
	}

	border := BoldCyan(strings.Repeat("=", 80))
	titleColored := BoldWhite(title)

	return fmt.Sprintf("\n%s\n%s\n%s", border, titleColored, border)
}

// Sub-section headers
func SubSectionHeader(title string) string {
	if !colorEnabled() {
		border := strings.Repeat("-", 107) // Match table width: 16+18+18+10+10+12+8+8+7=107
		return fmt.Sprintf("%s\n%s\n%s", title, border, "")
	}

	border := Cyan(strings.Repeat("-", 107)) // Match table width: 16+18+18+10+10+12+8+8+7=107
	titleColored := BoldCyan(title)

	return fmt.Sprintf("%s\n%s", titleColored, border)
}

// Progress indicators
func ProgressDot() string {
	if colorConfig.UseEmoji {
		return Green("●")
	}
	return Green("✓")
}

func ProcessingIndicator(message string) string {
	if colorConfig.UseEmoji {
		return Info("🔄 " + message)
	}
	return Info("⟳ " + message)
}

// IP address highlighting
func HighlightIP(ip string) string {
	// Highlight private network IPs differently
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.") {
		return BoldBlue(ip)
	}
	return BoldYellow(ip) // Public IPs
}

// Domain highlighting based on type
func HighlightDomain(domain string) string {
	// Highlight different types of domains
	if strings.Contains(domain, "google") || strings.Contains(domain, "microsoft") {
		return BoldGreen(domain) // Major services
	} else if strings.Contains(domain, "ads") || strings.Contains(domain, "tracking") ||
		strings.Contains(domain, "doubleclick") || strings.Contains(domain, "telemetry") {
		return BoldRed(domain) // Ads/tracking
	} else if strings.Contains(domain, "github") || strings.Contains(domain, "stackoverflow") {
		return BoldCyan(domain) // Development
	}
	return domain // Default - no color
}

// Configuration functions
func EnableColors() {
	colorConfig.Enabled = true
	colorConfig.ForceDisabled = false
}

func DisableColors() {
	colorConfig.ForceDisabled = true
}

func EnableEmojis() {
	colorConfig.UseEmoji = true
}

func DisableEmojis() {
	colorConfig.UseEmoji = false
}

func IsColorEnabled() bool {
	return colorEnabled()
}

// EnableTestMode enables colors for unit testing (bypasses terminal detection)
func EnableTestMode() {
	colorConfig.TestMode = true
	colorConfig.Enabled = true
	colorConfig.ForceDisabled = false
}

// DisableTestMode disables test mode
func DisableTestMode() {
	colorConfig.TestMode = false
}

// stripColorCodes removes ANSI color codes from a string to calculate its actual display length
func stripColorCodes(text string) string {
	// Simple regex would be better, but this avoids additional dependencies
	result := text
	colorCodes := []string{
		ColorReset, ColorRed, ColorGreen, ColorYellow, ColorBlue, ColorPurple, ColorCyan, ColorWhite, ColorGray,
		ColorBoldRed, ColorBoldGreen, ColorBoldYellow, ColorBoldBlue, ColorBoldPurple, ColorBoldCyan, ColorBoldWhite,
		ColorBgRed, ColorBgGreen, ColorBgYellow, ColorBgBlue, ColorBold, ColorUnderline, ColorReverse,
	}

	for _, code := range colorCodes {
		result = strings.ReplaceAll(result, code, "")
	}

	return result
}

// getDisplayLength calculates the actual display length of a string (excluding color codes)
func getDisplayLength(text string) int {
	return len(stripColorCodes(text))
}

// padColoredText pads a colored string to a specific width, accounting for color codes
func padColoredText(text string, width int, leftAlign bool) string {
	displayLen := getDisplayLength(text)

	if displayLen >= width {
		return text
	}

	padding := strings.Repeat(" ", width-displayLen)

	if leftAlign {
		return text + padding
	} else {
		return padding + text
	}
}

// formatTableColumn formats a colored text for table display with proper padding
func formatTableColumn(text string, width int) string {
	return padColoredText(text, width, true) // left-align by default
}

// formatTableColumnRight formats a colored text for table display with right alignment
func formatTableColumnRight(text string, width int) string {
	return padColoredText(text, width, false)
}
