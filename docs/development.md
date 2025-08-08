# Development Guide

Welcome to the Pi-hole Network Analyzer development guide! This document will help you set up a development environment, understand the codebase, and contribute to the project.

## Development Environment Setup

### Prerequisites

- **Go 1.23+** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation
- **SSH client** - For Pi-hole connectivity testing
- **Text editor/IDE** - VS Code, GoLand, or vim

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run integration tests
make ci-test
```

## Project Structure

### Standard Go Project Layout

```
pihole-network-analyzer/
├── cmd/
│   └── pihole-analyzer/           # Main application entry point
│       └── main.go
├── internal/                      # Private application packages
│   ├── analyzer/                  # Pi-hole data analysis engine
│   │   └── analyzer.go
│   ├── cli/                       # Command-line interface
│   │   └── flags.go
│   ├── colors/                    # Terminal colorization
│   │   ├── colors.go
│   │   └── colors_test.go
│   ├── config/                    # Configuration management
│   │   └── config.go
│   ├── network/                   # Network analysis & ARP
│   │   └── network.go
│   ├── reporting/                 # Output formatting & display
│   │   └── display.go
│   ├── ssh/                       # Pi-hole SSH connectivity
│   │   └── pihole.go
│   ├── testutils/                 # Testing utilities
│   │   └── test_data.go
│   └── types/                     # Core data structures
│       └── types.go
├── docs/                          # Documentation
├── scripts/                       # Build and automation scripts
├── test_data/                     # Mock data for testing
├── .github/                       # CI/CD workflows
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
└── Makefile                       # Build automation
```

### Key Components

#### Core Data Structures (`internal/types/`)

```go
// PiholeRecord represents a single DNS query record
type PiholeRecord struct {
    Timestamp string
    Client    string
    HWAddr    string
    Domain    string
    Status    int
}

// ClientStats contains aggregated statistics for a client
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

// Config represents the application configuration
type Config struct {
    Pihole     PiholeConfig     `json:"pihole"`
    Exclusions ExclusionConfig  `json:"exclusions"`
    Output     OutputConfig     `json:"output"`
}
```

#### Data Flow Architecture

```
SSH Connection → Pi-hole Database → Query Parsing → Analysis → Output Formatting
     ↓               ↓                    ↓            ↓            ↓
[ssh/pihole.go] [analyzer/analyzer.go] [types/] [analyzer/] [reporting/display.go]
```

## Development Workflow

### 1. Setting Up Your Development Branch

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ... code changes ...

# Run tests
make test
make ci-test

# Commit your changes
git add .
git commit -m "Add your feature description"

# Push and create PR
git push origin feature/your-feature-name
```

### 2. Running Tests

```bash
# Unit tests only
go test ./...

# Unit tests with coverage
go test -cover ./...

# Integration tests
./scripts/integration-test.sh

# All tests (CI compatible)
make ci-test

# Specific package tests
go test ./internal/colors/
go test ./internal/analyzer/
```

### 3. Building and Testing

```bash
# Build for development
make build

# Build with race detection
go build -race -o pihole-analyzer ./cmd/pihole-analyzer

# Test with mock data
./pihole-analyzer --test

# Test with custom config
./pihole-analyzer --config test-config.json --test
```

## Code Style and Standards

### Go Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Lint code
golangci-lint run
```

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Functions**: camelCase, exported functions start with uppercase
- **Variables**: camelCase, descriptive names
- **Constants**: ALL_CAPS for package-level constants

### Documentation

```go
// Package analyzer provides Pi-hole database analysis functionality.
package analyzer

// AnalyzePiholeData processes Pi-hole records and generates client statistics.
// It takes a slice of PiholeRecord and returns aggregated ClientStats.
func AnalyzePiholeData(records []types.PiholeRecord) ([]types.ClientStats, error) {
    // Implementation
}
```

## Testing Strategy

### Unit Tests

```go
// Example unit test
func TestAnalyzePiholeData(t *testing.T) {
    // Arrange
    records := []types.PiholeRecord{
        {
            Timestamp: "1691520001",
            Client:    "192.168.1.100",
            HWAddr:    "aa:bb:cc:dd:ee:ff",
            Domain:    "google.com",
            Status:    2,
        },
    }

    // Act
    stats, err := AnalyzePiholeData(records)

    // Assert
    assert.NoError(t, err)
    assert.Len(t, stats, 1)
    assert.Equal(t, "192.168.1.100", stats[0].Client)
    assert.Equal(t, 1, stats[0].TotalQueries)
}
```

### Integration Tests

```bash
# Test scenarios defined in scripts/integration-test.sh
./scripts/integration-test.sh pihole-db         # Test Pi-hole functionality
./scripts/integration-test.sh colorized-output # Test color output
./scripts/integration-test.sh all-features     # Comprehensive tests
```

### Mock Data

The project includes comprehensive mock data for development:

```bash
# Mock Pi-hole database
test_data/mock_pihole.db

# Mock configuration
test_data/mock_pihole_config.json

# Use mock data in tests
./pihole-analyzer --test
```

## Adding New Features

### 1. Adding a New Analysis Feature

Example: Adding average query response time analysis

#### Step 1: Extend Data Structures

```go
// internal/types/types.go
type ClientStats struct {
    // ... existing fields ...
    AvgResponseTime  float64 `json:"avg_response_time"`
    MaxResponseTime  float64 `json:"max_response_time"`
}
```

#### Step 2: Implement Analysis Logic

```go
// internal/analyzer/analyzer.go
func calculateResponseTimes(records []types.PiholeRecord) map[string]ResponseTimeStats {
    stats := make(map[string]ResponseTimeStats)
    
    for _, record := range records {
        // Calculate response time statistics
        // Implementation details...
    }
    
    return stats
}
```

#### Step 3: Update Output Display

```go
// internal/reporting/display.go
func formatClientStats(stats []types.ClientStats) string {
    // Add response time column to output table
    // Update formatting logic...
}
```

#### Step 4: Add Tests

```go
// internal/analyzer/analyzer_test.go
func TestCalculateResponseTimes(t *testing.T) {
    // Test the new response time calculation
}
```

#### Step 5: Update Documentation

Update relevant documentation files:
- `docs/usage.md` - Document new output columns
- `docs/api.md` - Update API documentation
- `README.md` - Update example output

### 2. Adding a New Command-Line Option

Example: Adding `--top-domains` flag

#### Step 1: Define Flag

```go
// internal/cli/flags.go
type Flags struct {
    // ... existing flags ...
    TopDomains int `json:"top_domains"`
}

func ParseFlags() *Flags {
    flags := &Flags{}
    
    // ... existing flag parsing ...
    flag.IntVar(&flags.TopDomains, "top-domains", 10, "Number of top domains to display")
    
    flag.Parse()
    return flags
}
```

#### Step 2: Implement Feature Logic

```go
// internal/analyzer/analyzer.go
func getTopDomains(stats []types.ClientStats, limit int) []DomainStats {
    // Implementation for getting top domains
}
```

#### Step 3: Update Main Application

```go
// cmd/pihole-analyzer/main.go
func main() {
    flags := cli.ParseFlags()
    
    // ... existing logic ...
    
    if flags.TopDomains > 0 {
        topDomains := analyzer.GetTopDomains(clientStats, flags.TopDomains)
        reporting.DisplayTopDomains(topDomains)
    }
}
```

## Debugging

### Debug Mode

```bash
# Enable verbose output
export PIHOLE_ANALYZER_DEBUG=1
./pihole-analyzer --pihole config.json

# SSH debugging
export SSH_DEBUG=1
./pihole-analyzer --pihole config.json
```

### Common Debug Scenarios

#### Database Connection Issues

```go
// Add debug logging in ssh/pihole.go
log.Printf("Connecting to Pi-hole at %s:%d", config.Host, config.Port)
log.Printf("Using SSH key: %s", config.KeyPath)
log.Printf("Database path: %s", config.DbPath)
```

#### Data Processing Issues

```go
// Add debug logging in analyzer/analyzer.go
log.Printf("Processing %d records", len(records))
log.Printf("Found %d unique clients", len(clientMap))
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/analyzer/

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/analyzer/

# View profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Performance Optimization

### Database Queries

```go
// Optimize SQLite queries
query := `
    SELECT timestamp, client, domain, status
    FROM queries 
    WHERE timestamp > ? 
    ORDER BY timestamp DESC 
    LIMIT ?
`
```

### Memory Management

```go
// Use sync.Pool for frequently allocated objects
var recordPool = sync.Pool{
    New: func() interface{} {
        return make([]types.PiholeRecord, 0, 1000)
    },
}

func processRecords() {
    records := recordPool.Get().([]types.PiholeRecord)
    defer recordPool.Put(records[:0])
    
    // Process records...
}
```

### Concurrent Processing

```go
// Process clients concurrently
func analyzeClientsParallel(records []types.PiholeRecord) []types.ClientStats {
    clientChan := make(chan types.ClientStats, len(records))
    var wg sync.WaitGroup
    
    for client, clientRecords := range groupByClient(records) {
        wg.Add(1)
        go func(client string, records []types.PiholeRecord) {
            defer wg.Done()
            stats := analyzeClient(client, records)
            clientChan <- stats
        }(client, clientRecords)
    }
    
    go func() {
        wg.Wait()
        close(clientChan)
    }()
    
    var results []types.ClientStats
    for stats := range clientChan {
        results = append(results, stats)
    }
    
    return results
}
```

## CI/CD Integration

### GitHub Actions

The project uses GitHub Actions for CI/CD:

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.23'
    - run: make ci-test
```

### Pre-commit Hooks

```bash
# Install pre-commit hooks
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e

# Run tests
make test

# Run linting
golangci-lint run

# Run formatting
go fmt ./...
EOF

chmod +x .git/hooks/pre-commit
```

## Release Process

### Version Management

```bash
# Create a new release
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0

# Build release binaries
make release
```

### Build Automation

```bash
# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o pihole-analyzer-linux-amd64 ./cmd/pihole-analyzer
GOOS=darwin GOARCH=amd64 go build -o pihole-analyzer-darwin-amd64 ./cmd/pihole-analyzer
GOOS=windows GOARCH=amd64 go build -o pihole-analyzer-windows-amd64.exe ./cmd/pihole-analyzer
```

## Contributing Guidelines

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes with tests
4. **Run** the test suite (`make ci-test`)
5. **Update** documentation if needed
6. **Submit** a pull request

### Code Review Checklist

- [ ] Code follows Go conventions
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No breaking changes (or properly documented)
- [ ] Performance impact considered
- [ ] Error handling is appropriate

### Issue Reporting

When reporting issues, include:
- Operating system and version
- Go version
- Pi-hole version and setup
- Steps to reproduce
- Expected vs. actual behavior
- Relevant log output

## Getting Help

### Resources

- **Documentation**: [docs/](../docs/)
- **GitHub Issues**: [Issues](https://github.com/GrammaTonic/pihole-network-analyzer/issues)
- **GitHub Discussions**: [Discussions](https://github.com/GrammaTonic/pihole-network-analyzer/discussions)

### Community

- Follow the [Code of Conduct](CODE_OF_CONDUCT.md)
- Be respectful and constructive
- Help others learn and contribute

## Next Steps

- **[API Reference](api.md)** - Detailed package documentation
- **[Troubleshooting](troubleshooting.md)** - Common development issues
- **[Configuration Guide](configuration.md)** - Advanced configuration

---

**Happy coding!** Thank you for contributing to the Pi-hole Network Analyzer project.
