# Development Guide

This guide covers development setup, coding standards, testing procedures, and contribution guidelines for the Pi-hole Network Analyzer.

## Table of Contents
- [Development Environment](#development-environment)
- [Project Architecture](#project-architecture)
- [Coding Standards](#coding-standards)
- [Testing Framework](#testing-framework)
- [Build System](#build-system)
- [Contributing Guidelines](#contributing-guidelines)
- [Debugging & Troubleshooting](#debugging--troubleshooting)

## Development Environment

### Prerequisites
- **Go 1.23.12+** - Latest Go version for modern features
- **Docker** - Container development and testing
- **Make** - Build automation (40+ targets)
- **Git** - Version control
- **VS Code** (recommended) - IDE with Go extension

### Quick Setup

```bash
# Clone repository
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer

# Setup development environment
make dev-setup

# Fast development build
make fast-build

# Run with test data
./pihole-analyzer-test --test
```

### Development Tools Installation

```bash
# Install Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install entr for file watching (optional)
# macOS
brew install entr

# Linux
sudo apt-get install entr

# Enable file watching builds
make watch
```

## Project Architecture

### Directory Structure
```
/
├── cmd/
│   ├── pihole-analyzer/          # Production binary entry point
│   └── pihole-analyzer-test/     # Test/development binary with mock data
├── internal/                     # Private application packages
│   ├── analyzer/                 # Pi-hole data analysis engine
│   ├── cli/                      # Command-line interface and flag management
│   ├── colors/                   # Terminal colorization with cross-platform support
│   ├── config/                   # Configuration management
│   ├── interfaces/               # Data source abstraction and factory pattern
│   ├── logger/                   # Structured logging with slog integration
│   ├── network/                  # Network analysis & ARP integration
│   ├── pihole/                   # Pi-hole API client implementation
│   ├── reporting/                # Output display & formatted reports
│   └── types/                    # Core data structures
├── docs/                         # Comprehensive documentation
├── scripts/                      # Build automation & cache warming
├── testing/                      # Test utilities and fixtures
├── .github/workflows/           # CI/CD with advanced caching
├── Dockerfile                   # Multi-stage, multi-arch container builds
├── docker-compose*.yml         # Development and production environments
└── Makefile                     # Enhanced build system (40+ targets)
```

### Data Flow Architecture

#### API-Only Data Flow
```
Pi-hole API → Client Authentication → Query Processing → Analysis → Report Generation
[pihole/client.go] [analyzer/analyzer.go] [types/] [analyzer/] [reporting/display.go]
```

#### Key Components
1. **`cmd/pihole-analyzer/main.go`** - Production entry point with API connectivity
2. **`cmd/pihole-analyzer-test/main.go`** - Development/testing entry point with mock data
3. **`internal/pihole/client.go`** - Pi-hole API client with session management
4. **`internal/interfaces/data_source.go`** - Data source abstraction interface
5. **`internal/logger/logger.go`** - Structured logging with slog, colors, and emojis

### Core Data Structures

#### PiholeRecord
```go
type PiholeRecord struct {
    ID        int
    DateTime  string
    Domain    string
    Client    string
    QueryType string
    Status    int
    Timestamp string   // Unix timestamp
    HWAddr    string   // Hardware/MAC address
}
```

#### ClientStats
```go
type ClientStats struct {
    IP            string
    Hostname      string
    QueryCount    int
    Domains       map[string]int
    DomainCount   int
    MACAddress    string
    IsOnline      bool
    LastSeen      string
    TopDomains    []DomainStat
    Status        string
    UniqueQueries int
    TotalQueries  int
    // Additional analysis fields...
}
```

#### Configuration Structure
```go
type Config struct {
    OnlineOnly bool            `json:"online_only"`
    NoExclude  bool            `json:"no_exclude"`
    TestMode   bool            `json:"test_mode"`
    Quiet      bool            `json:"quiet"`
    Pihole     PiholeConfig    `json:"pihole"`
    Output     OutputConfig    `json:"output"`
    Exclusions ExclusionConfig `json:"exclusions"`
    Logging    LoggingConfig   `json:"logging"`
}

type PiholeConfig struct {
    Host        string `json:"host"`
    Port        int    `json:"port"`
    APIEnabled  bool   `json:"api_enabled"`
    APIPassword string `json:"api_password"`
    APITOTP     string `json:"api_totp"`
    UseHTTPS    bool   `json:"use_https"`
    APITimeout  int    `json:"api_timeout"`
}
```

## Coding Standards

### Go Conventions

#### Structured Logging (CRITICAL)
**Never use `fmt.Printf`** - Always use structured logging:

```go
// ❌ NEVER do this
fmt.Printf("Error: %v\n", err)

// ✅ ALWAYS do this
logger := logger.New(&logger.Config{
    Level:        logger.LevelInfo,
    EnableColors: true,
    EnableEmojis: true,
    Component:    "analyzer",
})

logger.Error("Operation failed", 
    slog.String("error", err.Error()),
    slog.String("context", "additional context"))
```

#### Error Handling
```go
// Structured error logging with context
if err != nil {
    logger.Error("Pi-hole API connection failed",
        slog.String("host", config.Host),
        slog.Int("port", config.Port),
        slog.String("error", err.Error()))
    return fmt.Errorf("API connection failed: %w", err)
}
```

#### Package Organization
- **`internal/`** - Private application code
- **`cmd/`** - Application entry points
- **`testing/`** - Test utilities and fixtures
- **`docs/`** - Documentation

#### Code Quality Standards
```bash
# Format code
go fmt ./...

# Import organization
goimports -w .

# Lint checking
golint ./...

# Static analysis
staticcheck ./...

# Vet analysis
go vet ./...
```

### Documentation Standards

#### Function Documentation
```go
// NewClient creates a new Pi-hole API client with session management.
// It initializes HTTP client, validates configuration, and prepares
// for Pi-hole API authentication.
//
// Parameters:
//   - config: Pi-hole connection configuration
//   - logger: Structured logger for API operations
//
// Returns:
//   - *Client: Configured Pi-hole API client
//   - error: Configuration validation error, if any
func NewClient(config *Config, logger *logger.Logger) (*Client, error) {
    // Implementation...
}
```

#### Package Documentation
```go
// Package pihole provides Pi-hole API client functionality with session
// management, 2FA support, and structured logging integration.
//
// The package supports:
//   - Session-based authentication with CSRF protection
//   - TOTP 2FA authentication
//   - HTTPS/HTTP with certificate validation
//   - Comprehensive error handling and retry logic
//   - Structured logging throughout all operations
package pihole
```

### File Organization

#### Import Groups
```go
package main

import (
    // Standard library
    "context"
    "fmt"
    "log/slog"
    "os"
    
    // Third-party
    "modernc.org/sqlite"
    
    // Internal packages
    "pihole-analyzer/internal/analyzer"
    "pihole-analyzer/internal/config"
    "pihole-analyzer/internal/logger"
    "pihole-analyzer/internal/pihole"
)
```

## Testing Framework

### Test Structure

#### Unit Tests
```bash
# Run all unit tests
make test

# Run specific package tests
go test ./internal/logger -v

# Run tests with coverage
make test-coverage

# Cached test execution
make test-cached
```

#### Integration Tests
```bash
# Full integration test suite
make integration-test

# Container-based integration tests
make docker-test

# CI-compatible test execution
make ci-test
```

#### Test File Organization
```
tests/
├── unit/                    # Unit tests alongside source
│   ├── colors_test.go
│   └── logger_test.go
├── integration/             # Integration test scenarios
│   └── integration-test.sh
└── scripts/                 # Test automation scripts
    ├── ci-test.sh
    ├── integration-test.sh
    └── test.sh
```

### Writing Tests

#### Unit Test Example
```go
// internal/logger/logger_test.go
func TestLoggerWithColors(t *testing.T) {
    buf := &bytes.Buffer{}
    logger := New(&Config{
        Level:        LevelInfo,
        EnableColors: true,
        EnableEmojis: true,
        Component:    "test",
        Writer:       buf,
    })

    logger.Info("test message", slog.String("key", "value"))
    
    output := buf.String()
    assert.Contains(t, output, "test message")
    assert.Contains(t, output, "key=value")
}
```

#### Integration Test Example
```bash
#!/bin/bash
# tests/integration/api-test.sh

# Test Pi-hole API connectivity
echo "Testing Pi-hole API connection..."

# Use test binary with mock data
./pihole-analyzer-test --test --config testing/fixtures/mock_pihole_config.json

if [ $? -eq 0 ]; then
    echo "✅ Integration test passed"
else
    echo "❌ Integration test failed"
    exit 1
fi
```

### Mock Data & Fixtures

#### Test Fixtures
```
testing/fixtures/
├── mock_pihole_config.json  # Test configuration
├── mock_pihole.db          # Sample database
└── sample_configs/         # Various config scenarios
```

#### Mock Data Generation
```go
// testing/testutils/mock_data.go
func GenerateMockPiholeRecords(count int) []types.PiholeRecord {
    records := make([]types.PiholeRecord, count)
    for i := 0; i < count; i++ {
        records[i] = types.PiholeRecord{
            ID:        i + 1,
            DateTime:  time.Now().Format("2006-01-02 15:04:05"),
            Domain:    fmt.Sprintf("example%d.com", i),
            Client:    fmt.Sprintf("192.168.1.%d", 100+i%50),
            QueryType: "A",
            Status:    2, // ANSWERED
            Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
            HWAddr:    fmt.Sprintf("00:11:22:33:44:%02x", i%256),
        }
    }
    return records
}
```

## Build System

### Makefile Targets (40+ available)

#### Fast Development Builds
```bash
# Optimized incremental build with timing
make fast-build

# Build only if sources changed
make build-cached

# Auto-rebuild on file changes (requires entr)
make watch
```

#### Cache Management
```bash
# Pre-populate build caches for faster builds  
make cache-warm

# Display cache sizes and status
make cache-info

# Clean Go build and module caches
make cache-clean
```

#### Development Environment
```bash
# Complete development environment setup
make dev-setup

# Performance benchmarking
make benchmark

# Binary size analysis
make analyze-size
```

#### Container Workflows
```bash
# Build optimized Docker image
make docker-build

# Start development container environment
make docker-dev

# Multi-architecture builds (AMD64, ARM64, ARMv7)
make docker-multi

# Build API-only container variant
make docker-api-only
```

### Build Performance

#### Cache Strategy
- **Go Modules**: Cached by `go.sum` hash
- **Build Cache**: Preserved between builds for 60-80% speedup
- **Container Layers**: Multi-stage builds with dependency caching
- **CI/CD**: Comprehensive caching in GitHub Actions

#### Performance Metrics
```bash
# Example timing outputs
make cache-info     # "Build cache: 323M, Module cache: 2.2G"
make fast-build     # "✅ Fast build completed in 3s"
make docker-build   # "✅ Docker build completed in 21s"
```

### Binary Variants

#### Production Binary
```bash
# Build production binary
make build

# Run with Pi-hole API
./pihole-analyzer --config config.json
```

#### Test Binary
```bash
# Build test binary
make build-test

# Run with mock data
./pihole-analyzer-test --test
```

## Contributing Guidelines

### Development Workflow

#### 1. Feature Development
```bash
# Create feature branch
git checkout -b feature/api-enhancement

# Setup development environment
make dev-setup

# Implement changes with structured logging
# (Always use internal/logger, never fmt.Printf)

# Test changes
make test
make integration-test

# Build and validate
make fast-build
./pihole-analyzer-test --test
```

#### 2. Code Quality Checks
```bash
# Format and organize imports
go fmt ./...
goimports -w .

# Run linting
golint ./...
staticcheck ./...
go vet ./...

# Validate tests pass
make ci-test
```

#### 3. Container Testing
```bash
# Test containerized build
make docker-build

# Test multi-architecture
make docker-multi

# Validate development environment
make docker-dev
```

#### 4. Pull Request Preparation
```bash
# Ensure clean build cache
make cache-clean
make cache-warm

# Final test run
make ci-test

# Commit with descriptive message
git add .
git commit -m "feat: add API retry logic with structured logging"

# Push and create PR
git push origin feature/api-enhancement
```

### Code Review Guidelines

#### Required Checks
1. **Structured Logging**: No `fmt.Printf` usage
2. **Error Handling**: Proper error wrapping and context
3. **Testing**: Unit and integration tests included
4. **Documentation**: Code comments and README updates
5. **Performance**: Build cache impact considered

#### Review Checklist
- [ ] Structured logging used throughout
- [ ] Tests added/updated for new functionality
- [ ] Error handling includes proper context
- [ ] Container builds successfully
- [ ] Documentation updated
- [ ] No performance regression in builds

## Debugging & Troubleshooting

### Development Debugging

#### Local Debugging
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run with verbose output
./pihole-analyzer --config config.json --debug

# Debug Pi-hole API connectivity
./pihole-analyzer --pihole-setup
```

#### Container Debugging
```bash
# Interactive container shell
make docker-shell

# View container logs
make docker-logs

# Debug container networking
docker exec -it pihole-analyzer-dev ping 192.168.1.100
```

### Common Development Issues

#### Build Problems
```bash
# Check Go version
go version

# Verify dependencies
go mod verify
go mod tidy

# Clear and rebuild caches
make cache-clean
make cache-warm
make fast-build
```

#### Test Failures
```bash
# Run specific test with verbose output
go test -v ./internal/logger

# Run integration tests separately
make integration-test

# Check test fixtures
ls -la testing/fixtures/
```

#### Performance Issues
```bash
# Profile application
go build -o pihole-analyzer-profile ./cmd/pihole-analyzer
./pihole-analyzer-profile --config config.json --cpuprofile cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Logging Configuration

#### Development Logging
```json
{
  "logging": {
    "level": "debug",
    "colors": true,
    "emoji": true,
    "component": "dev",
    "file": "/tmp/pihole-analyzer-dev.log"
  }
}
```

#### Production Logging
```json
{
  "logging": {
    "level": "info",
    "colors": false,
    "emoji": false,
    "component": "prod",
    "file": "/var/log/pihole-analyzer.log"
  }
}
```

### Performance Monitoring

#### Build Performance
```bash
# Monitor build cache effectiveness
make cache-info

# Time build operations
time make fast-build

# Analyze binary size
make analyze-size
```

#### Runtime Performance
```bash
# Memory usage profiling
go test -memprofile mem.prof -bench .

# CPU usage profiling
go test -cpuprofile cpu.prof -bench .

# Container resource monitoring
docker stats pihole-analyzer-dev
```

## Best Practices Summary

### Development Standards
1. **Structured Logging**: Always use `internal/logger` with `slog`
2. **Error Context**: Include relevant context in error messages
3. **Testing**: Write tests for all new functionality
4. **Documentation**: Keep code documentation current
5. **Performance**: Consider build cache impact of changes

### Container Development
1. **Multi-Stage Builds**: Use optimized Docker builds
2. **Cache Optimization**: Preserve Go module and build caches
3. **Security**: Run as non-root user, minimal attack surface
4. **Multi-Architecture**: Test AMD64, ARM64, ARMv7 builds

### Code Quality
1. **Formatting**: Use `go fmt` and `goimports`
2. **Linting**: Run `golint` and `staticcheck`
3. **Testing**: Maintain high test coverage
4. **Review**: Follow code review guidelines
5. **CI/CD**: Ensure all checks pass in automation

This development guide provides comprehensive coverage of the modern development workflow for the Pi-hole Network Analyzer with emphasis on API-only architecture, structured logging, and container-first development.
