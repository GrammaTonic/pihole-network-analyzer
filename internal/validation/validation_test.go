package validation

import (
	"testing"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

func TestNewValidator(t *testing.T) {
	// Test with nil logger
	validator := NewValidator(nil)
	if validator == nil {
		t.Error("NewValidator should not return nil")
	}
	if validator.logger == nil {
		t.Error("NewValidator should create a logger when nil is passed")
	}

	// Test with provided logger
	log := logger.New(&logger.Config{Component: "test"})
	validator = NewValidator(log)
	if validator.logger != log {
		t.Error("NewValidator should use the provided logger")
	}
}

func TestValidateConfig_ValidConfiguration(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	result := validator.ValidateConfig(config)

	if !result.Valid {
		t.Errorf("Valid configuration should pass validation, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s", err.Error())
		}
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestValidateConfig_InvalidPiholeHost(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Pihole.Host = ""

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with empty host should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "pihole.host" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for pihole.host field")
	}
}

func TestValidateConfig_InvalidPiholePort(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Pihole.Port = -1

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with invalid port should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "pihole.port" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for pihole.port field")
	}
}

func TestValidateConfig_InvalidOutputMaxClients(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Output.MaxClients = 0

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with zero max clients should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "output.max_clients" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for output.max_clients field")
	}
}

func TestValidateConfig_InvalidFormat(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Output.Format = "invalid_format"

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with invalid format should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "output.format" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for output.format field")
	}
}

func TestValidateConfig_InvalidCIDR(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Exclusions.ExcludeNetworks = []string{"invalid_cidr"}

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with invalid CIDR should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "exclusions.exclude_networks[0]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for exclusions.exclude_networks[0] field")
	}
}

func TestValidateConfig_InvalidIP(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Exclusions.ExcludeIPs = []string{"invalid_ip"}

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with invalid IP should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "exclusions.exclude_ips[0]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for exclusions.exclude_ips[0] field")
	}
}

func TestValidateConfig_InvalidLogLevel(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Logging.Level = "INVALID"

	result := validator.ValidateConfig(config)

	if result.Valid {
		t.Error("Configuration with invalid log level should fail validation")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Field == "logging.level" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for logging.level field")
	}
}

func TestValidateConfig_Warnings(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := GetDefaultValidationConfig()
	config.Pihole.APITimeout = 600  // Very high timeout
	config.Output.MaxClients = 2000 // Very high max clients

	result := validator.ValidateConfig(config)

	// Should still be valid but have warnings
	if !result.Valid {
		t.Error("Configuration with high values should still be valid")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for high timeout and max clients values")
	}
}

func TestIsValidHost(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	validHosts := []string{
		"192.168.1.1",
		"10.0.0.1",
		"example.com",
		"pi.hole",
		"localhost",
	}

	invalidHosts := []string{
		"",
		"256.256.256.256",
		"invalid..host",
	}

	for _, host := range validHosts {
		if !validator.isValidHost(host) {
			t.Errorf("Host %s should be valid", host)
		}
	}

	for _, host := range invalidHosts {
		if validator.isValidHost(host) {
			t.Errorf("Host %s should be invalid", host)
		}
	}
}

func TestIsValidIP(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"127.0.0.1",
		"::1",
		"2001:db8::1",
	}

	invalidIPs := []string{
		"",
		"256.256.256.256",
		"invalid.ip",
		"192.168.1",
	}

	for _, ip := range validIPs {
		if !validator.isValidIP(ip) {
			t.Errorf("IP %s should be valid", ip)
		}
	}

	for _, ip := range invalidIPs {
		if validator.isValidIP(ip) {
			t.Errorf("IP %s should be invalid", ip)
		}
	}
}

func TestIsValidCIDR(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	validCIDRs := []string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"127.0.0.0/8",
		"2001:db8::/32",
	}

	invalidCIDRs := []string{
		"",
		"192.168.1.0",
		"192.168.1.0/33",
		"invalid/24",
		"192.168.1.0/",
	}

	for _, cidr := range validCIDRs {
		if !validator.isValidCIDR(cidr) {
			t.Errorf("CIDR %s should be valid", cidr)
		}
	}

	for _, cidr := range invalidCIDRs {
		if validator.isValidCIDR(cidr) {
			t.Errorf("CIDR %s should be invalid", cidr)
		}
	}
}

func TestContains(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	slice := []string{"apple", "banana", "orange"}

	if !validator.contains(slice, "apple") {
		t.Error("Should find 'apple' in slice")
	}

	if validator.contains(slice, "grape") {
		t.Error("Should not find 'grape' in slice")
	}

	if validator.contains([]string{}, "anything") {
		t.Error("Should not find anything in empty slice")
	}
}

func TestValidateAndWarn(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})

	// Valid config
	config := GetDefaultValidationConfig()
	if !ValidateAndWarn(config, log) {
		t.Error("Valid configuration should return true")
	}

	// Invalid config
	config.Pihole.Host = ""
	if ValidateAndWarn(config, log) {
		t.Error("Invalid configuration should return false")
	}
}

func TestApplyDefaults(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	config := &types.Config{
		Pihole: types.PiholeConfig{
			Host:       "", // Invalid
			Port:       -1, // Invalid
			APITimeout: 0,  // Invalid
		},
		Output: types.OutputConfig{
			MaxClients:  0, // Invalid
			MaxDomains:  0, // Invalid
			SaveReports: true,
			ReportDir:   "", // Invalid for saving reports
		},
		Logging: types.LoggingConfig{
			Level: "INVALID", // Invalid
		},
	}

	validator.ApplyDefaults(config)

	// Check that defaults were applied
	if config.Pihole.Host == "" {
		t.Error("Default host should have been applied")
	}
	if config.Pihole.Port <= 0 {
		t.Error("Default port should have been applied")
	}
	if config.Pihole.APITimeout <= 0 {
		t.Error("Default API timeout should have been applied")
	}
	if config.Output.MaxClients <= 0 {
		t.Error("Default max clients should have been applied")
	}
	if config.Output.MaxDomains <= 0 {
		t.Error("Default max domains should have been applied")
	}
	if config.Output.ReportDir == "" {
		t.Error("Default report directory should have been applied")
	}

	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	found := false
	for _, level := range validLevels {
		if config.Logging.Level == level {
			found = true
			break
		}
	}
	if !found {
		t.Error("Default log level should have been applied")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test.field",
		Value:   "test_value",
		Message: "test message",
	}

	expected := "validation error in field 'test.field': test message (value: test_value)"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// Benchmark tests
func BenchmarkValidateConfig(b *testing.B) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)
	config := GetDefaultValidationConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateConfig(config)
	}
}

func BenchmarkIsValidHost(b *testing.B) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.isValidHost("192.168.1.1")
	}
}

func BenchmarkIsValidCIDR(b *testing.B) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := NewValidator(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.isValidCIDR("192.168.1.0/24")
	}
}
