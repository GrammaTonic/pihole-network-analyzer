# Pi-hole Network Analyzer - AI Coding Assistant Instructions

This file provides comprehensive guidance for AI coding assistants working on the Pi-hole Network Analyzer project.

## Project Overview

**Repository**: Pi-hole Network Analyzer  
**Language**: Go 1.23.12+  
**Module**: `pihole-analyzer`  
**Binary Names**: `pihole-analyzer` (production), `pihole-analyzer-test` (development)
**Commands Directory**: `cmd/pihole-analyzer/`, `cmd/pihole-analyzer-test/`  
**Main Files**: `cmd/pihole-analyzer/main.go`, `cmd/pihole-analyzer-test/main.go`
**Help Command**: `pihole-analyzer --help`  
**Architecture**: Standard Go Project Layout with API-only Pi-hole integration
**Pi-hole API Documentation**: https://ftl.pi-hole.net/master/docs/#

### Core Purpose
**Pi-hole-focused** DNS usage analysis tool with **API-only connectivity**. Features structured logging, colorized terminal output, network analysis, and comprehensive reporting. Includes production-ready containerization and optimized build system.

**🚨 IMPORTANT ARCHITECTURE CHANGES**: 
- **API-Only Implementation**: Direct Pi-hole API integration - no SSH dependencies
- **SSH Functionality Removed**: All SSH connectivity and database access removed
- **Migration Logic Removed**: No fallback or transition mechanisms needed
- **Simplified Configuration**: Pure API-based configuration structure
- **Enhanced Security**: No SSH access required, controlled API permissions only
- **Structured Logging**: Comprehensive `log/slog` implementation throughout
- **Container-First**: Production-ready multi-architecture containerization
- **Fast Builds**: Advanced caching strategies for 60-80% build speed improvement

## Project Structure & Architecture

### Documentation Structure
The project maintains comprehensive documentation in the `docs/` directory:

**Core Documentation Files**:
- `docs/api.md` - API reference guide
- `docs/development.md` - Development setup and guidelines
- `docs/configuration.md` - Configuration guide  
- `docs/installation.md` - Installation instructions
- `docs/usage.md` - Usage examples and CLI reference
- `docs/troubleshooting.md` - Troubleshooting guide
- `docs/index.md` - Main documentation index
- `docs/container-registry.md` - Container deployment guide

**Specialized Documentation**:
- `docs/fast-builds.md` - Build optimization strategies
- `docs/container-usage.md` - Docker deployment guide
- `docs/README.md` - Documentation overview

**Documentation Standards**:
- Enhanced with structured examples and code snippets
- Container-first approach with deployment examples
- API-only architecture focus (no SSH references)
- Comprehensive error handling and troubleshooting guides

### Directory Layout
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
│   ├── api.md                   # Pi-hole API integration guide
│   ├── development.md           # Development setup and guidelines
│   ├── configuration.md         # Configuration documentation
│   ├── installation.md          # Installation instructions
│   ├── usage.md                 # Usage examples and CLI reference
│   ├── troubleshooting.md       # Troubleshooting guide
│   ├── index.md                 # Main documentation index
│   ├── container-registry.md    # Container deployment guide
│   ├── fast-builds.md           # Build optimization strategies
│   ├── container-usage.md       # Docker deployment guide
│   └── README.md                # Documentation overview
├── scripts/                      # Build automation & cache warming
├── testing/                      # Test utilities and fixtures
│   ├── fixtures/                 # Mock data and test configs
│   ├── integration/              # Integration test helpers
│   └── testutils/               # Test utilities and mock data
├── tests/                        # Test scripts and CI validation
│   ├── integration/              # Integration test scripts
│   ├── scripts/                  # CI/CD test automation
│   └── unit/                     # Unit test files
├── .github/workflows/           # CI/CD with advanced caching
│   ├── ci.yml                   # Main CI/CD pipeline
│   └── container.yml            # Container build and publish
├── Dockerfile                   # Multi-stage, multi-arch container builds
├── Dockerfile.api-only         # API-only container variant
├── docker-compose*.yml         # Multiple environment configurations
│   ├── docker-compose.yml      # Default development
│   ├── docker-compose.dev.yml  # Development environment
│   ├── docker-compose.prod.yml # Production environment
│   └── docker-compose.api-only.yml  # API-only deployment
└── Makefile                     # Enhanced build system (40+ targets)
```

### Key Files & Their Roles
- **`cmd/pihole-analyzer/main.go`**: Production application entry point with API-only connectivity
- **`cmd/pihole-analyzer-test/main.go`**: Development/testing entry point with mock data
- **`internal/logger/logger.go`**: Structured logging with slog, colors, and emojis
- **`internal/pihole/client.go`**: Pi-hole API client with session management
- **`internal/interfaces/data_source.go`**: Data source abstraction interface
- **`internal/cli/flags.go`**: Command-line flag parsing and validation
- **`internal/analyzer/analyzer.go`**: Pi-hole data analysis engine
- **`internal/types/types.go`**: Core data structures (ClientStats, PiholeRecord, Config)
- **`.github/workflows/ci.yml`**: Enhanced CI/CD with multi-layer caching
- **`Dockerfile`**: Multi-architecture container builds (AMD64, ARM64, ARMv7)
- **`Makefile`**: 40+ targets including fast builds, caching, and container management

## Modern Architecture Patterns

### Structured Logging System (NEW)
**Package**: `internal/logger` - Replaces all `fmt.Printf` statements  
**Implementation**: Go's `log/slog` with color and emoji support

```go
// Logger usage pattern throughout codebase
logger := logger.New(&logger.Config{
    Level:        logger.LevelInfo,
    EnableColors: true,
    EnableEmojis: true,
    Component:    "analyzer",
})

logger.Info("Analysis complete", 
    slog.Int("clients", clientCount),
    slog.String("status", "success"))
```

### Data Flow (API-Only)
1. **Input**: Pi-hole configuration file with API credentials
2. **Processing**: Use Pi-hole API to query DNS records with structured logging
3. **Analysis**: Aggregate into `types.ClientStats` with network analysis
4. **Output**: Colorized terminal display + optional file reports
5. **Logging**: Structured logs with contextual information throughout

### Fast Builds System (NEW)
**Performance Improvements**:
- **Cold builds**: 20-30% faster through optimized flags
- **Warm builds**: 60-80% faster through comprehensive caching  
- **CI builds**: 50-70% faster through cache restoration
- **Docker builds**: 40-60% faster through multi-stage optimization

**Key Build Targets**:
```bash
make fast-build    # Optimized incremental build with timing
make cache-warm    # Pre-populate build caches
make cache-info    # Display cache status and sizes
make dev-setup     # Complete development environment setup
```

### Container Infrastructure (NEW)
**Multi-Architecture Support**: AMD64, ARM64, ARMv7 (Raspberry Pi compatible)  
**Container Variants**:
- **Production**: `ghcr.io/grammatonic/pihole-analyzer:latest` (~44MB)
- **Development**: `ghcr.io/grammatonic/pihole-analyzer:latest-development` (~45MB)

**Container Deployment Patterns**:
```bash
# Quick deployment
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help

# Development environment
make docker-dev

# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Core Data Structures

#### `types.PiholeRecord`
```go
type PiholeRecord struct {
    ID        int    `json:"id"`        // Query ID
    DateTime  string `json:"datetime"`  // Query timestamp
    Domain    string `json:"domain"`    // Queried domain name
    Client    string `json:"client"`    // Client IP address
    QueryType string `json:"querytype"` // DNS query type (A, AAAA, etc)
    Status    int    `json:"status"`    // Pi-hole status code
    Timestamp string `json:"timestamp"` // Unix timestamp
    HWAddr    string `json:"hwaddr"`    // Hardware/MAC address
}
```

#### `types.ClientStats`
```go
type ClientStats struct {
    IP            string            // Client IP address
    Hostname      string            // Client hostname
    QueryCount    int               // Total number of queries
    Domains       map[string]int    // Domain query counts
    DomainCount   int               // Number of unique domains
    MACAddress    string            // MAC address
    IsOnline      bool              // Online status from ARP
    LastSeen      string            // Last seen timestamp
    TopDomains    []DomainStat      // Top queried domains
    Status        string            // Connection status
    UniqueQueries int               // Unique query count
    TotalQueries  int               // Total query count
    FirstSeen     string            // First seen timestamp
    DeviceType    string            // Device type identification
    // Additional analysis fields
    Client         string           // Client identifier
    QueryTypes     map[int]int      // Query type counts
    StatusCodes    map[int]int      // Status code counts
    HWAddr         string           // Hardware address
    ARPStatus      string           // ARP table status
    TotalReplyTime float64          // Total reply time
    AvgReplyTime   float64          // Average reply time
    Uniquedomains  int              // Unique domain count
}
```

#### `types.Config` (API-Only)
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
    Host string `json:"host"`
    Port int    `json:"port"`
    
    // API Configuration (only method)
    APIEnabled  bool   `json:"api_enabled"`
    APIPassword string `json:"api_password"`
    APITOTP     string `json:"api_totp"`
    UseHTTPS    bool   `json:"use_https"`
    APITimeout  int    `json:"api_timeout"`
}
```

### Configuration Management
- **File**: `~/.pihole-analyzer/config.json` (default)
- **Structure**: `types.Config` with nested structs for Pi-hole API, exclusions, output, logging
- **Defaults**: Comprehensive defaults in `config.DefaultConfig()`
- **Logging Config**: Structured logging levels, colors, emoji, output file support

## Enhanced Development Workflow

### Data Source Interface (NEW)
**Package**: `internal/interfaces` - Provides abstraction for Pi-hole data access

```go
type DataSource interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    IsConnected() bool
    
    // Core data retrieval (API-only implementation)
    GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
    GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error)
    GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error)
    GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error)
    
    // Performance and metadata
    GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error)
    GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error)
    
    // Configuration and metadata
    GetDataSourceType() DataSourceType
    GetConnectionInfo() *ConnectionInfo
}
```

### Pi-hole API Client (NEW)
**Package**: `internal/pihole` - Direct Pi-hole API integration

```go
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    SID        string
    CSRFToken  string
    Logger     *logger.Logger
    config     *Config
}

// API client with session management and 2FA support
func NewClient(config *Config, log *logger.Logger) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
```

### CLI Package (NEW)
**Package**: `internal/cli` - Centralized command-line flag management

```go
type Flags struct {
    OnlineOnly   *bool
    NoExclude    *bool
    Pihole       *string
    Config       *string
    NoColor      *bool
    NoEmoji      *bool
    Quiet        *bool
    CreateConfig *bool
    ShowConfig   *bool
    PiholeSetup  *bool
}
```

### Build System (Advanced Makefile - 40+ Targets)
```bash
# Fast development builds
make fast-build         # Optimized incremental build with timing
make build-cached       # Build only if sources changed
make watch             # Auto-rebuild on file changes (requires entr)

# Cache management
make cache-warm        # Pre-populate build caches for faster builds  
make cache-info        # Display cache sizes and status
make cache-clean       # Clean Go build and module caches

# Development environment
make dev-setup         # Complete development environment setup
make benchmark         # Performance benchmarking
make analyze-size      # Binary size analysis

# Container workflows
make docker-build      # Build optimized Docker image
make docker-dev        # Start development container environment
make docker-multi      # Multi-architecture builds (AMD64, ARM64, ARMv7)
make docker-api-only   # Build API-only container variant

# Testing (enhanced)
make ci-test          # CI-compatible test suite with caching
make test-cached      # Cached test execution
make pre-push         # Comprehensive pre-push tests
make release-build    # Full release build validation

# Pi-hole specific
make analyze-pihole   # Analyze Pi-hole live data
make setup-pihole     # Setup Pi-hole API configuration
make test-pihole      # Test Pi-hole connection
```

### Testing Strategy (Enhanced)
- **Dual Binary System**: `pihole-analyzer` (production), `pihole-analyzer-test` (development)
- **Unit Tests**: Go standard testing with structured logging validation
- **Integration Tests**: `scripts/integration-test.sh` with Pi-hole scenarios
- **Container Tests**: Multi-architecture validation in CI/CD
- **CI/CD**: GitHub Actions with comprehensive caching and parallel builds
- **Mock Data**: Comprehensive test fixtures in `testing/fixtures/`

### Code Quality Standards
- **Current Grade**: A- (significantly improved from previous B+)
- **Structured Logging**: Complete migration from `fmt.Printf` to `log/slog`
- **Modular Architecture**: Separated production and test binaries
- **Container Ready**: Production-grade containerization implemented
- **Standards**: Go formatting, comprehensive error handling, structured logging
- **Dependencies**: Minimal external dependencies (sqlite, crypto)

## Key Features & Modern Implementation

### Structured Logging System (CRITICAL)
**Package**: `internal/logger` - **Replaces all `fmt.Printf` statements**  
**Migration Complete**: No more direct fmt.Printf usage in codebase

```go
// Correct logging pattern (use throughout codebase)
logger := logger.New(&logger.Config{
    Level:        logger.LevelInfo,
    EnableColors: true,
    EnableEmojis: true,
    Component:    "pihole-api",
})

// Structured logging with context
logger.Info("Pi-hole API connection established",
    slog.String("host", config.Host),
    slog.Int("port", config.Port),
    slog.Bool("https", config.UseHTTPS))

// Error logging with context
logger.Error("API query failed",
    slog.String("query", apiQuery),
    slog.String("error", err.Error()))
```

### Colorized Output System (Enhanced)
- **Package**: `internal/colors`
- **Features**: Cross-platform terminal colors, emoji support, smart domain highlighting
- **Integration**: Works seamlessly with structured logging
- **Flags**: `--no-color`, `--no-emoji` for compatibility
- **Configuration**: Configurable via JSON config and logger configuration

### Container Infrastructure (Production Ready)
**Multi-Architecture Dockerfile**: 
```dockerfile
# Cross-platform build support
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build...
```

**Container Variants**:
- **Production**: Minimal runtime, optimized for deployment
- **Development**: Includes test utilities and debugging tools
- **Registry**: `ghcr.io/grammatonic/pihole-analyzer` with automated publishing

### Fast Builds & Caching (Performance Critical)
**Cache Strategy**:
- **Go Modules**: Cached by `go.sum` hash in CI/CD
- **Build Cache**: Preserved between builds for 60-80% speedup
- **Container Layers**: Multi-stage builds with dependency caching
- **Local Development**: Persistent caches via Docker volumes

**Build Performance**:
```bash
# Timing examples
make cache-info     # Shows: "Build cache: 323M, Module cache: 2.2G"
make fast-build     # Typical output: "✅ Fast build completed in 3s"
make docker-build   # Typical output: "✅ Docker build completed in 21s"
```

### Pi-hole API Connection (Enhanced)
```go
// Modern API connection pattern with structured logging
logger := logger.New(&logger.Config{Component: "pihole-api"})

client := pihole.NewClient(&pihole.Config{
    Host:        config.Pihole.Host,
    Port:        config.Pihole.Port,
    Password:    config.Pihole.APIPassword,
    UseHTTPS:    config.Pihole.UseHTTPS,
    Timeout:     time.Duration(config.Pihole.APITimeout) * time.Second,
}, logger)

// Always log connection attempts
logger.Info("Establishing Pi-hole API connection",
    slog.String("host", config.Pihole.Host),
    slog.Int("port", config.Pihole.Port),
    slog.Bool("https", config.Pihole.UseHTTPS))
```

### Network Analysis (Enhanced)
- **ARP Table**: Determines online/offline status via MAC address lookup
- **Exclusions**: Configurable network/IP/hostname exclusions (Docker, loopback, etc.)
- **DNS Query Analysis**: Status codes, query types, domain categorization
- **Logging**: All network operations logged with structured logging

## Common Tasks & Modern Patterns

### Implementing New Features (Updated Process)
1. **Use Structured Logging**: Never use `fmt.Printf` - always use `internal/logger`
2. **Container Awareness**: Consider both production and development container usage
3. **Build Performance**: Ensure changes don't break caching strategies
4. **Multi-Binary Support**: Consider impact on both `pihole-analyzer` and `pihole-analyzer-test`

### Adding New CLI Flags
1. Declare in `internal/cli/flags.go` for centralized flag management
2. Add to flag parsing with structured logging for validation
3. Update help text consistently
4. Handle in configuration logic with logging
5. Test in both container and local environments

### Adding New Analysis Features
1. Extend `types.ClientStats` or `types.PiholeRecord` if needed
2. Implement analysis logic in analyzer package with structured logging
3. Add colorized output in reporting package
4. Include in report generation with appropriate logging
5. Test with both binaries and in container environments

### Configuration Changes (Enhanced)
1. Update `types.Config` structure
2. Modify `config.DefaultConfig()` 
3. Handle in JSON marshaling/unmarshaling
4. Update validation logic with structured logging
5. Test in container environments and with persistent volumes

### Testing New Features (Comprehensive)
1. Add unit tests in appropriate package with logging validation
2. Update integration test scenarios for both binaries
3. Test Pi-hole API connectivity and mock environments
4. Verify colorized output and structured logging
5. Test container builds and multi-architecture support
6. Validate caching doesn't break with changes

## Dependencies & External Libraries (Updated)

### Core Dependencies
- **`modernc.org/sqlite`**: SQLite database access (pure Go)
- **`log/slog`**: Structured logging (Go 1.23+ standard library)
- **Standard Library**: Heavy use of net, io, encoding packages

### Development Dependencies
- **Testing**: Standard Go testing framework
- **CI/CD**: GitHub Actions with advanced caching and multi-architecture builds
- **Containers**: Docker with BuildKit, multi-stage builds
- **Build**: Make-based build system with 40+ targets

## Common Pitfalls & Modern Solutions

### 1. Logging Anti-Patterns (CRITICAL)
**Problem**: Using `fmt.Printf` for logging  
**Solution**: **ALWAYS** use `internal/logger` with structured logging
```go
// ❌ NEVER do this
fmt.Printf("Error: %v\n", err)

// ✅ ALWAYS do this
logger.Error("Operation failed", slog.String("error", err.Error()))
```

### 2. Container Development Workflow
**Problem**: Inconsistent development between local and container environments  
**Solution**: Use dual development approach
```bash
# Local development with caching
make dev-setup && make fast-build

# Container development
make docker-dev  # Persistent Go caches, live development
```

### 3. Build Performance Degradation
**Problem**: Changes breaking cache effectiveness  
**Solution**: Always validate build performance impact
```bash
make cache-info      # Check cache status before/after changes
make cache-warm      # Restore optimal cache state
make fast-build     # Verify performance maintained
```

### 4. Binary Confusion (NEW)
**Problem**: Confusion between production and test binaries  
**Solution**: Clear binary separation
- `pihole-analyzer`: Production use, requires real Pi-hole
- `pihole-analyzer-test`: Development/testing, includes mock data

### 5. Multi-Architecture Compatibility
**Problem**: Code working on one architecture but failing on others  
**Solution**: Container-based testing
```bash
make docker-build-multi  # Test AMD64, ARM64, ARMv7
```

## Refactoring Opportunities & Current Status

### Completed Improvements ✅
1. **✅ Structured Logging Migration**: Complete replacement of `fmt.Printf` with `log/slog`
2. **✅ Binary Separation**: Production and test binaries properly separated
3. **✅ Fast Builds Implementation**: 60-80% build speed improvement achieved
4. **✅ Container Infrastructure**: Production-ready multi-architecture containers
5. **✅ Enhanced CI/CD**: Advanced caching and parallel builds implemented

### High Priority (Remaining)
1. **Enhanced Configuration Validation** - Add comprehensive config validation with structured logging
2. **Performance Metrics Collection** - Built-in metrics for monitoring
3. **Multi-Pi-hole Support** - Connect to multiple Pi-hole instances

## Current API-Only Architecture

### Overview
**Implementation Status**: Complete transition to Pi-hole API-only architecture. SSH functionality has been completely removed and replaced with direct Pi-hole API integration.

### Current Architecture Benefits
- **Enhanced Security**: No SSH access required, controlled API permissions only
- **Improved Reliability**: Direct API communication with built-in retry logic
- **Better Performance**: Optimized API queries vs direct database access
- **Simplified Configuration**: Single API-based configuration structure
- **Real-time Data**: More up-to-date information than database snapshots

### API Client Implementation
**Package**: `internal/pihole` - Complete Pi-hole API client with session management

```go
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    SID        string
    CSRFToken  string
    Logger     *logger.Logger
    config     *Config
}

// API client features:
// - Session-based authentication with automatic refresh
// - 2FA TOTP support
// - HTTPS/HTTP with certificate validation
// - Comprehensive error handling and retry logic
// - Structured logging throughout
```

### Data Source Interface
**Package**: `internal/interfaces` - Abstraction layer for data access

```go
type DataSource interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    IsConnected() bool
    
    // Core data retrieval (API implementation)
    GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
    GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error)
    GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error)
    GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error)
    
    // Performance and metadata
    GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error)
    GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error)
    
    // Configuration and metadata
    GetDataSourceType() DataSourceType
    GetConnectionInfo() *ConnectionInfo
}
```

### Current Configuration Structure
```go
type PiholeConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
    
    // API Configuration (only method)
    APIEnabled  bool   `json:"api_enabled"`
    APIPassword string `json:"api_password"`
    APITOTP     string `json:"api_totp"`
    UseHTTPS    bool   `json:"use_https"`
    APITimeout  int    `json:"api_timeout"`
}
```

### CLI Integration
**Package**: `internal/cli` - Centralized flag management for API-only operation

```bash
# Current CLI flags (API-only)
--pihole          # Pi-hole config file path
--config          # Application config file
--create-config   # Create default configuration
--show-config     # Display current configuration
--pihole-setup    # Pi-hole configuration wizard
--online-only     # Show only online clients
--no-exclude      # Disable default exclusions
--no-color        # Disable colored output
--no-emoji        # Disable emoji output
--quiet           # Suppress non-essential output
```

### Medium Priority
1. **Add Prometheus metrics endpoints** for monitoring
2. **Support multiple output formats** (JSON, XML) beyond terminal
3. **Enhanced network analysis** capabilities

## Integration Points (Enhanced)

### CI/CD Pipeline (Production Ready)
- **Speed**: 50-70% faster builds with multi-layer caching
- **Artifacts**: Binary artifacts shared between jobs
- **Container Publishing**: Automated GHCR publishing with security scanning
- **GitHub Actions**: `.github/workflows/ci.yml` and `.github/workflows/container.yml`
- **Multi-Architecture**: Parallel AMD64, ARM64, ARMv7 builds
- **Performance Monitoring**: Build timing and cache hit rate reporting

### External Systems (Enhanced)
- **Pi-hole**: Direct API access with session management and structured logging
- **Container Registries**: GitHub Container Registry (GHCR) with automated publishing
- **Development Environments**: Docker Compose with persistent caches
- **Build Systems**: Enhanced Makefile with 40+ targets and performance monitoring
- **ARP Tables**: System ARP command execution with logging
- **File System**: Configuration, reports, and persistent container volumes

### Container Orchestration (NEW)
- **Docker**: Multi-stage builds with BuildKit optimization
- **Docker Compose**: Separate dev/prod configurations with persistent caches
- **Registry**: GHCR with automated security scanning and SBOM generation

## Debugging & Troubleshooting (Enhanced)

### Common Debug Flags
```bash
# Local debugging
--quiet              # Suppress non-essential output
--no-color           # Disable colors for log analysis
--no-emoji           # Disable emojis for cleaner logs
--test               # Use mock data (pihole-analyzer-test binary)
--show-config        # Display current configuration

# Container debugging
make docker-shell    # Access development container
make docker-logs     # View container logs
docker exec -it pihole-analyzer-dev sh  # Interactive container access
```

### Structured Log Analysis
```bash
# View structured logs with context
./pihole-analyzer --pihole config.json 2>&1 | grep "level=ERROR"

# Container log analysis
docker logs pihole-analyzer-prod 2>&1 | jq '.level' | sort | uniq -c
```

### Performance Debugging
```bash
# Build performance analysis
make cache-info      # Check cache utilization
make analyze-size    # Binary size analysis
make benchmark       # Performance benchmarking

# Container performance
docker stats pihole-analyzer-prod
```

### Log Analysis Patterns
- **Structured Logs**: All logs include level, timestamp, component, and context
- **Color Support**: Automatic detection with fallback for CI/containers
- **File Output**: Optional log file output for persistence
- **Container Logs**: Aggregated via Docker logging drivers

## Future Roadmap (Updated)

### Planned Features (Short Term)
1. **Enhanced Configuration Validation** - Add comprehensive config validation with structured logging
2. **Performance Metrics Collection** - Built-in metrics for monitoring
3. **Multi-Pi-hole Support** - Connect to multiple Pi-hole instances
4. **Multi-Pi-hole Support** - Connect to multiple Pi-hole instances

### Architecture Evolution (Medium Term)
1. **Real-time Monitoring** - Live dashboard capabilities
4. **Plugin System** - Extensible analysis modules

### Container Ecosystem (Advanced)
1. **Multi-Registry Support** - Docker Hub, ACR, ECR publishing
2. **ARM32 Support** - Additional Raspberry Pi architectures

---

## Quick Reference for AI Assistants (Updated)

When working on this project:

### CRITICAL Requirements
1. **🚨 NEVER use `fmt.Printf`** - Always use `internal/logger` with structured logging
2. **🔧 Test both binaries** - `pihole-analyzer` (production) and `pihole-analyzer-test` (development)
3. **🐳 Validate container builds** - Changes must work in containerized environments
4. **⚡ Preserve build performance** - Check cache impact with `make cache-info`
5. **🏗️ Follow dual-binary architecture** - Production/test separation is intentional
6. **🔒 Security first** - Non-root containers, minimal attack surface
7. **🌐 Cross-platform support** - AMD64, ARM64, ARMv7 architectures
8. **📜 Maintain structured logging** - Use `slog` with colors and emojis
9. **🚀 Fast builds** - Use `make fast-build` for quick iterations
10. **📦 Container first** - All features must work in containerized environments
11. **📋 Comprehensive unit tests** - Ensure all new features are covered
12. **🔗 Robust integration tests** - Validate Pi-hole connectivity and mock environments
13. **📚 Documentation updates** - Keep documentation current and comprehensive
14. **🧪 Test directory structure** - Use both `tests/` and `testing/` directories appropriately

### Development Workflow
1. **Start with**: `make dev-setup` for complete environment preparation
2. **Fast iteration**: `make fast-build` for quick development cycles
3. **Container testing**: `make docker-dev` for containerized development
4. **Before committing**: `make pre-push` to validate all tests pass
5. **Performance check**: `make cache-info` to verify cache effectiveness
6. **Documentation updates**: Keep documentation files current and comprehensive

### Code Patterns
- **Logging**: Use structured logging with context (`slog.String()`, `slog.Int()`, etc.)
- **Configuration**: Always validate config with appropriate logging
- **API**: Include connection logging with host/port context
- **Errors**: Structured error logging with full context
- **Colors**: Respect `--no-color` and `--no-emoji` flags

### Architecture Principles
- **Modular Design**: Use internal packages for separation of concerns  
- **Container First**: All features must work in containerized environments
- **Performance Aware**: Consider build cache impact of changes
- **Security Focused**: Non-root containers, minimal attack surface
- **Cross-Platform**: Support AMD64, ARM64, ARMv7 architectures

This project emphasizes **structured logging**, **fast builds**, **containerization**, and **beautiful terminal output** - maintain these core values when implementing changes.
