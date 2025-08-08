# Pi-hole Network Analyzer - AI Coding Assistant Instructions

This file provides comprehensive guidance for AI coding assistants working on the Pi-hole Network Analyzer project.

## Project Overview

**Repository**: Pi-hole Network Analyzer  
**Language**: Go 1.23+  
**Module**: `pihole-analyzer`  
**Binary Names**: `pihole-analyzer` (production), `pihole-analyzer-test` (development)
**Commands Directory**: `cmd/pihole-analyzer/`, `cmd/pihole-analyzer-test/`  
**Main Files**: `cmd/pihole-analyzer/main.go`, `cmd/pihole-analyzer-test/main.go`
**Help Command**: `pihole-analyzer --help`  
**Architecture**: Standard Go Project Layout with comprehensive containerization

### Core Purpose
**Pi-hole-focused** DNS usage analysis tool with SSH connectivity. Features structured logging, colorized terminal output, network analysis, and comprehensive reporting. Now includes production-ready containerization and optimized build system.

**🚨 IMPORTANT**: 
- CSV functionality completely removed as of August 2025
- SSH functionality is completely removed as of August 2025
- Getting data from Pi-Hole is now done via API calls instead of direct database access
- Structured logging implemented with Go's `log/slog` package
- Comprehensive Docker/container support added
- Fast builds with advanced caching strategies implemented

## Project Structure & Modern Enhancements

### Directory Layout
```
/
├── cmd/
│   ├── pihole-analyzer/          # Production binary entry point
│   └── pihole-analyzer-test/     # Test/development binary with mock data
├── internal/                     # Private application packages
│   ├── analyzer/                 # Pi-hole data analysis engine
│   ├── cli/                      # Command-line interface
│   ├── colors/                   # Terminal colorization with cross-platform support
│   ├── config/                   # Configuration management
│   ├── logger/                   # Structured logging with slog integration
│   ├── network/                  # Network analysis & ARP integration
│   ├── reporting/                # Output display & formatted reports
│   ├── ssh/                      # Pi-hole SSH connectivity & database access
│   └── types/                    # Core data structures
├── docs/                         # Comprehensive documentation
│   ├── fast-builds.md           # Build optimization guide
│   ├── container-registry.md    # Container deployment strategy
│   └── container-usage.md       # Docker usage guide
├── scripts/                      # Build automation & cache warming
├── testing/                      # Test utilities and fixtures
├── .github/workflows/           # CI/CD with advanced caching
├── Dockerfile                   # Multi-stage, multi-arch container builds
├── docker-compose*.yml         # Development and production environments
└── Makefile                     # Enhanced build system (40+ targets)
```

### Key Files & Their Roles
- **`cmd/pihole-analyzer/main.go`**: Production application entry point
- **`cmd/pihole-analyzer-test/main.go`**: Development/testing entry point with mock data
- **`internal/logger/logger.go`**: Structured logging with slog, colors, and emojis
- **`internal/analyzer/analyzer.go`**: Pi-hole data analysis engine
- **`internal/ssh/pihole.go`**: SSH connection and database analysis
- **`internal/types/types.go`**: Core data structures (ClientStats, PiholeRecord)
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

### Data Flow (Pi-hole Only)
1. **Input**: SSH connection to Pi-hole SQLite database
2. **Processing**: Query Pi-hole database for DNS records with structured logging
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
    Timestamp string   // Unix timestamp
    Client    string   // Client IP address
    HWAddr    string   // Hardware/MAC address
    Domain    string   // Queried domain
    Status    int      // Pi-hole status code
}
```

#### `types.ClientStats`
```go
type ClientStats struct {
    Client       string
    Hostname     string
    HardwareAddr string
    IsOnline     bool
    TotalQueries int
    UniqueQueries int
    AvgReplyTime float64
    Domains      map[string]int
    QueryTypes   map[int]int
    StatusCodes  map[int]int
    TopDomains   []DomainCount
}
```

### Configuration Management
- **File**: `~/.pihole-analyzer/config.json` (default)
- **Structure**: `types.Config` with nested structs for Pi-hole, exclusions, output, logging
- **Defaults**: Comprehensive defaults in `config.DefaultConfig()`
- **SSH Support**: Key-based and password authentication
- **Logging Config**: Structured logging levels, colors, emoji, output file support

## Enhanced Development Workflow

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

# Testing (enhanced)
make ci-test          # CI-compatible test suite with caching
make test-cached      # Cached test execution
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
- **Dependencies**: Minimal external dependencies (ssh, sqlite, crypto)

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
    Component:    "ssh",
})

// Structured logging with context
logger.Info("SSH connection established",
    slog.String("host", config.Host),
    slog.Int("port", config.Port),
    slog.String("user", config.Username))

// Error logging with context
logger.Error("Database query failed",
    slog.String("query", sqlQuery),
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

### SSH Pi-hole Connection (Enhanced)
```go
// Modern SSH connection pattern with structured logging
logger := logger.New(&logger.Config{Component: "ssh"})

sshConfig := &ssh.ClientConfig{
    User: config.Pihole.Username,
    Auth: []ssh.AuthMethod{
        ssh.PublicKeys(signer),
        ssh.Password(config.Pihole.Password),
    },
    HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}

// Always log connection attempts
logger.Info("Establishing SSH connection",
    slog.String("host", config.Pihole.Host),
    slog.Int("port", config.Pihole.Port))
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
1. Declare in appropriate main.go (production or test binary)
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
3. Test Pi-hole connectivity and mock environments
4. Verify colorized output and structured logging
5. Test container builds and multi-architecture support
6. Validate caching doesn't break with changes

## Dependencies & External Libraries (Updated)

### Core Dependencies
- **`golang.org/x/crypto/ssh`**: SSH client for Pi-hole connections
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
1. **Use PiHole Official API** - Integrate with Pi-hole official API for enhanced data access
2. **Configuration validation enhancement** - Add comprehensive config validation
3. **Performance monitoring integration** - Add metrics collection

## Pi-hole API Integration Implementation Plan

### Overview
**Primary Goal**: Replace SSH-based database access with Pi-hole's official REST API while maintaining 100% feature parity and backward compatibility during transition. This modernization will provide enhanced security, reliability, and performance while eliminating the need for SSH access to Pi-hole systems.

### Current SSH Architecture Analysis
- **Current Method**: SSH connection → Direct SQLite database queries → Manual data parsing
- **Data Access**: Raw access to `/etc/pihole/pihole-FTL.db` via SSH tunnel
- **Authentication**: SSH key-based or password authentication with server access requirements
- **Data Processing**: Custom SQL queries for DNS records, client statistics, and network analysis
- **Limitations**: 
  - Requires SSH server access and credentials
  - Direct database dependency creates fragility
  - Manual query optimization and data parsing
  - Network connectivity issues with SSH tunneling
  - Security concerns with database-level access

### Pi-hole API Capabilities & Feature Mapping
Based on official documentation (https://docs.pi-hole.net/api/):
- **REST API**: Standard HTTP/HTTPS with structured JSON responses
- **Authentication**: Session-based SID tokens with optional 2FA support
- **Complete Data Access**: All SSH functionality available through API endpoints
- **Enhanced Security**: Controlled permissions, no direct database access required
- **Built-in Features**: Rate limiting, session management, CSRF protection
- **Self-Documentation**: Complete API docs at `http://pi.hole/api/docs`

**Critical Requirement**: The API implementation must provide 100% of current SSH functionality including:
- Complete DNS query history access
- Client statistics and analysis
- Domain resolution data
- Network device information
- Real-time and historical data
- All current filtering and analysis capabilities

### Implementation Strategy

#### Phase 1: Complete SSH Feature Analysis & API Client Foundation
**Goal**: Catalog all current SSH functionality and build API client with equivalent capabilities

**SSH Feature Inventory**:
1. **Database Queries** (`internal/ssh/pihole.go`):
   - DNS query history retrieval
   - Client-based query filtering
   - Domain statistics aggregation
   - Time-based query filtering
   - Query type and status analysis

2. **Data Processing** (`internal/analyzer/analyzer.go`):
   - Client statistics calculation
   - Network device identification
   - ARP table correlation
   - Domain categorization and ranking
   - Performance metrics (reply times, query counts)

3. **Network Analysis**:
   - MAC address resolution
   - Online/offline status determination
   - Hardware address mapping
   - Device type identification

**API Client Implementation** (`internal/pihole/client.go`):
```go
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    SID        string
    CSRFToken  string
    Logger     *logger.Logger
    config     *Config
}

// Must provide 100% SSH functionality equivalents
func (c *Client) GetDNSQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
func (c *Client) GetClientStatistics(ctx context.Context) (map[string]*types.ClientStats, error)
func (c *Client) GetNetworkDevices(ctx context.Context) ([]types.NetworkDevice, error)
func (c *Client) GetDomainStatistics(ctx context.Context) (*types.DomainStats, error)
```

#### Phase 2: Data Source Interface & SSH Replacement Strategy
**Goal**: Create unified interface that abstracts SSH and API, ensuring identical data output

**Data Source Abstraction** (`internal/interfaces/data_source.go`):
```go
type DataSource interface {
    Connect(ctx context.Context) error
    
    // Core SSH functionality - must be 100% equivalent
    GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
    GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error)
    GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error)
    GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error)
    
    // Performance and metadata - must match SSH implementation
    GetQueryPerformance(ctx context.Context) (*types.QueryPerformance, error)
    GetConnectionStatus(ctx context.Context) (*types.ConnectionStatus, error)
    
    Close() error
}

// Data format compatibility - ensure identical output structures
type QueryParams struct {
    StartTime    time.Time
    EndTime      time.Time
    ClientFilter string
    DomainFilter string
    Limit        int
    StatusFilter []int    // Match SSH status code filtering
    TypeFilter   []int    // Match SSH query type filtering
}
```

**Implementation Requirements**:
1. **SSH Data Source** (`internal/ssh/datasource.go`) - Wrapper around existing SSH code
2. **API Data Source** (`internal/pihole/datasource.go`) - New API implementation with identical output
3. **Data Validation** - Ensure API responses match SSH data structures exactly

#### Phase 3: Complete SSH Feature Replacement
**Goal**: Implement every SSH database query through Pi-hole API endpoints

**SSH Function Mapping to API Endpoints**:

1. **DNS Query History** (Replace SSH database queries):
   ```sql
   -- Current SSH Query (to be replaced)
   SELECT timestamp, client, domain, status FROM queries 
   WHERE timestamp BETWEEN ? AND ?
   ```
   ```go
   // API Replacement - must return identical data structure
   GET /api/queries?from={timestamp}&until={timestamp}
   ```

2. **Client Statistics** (Replace SSH aggregation):
   ```sql
   -- Current SSH Queries (to be replaced)
   SELECT client, COUNT(*) as total FROM queries GROUP BY client
   SELECT client, domain, COUNT(*) FROM queries GROUP BY client, domain
   ```
   ```go
   // API Replacement - must provide same statistics
   GET /api/stats/query_types_over_time
   GET /api/stats/top_clients
   ```

3. **Domain Analysis** (Replace SSH domain queries):
   ```sql
   -- Current SSH Query (to be replaced)
   SELECT domain, COUNT(*) FROM queries GROUP BY domain ORDER BY COUNT(*) DESC
   ```
   ```go
   // API Replacement - must match sorting and counting
   GET /api/stats/top_domains
   GET /api/stats/query_types
   ```

4. **Network Device Information** (Replace SSH network queries):
   ```sql
   -- Current SSH Query (to be replaced)
   SELECT DISTINCT client FROM queries
   ```
   ```go
   // API Replacement - must include all client information
   GET /api/network
   GET /api/clients
   ```

**Critical Implementation Requirements**:
- Each API implementation must return data in **identical format** to SSH version
- All filtering, sorting, and aggregation logic must produce **same results**
- Performance metrics (query counts, response times) must be **equivalent**
- Network analysis (ARP correlation, device detection) must be **preserved**

#### Phase 4: SSH-to-API Migration Strategy
**Goal**: Complete replacement of SSH functionality with seamless transition

**Enhanced Config Structure** (Transition-focused):
```go
type PiholeConfig struct {
    // Existing SSH fields (to be deprecated)
    Host         string `json:"host"`
    Port         int    `json:"port"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    
    // New API fields (primary method)
    APIEnabled   bool   `json:"api_enabled"`
    APIPassword  string `json:"api_password"`
    APITOTP      string `json:"api_totp"`
    UseHTTPS     bool   `json:"use_https"`
    APITimeout   int    `json:"api_timeout"`
    
    // Migration control
    UseSSH       bool   `json:"use_ssh"`        // Fallback only
    DatabasePath string `json:"database_path"`  // Legacy support
    MigrationMode string `json:"migration_mode"` // "api-first", "ssh-only", "auto"
}
```

**Migration Phases**:
1. **Phase 4a**: API implementation with SSH fallback
2. **Phase 4b**: API-first with SSH backup
3. **Phase 4c**: API-only with SSH deprecation warnings
4. **Phase 4d**: Complete SSH removal (future release)

#### Phase 5: Complete SSH Replacement & Analyzer Integration
**Goal**: Finalize API-first architecture with SSH removal path

**Enhanced Analyzer** (`internal/analyzer/analyzer.go`):
```go
type Analyzer struct {
    dataSource interfaces.DataSource
    config     *types.Config
    logger     *logger.Logger
}

func (a *Analyzer) AnalyzeData(ctx context.Context) (*types.AnalysisResult, error) {
    // Universal analysis logic - identical results regardless of data source
    // Must produce identical output whether using API or SSH
}
```

**SSH Deprecation Strategy**:
1. **Warning Messages**: Structured logging warnings about SSH deprecation
2. **Feature Flags**: Gradual removal of SSH-specific code paths
3. **Documentation Updates**: Migration guides and API-first examples
4. **Container Builds**: API-only variants for new deployments

### Implementation Details

#### Error Handling & Resilience
1. **Connection Retry Logic**:
   ```go
   type RetryConfig struct {
       MaxRetries int
       Backoff    time.Duration
       MaxBackoff time.Duration
   }
   ```

2. **Fallback Strategy**:
   - Primary: Pi-hole API
   - Fallback: SSH database access
   - Configuration-driven priority

3. **Graceful Degradation**:
   - Handle API endpoint unavailability
   - Partial data scenarios
   - Network connectivity issues

#### Security Implementation
1. **Session Management**:
   - Automatic session refresh
   - Secure token storage
   - Session cleanup on exit

2. **HTTPS Support**:
   - TLS certificate validation
   - Option for self-signed certificates
   - Secure communication channels

3. **Authentication Flows**:
   - Standard password authentication
   - 2FA TOTP integration
   - Application password support

#### Testing Strategy
1. **Mock API Server** (`testing/mock_pihole_api.go`):
   ```go
   type MockAPIServer struct {
       responses map[string]interface{}
       delays    map[string]time.Duration
   }
   ```

2. **Integration Tests**:
   - Real Pi-hole API testing
   - SSH fallback scenarios
   - Configuration validation

3. **Test Data**:
   - Enhanced mock data for API responses
   - Realistic query patterns
   - Error scenario simulation

### Migration Strategy

#### Backward Compatibility
1. **Configuration Migration**:
   - Auto-detect SSH vs API configuration
   - Configuration file migration tool
   - Validation of mixed configurations

2. **CLI Flag Enhancement**:
   ```bash
   --api-mode           # Force API mode
   --ssh-mode           # Force SSH mode (current behavior)
   --auto-detect        # Auto-detect best method (default)
   ```

3. **Feature Parity**:
   - Ensure API implementation provides same data as SSH
   - Maintain output format consistency
   - Preserve existing CLI behavior

#### Rollout Plan
1. **Phase 1**: API client implementation with tests
2. **Phase 2**: Data source abstraction and API integration
3. **Phase 3**: Configuration enhancement and migration tools
4. **Phase 4**: Integration testing and documentation
5. **Phase 5**: Default to API with SSH fallback

### Documentation Requirements

#### User Documentation
1. **API Setup Guide** (`docs/api-setup.md`):
   - Pi-hole API configuration
   - Authentication setup
   - 2FA configuration

2. **Migration Guide** (`docs/migration-ssh-to-api.md`):
   - Configuration file updates
   - Testing connectivity
   - Troubleshooting guide

3. **Configuration Reference** (`docs/configuration.md` updates):
   - New API configuration options
   - Security best practices
   - Performance considerations

#### Developer Documentation
1. **API Client Usage** (`docs/development.md` updates):
   - Data source interface usage
   - Adding new endpoints
   - Testing with mock server

2. **Architecture Documentation**:
   - Data flow diagrams
   - Security architecture
   - Error handling patterns

### Benefits of API Integration

#### Enhanced Capabilities
1. **Real-time Data**: More up-to-date information than database snapshots
2. **Better Security**: No SSH access required, controlled API permissions
3. **Reliability**: Built-in retry logic, session management
4. **Performance**: Optimized data queries vs direct database access

#### Future Enhancements Enabled
1. **Live Monitoring**: Real-time query streaming
2. **Configuration Management**: Modify Pi-hole settings via API
3. **Multi-Pi-hole Support**: Easier to connect to multiple instances
4. **Advanced Analytics**: Access to additional Pi-hole metrics

### Implementation Checklist

#### Phase 1: Foundation
- [ ] Create `internal/pihole` package structure
- [ ] Implement basic HTTP client with authentication
- [ ] Add session management with automatic refresh
- [ ] Create comprehensive error handling

#### Phase 2: Data Integration
- [ ] Enhance `interfaces.DataSource` interface
- [ ] Implement API-based data source
- [ ] Create data transformation utilities
- [ ] Add structured logging throughout

#### Phase 3: Configuration
- [ ] Extend configuration structures
- [ ] Add configuration validation
- [ ] Implement auto-detection logic
- [ ] Create migration utilities

#### Phase 4: Testing
- [ ] Build mock API server
- [ ] Create comprehensive test suite
- [ ] Add integration tests
- [ ] Performance benchmarking

#### Phase 5: Integration
- [ ] Update analyzer to use data source interface
- [ ] Enhance CLI with new options
- [ ] Update documentation
- [ ] Container support validation

### Success Metrics
1. **Functionality**: Feature parity with SSH implementation
2. **Performance**: ≤ 20% performance overhead vs SSH
3. **Reliability**: 99%+ successful API connections
4. **Security**: No credential storage in logs/memory
5. **Usability**: Zero-config migration for existing users

This implementation will modernize the Pi-hole integration while maintaining full backward compatibility and providing a foundation for future enhancements.

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
- **Pi-hole**: SQLite database access via SSH with structured logging
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
1. **Pi-hole API Integration** - Replace SSH with official Pi-hole REST API
2. **Enhanced Configuration Validation** - Comprehensive config validation with structured logging
3. **Performance Metrics Collection** - Built-in metrics for monitoring
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
11. **Make unit tests comprehensive** - Ensure all new features are covered
12. **Make intergration tests robust** - Validate Pi-hole connectivity and mock environments

### Development Workflow
1. **Start with**: `make dev-setup` for complete environment preparation
2. **Fast iteration**: `make fast-build` for quick development cycles
3. **Container testing**: `make docker-dev` for containerized development
4. **Before committing**: `make ci-test` to validate all tests pass
5. **Performance check**: `make cache-info` to verify cache effectiveness

### Code Patterns
- **Logging**: Use structured logging with context (`slog.String()`, `slog.Int()`, etc.)
- **Configuration**: Always validate config with appropriate logging
- **SSH**: Include connection logging with host/port context
- **Errors**: Structured error logging with full context
- **Colors**: Respect `--no-color` and `--no-emoji` flags

### Architecture Principles
- **Modular Design**: Use internal packages for separation of concerns  
- **Container First**: All features must work in containerized environments
- **Performance Aware**: Consider build cache impact of changes
- **Security Focused**: Non-root containers, minimal attack surface
- **Cross-Platform**: Support AMD64, ARM64, ARMv7 architectures

This project emphasizes **structured logging**, **fast builds**, **containerization**, and **beautiful terminal output** - maintain these core values when implementing changes.
