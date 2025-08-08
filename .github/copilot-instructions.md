# Pi-hole Network Analyzer - AI Coding Assistant Instructions

This file provides comprehensive guidance for AI coding assistants working on the Pi-hole Network Analyzer project.

## Project Overview

**Repository**: Pi-hole Network Analyzer  
**Language**: Go 1.24.5  
**Module**: `pihole-analyzer`  
**Binary Name**: `pihole-analyzer` 
**Command Directory**: `cmd/pihole-analyzer/`  
**Main File**: `cmd/pihole-analyzer/main.go`
**Help Command**: `pihole-analyzer --help`  
**Architecture**: Standard Go Project Layout

### Core Purpose
**Pi-hole-only** DNS usage analysis tool that connects to Pi-hole servers via SSH. Features colorized terminal output, network analysis, and comprehensive reporting. 

**🚨 IMPORTANT**: CSV functionality has been completely removed as of August 2025. This is now a dedicated Pi-hole analyzer.

## Project Structure & Conventions

### Directory Layout
```
/
├── cmd/pihole-analyzer/          # Main application entry point
├── internal/                     # Private application packages
│   ├── analyzer/                 # Pi-hole data analysis
│   ├── cli/                      # Command-line interface
│   ├── colors/                   # Terminal colorization utilities
│   ├── config/                   # Configuration management
│   ├── network/                  # Network analysis & ARP
│   ├── reporting/                # Output display & reports
│   ├── ssh/                      # Pi-hole SSH connectivity
│   ├── testutils/                # Testing utilities
│   └── types/                    # Core data structures
├── docs/                         # Comprehensive documentation
├── scripts/                      # Build and testing automation
└── test_data/                    # Mock databases for testing
```

### Key Files & Their Roles
- **`cmd/pihole-analyzer/main.go`**: Refactored main application (Pi-hole only)
- **`internal/analyzer/analyzer.go`**: Pi-hole data analysis engine
- **`internal/ssh/pihole.go`**: SSH connection and database analysis
- **`internal/cli/flags.go`**: Command-line interface and validation
- **`internal/types/types.go`**: Core data structures (ClientStats, PiholeRecord)
- **`internal/colors/colors.go`**: Terminal color/emoji system with cross-platform support
- **`internal/config/config.go`**: JSON configuration, SSH settings, exclusion rules
- **`Makefile`**: Build system with Pi-hole-focused targets

## Naming Consistency Status ✅

**RESOLVED**: Naming has been standardized throughout the project:
- Module name: `pihole-analyzer`
- Binary name: `pihole-analyzer`
- Command directory: `cmd/pihole-analyzer/`
- All references now use consistent `pihole-analyzer` naming

## Architecture Patterns

### Data Flow (Pi-hole Only)
1. **Input**: SSH connection to Pi-hole SQLite database
2. **Processing**: Query Pi-hole database for DNS records
3. **Analysis**: Aggregate into `types.ClientStats` with network analysis
4. **Output**: Colorized terminal display + optional file reports

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
- **Structure**: `types.Config` with nested structs for Pi-hole, exclusions, output
- **Defaults**: Comprehensive defaults in `config.DefaultConfig()`
- **SSH Support**: Key-based and password authentication

## Development Workflow

### Build System (Makefile)
```bash
make build      # Build binary (BINARY_NAME=pihole-analyzer)
make run        # Run with test data
make test       # Run all tests
make ci-test    # CI-compatible test suite
make clean      # Clean build artifacts
```

### Testing Strategy
- **Unit Tests**: Go standard testing in each package
- **Integration Tests**: `scripts/integration-test.sh` with Pi-hole scenarios
- **CI/CD**: GitHub Actions with cross-platform builds
- **Test Data**: Mock Pi-hole database environment

### Code Quality Standards
- **Current Grade**: B+ (identified in prior analysis)
- **Main Issue**: Monolithic `main.go` needs refactoring
- **Standards**: Go formatting, no unused variables/functions
- **Dependencies**: Minimal external dependencies (ssh, sqlite, crypto)

## Key Features & Implementation

### Colorized Output System
- **Package**: `internal/colors`
- **Features**: Cross-platform terminal colors, emoji support, smart domain highlighting
- **Flags**: `--no-color`, `--no-emoji` for compatibility, configurable in `config.json`
- **Patterns**: Color-coded statistics, progress indicators, status messages

### SSH Pi-hole Connection
```go
// SSH connection pattern used throughout
sshConfig := &ssh.ClientConfig{
    User: config.Pihole.Username,
    Auth: []ssh.AuthMethod{
        ssh.PublicKeys(signer),
        ssh.Password(config.Pihole.Password),
    },
    HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}
```

### Network Analysis
- **ARP Table**: Determines online/offline status via MAC address lookup
- **Exclusions**: Configurable network/IP/hostname exclusions (Docker, loopback, etc.)
- **DNS Query Analysis**: Status codes, query types, domain categorization

## Common Tasks & Patterns

### Adding New CLI Flags
1. Declare in main.go variable section
2. Add to flag parsing
3. Update help text consistently
4. Handle in configuration logic

### Adding New Analysis Features
1. Extend `types.ClientStats` or `types.PiholeRecord` if needed
2. Implement analysis logic in analyzer package
3. Add colorized output in reporting package
4. Include in report generation

### Configuration Changes
1. Update `types.Config` structure
2. Modify `config.DefaultConfig()`
3. Handle in JSON marshaling/unmarshaling
4. Update validation logic

### Testing New Features
1. Add unit tests in appropriate package
2. Update integration test scenarios
3. Test Pi-hole connectivity and mock environments
4. Verify colorized output works correctly

## Dependencies & External Libraries

### Core Dependencies
- **`golang.org/x/crypto/ssh`**: SSH client for Pi-hole connections
- **`modernc.org/sqlite`**: SQLite database access (pure Go)
- **Standard Library**: Heavy use of net, io, encoding packages

### Development Dependencies
- **Testing**: Standard Go testing framework
- **CI/CD**: GitHub Actions with cross-platform builds
- **Build**: Make-based build system

## Common Pitfalls & Solutions

### 1. Binary Name Confusion
**Problem**: Inconsistent naming between module name and binary
**Solution**: Standardized on `pihole-analyzer` throughout the project

### 2. Large main.go File
**Problem**: Monolithic main.go is difficult to maintain
**Solution**: Refactored into modular packages (analyzer, cli, reporting)

### 3. Color Output in CI
**Problem**: Terminal colors can break CI output parsing
**Solution**: Always test with `--no-color` flag in CI environments

### 4. SSH Connection Handling
**Problem**: SSH connections can timeout or fail
**Solution**: Implement proper error handling and connection retries

## Refactoring Opportunities

### High Priority
1. **🚧 ACTIVE: Separate testing utilities** - Remove testing code from production binary (see TESTING_SEPARATION_PLAN.md)
2. **Implement configuration validation**

### Medium Priority
1. **Add metrics/monitoring endpoints**
2. **Support multiple output formats** (JSON, XML)
3. **Performance optimization** for large datasets
4. **Enhanced network analysis** capabilities

## Integration Points

### CI/CD Pipeline
- **Speed**: Fast builds with caching
- **Artifacts**: Store build artifacts for easy access
- **GitHub Actions**: `.github/workflows/` (when created)
- **Test Commands**: `make ci-test`, `scripts/integration-test.sh`
- **Cross-Platform**: Linux, macOS, Windows builds

### External Systems
- **Pi-hole**: SQLite database access via SSH
- **ARP Tables**: System ARP command execution
- **File System**: Configuration and reports

## Debugging & Troubleshooting

### Common Debug Flags
```bash
--quiet          # Suppress verbose output
--no-color       # Disable colors for log analysis
--test-mode      # Use mock data for development
--show-config    # Display current configuration
```

### Log Analysis
- **Terminal Output**: Colorized progress and status
- **File Reports**: Timestamped analysis saved to `reports/`
- **Error Handling**: Graceful degradation with informative messages

## Future Roadmap

### Planned Features
1. **Docker Support**: Containerization with Docker Compose
2. **Prometheus Metrics**: Real-time monitoring and alerting
3. **Grafana Dashboards**: Visualization and analytics
4. **Multi-Platform ARM64**: Raspberry Pi deployment support

### Architecture Evolution

---

## Quick Reference for AI Assistants

When working on this project:

1. **Always check** the naming consistency `pihole-analyzer`
2. **Use the internal packages** for modular functionality
3. **Test Pi-hole SSH connectivity** for any changes
4. **Maintain colorized output compatibility** with `--no-color` flag
5. **Follow the Standard Go Project Layout** conventions
6. **Run `make ci-test`** before suggesting major changes
7. **Consider the modular architecture** when making changes
8. **Preserve the comprehensive testing framework** when making changes

This project emphasizes **beautiful terminal output**, **robust SSH connectivity**, and **comprehensive DNS analysis** - keep these core values when suggesting improvements.
