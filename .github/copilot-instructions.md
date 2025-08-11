# Pi-hole Network Analyzer - AI Assistant Guide

A Go-based DNS analysis tool that connects to Pi-hole via API to generate colorized terminal reports with network insights, featuring web UI with real-time WebSocket updates, metrics collection, and daemon mode.

## Architecture Overview

**Dual Binary System**: `pihole-analyzer` (production) and `pihole-analyzer-test` (development with mock data)  
**API-Only**: Direct Pi-hole REST API integration - no SSH or database access  
**Structured Logging**: Complete migration from `fmt.Printf` to `log/slog` with colors/emojis  
**Container-First**: Multi-architecture Docker builds (AMD64, ARM64, ARMv7) with optimized caching  
**Real-time Web UI**: HTTP dashboard with WebSocket live updates and embedded templates  
**WebSocket Infrastructure**: gorilla/websocket for real-time dashboard updates with auto-reconnection  
**Metrics**: Prometheus endpoints for monitoring and observability  
**Enhanced Network Analysis**: Deep packet inspection, traffic patterns, security analysis, and performance monitoring ✅  
**Alert System**: Configurable alerts for network anomalies with ML integration, Slack/Email notifications ✅  
**Factory Pattern**: `interfaces.DataSourceFactory` abstracts Pi-hole vs mock data sources

## Key Components

```
cmd/                          # Binary entry points
├── pihole-analyzer/         # Production binary
└── pihole-analyzer-test/    # Development binary with mock data

internal/
├── interfaces/              # DataSource abstraction (API vs mock)
├── pihole/                  # Pi-hole API client with session management
├── logger/                  # Structured logging (slog) - USE THIS, NOT fmt.Printf
├── types/                   # Core data structures (PiholeRecord, ClientStats, MLConfig, NetworkAnalysisConfig, AlertConfig)
├── analyzer/                # Pi-hole data processing engine with ML and alert integration
├── reporting/               # Colorized terminal output
├── cli/                     # Centralized flag management
├── web/                     # Web UI server and dashboard templates with WebSocket support
├── metrics/                 # Prometheus metrics collection and server
├── ml/                      # Machine learning (anomaly detection, trend analysis)
├── network/                 # Enhanced network analysis (DPI, traffic patterns, security, performance) ✅
├── alerts/                  # Alert system (rules, notifications, storage) ✅
├── integrations/            # External integrations (Grafana, Prometheus, Loki) ✅
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

### Configuration Merging Pattern
Flags override config file values via `config.MergeFlags()`:
```go
// Load config file first, then apply CLI flags
cfg, err := config.LoadConfig(configPath)
config.MergeFlags(cfg, *flags.OnlineOnly, *flags.NoExclude, testMode, *flags.Pihole)
```

### Configuration Structure
Multi-service config in `types.Config`:
```go
type Config struct {
    Pihole          PiholeConfig          `json:"pihole"`           // Pi-hole API settings
    Web             WebConfig             `json:"web"`              // Web UI configuration with WebSocket
    Metrics         MetricsConfig         `json:"metrics"`          // Prometheus metrics
    ML              MLConfig              `json:"ml"`               // Machine learning features
    NetworkAnalysis NetworkAnalysisConfig `json:"network_analysis"` // Enhanced network analysis ✅
    Alerts          AlertConfig           `json:"alerts"`           // Alert system configuration ✅
    Integrations    IntegrationsConfig    `json:"integrations"`     // External integrations (Grafana, Prometheus, Loki) ✅
    Logging         LoggingConfig         `json:"logging"`          // Structured logging
    // No SSH fields - API only
}
```

### External Dependencies
The project uses minimal external dependencies:
- **gorilla/websocket v1.5.3**: WebSocket server implementation for real-time updates
- **Go standard library**: Core functionality (net/http, log/slog, etc.)
- **Node.js (dev only)**: Release automation, NOT runtime dependency

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

### WebSocket Real-Time Updates Architecture
Complete WebSocket system in `internal/web/`:
```go
// WebSocket manager with connection pooling
type WebSocketManager struct {
    connections map[string]*WebSocketConnection
    updateChan  chan *UpdateMessage
    // Connection lifecycle, broadcasting, cleanup
}

// Real-time update broadcasting
func (s *Server) BroadcastClientUpdate(clientData interface{})
func (s *Server) BroadcastQueryUpdate(queryData interface{})  
func (s *Server) BroadcastMetricsUpdate(metricsData interface{})

// WebSocket connection with auto-reconnection
class WebSocketClient {
    connect()           // Establish connection
    handleMessage()     // Process updates
    reconnect()         // Auto-reconnection with backoff
    updateDOM()         // Live DOM updates
}
```

### Enhanced Network Analysis Architecture
Complete network analysis system in `internal/network/`:
```go
// Four-tier network analysis system
type NetworkAnalyzer interface {
    DeepPacketInspector   // Protocol analysis, packet inspection
    TrafficPatternAnalyzer // Bandwidth patterns, temporal analysis, client behavior
    SecurityAnalyzer      // Threat detection, DNS anomalies, port scanning
    PerformanceAnalyzer   // Latency, throughput, quality assessment
}

// Factory pattern for analyzer creation
factory := network.NewAnalyzerFactory(logger)
analyzer, err := factory.CreateNetworkAnalyzer(config.NetworkAnalysis)
result, err := analyzer.AnalyzeTraffic(ctx, records, clientStats)
```

### Alert System Architecture
Complete alert system in `internal/alerts/`:
```go
// Three-tier alert system
type AlertManager interface {
    AlertEvaluator    // Rule evaluation and condition checking
    NotificationSender // Multi-channel notifications (Slack, Email, Log)
    AlertStorage      // Alert persistence and retrieval
}

// Factory pattern for alert manager creation
alertConfig := alerts.AlertConfig{...}
manager := alerts.NewManager(alertConfig, logger)
err := manager.ProcessData(ctx, analysisResult, mlResults)
```

## Essential Commands

### Semantic Versioning & Release Management
**IMPORTANT**: Node.js is ONLY used for release automation and development workflow tools. The Go application itself has ZERO Node.js dependencies and runs completely independently.

```bash
make version           # Show current version information
make release-setup     # Install semantic-release dependencies (requires Node.js - DEVELOPMENT ONLY)
make commit           # Interactive conventional commit creation
make release-dry-run  # Test what the next release would be

# Conventional commit examples
git commit -m "feat(api): add user authentication"     # MINOR bump
git commit -m "fix(network): resolve memory leak"      # PATCH bump
git commit -m "docs: update API documentation"         # No version bump

# Breaking change (MAJOR bump)
git commit -m "feat(api): redesign configuration interface

BREAKING CHANGE: Configuration structure has changed. See migration guide."
```

**Node.js Scope**: Release automation, commit validation, changelog generation (CI/CD only)  
**Go Application**: Pure Go - no Node.js runtime dependencies whatsoever

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

### Web UI and WebSocket Development
```bash
# Start web UI server with WebSocket support (default: localhost:8080)
./pihole-analyzer --web --config config.json

# Run in daemon mode with web UI and real-time updates
./pihole-analyzer --daemon --web --config config.json

# Custom web configuration with WebSocket
./pihole-analyzer --web --web-host 0.0.0.0 --web-port 3000 --config config.json
```

### Metrics Collection
```bash
# Enable Prometheus metrics (default: localhost:9090)
./pihole-analyzer --metrics --config config.json

# Combined web UI and metrics
./pihole-analyzer --web --metrics --daemon --config config.json
```

### Enhanced Network Analysis Commands
```bash
# Enable all network analysis features
./pihole-analyzer --network-analysis --pihole config.json

# Enable specific analysis components
./pihole-analyzer --enable-dpi --enable-security-analysis --pihole config.json

# Combined with web UI for real-time visualization
./pihole-analyzer --network-analysis --web --pihole config.json

# Test network analysis with mock data
./pihole-analyzer-test --network-analysis --test
```

### Alert System Commands
```bash
# Enable alert system with default rules
./pihole-analyzer --alerts --config config.json

# Enable alerts with specific configuration file
./pihole-analyzer --alerts --config config-with-alerts.json

# Combined alerts with ML anomaly detection
./pihole-analyzer --alerts --ml --config config.json

# Test alert system with mock data and ML
./pihole-analyzer-test --alerts --ml --test

# View alert status in daemon mode
./pihole-analyzer --alerts --daemon --web --config config.json
```

### Container Development
```bash
make docker-dev        # Development container with persistent caches
make docker-build      # Production container build
make docker-multi      # Multi-architecture builds
make docker-push-ghcr  # Push to GitHub Container Registry
```

## Build System (40+ Makefile Targets)

**Performance**: Build cache reduces build times by 60-80%  
**Versioning**: Automatic VERSION and BUILD_TIME injection into binaries  
**Testing**: Separate test infrastructure with mock Pi-hole data in `testing/fixtures/`  
**Embedded Assets**: Web templates embedded using `//go:embed` for single binary distribution

### Fast Build Commands (Incremental)
```bash
make build-cached     # Only rebuilds if Go sources changed
make cache-warm       # Pre-populate build caches for CI
make fast-build       # Optimized build with aggressive caching
```

## Critical Rules

1. **Single Pi-hole Instance Only**: Current architecture supports one Pi-hole connection per execution
2. **Structured Logging Only**: Use `internal/logger`, never `fmt.Printf`
3. **Test Both Binaries**: Production and test versions have different data sources
4. **Container Compatibility**: All features must work in containerized environments
5. **Build Cache Awareness**: Check `make cache-info` after changes
6. **API-Only Architecture**: No SSH, database, or legacy connectivity code
7. **Web UI Foundation**: Use `internal/web` for dashboard features, WebSocket real-time updates, ensure localhost:8080 default
8. **Metrics Integration**: Prometheus endpoints at localhost:9090, use `internal/metrics`
9. **Configuration Validation**: Use `internal/validation` with structured logging for all config checks
10. **ML Algorithm Calibration**: Always test threshold values - confidence (0.75), score normalization (≤1.0), sensitivity (0.01-0.1)
11. **Enhanced Network Analysis**: Use `internal/network` for DPI, traffic patterns, security, and performance analysis - all components integrate via factory pattern
12. **Alert System Integration**: Use `internal/alerts` for alert management, supports ML integration, multi-channel notifications (Slack/Email/Log), and configurable rules
13. **Node.js Scope Limitation**: Node.js is STRICTLY for release automation and development tooling ONLY - the Go application has zero Node.js runtime dependencies
14. **Pure Go Runtime**: All production binaries, Docker containers, and deployed applications run on Go only - no Node.js required in production environments
15. **Conventional Commits**: Use conventional commit format for all commits - enables automated versioning and changelog generation
16. **GitLab Flow**: Follow GitLab Flow with release branches - main for development, release/vX.Y for stable releases, feature branches for development
17. **Semantic Versioning**: Automated SemVer based on commit messages - feat: (minor), fix: (patch), BREAKING CHANGE: (major)
18. **Embedded Assets**: All web templates and static files must use `//go:embed` for single binary distribution
19. **Fix Merge Conflicts**: Always resolve merge conflicts in your own branch before submitting changes
20. **Documentation Updates**: Keep all documentation up to date with code changes
21. **Inconsistent Code Styles**: Ensure consistent code styles across the codebase
22. **Unoptimized Code Paths**: Identify and optimize any performance bottlenecks in the codebase
23. **Error Handling Improvements**: Enhance error handling and logging throughout the codebase
24. **Testing Coverage**: Ensure all new features and bug fixes have corresponding tests


## Common Tasks

**Adding New Features**: Extend `interfaces.DataSource`, implement in `internal/pihole`, add structured logging  
**CLI Changes**: Modify `internal/cli/flags.go` for centralized flag management  
**Data Structures**: Update `internal/types/types.go` for Pi-hole record structures  
**Output Formatting**: Use `internal/reporting` with color support (`--no-color` flag)  
**Web UI Development**: Use `internal/web` templates and server, WebSocket manager for real-time updates, ensure daemon mode compatibility  
**Metrics Addition**: Extend `internal/metrics` for new Prometheus endpoints  
**Configuration Updates**: Add validation in `internal/validation` with proper error handling  
**ML Development**: Implement `ml.AnomalyDetector` or `ml.TrendAnalyzer` interfaces, test with `go test ./internal/ml/...`  
**Enhanced Network Analysis**: Extend `network.NetworkAnalyzer` interfaces, implement in `internal/network`, integrate via factory pattern  
**Alert System Development**: Extend `alerts.AlertManager` interfaces, implement new notification channels, configure alert rules  
**Release Management**: Use conventional commits (`make commit`), follow GitLab Flow with release branches, semantic versioning automation

## Testing Infrastructure Patterns

### Mock Data Architecture
```go
// Use testing/testutils for test mode
testutils.RunTestMode(cfg)  // Generates mock client stats
// Mock data in testing/fixtures/ directory
```

### Dual Binary Testing Strategy
```bash
# Always test both production and mock binaries
./pihole-analyzer --pihole real-config.json    # Production
./pihole-analyzer-test --test                  # Mock data
```

### WebSocket Testing Patterns
```go
// Test WebSocket manager and connections
go test -v ./internal/web/ -run TestWebSocketManager
go test -v ./internal/web/ -run TestWebSocketConnection

// Test real-time broadcasting
manager.BroadcastUpdate("test", map[string]string{"key": "value"})
// Verify message received in connection's Send channel
```

## Web UI Development Patterns

### WebSocket Real-Time Updates
```go
// Initialize WebSocket manager in server
logger := logger.New(&logger.Config{Component: "web-server"})
wsManager := NewWebSocketManager(logger, DefaultWebSocketConfig())
server := &Server{wsManager: wsManager, ...}

// Broadcast real-time updates
server.BroadcastQueryUpdate(queryData)
server.BroadcastClientUpdate(clientData)
server.BroadcastMetricsUpdate(metricsData)
```

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
// Use embedded templates with WebSocket support
//go:embed templates/*
var templatesFS embed.FS
template := template.Must(template.ParseFS(templatesFS, "templates/dashboard.html"))

// WebSocket endpoint for real-time updates
http.HandleFunc("/ws", server.HandleWebSocket)
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

## Enhanced Network Analysis Development Patterns

### Network Analyzer Usage
```go
// Initialize network analyzer with configuration
logger := logger.New(&logger.Config{Component: "network-analyzer"})
factory := network.NewAnalyzerFactory(logger)
analyzer, err := factory.CreateNetworkAnalyzer(config.NetworkAnalysis)
if err != nil {
    logger.Error("Failed to create network analyzer", slog.String("error", err.Error()))
}

// Perform comprehensive network analysis
result, err := analyzer.AnalyzeTraffic(ctx, piholeRecords, clientStats)
```

### Component-Specific Analysis
```go
// Deep Packet Inspection
dpi, err := factory.CreateDPIAnalyzer(config.DeepPacketInspection)
packetResult, err := dpi.InspectPackets(ctx, records, config.DeepPacketInspection)

// Traffic Pattern Analysis
trafficAnalyzer, err := factory.CreateTrafficAnalyzer(config.TrafficPatterns)
patternResult, err := trafficAnalyzer.AnalyzePatterns(ctx, records, clientStats, config.TrafficPatterns)

// Security Analysis
securityAnalyzer, err := factory.CreateSecurityAnalyzer(config.SecurityAnalysis)
securityResult, err := securityAnalyzer.AnalyzeSecurity(ctx, records, clientStats, config.SecurityAnalysis)

// Performance Analysis
perfAnalyzer, err := factory.CreatePerformanceAnalyzer(config.Performance)
perfResult, err := perfAnalyzer.AnalyzePerformance(ctx, records, clientStats, config.Performance)
```

### Visualization Integration
```go
// Generate visualization data for web UI
visualizer, err := factory.CreateVisualizer()
trafficViz, err := visualizer.GenerateTrafficVisualization(analysisResult)
topologyViz, err := visualizer.GenerateTopologyVisualization(records, clientStats)
timeSeriesViz, err := visualizer.GenerateTimeSeriesData(records, "query_count", time.Hour)
```

### Network Analysis Testing Patterns
```go
// Test complete network analysis workflow
go test -v ./internal/network/ -run TestEnhancedNetworkAnalyzer
go test -v ./tests/integration/ -run TestNetworkAnalysis_Integration

// Test individual components
go test -v ./internal/network/ -run TestDeepPacketInspector
go test -v ./internal/network/ -run TestTrafficPatternAnalyzer
go test -v ./internal/network/ -run TestSecurityAnalyzer
go test -v ./internal/network/ -run TestPerformanceAnalyzer
```

## Alert System Development Patterns

### Alert Manager Usage
```go
// Initialize alert manager with configuration
logger := logger.New(&logger.Config{Component: "alert-manager"})
alertConfig := alerts.AlertConfig{...}
manager := alerts.NewManager(alertConfig, logger)
if err := manager.Initialize(ctx, alertConfig); err != nil {
    logger.Error("Failed to initialize alert manager", slog.String("error", err.Error()))
}

// Process data for alert evaluation
err := manager.ProcessData(ctx, analysisResult, mlResults)
```

### Component-Specific Alert Development
```go
// Rule evaluation
evaluator := alerts.NewEvaluator(config, logger)
triggeredRules, err := evaluator.EvaluateRules(ctx, data, rules)

// Notification handling
notifier := alerts.NewNotificationSender(config.Notifications, logger)
err := notifier.SendAlert(ctx, alert, []alerts.NotificationChannel{alerts.ChannelSlack, alerts.ChannelEmail})

// Alert storage
storage := alerts.NewStorage(config.Storage, logger)
err := storage.StoreAlert(ctx, alert)
```

### Alert System Testing Patterns
```go
// Test complete alert system workflow
go test -v ./internal/alerts/ -run TestIntegrationAlertSystemWorkflow
go test -v ./tests/integration/ -run TestAlerts_Integration

// Test individual components
go test -v ./internal/alerts/ -run TestEvaluator
go test -v ./internal/alerts/ -run TestNotifications
go test -v ./internal/alerts/ -run TestStorage
go test -v ./internal/alerts/ -run TestManager
```

## Roadmap

### Current Focus (Q1 2025)
- **Web UI Foundation**: HTTP dashboard at localhost:8080 with real-time Pi-hole data ✅
- **Prometheus Metrics**: Built-in metrics endpoints at localhost:9090 for monitoring ✅
- **Enhanced Configuration Validation**: Comprehensive config validation with structured logging ✅
- **Daemon Mode**: Background service for continuous Pi-hole monitoring ✅
- **Machine Learning**: AI-powered anomaly detection and trend analysis ✅
- **Enhanced Network Analysis**: Deep packet inspection, traffic patterns, security analysis, and performance monitoring ✅
- **Alert System**: Configurable alerts for network anomalies with ML integration, Slack/Email notifications ✅
- **Integration Ecosystem**: Grafana, Prometheus, and monitoring platform connectors and logging to Loki ✅
- **Semantic Versioning & Release Automation**: GitLab Flow with conventional commits and automated releases ✅

### Near Term (Q2 2025)
- **REST API**: HTTP API for programmatic access to analysis data
- **Advanced Filtering**: Complex query filters and time-based analysis
- **WebSocket Updates**: Real-time dashboard updates without page refresh ✅

### Medium Term (Q3-Q4 2025)
- **Enhanced Dashboard**: Advanced web UI with interactive charts and graphs
- **Plugin Architecture**: Extensible analysis modules and custom reporters
- **Multi-format Export**: JSON, XML, CSV export capabilities

### Long Term (2026+)
- **Enhanced ML Models**: Advanced machine learning with custom model training
- **Mobile App**: Companion mobile application for network monitoring

This project prioritizes **API-only Pi-hole integration**, **structured logging**, **real-time web UI with WebSocket updates**, **Prometheus metrics**, **fast containerized builds**, **ML-powered analysis**, **enhanced network analysis**, **alert system with notifications**, **integration ecosystem**, **automated semantic versioning**, and **beautiful terminal output**.

**Architecture Note**: This is a **pure Go application** with zero runtime dependencies. Node.js is used exclusively for release automation and development workflow tools - it is never required in production environments or for running the actual Pi-hole analyzer.
