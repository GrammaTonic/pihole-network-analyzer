package main

import (
	"strings"
	"testing"

	"pihole-network-analyzer/internal/colors"
)

func TestBasicColors(t *testing.T) {
	colors.EnableTestMode()
	defer colors.DisableColors()

	result := colors.Red("test")
	if !strings.Contains(result, "test") {
		t.Errorf("Red() should contain input text")
	}
}

func TestStatusColors(t *testing.T) {
	colors.EnableTestMode()
	defer colors.DisableColors()

	result := colors.Success("OK")
	if !strings.Contains(result, "OK") {
		t.Errorf("Success() should contain input text")
	}
}
