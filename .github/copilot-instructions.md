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
1. **Configuration validation enhancement** - Add comprehensive config validation
2. **Performance monitoring integration** - Add metrics collection
3. **Enhanced error recovery** - Improve SSH connection resilience

### Medium Priority
1. **Add Prometheus metrics endpoints** for monitoring
2. **Support multiple output formats** (JSON, XML) beyond terminal
3. **Enhanced network analysis** capabilities
4. **Kubernetes deployment manifests** for container orchestration

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
- **Kubernetes**: Ready for deployment with provided manifests (future enhancement)

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
1. **Enhanced Configuration Validation** - Comprehensive config validation with structured logging
2. **Performance Metrics Collection** - Built-in metrics for monitoring
3. **Kubernetes Manifests** - Ready-to-deploy K8s configurations
4. **Multi-Pi-hole Support** - Connect to multiple Pi-hole instances

### Architecture Evolution (Medium Term)
1. **Microservices Architecture** - Break into focused services
2. **REST API Endpoints** - Web API for remote access
3. **Real-time Monitoring** - Live dashboard capabilities
4. **Plugin System** - Extensible analysis modules

### Container Ecosystem (Advanced)
1. **Helm Charts** - Kubernetes package management
2. **Operator Pattern** - Kubernetes custom resources
3. **Multi-Registry Support** - Docker Hub, ACR, ECR publishing
4. **ARM32 Support** - Additional Raspberry Pi architectures

---

## Quick Reference for AI Assistants (Updated)

When working on this project:

### CRITICAL Requirements
1. **🚨 NEVER use `fmt.Printf`** - Always use `internal/logger` with structured logging
2. **🔧 Test both binaries** - `pihole-analyzer` (production) and `pihole-analyzer-test` (development)
3. **🐳 Validate container builds** - Changes must work in containerized environments
4. **⚡ Preserve build performance** - Check cache impact with `make cache-info`
5. **🏗️ Follow dual-binary architecture** - Production/test separation is intentional

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
