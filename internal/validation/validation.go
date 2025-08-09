package validation

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationResult holds the results of configuration validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
	Warnings []ValidationWarning
}

// ValidationWarning represents a validation warning with context  
type ValidationWarning struct {
	Field   string
	Value   interface{}
	Message string
}

// Validator provides configuration validation functionality
type Validator struct {
	logger *logger.Logger
}

// NewValidator creates a new configuration validator
func NewValidator(log *logger.Logger) *Validator {
	if log == nil {
		log = logger.Component("validator")
	}
	return &Validator{
		logger: log,
	}
}

// ValidateConfig performs comprehensive validation of the entire configuration
func (v *Validator) ValidateConfig(config *types.Config) *ValidationResult {
	v.logger.Info("Starting comprehensive configuration validation")
	
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate each configuration section
	v.validatePiholeConfig(&config.Pihole, result)
	v.validateOutputConfig(&config.Output, result)
	v.validateExclusionConfig(&config.Exclusions, result)
	v.validateLoggingConfig(&config.Logging, result)
	v.validateGlobalFlags(config, result)

	// Update overall validity
	result.Valid = len(result.Errors) == 0

	// Log validation summary
	v.logValidationSummary(result)

	return result
}

// validatePiholeConfig validates Pi-hole specific configuration
func (v *Validator) validatePiholeConfig(config *types.PiholeConfig, result *ValidationResult) {
	v.logger.Debug("Validating Pi-hole configuration")

	// Validate host
	if config.Host == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "pihole.host",
			Value:   config.Host,
			Message: "host cannot be empty",
		})
	} else {
		// Validate host format (IP or hostname)
		if !v.isValidHost(config.Host) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "pihole.host",
				Value:   config.Host,
				Message: "host must be a valid IP address or hostname",
			})
		}
	}

	// Validate port
	if config.Port <= 0 || config.Port > 65535 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "pihole.port",
			Value:   config.Port,
			Message: "port must be between 1 and 65535",
		})
	}

	// Validate API timeout
	if config.APITimeout <= 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "pihole.api_timeout",
			Value:   config.APITimeout,
			Message: "API timeout should be positive, using default value",
		})
	} else if config.APITimeout > 300 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "pihole.api_timeout",
			Value:   config.APITimeout,
			Message: "API timeout is very high, may cause performance issues",
		})
	}

	// Validate API configuration
	if config.APIEnabled {
		if config.APIPassword == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "pihole.api_password",
				Value:   "empty",
				Message: "API password is empty, some features may not work",
			})
		}
	}

	// Log Pi-hole config validation result
	errorCount := 0
	warningCount := 0
	for _, err := range result.Errors {
		if strings.HasPrefix(err.Field, "pihole.") {
			errorCount++
		}
	}
	for _, warn := range result.Warnings {
		if strings.HasPrefix(warn.Field, "pihole.") {
			warningCount++
		}
	}

	v.logger.InfoFields("Pi-hole configuration validation completed", map[string]any{
		"errors":   errorCount,
		"warnings": warningCount,
		"host":     config.Host,
		"port":     config.Port,
		"api_enabled": config.APIEnabled,
	})
}

// validateOutputConfig validates output configuration
func (v *Validator) validateOutputConfig(config *types.OutputConfig, result *ValidationResult) {
	v.logger.Debug("Validating output configuration")

	// Validate MaxClients
	if config.MaxClients <= 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "output.max_clients",
			Value:   config.MaxClients,
			Message: "max_clients must be positive",
		})
	} else if config.MaxClients > 1000 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "output.max_clients",
			Value:   config.MaxClients,
			Message: "max_clients is very high, may impact performance",
		})
	}

	// Validate MaxDomains
	if config.MaxDomains <= 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "output.max_domains_display",
			Value:   config.MaxDomains,
			Message: "max_domains_display must be positive",
		})
	}

	// Validate format if specified
	if config.Format != "" {
		validFormats := []string{"table", "json", "csv", "text"}
		if !v.contains(validFormats, config.Format) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "output.format",
				Value:   config.Format,
				Message: fmt.Sprintf("format must be one of: %s", strings.Join(validFormats, ", ")),
			})
		}
	}

	// Validate report directory
	if config.SaveReports && config.ReportDir != "" {
		if !v.isValidDirectory(config.ReportDir) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "output.report_dir",
				Value:   config.ReportDir,
				Message: "report directory does not exist or is not writable",
			})
		}
	}

	v.logger.InfoFields("Output configuration validation completed", map[string]any{
		"max_clients": config.MaxClients,
		"max_domains": config.MaxDomains,
		"save_reports": config.SaveReports,
		"report_dir": config.ReportDir,
	})
}

// validateExclusionConfig validates exclusion configuration
func (v *Validator) validateExclusionConfig(config *types.ExclusionConfig, result *ValidationResult) {
	v.logger.Debug("Validating exclusion configuration")

	// Validate network exclusions
	for i, network := range config.ExcludeNetworks {
		if !v.isValidCIDR(network) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("exclusions.exclude_networks[%d]", i),
				Value:   network,
				Message: "invalid CIDR notation",
			})
		}
	}

	// Validate IP exclusions
	for i, ip := range config.ExcludeIPs {
		if !v.isValidIP(ip) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("exclusions.exclude_ips[%d]", i),
				Value:   ip,
				Message: "invalid IP address format",
			})
		}
	}

	// Validate host exclusions (basic format check)
	for i, host := range config.ExcludeHosts {
		if host == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("exclusions.exclude_hosts[%d]", i),
				Value:   host,
				Message: "hostname cannot be empty",
			})
		}
	}

	v.logger.InfoFields("Exclusion configuration validation completed", map[string]any{
		"exclude_networks_count": len(config.ExcludeNetworks),
		"exclude_ips_count":     len(config.ExcludeIPs),
		"exclude_hosts_count":   len(config.ExcludeHosts),
	})
}

// validateLoggingConfig validates logging configuration
func (v *Validator) validateLoggingConfig(config *types.LoggingConfig, result *ValidationResult) {
	v.logger.Debug("Validating logging configuration")

	// Validate log level
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if !v.contains(validLevels, strings.ToUpper(config.Level)) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "logging.level",
			Value:   config.Level,
			Message: fmt.Sprintf("log level must be one of: %s", strings.Join(validLevels, ", ")),
		})
	}

	// Validate output file if specified
	if config.OutputFile != "" {
		dir := filepath.Dir(config.OutputFile)
		if !v.isValidDirectory(dir) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "logging.output_file",
				Value:   config.OutputFile,
				Message: "log file directory does not exist or is not writable",
			})
		}
	}

	v.logger.InfoFields("Logging configuration validation completed", map[string]any{
		"level":       config.Level,
		"output_file": config.OutputFile,
		"colors":     config.EnableColors,
		"emojis":     config.EnableEmojis,
	})
}

// validateGlobalFlags validates global configuration flags
func (v *Validator) validateGlobalFlags(config *types.Config, result *ValidationResult) {
	v.logger.Debug("Validating global configuration flags")

	// Log global flags
	v.logger.InfoFields("Global configuration flags validated", map[string]any{
		"online_only": config.OnlineOnly,
		"no_exclude":  config.NoExclude,
		"test_mode":   config.TestMode,
		"quiet":       config.Quiet,
	})
}

// logValidationSummary logs a summary of the validation results
func (v *Validator) logValidationSummary(result *ValidationResult) {
	if result.Valid {
		v.logger.Success("Configuration validation completed successfully")
	} else {
		v.logger.Error("Configuration validation failed with %d errors", len(result.Errors))
	}

	if len(result.Warnings) > 0 {
		v.logger.Warn("Configuration validation completed with %d warnings", len(result.Warnings))
	}

	// Log detailed errors
	for _, err := range result.Errors {
		v.logger.ErrorFields("Validation error", map[string]any{
			"field":   err.Field,
			"value":   err.Value,
			"message": err.Message,
		})
	}

	// Log detailed warnings
	for _, warn := range result.Warnings {
		v.logger.InfoFields("Validation warning", map[string]any{
			"field":   warn.Field,
			"value":   warn.Value,
			"message": warn.Message,
		})
	}

	// Summary log
	v.logger.InfoFields("Validation summary", map[string]any{
		"valid":    result.Valid,
		"errors":   len(result.Errors),
		"warnings": len(result.Warnings),
	})
}

// Helper validation methods

// isValidHost checks if a string is a valid hostname or IP address
func (v *Validator) isValidHost(host string) bool {
	// Try to parse as IP address first
	if net.ParseIP(host) != nil {
		return true
	}

	// Basic hostname validation
	if len(host) == 0 || len(host) > 253 {
		return false
	}

	// Check for invalid patterns
	if strings.Contains(host, "..") || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}

	// Check if it looks like an invalid IP address (e.g., 256.256.256.256)
	parts := strings.Split(host, ".")
	allNumeric := true
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		isNumeric := true
		for _, char := range part {
			if char < '0' || char > '9' {
				isNumeric = false
				break
			}
		}
		if !isNumeric {
			allNumeric = false
		}
	}
	
	// If all parts are numeric but it's not a valid IP, it's invalid
	if allNumeric && len(parts) == 4 {
		return false // Already checked by net.ParseIP and failed
	}

	// Simple hostname validation - allow alphanumeric, dots, and hyphens
	for _, char := range host {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// isValidIP checks if a string is a valid IP address
func (v *Validator) isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidCIDR checks if a string is valid CIDR notation
func (v *Validator) isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// isValidDirectory checks if a directory exists and is writable
func (v *Validator) isValidDirectory(dir string) bool {
	// Handle relative paths
	if !filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			return false
		}
		dir = filepath.Join(wd, dir)
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		// Try to create the directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false
		}
		return true
	}

	// Check if it's a directory
	if !info.IsDir() {
		return false
	}

	// Check if it's writable by trying to create a temp file
	tempFile := filepath.Join(dir, ".write_test")
	if file, err := os.Create(tempFile); err == nil {
		file.Close()
		os.Remove(tempFile)
		return true
	}

	return false
}

// contains checks if a slice contains a specific string
func (v *Validator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetDefaultValidationConfig returns a default validation configuration
func GetDefaultValidationConfig() *types.Config {
	return &types.Config{
		Pihole: types.PiholeConfig{
			Host:        "192.168.1.100",
			Port:        80,
			APIEnabled:  true,
			APIPassword: "",
			UseHTTPS:    false,
			APITimeout:  30,
		},
		Output: types.OutputConfig{
			Colors:        true,
			Emojis:        true,
			Verbose:       false,
			Format:        "table",
			MaxClients:    20,
			MaxDomains:    10,
			SaveReports:   true,
			ReportDir:     "./reports",
			VerboseOutput: false,
		},
		Exclusions: types.ExclusionConfig{
			ExcludeNetworks: []string{"172.16.0.0/12", "127.0.0.0/8"},
			ExcludeIPs:      []string{},
			ExcludeHosts:    []string{"pi.hole"},
		},
		Logging: types.LoggingConfig{
			Level:         "INFO",
			EnableColors:  true,
			EnableEmojis:  true,
			OutputFile:    "",
			ShowTimestamp: true,
			ShowCaller:    false,
		},
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,
		Quiet:      false,
	}
}

// ValidateAndWarn validates configuration and logs warnings/errors
func ValidateAndWarn(config *types.Config, log *logger.Logger) bool {
	validator := NewValidator(log)
	result := validator.ValidateConfig(config)
	
	return result.Valid
}

// ApplyDefaults applies default values to missing or invalid configuration fields
func (v *Validator) ApplyDefaults(config *types.Config) {
	v.logger.Info("Applying default values to configuration")
	
	defaults := GetDefaultValidationConfig()
	
	// Apply Pi-hole defaults
	if config.Pihole.Host == "" {
		config.Pihole.Host = defaults.Pihole.Host
		v.logger.Warn("Applied default Pi-hole host: %s", defaults.Pihole.Host)
	}
	if config.Pihole.Port <= 0 || config.Pihole.Port > 65535 {
		config.Pihole.Port = defaults.Pihole.Port
		v.logger.Warn("Applied default Pi-hole port: %d", defaults.Pihole.Port)
	}
	if config.Pihole.APITimeout <= 0 {
		config.Pihole.APITimeout = defaults.Pihole.APITimeout
		v.logger.Warn("Applied default API timeout: %d", defaults.Pihole.APITimeout)
	}
	
	// Apply output defaults
	if config.Output.MaxClients <= 0 {
		config.Output.MaxClients = defaults.Output.MaxClients
		v.logger.Warn("Applied default max clients: %d", defaults.Output.MaxClients)
	}
	if config.Output.MaxDomains <= 0 {
		config.Output.MaxDomains = defaults.Output.MaxDomains
		v.logger.Warn("Applied default max domains: %d", defaults.Output.MaxDomains)
	}
	if config.Output.ReportDir == "" && config.Output.SaveReports {
		config.Output.ReportDir = defaults.Output.ReportDir
		v.logger.Warn("Applied default report directory: %s", defaults.Output.ReportDir)
	}
	
	// Apply logging defaults
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if !v.contains(validLevels, strings.ToUpper(config.Logging.Level)) {
		config.Logging.Level = defaults.Logging.Level
		v.logger.Warn("Applied default log level: %s", defaults.Logging.Level)
	}
	
	v.logger.Success("Default values applied to configuration")
}