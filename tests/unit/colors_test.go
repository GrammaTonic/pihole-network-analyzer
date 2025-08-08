package colors_test

import (
	"os"
	"strings"
	"testing"

	"pihole-analyzer/internal/colors"
)

func TestBasicColors(t *testing.T) {
	// Ensure terminal environment for color detection
	os.Setenv("TERM", "xterm-256color")
	defer os.Unsetenv("TERM")

	result := colors.Red("test")
	if !strings.Contains(result, "test") {
		t.Errorf("Red() should contain input text")
	}
}

func TestStatusColors(t *testing.T) {
	// Ensure terminal environment for color detection
	os.Setenv("TERM", "xterm-256color")
	defer os.Unsetenv("TERM")

	result := colors.Success("OK")
	if !strings.Contains(result, "OK") {
		t.Errorf("Success() should contain input text")
	}
}
