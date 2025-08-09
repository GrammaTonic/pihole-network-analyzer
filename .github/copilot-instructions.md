# Pi-hole Network Analyzer - AI Assistant Guide

A Go-based DNS analysis tool that connects to Pi-hole via API to generate colorized terminal reports with network insights.

## Architecture Overview

**Dual Binary System**: `pihole-analyzer` (production) and `pihole-analyzer-test` (development with mock data)  
**API-Only**: Direct Pi-hole REST API integration - no SSH or database access  
**Structured Logging**: Complete migration from `fmt.Printf` to `log/slog` with colors/emojis  
**Container-First**: Multi-architecture Docker builds (AMD64, ARM64, ARMv7) with optimized caching

## Key Components

```
cmd/                          # Binary entry points
├── pihole-analyzer/         # Production binary
└── pihole-analyzer-test/    # Development binary with mock data

internal/
├── interfaces/              # DataSource abstraction (API vs mock)
├── pihole/                  # Pi-hole API client with session management
├── logger/                  # Structured logging (slog) - USE THIS, NOT fmt.Printf
├── types/                   # Core data structures (PiholeRecord, ClientStats)
├── analyzer/                # Pi-hole data processing engine
├── reporting/               # Colorized terminal output
└── cli/                     # Centralized flag management
```

## Critical Development Patterns

### 🚨 NEVER Use fmt.Printf - Always Use Structured Logging
```go
// ❌ Wrong
fmt.Printf("Error: %v\n", err)

// ✅ Correct
logger := logger.New(&logger.Config{Component: "pihole-api"})
logger.Error("API connection failed", slog.String("error", err.Error()))
```

### DataSource Interface Pattern
All Pi-hole data access goes through `interfaces.DataSource`:
```go
// Production: uses internal/pihole (API client)
// Testing: uses mock implementation
client := pihole.NewClient(config, logger)
records, err := client.GetQueries(ctx, params)
```

### Configuration Structure
API-only config in `types.Config.Pihole`:
```go
type PiholeConfig struct {
    Host        string `json:"host"`
    APIPassword string `json:"api_password"`
    UseHTTPS    bool   `json:"use_https"`
    // No SSH fields - API only
}
```

## Essential Commands

### Fast Development Workflow
```bash
make fast-build        # Optimized incremental build (60-80% faster)
make cache-info        # Check build cache effectiveness
make docker-test-quick # Rapid container verification
make dev-setup         # Complete development environment prep
```

### Testing Both Binaries
```bash
# Production binary (requires real Pi-hole)
./pihole-analyzer --pihole config.json

# Test binary (uses mock data)
./pihole-analyzer-test --test
```

### Container Development
```bash
make docker-dev        # Development container with persistent caches
make docker-build      # Production container build
make docker-multi      # Multi-architecture builds
```

## Build System (40+ Makefile Targets)

**Performance**: Build cache reduces build times by 60-80%  
**Versioning**: Automatic VERSION and BUILD_TIME injection into binaries  
**Testing**: Separate test infrastructure with mock Pi-hole data in `testing/fixtures/`

## Critical Rules

1. **Single Pi-hole Instance Only**: Current architecture supports one Pi-hole connection per execution
2. **Structured Logging Only**: Use `internal/logger`, never `fmt.Printf`
2. **Test Both Binaries**: Production and test versions have different data sources
3. **Container Compatibility**: All features must work in containerized environments
4. **Build Cache Awareness**: Check `make cache-info` after changes
5. **API-Only Architecture**: No SSH, database, or legacy connectivity code

## Common Tasks

**Adding New Features**: Extend `interfaces.DataSource`, implement in `internal/pihole`, add structured logging  
**CLI Changes**: Modify `internal/cli/flags.go` for centralized flag management  
**Data Structures**: Update `internal/types/types.go` for Pi-hole record structures  
**Output Formatting**: Use `internal/reporting` with color support (`--no-color` flag)

## Roadmap

### Current Focus (Q3 2025)
- **Enhanced Configuration Validation**: Comprehensive config validation with structured logging
- **Performance Metrics**: Built-in Prometheus metrics endpoints for monitoring
- **Web UI Foundation**: Basic web interface for daemon mode preparation

### Near Term (Q4 2025)
- **Daemon Mode**: Background service for continuous Pi-hole monitoring
- **Multi-Pi-hole Support**: Connect and analyze multiple Pi-hole instances
- **REST API**: HTTP API for programmatic access to analysis data
- **Advanced Filtering**: Complex query filters and time-based analysis

### Medium Term (2026)
- **Real-time Dashboard**: Live monitoring interface with WebSocket updates
- **Alert System**: Configurable alerts for network anomalies
- **Plugin Architecture**: Extensible analysis modules and custom reporters
- **Enhanced Network Analysis**: Deep packet inspection and traffic patterns

### Long Term (Future)
- **Machine Learning**: AI-powered anomaly detection and trend analysis
- **Multi-format Export**: JSON, XML, CSV export capabilities
- **Integration Ecosystem**: Grafana, InfluxDB, and monitoring platform connectors
- **Mobile App**: Companion mobile application for network monitoring
- **Enhanced Network Analysis**: Deep packet inspection and traffic patterns

### Long Term (Future)
- **Machine Learning**: AI-powered anomaly detection and trend analysis
- **Multi-format Export**: JSON, XML, CSV export capabilities
- **Integration Ecosystem**: Grafana, InfluxDB, and monitoring platform connectors
- **Mobile App**: Companion mobile application for network monitoring

This project prioritizes **API-only Pi-hole integration**, **structured logging**, **fast containerized builds**, and **beautiful terminal output**.
