# Pi-hole Network Analyzer - AI Assistant Guide

A Go-based DNS analysis tool that connects to Pi-hole via API to generate colorized terminal reports with network insights, featuring web UI, metrics collection, and daemon mode.

## Architecture Overview

**Dual Binary System**: `pihole-analyzer` (production) and `pihole-analyzer-test` (development with mock data)  
**API-Only**: Direct Pi-hole REST API integration - no SSH or database access  
**Structured Logging**: Complete migration from `fmt.Printf` to `log/slog` with colors/emojis  
**Container-First**: Multi-architecture Docker builds (AMD64, ARM64, ARMv7) with optimized caching  
**Web UI**: Built-in HTTP dashboard for real-time monitoring and daemon mode  
**Metrics**: Prometheus endpoints for monitoring and observability

## Key Components

```
cmd/                          # Binary entry points
├── pihole-analyzer/         # Production binary
└── pihole-analyzer-test/    # Development binary with mock data

internal/
├── interfaces/              # DataSource abstraction (API vs mock)
├── pihole/                  # Pi-hole API client with session management
├── logger/                  # Structured logging (slog) - USE THIS, NOT fmt.Printf
├── types/                   # Core data structures (PiholeRecord, ClientStats, MLConfig)
├── analyzer/                # Pi-hole data processing engine
├── reporting/               # Colorized terminal output
├── cli/                     # Centralized flag management
├── web/                     # Web UI server and dashboard templates
├── metrics/                 # Prometheus metrics collection and server
├── ml/                      # Machine learning (anomaly detection, trend analysis)
└── validation/              # Configuration validation with structured logging
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
Multi-service config in `types.Config`:
```go
type Config struct {
    Pihole     PiholeConfig    `json:"pihole"`     // Pi-hole API settings
    Web        WebConfig       `json:"web"`        // Web UI configuration  
    Metrics    MetricsConfig   `json:"metrics"`    // Prometheus metrics
    ML         MLConfig        `json:"ml"`         // Machine learning features
    Logging    LoggingConfig   `json:"logging"`    // Structured logging
    // No SSH fields - API only
}
```

### ML System Architecture
Complete ML interfaces in `internal/ml/interfaces.go`:
```go
// Three-tier ML system
type MLEngine interface {
    AnomalyDetector   // Statistical anomaly detection
    TrendAnalyzer     // Trend analysis and forecasting
    // ProcessData combines both capabilities
}

// Engine creation with configuration
engine := ml.NewEngine(config.ML, logger)
results, err := engine.ProcessData(ctx, piholeRecords)
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

### ML Development Testing
```bash
# Run ML tests specifically
go test -v ./internal/ml/...

# Test ML engine integration
go test -v ./internal/ml/ -run TestEngine

# Debug ML algorithms with test data
go run debug_ml.go  # Uses internal/ml test fixtures
```

### Web UI and Daemon Mode
```bash
# Start web UI server (default: localhost:8080)
./pihole-analyzer --web --config config.json

# Run in daemon mode with web UI
./pihole-analyzer --daemon --web --config config.json

# Custom web configuration
./pihole-analyzer --web --web-host 0.0.0.0 --web-port 3000 --config config.json
```

### Metrics Collection
```bash
# Enable Prometheus metrics (default: localhost:9090)
./pihole-analyzer --metrics --config config.json

# Combined web UI and metrics
./pihole-analyzer --web --metrics --daemon --config config.json
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
3. **Test Both Binaries**: Production and test versions have different data sources
4. **Container Compatibility**: All features must work in containerized environments
5. **Build Cache Awareness**: Check `make cache-info` after changes
6. **API-Only Architecture**: No SSH, database, or legacy connectivity code
7. **Web UI Foundation**: Use `internal/web` for dashboard features, ensure localhost:8080 default
8. **Metrics Integration**: Prometheus endpoints at localhost:9090, use `internal/metrics`
9. **Configuration Validation**: Use `internal/validation` with structured logging for all config checks
10. **ML Algorithm Calibration**: Always test threshold values - confidence (0.75), score normalization (≤1.0), sensitivity (0.01-0.1)

## Common Tasks

**Adding New Features**: Extend `interfaces.DataSource`, implement in `internal/pihole`, add structured logging  
**CLI Changes**: Modify `internal/cli/flags.go` for centralized flag management  
**Data Structures**: Update `internal/types/types.go` for Pi-hole record structures  
**Output Formatting**: Use `internal/reporting` with color support (`--no-color` flag)  
**Web UI Development**: Use `internal/web` templates and server, ensure daemon mode compatibility  
**Metrics Addition**: Extend `internal/metrics` for new Prometheus endpoints  
**Configuration Updates**: Add validation in `internal/validation` with proper error handling  
**ML Development**: Implement `ml.AnomalyDetector` or `ml.TrendAnalyzer` interfaces, test with `go test ./internal/ml/...`

## Web UI Development Patterns

### Server Lifecycle
```go
// Start web server with proper logging
logger := logger.New(&logger.Config{Component: "web-server"})
server := web.NewServer(config.Web, dataSource, logger)
if err := server.Start(ctx); err != nil {
    logger.Error("Failed to start web server", slog.String("error", err.Error()))
}
```

### Dashboard Integration
```go
// Use templates in internal/web/templates/
template := template.Must(template.ParseFiles("internal/web/templates/dashboard.html"))
```

### Daemon Mode Patterns
```go
// Long-running service with graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle signals for graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
```

## Metrics Development Patterns

### Adding New Metrics
```go
// Use internal/metrics package
metricsServer := metrics.NewServer(config.Metrics, logger)
metricsServer.RecordQuery(queryType, clientIP, timestamp)
```

### Prometheus Integration
```go
// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
logger.Info("Metrics server starting", slog.String("addr", addr))
```

## ML Development Patterns

### ML Engine Usage
```go
// Initialize ML engine with configuration
logger := logger.New(&logger.Config{Component: "ml-engine"})
engine := ml.NewEngine(config.ML, logger)
if err := engine.Initialize(ctx, config.ML); err != nil {
    logger.Error("Failed to initialize ML engine", slog.String("error", err.Error()))
}

// Process data for anomaly detection and trend analysis
results, err := engine.ProcessData(ctx, piholeRecords)
```

### Algorithm Threshold Calibration
```go
// Critical ML algorithm settings - test these values:
// - Confidence thresholds: 0.75 (not 0.6) to prevent false positives
// - Score normalization: Always use math.Min(score, 1.0) to cap at 1.0
// - Sensitivity settings: 0.01-0.1 range for trend analysis
// - Window sizes: Use time.Duration for consistency
```

### ML Testing Patterns
```go
// Always test ML algorithms with expected behavior
go test -v ./internal/ml/ -run TestAnomalyDetector
go test -v ./internal/ml/ -run TestTrendAnalyzer
// Check that confidence filtering works: score ≥ confidence threshold
// Verify score normalization: all scores ≤ 1.0
```

## Roadmap

### Current Focus (Q1 2025)
- **Web UI Foundation**: HTTP dashboard at localhost:8080 with real-time Pi-hole data ✅
- **Prometheus Metrics**: Built-in metrics endpoints at localhost:9090 for monitoring ✅
- **Enhanced Configuration Validation**: Comprehensive config validation with structured logging ✅
- **Daemon Mode**: Background service for continuous Pi-hole monitoring ✅
- **Machine Learning**: AI-powered anomaly detection and trend analysis ✅

### Near Term (Q2 2025)
- **Multi-Pi-hole Support**: Connect and analyze multiple Pi-hole instances
- **REST API**: HTTP API for programmatic access to analysis data
- **Advanced Filtering**: Complex query filters and time-based analysis
- **WebSocket Updates**: Real-time dashboard updates without page refresh

### Medium Term (Q3-Q4 2025)
- **Enhanced Dashboard**: Advanced web UI with interactive charts and graphs
- **Alert System**: Configurable alerts for network anomalies
- **Plugin Architecture**: Extensible analysis modules and custom reporters
- **Enhanced Network Analysis**: Deep packet inspection and traffic patterns

### Long Term (2026+)
- **Enhanced ML Models**: Advanced machine learning with custom model training
- **Multi-format Export**: JSON, XML, CSV export capabilities
- **Integration Ecosystem**: Grafana, InfluxDB, and monitoring platform connectors
- **Mobile App**: Companion mobile application for network monitoring

This project prioritizes **API-only Pi-hole integration**, **structured logging**, **web UI Foundation**, **Prometheus metrics**, **fast containerized builds**, **ML-powered analysis**, and **beautiful terminal output**.
