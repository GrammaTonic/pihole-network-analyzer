# Pi-hole Network Analyzer - AI Coding Assistant Instructions

This file provides comprehensive guidance for AI coding assistants working on the Pi-hole Network Analyzer project.

## Project Overview

**Repository**: Pi-hole Network Analyzer  
**Language**: Go 1.24.5  
**Module**: `pihole-network-analyzer`  
**Binary Name**: `dns-analyzer` (Makefile) / `pihole-analyzer` (cmd directory)  
**Architecture**: Standard Go Project Layout  

### Core Purpose
DNS usage analysis tool that connects to Pi-hole servers via SSH or analyzes CSV log files. Features colorized terminal output, network analysis, and comprehensive reporting.

## Project Structure & Conventions

### Directory Layout
```
/
├── cmd/pihole-analyzer/          # Main application entry point (1693 lines)
├── internal/                     # Private application packages
│   ├── colors/                   # Terminal colorization utilities
│   ├── config/                   # Configuration management
│   └── types/                    # Core data structures
├── docs/                         # Comprehensive documentation
├── scripts/                      # Build and testing automation
└── reports/                      # Generated analysis reports
```

### Key Files & Their Roles
- **`cmd/pihole-analyzer/main.go`**: Monolithic main application (1693 lines) - PRIMARY REFACTORING TARGET
- **`internal/types/types.go`**: Core data structures (DNSRecord, ClientStats, PiholeRecord)
- **`internal/colors/colors.go`**: Terminal color/emoji system with cross-platform support
- **`internal/config/config.go`**: JSON configuration, SSH settings, exclusion rules
- **`Makefile`**: Build system with comprehensive targets (build, test, run, ci-test)

## Naming Consistency Issues ⚠️

**CRITICAL**: There's an ongoing naming inconsistency that needs resolution:
- Module name: `pihole-network-analyzer`
- Makefile binary: `dns-analyzer`
- Command directory: `cmd/pihole-analyzer/`
- Help text: References both names inconsistently

When making changes, be aware of this inconsistency and help resolve it by choosing one consistent name throughout.

## Architecture Patterns

### Data Flow
1. **Input**: CSV files or SSH connection to Pi-hole database
2. **Processing**: Parse DNS records into `types.DNSRecord` structures
3. **Analysis**: Aggregate into `types.ClientStats` with network analysis
4. **Output**: Colorized terminal display + optional file reports

### Core Data Structures

#### `types.DNSRecord`
```go
type DNSRecord struct {
    ID             int
    DateTime       string
    Domain         string
    Type           int      // Query type (A, AAAA, etc.)
    Status         int      // Pi-hole status codes
    Client         string   // Client IP address
    Forward        string
    AdditionalInfo string
    ReplyType      int
    ReplyTime      float64
    DNSSEC         bool
    ListID         int
    EDE            int
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
- **File**: `~/.dns-analyzer/config.json` (default)
- **Structure**: `types.Config` with nested structs for Pi-hole, exclusions, output
- **Defaults**: Comprehensive defaults in `config.DefaultConfig()`
- **SSH Support**: Key-based and password authentication

## Development Workflow

### Build System (Makefile)
```bash
make build      # Build binary (BINARY_NAME=dns-analyzer)
make run        # Run with test data
make test       # Run all tests
make ci-test    # CI-compatible test suite
make clean      # Clean build artifacts
```

### Testing Strategy
- **Unit Tests**: Go standard testing in each package
- **Integration Tests**: `scripts/integration-test.sh` with multiple scenarios
- **CI/CD**: GitHub Actions with cross-platform builds
- **Test Data**: `test.csv` and mock Pi-hole environment

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
1. Extend `types.ClientStats` or `types.DNSRecord` if needed
2. Implement analysis logic in main processing loop
3. Add colorized output in display functions
4. Include in report generation

### Configuration Changes
1. Update `types.Config` structure
2. Modify `config.DefaultConfig()`
3. Handle in JSON marshaling/unmarshaling
4. Update validation logic

### Testing New Features
1. Add unit tests in appropriate package
2. Update integration test scenarios
3. Test both CSV and Pi-hole modes
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
**Problem**: Inconsistent naming between `dns-analyzer` and `pihole-analyzer`
**Solution**: Choose one name and update all references consistently

### 2. Large main.go File
**Problem**: 1693-line main.go is difficult to maintain
**Solution**: Extract functions into internal packages (planned refactoring)

### 3. Color Output in CI
**Problem**: Terminal colors can break CI output parsing
**Solution**: Always test with `--no-color` flag in CI environments

### 4. SSH Connection Handling
**Problem**: SSH connections can timeout or fail
**Solution**: Implement proper error handling and connection retries

## Refactoring Opportunities

### High Priority
1. **Extract main.go functions** into logical packages
2. **Resolve naming inconsistency** throughout project
3. **Improve error handling** in SSH operations
4. **Add connection pooling** for multiple Pi-hole servers

### Medium Priority
1. **Add structured logging** (replace fmt.Printf)
2. **Implement configuration validation**
3. **Add metrics/monitoring endpoints**
4. **Support multiple output formats** (JSON, XML)

## Integration Points

### CI/CD Pipeline
- **GitHub Actions**: `.github/workflows/` (when created)
- **Test Commands**: `make ci-test`, `scripts/integration-test.sh`
- **Cross-Platform**: Linux, macOS, Windows builds

### External Systems
- **Pi-hole**: SQLite database access via SSH
- **ARP Tables**: System ARP command execution
- **File System**: Configuration, reports, CSV input

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
1. **Microservices**: Split monolithic main.go
2. **API Endpoints**: REST API for external integration
3. **Database Support**: PostgreSQL/MySQL alternatives
4. **Real-time Processing**: WebSocket streams for live analysis

---

## Quick Reference for AI Assistants

When working on this project:

1. **Always check** the naming consistency between `dns-analyzer` and `pihole-analyzer`
2. **Use the internal packages** for colors, config, and types
3. **Test both CSV and Pi-hole modes** for any changes
4. **Maintain colorized output compatibility** with `--no-color` flag
5. **Follow the Standard Go Project Layout** conventions
6. **Run `make ci-test`** before suggesting major changes
7. **Consider the 1693-line main.go** as the primary refactoring target
8. **Preserve the comprehensive testing framework** when making changes

This project emphasizes **beautiful terminal output**, **robust SSH connectivity**, and **comprehensive DNS analysis** - keep these core values when suggesting improvements.
