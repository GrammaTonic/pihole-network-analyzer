# API Reference

This document provides detailed reference for the internal packages and APIs used in the Pi-hole Network Analyzer.

## Package Overview

The Pi-hole Network Analyzer is built using modular packages following Go best practices:

```go
pihole-analyzer/
├── cmd/pihole-analyzer/        // Main application
├── internal/                   // Private packages
│   ├── analyzer/              // Core analysis logic
│   ├── cli/                   // Command-line interface
│   ├── colors/                // Terminal colorization
│   ├── config/                // Configuration management
│   ├── network/               // Network analysis
│   ├── reporting/             // Output formatting
│   ├── ssh/                   // SSH connectivity
│   ├── testutils/             // Testing utilities
│   └── types/                 // Data structures
```

## Core Data Types

### package `internal/types`

#### `PiholeRecord`

Represents a single DNS query record from the Pi-hole database.

```go
type PiholeRecord struct {
    Timestamp string `json:"timestamp"` // Unix timestamp as string
    Client    string `json:"client"`    // Client IP address
    HWAddr    string `json:"hwaddr"`    // Hardware/MAC address
    Domain    string `json:"domain"`    // Queried domain name
    Status    int    `json:"status"`    // Pi-hole status code
}
```

**Status Codes:**
- `1`: Query blocked by exact blacklist
- `2`: Query allowed (forwarded)
- `3`: Query blocked by CNAME
- `4`: Query blocked by gravity
- `9`: Query blocked by regex

#### `ClientStats`

Contains aggregated statistics for a network client.

```go
type ClientStats struct {
    Client        string            `json:"client"`         // IP address
    Hostname      string            `json:"hostname"`       // Resolved hostname
    HardwareAddr  string            `json:"hardware_addr"`  // MAC address
    IsOnline      bool              `json:"is_online"`      // Current online status
    TotalQueries  int               `json:"total_queries"`  // Total DNS queries
    UniqueQueries int               `json:"unique_queries"` // Unique domains queried
    AvgReplyTime  float64           `json:"avg_reply_time"` // Average response time
    Domains       map[string]int    `json:"domains"`        // Domain query counts
    QueryTypes    map[int]int       `json:"query_types"`    // Query type distribution
    StatusCodes   map[int]int       `json:"status_codes"`   // Status code distribution
    TopDomains    []DomainCount     `json:"top_domains"`    // Most queried domains
}
```

#### `DomainCount`

Represents domain query statistics.

```go
type DomainCount struct {
    Domain string  `json:"domain"` // Domain name
    Count  int     `json:"count"`  // Number of queries
    Percent float64 `json:"percent"` // Percentage of total queries
}
```

#### Configuration Types

```go
type Config struct {
    Pihole     PiholeConfig     `json:"pihole"`
    Exclusions ExclusionConfig  `json:"exclusions"`
    Output     OutputConfig     `json:"output"`
}

type PiholeConfig struct {
    Host     string `json:"host"`      // Pi-hole server hostname/IP
    Port     int    `json:"port"`      // SSH port (default: 22)
    Username string `json:"username"`  // SSH username
    KeyPath  string `json:"keyPath"`   // SSH private key path
    Password string `json:"password"`  // SSH password (alternative to key)
    DbPath   string `json:"dbPath"`    // Pi-hole database path
    Timeout  int    `json:"timeout"`   // Connection timeout (seconds)
}

type ExclusionConfig struct {
    Networks  []string `json:"networks"`  // CIDR network ranges to exclude
    Hostnames []string `json:"hostnames"` // Hostname patterns to exclude
    IPs       []string `json:"ips"`       // Specific IP addresses to exclude
}

type OutputConfig struct {
    Colors      bool   `json:"colors"`      // Enable colored output
    Emoji       bool   `json:"emoji"`       // Enable emoji symbols
    SaveReports bool   `json:"saveReports"` // Save reports to files
    ReportPath  string `json:"reportPath"`  // Report save directory
    MaxClients  int    `json:"maxClients"`  // Maximum clients to display
    SortBy      string `json:"sortBy"`      // Sort order
}
```

## Package APIs

### package `internal/analyzer`

Core analysis functions for processing Pi-hole data.

#### `AnalyzePiholeData`

```go
func AnalyzePiholeData(records []types.PiholeRecord, config *types.Config) ([]types.ClientStats, error)
```

Analyzes Pi-hole records and generates client statistics.

**Parameters:**
- `records`: Slice of Pi-hole DNS query records
- `config`: Application configuration for exclusions and settings

**Returns:**
- `[]types.ClientStats`: Aggregated statistics per client
- `error`: Any error encountered during analysis

**Example:**
```go
records := []types.PiholeRecord{
    {
        Timestamp: "1691520001",
        Client:    "192.168.1.100",
        Domain:    "google.com",
        Status:    2,
    },
}

stats, err := analyzer.AnalyzePiholeData(records, config)
if err != nil {
    return fmt.Errorf("analysis failed: %w", err)
}
```

#### `GroupByClient`

```go
func GroupByClient(records []types.PiholeRecord) map[string][]types.PiholeRecord
```

Groups Pi-hole records by client IP address.

#### `CalculateTopDomains`

```go
func CalculateTopDomains(domains map[string]int, limit int) []types.DomainCount
```

Calculates the most frequently queried domains.

### package `internal/ssh`

SSH connectivity and Pi-hole database access.

#### `ConnectToPihole`

```go
func ConnectToPihole(config *types.PiholeConfig) (*ssh.Client, error)
```

Establishes SSH connection to Pi-hole server.

**Parameters:**
- `config`: Pi-hole connection configuration

**Returns:**
- `*ssh.Client`: SSH client connection
- `error`: Connection error if any

#### `QueryPiholeDatabase`

```go
func QueryPiholeDatabase(client *ssh.Client, config *types.PiholeConfig, query string) ([]types.PiholeRecord, error)
```

Executes SQL query against Pi-hole database via SSH.

**Parameters:**
- `client`: Established SSH client connection
- `config`: Pi-hole configuration with database path
- `query`: SQL query to execute

**Returns:**
- `[]types.PiholeRecord`: Query results
- `error`: Query error if any

**Example:**
```go
client, err := ssh.ConnectToPihole(config)
if err != nil {
    return fmt.Errorf("connection failed: %w", err)
}
defer client.Close()

query := "SELECT timestamp, client, domain, status FROM queries ORDER BY timestamp DESC LIMIT 1000"
records, err := ssh.QueryPiholeDatabase(client, config, query)
```

#### Standard Queries

```go
const (
    // Get recent DNS queries
    RecentQueriesQuery = `
        SELECT timestamp, client, domain, status 
        FROM queries 
        WHERE timestamp > strftime('%s', 'now', '-24 hours')
        ORDER BY timestamp DESC
    `
    
    // Get client statistics
    ClientStatsQuery = `
        SELECT client, COUNT(*) as query_count, COUNT(DISTINCT domain) as unique_domains
        FROM queries 
        GROUP BY client 
        ORDER BY query_count DESC
    `
    
    // Get top domains
    TopDomainsQuery = `
        SELECT domain, COUNT(*) as query_count
        FROM queries 
        GROUP BY domain 
        ORDER BY query_count DESC 
        LIMIT ?
    `
)
```

### package `internal/network`

Network analysis and ARP table processing.

#### `GetARPTable`

```go
func GetARPTable(client *ssh.Client) (map[string]string, error)
```

Retrieves ARP table from Pi-hole server to determine online status.

**Returns:**
- `map[string]string`: Map of IP addresses to MAC addresses
- `error`: Network command error if any

#### `DetermineOnlineStatus`

```go
func DetermineOnlineStatus(clients []types.ClientStats, arpTable map[string]string) []types.ClientStats
```

Updates client online status based on ARP table data.

#### `ApplyExclusions`

```go
func ApplyExclusions(clients []types.ClientStats, exclusions *types.ExclusionConfig) []types.ClientStats
```

Filters clients based on exclusion rules.

### package `internal/config`

Configuration management and defaults.

#### `LoadConfig`

```go
func LoadConfig(path string) (*types.Config, error)
```

Loads configuration from JSON file.

#### `DefaultConfig`

```go
func DefaultConfig() *types.Config
```

Returns default configuration settings.

#### `SaveConfig`

```go
func SaveConfig(config *types.Config, path string) error
```

Saves configuration to JSON file.

#### `ValidateConfig`

```go
func ValidateConfig(config *types.Config) error
```

Validates configuration settings.

**Example:**
```go
config, err := config.LoadConfig("~/.pihole-analyzer/config.json")
if err != nil {
    config = config.DefaultConfig()
}

if err := config.ValidateConfig(config); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

### package `internal/colors`

Terminal colorization and emoji support.

#### Color Constants

```go
const (
    Red     = "\033[31m"
    Green   = "\033[32m"
    Yellow  = "\033[33m"
    Blue    = "\033[34m"
    Magenta = "\033[35m"
    Cyan    = "\033[36m"
    White   = "\033[37m"
    Reset   = "\033[0m"
    Bold    = "\033[1m"
)
```

#### `Colorize`

```go
func Colorize(text string, color string) string
```

Applies color to text string.

#### `StatusEmoji`

```go
func StatusEmoji(isOnline bool) string
```

Returns appropriate emoji for online status.

#### `FormatWithColors`

```go
func FormatWithColors(stats []types.ClientStats, config *types.OutputConfig) string
```

Formats client statistics with colors and emojis.

### package `internal/reporting`

Output formatting and report generation.

#### `DisplayClientStats`

```go
func DisplayClientStats(stats []types.ClientStats, config *types.OutputConfig) error
```

Displays formatted client statistics to stdout.

#### `GenerateReport`

```go
func GenerateReport(stats []types.ClientStats, config *types.OutputConfig) (string, error)
```

Generates formatted report as string.

#### `SaveReport`

```go
func SaveReport(report string, path string) error
```

Saves report to file.

#### `FormatTable`

```go
func FormatTable(headers []string, rows [][]string, config *types.OutputConfig) string
```

Formats data as aligned table.

### package `internal/cli`

Command-line interface and flag parsing.

#### `Flags`

```go
type Flags struct {
    ConfigPath   string
    PiholeConfig string
    CreateConfig bool
    ShowConfig   bool
    PiholeSetup  bool
    Test         bool
    TestMode     bool
    Quiet        bool
    NoColor      bool
    NoEmoji      bool
    OnlineOnly   bool
    NoExclude    bool
}
```

#### `ParseFlags`

```go
func ParseFlags() *Flags
```

Parses command-line flags and returns configuration.

#### `ShowHelp`

```go
func ShowHelp()
```

Displays usage information.

### package `internal/testutils`

Testing utilities and mock data.

#### `GenerateMockRecords`

```go
func GenerateMockRecords(count int) []types.PiholeRecord
```

Generates mock Pi-hole records for testing.

#### `CreateMockDatabase`

```go
func CreateMockDatabase(path string, records []types.PiholeRecord) error
```

Creates SQLite database with mock data.

#### `MockConfig`

```go
func MockConfig() *types.Config
```

Returns configuration suitable for testing.

## Error Handling

### Error Types

```go
var (
    ErrSSHConnection    = errors.New("SSH connection failed")
    ErrDatabaseAccess   = errors.New("database access denied")
    ErrInvalidConfig    = errors.New("invalid configuration")
    ErrNetworkCommand   = errors.New("network command failed")
    ErrInvalidQuery     = errors.New("invalid SQL query")
)
```

### Error Handling Patterns

```go
// Wrap errors with context
func ConnectToPihole(config *types.PiholeConfig) (*ssh.Client, error) {
    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Pi-hole server %s:%d: %w", 
            config.Host, config.Port, err)
    }
    return client, nil
}

// Check for specific error types
if errors.Is(err, ErrSSHConnection) {
    log.Printf("SSH connection issue: %v", err)
    // Handle SSH-specific errors
}
```

## Performance Considerations

### Memory Usage

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
    clientGroups := GroupByClient(records)
    resultChan := make(chan types.ClientStats, len(clientGroups))
    var wg sync.WaitGroup
    
    for client, clientRecords := range clientGroups {
        wg.Add(1)
        go func(client string, records []types.PiholeRecord) {
            defer wg.Done()
            stats := analyzeClient(client, records)
            resultChan <- stats
        }(client, clientRecords)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    var results []types.ClientStats
    for stats := range resultChan {
        results = append(results, stats)
    }
    
    return results
}
```

## Testing APIs

### Unit Testing

```go
func TestAnalyzePiholeData(t *testing.T) {
    // Create test data
    records := testutils.GenerateMockRecords(100)
    config := testutils.MockConfig()
    
    // Run analysis
    stats, err := analyzer.AnalyzePiholeData(records, config)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotEmpty(t, stats)
    assert.Equal(t, len(stats), expectedClientCount)
}
```

### Integration Testing

```go
func TestPiholeConnection(t *testing.T) {
    // Skip if no test configuration
    if !hasTestConfig() {
        t.Skip("No test Pi-hole configuration available")
    }
    
    config := loadTestConfig()
    client, err := ssh.ConnectToPihole(&config.Pihole)
    require.NoError(t, err)
    defer client.Close()
    
    records, err := ssh.QueryPiholeDatabase(client, &config.Pihole, testQuery)
    assert.NoError(t, err)
    assert.NotEmpty(t, records)
}
```

## Extension Points

### Adding New Analysis Features

1. **Extend `ClientStats` structure:**
   ```go
   type ClientStats struct {
       // ... existing fields ...
       NewMetric float64 `json:"new_metric"`
   }
   ```

2. **Implement analysis logic:**
   ```go
   func calculateNewMetric(records []types.PiholeRecord) float64 {
       // Implementation
   }
   ```

3. **Update display formatting:**
   ```go
   func formatClientStats(stats []types.ClientStats) string {
       // Include new metric in output
   }
   ```

### Adding New Output Formats

```go
// Add to OutputConfig
type OutputConfig struct {
    // ... existing fields ...
    Format string `json:"format"` // "table", "json", "csv", etc.
}

// Implement formatter
func FormatAsJSON(stats []types.ClientStats) (string, error) {
    return json.MarshalIndent(stats, "", "  ")
}
```

## Version Compatibility

The API follows semantic versioning. Breaking changes will increment the major version.

### Current Version: v1.x.x

- Stable API for all public interfaces
- Internal packages may change between minor versions
- Configuration format is stable

### Planned Changes (v2.x.x)

- Enhanced query filtering
- Multi-Pi-hole support
- Streaming analysis mode
- Plugin architecture

---

For more examples and usage patterns, see the [Development Guide](development.md) and [Usage Guide](usage.md).
