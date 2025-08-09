# Configuration Validation

The Pi-hole Network Analyzer includes comprehensive configuration validation with structured logging to ensure reliable operation and clear error reporting.

## Overview

The validation system consists of two main packages:
- `internal/validation` - Core validation logic and error handling
- `internal/config` - Configuration management with integrated validation

## Features

### Comprehensive Validation

The system validates all configuration sections:

- **Pi-hole Configuration**
  - Host/IP address format validation
  - Port range validation (1-65535)
  - API timeout validation with performance warnings
  - API password presence warnings

- **Output Configuration**
  - Positive values for client and domain limits
  - Valid output format selection
  - Directory existence and write permissions
  - Performance warnings for very high limits

- **Exclusion Configuration**
  - CIDR notation validation for network exclusions
  - IP address format validation
  - Non-empty hostname validation

- **Logging Configuration**
  - Valid log level selection (DEBUG, INFO, WARN, ERROR)
  - Log file directory validation
  - Output file path verification

### Structured Logging

All validation processes use structured logging with contextual information:

```go
v.logger.InfoFields("Pi-hole configuration validation completed", map[string]any{
    "errors":   errorCount,
    "warnings": warningCount,
    "host":     config.Host,
    "port":     config.Port,
    "api_enabled": config.APIEnabled,
})
```

### Error Recovery

The system automatically applies sensible defaults for invalid configurations:

- Invalid hosts default to `192.168.1.100`
- Invalid ports default to `80`
- Invalid timeouts default to `30` seconds
- Invalid limits default to reasonable values (20 clients, 10 domains)
- Invalid log levels default to `INFO`

## Usage

### Basic Validation

```go
import (
    "pihole-analyzer/internal/logger"
    "pihole-analyzer/internal/validation"
)

// Create logger and validator
log := logger.Component("my-component")
validator := validation.NewValidator(log)

// Validate configuration
result := validator.ValidateConfig(config)

if !result.Valid {
    // Handle validation errors
    for _, err := range result.Errors {
        log.Error("Validation failed: %s", err.Error())
    }
}
```

### Configuration Loading with Validation

```go
import "pihole-analyzer/internal/config"

// Load config with automatic validation and error recovery
config, err := config.LoadConfig("path/to/config.json")
if err != nil {
    return err // Critical error, cannot proceed
}

// Config is now validated and has defaults applied for any issues
```

### Validation with Auto-fix

```go
import (
    "pihole-analyzer/internal/logger"
    "pihole-analyzer/internal/validation"
)

log := logger.Component("config-fixer")
validator := validation.NewValidator(log)

// Apply defaults to fix invalid values
validator.ApplyDefaults(config)

// Re-validate to ensure everything is correct
result := validator.ValidateConfig(config)
```

## Validation Results

The validation system returns detailed results:

```go
type ValidationResult struct {
    Valid    bool                 // Overall validation status
    Errors   []ValidationError    // Critical errors that prevent operation
    Warnings []ValidationWarning  // Non-critical issues with recommendations
}

type ValidationError struct {
    Field   string      // Configuration field that failed
    Value   interface{} // The invalid value
    Message string      // Human-readable error description
}
```

## Error Types

### Critical Errors (Prevent Operation)
- Empty Pi-hole host
- Invalid port numbers
- Zero or negative client/domain limits
- Invalid CIDR notation
- Invalid IP addresses
- Invalid log levels
- Non-existent/non-writable directories

### Warnings (Recommendations)
- Empty API password
- Very high timeout values
- Very high client/domain limits
- Performance considerations

## Configuration Examples

### Valid Configuration
```json
{
  "pihole": {
    "host": "192.168.1.100",
    "port": 80,
    "api_enabled": true,
    "api_password": "your-password",
    "api_timeout": 30
  },
  "output": {
    "max_clients": 20,
    "max_domains": 10,
    "format": "table",
    "save_reports": true,
    "report_dir": "./reports"
  },
  "exclusions": {
    "exclude_networks": ["172.16.0.0/12", "127.0.0.0/8"],
    "exclude_ips": ["192.168.1.1"],
    "exclude_hosts": ["pi.hole"]
  },
  "logging": {
    "level": "INFO",
    "enable_colors": true,
    "enable_emojis": true
  }
}
```

### Invalid Configuration (Will be Auto-fixed)
```json
{
  "pihole": {
    "host": "",          // Error: will default to 192.168.1.100
    "port": -1,          // Error: will default to 80
    "api_timeout": 0     // Warning: will default to 30
  },
  "output": {
    "max_clients": 0,    // Error: will default to 20
    "format": "invalid"  // Error: will default to table
  },
  "logging": {
    "level": "INVALID"   // Error: will default to INFO
  }
}
```

## Testing

The validation system includes comprehensive tests:

### Unit Tests
```bash
# Test validation logic
go test ./internal/validation -v

# Test config integration
go test ./internal/config -v
```

### Integration Tests
```bash
# Test full application workflow
go test ./tests/integration -v
```

### Benchmarks
```bash
# Performance testing
go test -bench=. ./internal/validation
go test -bench=. ./tests/integration
```

## Performance

The validation system is designed for efficiency:
- Validation completes in microseconds for typical configurations
- Structured logging adds minimal overhead
- Memory allocation is optimized for repeated validation calls

## Best Practices

1. **Always validate configuration at startup**
2. **Use structured logging for operational insights**
3. **Handle validation errors gracefully with defaults**
4. **Monitor validation warnings for configuration issues**
5. **Test configuration changes with the validation system**

## Migration Guide

### From Legacy Configuration

The new validation system is fully backward compatible. Existing configurations will:
- Be automatically validated on load
- Have defaults applied for any invalid values
- Generate warnings for deprecated or suboptimal settings
- Continue to work without changes

### Updating Applications

To integrate validation into existing code:

```go
// Before
config, err := config.LoadConfig(path)

// After (no changes needed - validation is automatic)
config, err := config.LoadConfig(path)
```

For custom validation needs:

```go
// Custom validation with specific logger
log := logger.New(&logger.Config{Component: "my-app"})
validator := validation.NewValidator(log)
result := validator.ValidateConfig(config)
```

## Troubleshooting

### Common Issues

1. **"Host cannot be empty"**
   - Ensure Pi-hole host is configured
   - Use IP address or valid hostname

2. **"Port must be between 1 and 65535"**
   - Check port configuration
   - Common ports: 80 (HTTP), 443 (HTTPS)

3. **"Invalid CIDR notation"**
   - Verify network exclusion format
   - Example: `192.168.1.0/24`

4. **"Directory not writable"**
   - Check file permissions
   - Ensure parent directories exist

### Debug Logging

Enable debug logging for detailed validation information:

```go
log := logger.New(&logger.Config{
    Level: logger.LevelDebug,
    Component: "validation",
})
```

This will provide step-by-step validation progress and detailed error analysis.