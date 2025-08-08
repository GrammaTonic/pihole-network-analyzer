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
│   ├── interfaces/            // Data source abstraction
│   ├── logger/                // Structured logging
│   ├── network/               // Network analysis
│   ├── pihole/                // Pi-hole API client
│   ├── reporting/             // Output formatting
│   └── types/                 // Data structures
```

## Core Data Types

### package `internal/types`

#### `PiholeRecord`

Represents a single DNS query record from the Pi-hole API.

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

**Status Codes:**
- `1`: Query blocked by exact blacklist
- `2`: Query allowed (forwarded)
- `3`: Query blocked by CNAME
- `4`: Query blocked by gravity
- `9`: Query blocked by regex

#### `ClientStats`

Aggregated statistics for a single network client.

```go
type ClientStats struct {
    IP            string            `json:"ip"`             // Client IP address
    Hostname      string            `json:"hostname"`       // Resolved hostname
    QueryCount    int               `json:"query_count"`    // Total queries
    Domains       map[string]int    `json:"domains"`        // Domain query counts
    DomainCount   int               `json:"domain_count"`   // Unique domains
    MACAddress    string            `json:"mac_address"`    // MAC address
    IsOnline      bool              `json:"is_online"`      // ARP table status
    LastSeen      string            `json:"last_seen"`      // Last query time
    TopDomains    []DomainStat      `json:"top_domains"`    // Most queried domains
    Status        string            `json:"status"`         // Connection status
    UniqueQueries int               `json:"unique_queries"` // Unique query count
    TotalQueries  int               `json:"total_queries"`  // Total query count
    // Additional analysis fields
    Client         string         `json:"client"`          // Client identifier
    QueryTypes     map[int]int    `json:"query_types"`     // Query type counts
    StatusCodes    map[int]int    `json:"status_codes"`    // Status code counts
    HWAddr         string         `json:"hwaddr"`          // Hardware address
    TotalReplyTime float64        `json:"total_reply_time"` // Total response time
    AvgReplyTime   float64        `json:"avg_reply_time"`   // Average response time
}
```

#### `DomainStat`

Represents domain query statistics.

```go
type DomainStat struct {
    Domain string `json:"domain"` // Domain name
    Count  int    `json:"count"`  // Query count
}
```

#### `Config`

Application configuration structure.

```go
type Config struct {
    OnlineOnly bool            `json:"online_only"` // Show only online clients
    NoExclude  bool            `json:"no_exclude"`  // Disable exclusions
    TestMode   bool            `json:"test_mode"`   // Use mock data
    Quiet      bool            `json:"quiet"`       // Suppress output
    Pihole     PiholeConfig    `json:"pihole"`      // Pi-hole connection
    Output     OutputConfig    `json:"output"`      // Output formatting
    Exclusions ExclusionConfig `json:"exclusions"`  // Exclusion rules
    Logging    LoggingConfig   `json:"logging"`     // Logging configuration
}
```

#### `PiholeConfig`

Pi-hole API connection configuration.

```go
type PiholeConfig struct {
    Host string `json:"host"` // Pi-hole hostname/IP
    Port int    `json:"port"` // Pi-hole web interface port
    
    // API Configuration (only method)
    APIEnabled  bool   `json:"api_enabled"`  // Enable API access
    APIPassword string `json:"api_password"` // API password/token
    APITOTP     string `json:"api_totp"`     // 2FA TOTP secret
    UseHTTPS    bool   `json:"use_https"`    // Force HTTPS
    APITimeout  int    `json:"api_timeout"`  // Request timeout (seconds)
}
```

## Package APIs

### package `internal/pihole`

Pi-hole API client implementation with session management.

#### `Client`

Main API client structure.

```go
type Client struct {
    BaseURL    string        // Pi-hole base URL
    HTTPClient *http.Client  // HTTP client
    SID        string        // Session ID
    CSRFToken  string        // CSRF token
    Logger     *logger.Logger // Structured logger
    config     *Config       // Client configuration
}
```

#### `NewClient`

Creates a new Pi-hole API client.

```go
func NewClient(config *Config, log *logger.Logger) *Client
```

**Parameters:**
- `config`: Pi-hole API configuration
- `log`: Structured logger instance

**Returns:**
- `*Client`: Configured API client

#### `Connect`

Establishes connection to Pi-hole API.

```go
func (c *Client) Connect(ctx context.Context) error
```

**Parameters:**
- `ctx`: Request context for timeout/cancellation

**Returns:**
- `error`: Connection error or nil on success

#### `GetQueries`

Retrieves DNS queries from Pi-hole API.

```go
func (c *Client) GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
```

**Parameters:**
- `ctx`: Request context
- `params`: Query filtering parameters

**Returns:**
- `[]types.PiholeRecord`: Query records
- `error`: Request error or nil

#### Example Usage

```go
// Create and configure client
config := &pihole.Config{
    Host:        "192.168.1.50",
    Port:        80,
    APIPassword: "your-api-password",
    UseHTTPS:    false,
}

logger := logger.New(&logger.Config{Component: "pihole-api"})
client := pihole.NewClient(config, logger)

// Connect to Pi-hole
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal("Failed to connect:", err)
}

// Query DNS records
params := pihole.QueryParams{
    StartTime: time.Now().Add(-24 * time.Hour),
    EndTime:   time.Now(),
    Limit:     1000,
}

records, err := client.GetQueries(ctx, params)
if err != nil {
    log.Fatal("Failed to get queries:", err)
}
```

### package `internal/interfaces`

Data source abstraction layer for Pi-hole connectivity.

#### `DataSource`

Interface for Pi-hole data access.

```go
type DataSource interface {
    Connect(ctx context.Context) error
    Close() error
    IsConnected() bool
    
    // Core data retrieval (API implementation)
    GetQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error)
    GetClientStats(ctx context.Context) (map[string]*types.ClientStats, error)
    GetNetworkInfo(ctx context.Context) ([]types.NetworkDevice, error)
    GetDomainAnalysis(ctx context.Context) (*types.DomainAnalysis, error)
    
    GetDataSourceType() DataSourceType
    GetConnectionInfo() *ConnectionInfo
}
```

#### `QueryParams`

Parameters for filtering Pi-hole queries.

```go
type QueryParams struct {
    StartTime    time.Time // Start time filter
    EndTime      time.Time // End time filter
    ClientFilter string    // Client IP filter
    DomainFilter string    // Domain name filter
    Limit        int       // Result limit
    StatusFilter []int     // Status code filter
    TypeFilter   []int     // Query type filter
}
```

### package `internal/analyzer`

Core analysis engine for Pi-hole data processing.

#### `Analyzer`

Main analysis engine.

```go
type Analyzer struct {
    dataSource interfaces.DataSource // Data source
    config     *types.Config         // Configuration
    logger     *logger.Logger        // Structured logger
}
```

#### `NewAnalyzer`

Creates a new analyzer instance.

```go
func NewAnalyzer(dataSource interfaces.DataSource, config *types.Config, logger *logger.Logger) *Analyzer
```

#### `AnalyzeData`

Performs comprehensive data analysis.

```go
func (a *Analyzer) AnalyzeData(ctx context.Context) (*types.AnalysisResult, error)
```

**Returns:**
- `*types.AnalysisResult`: Analysis results
- `error`: Analysis error or nil

### package `internal/logger`

Structured logging implementation using Go's `log/slog`.

#### `Logger`

Structured logger with colors and emoji support.

```go
type Logger struct {
    slogger *slog.Logger // Underlying slog logger
    config  *Config      // Logger configuration
}
```

#### `New`

Creates a new structured logger.

```go
func New(config *Config) *Logger
```

#### `Config`

Logger configuration.

```go
type Config struct {
    Level        Level  // Log level
    EnableColors bool   // Enable colored output
    EnableEmojis bool   // Enable emoji indicators
    Component    string // Component identifier
    OutputFile   string // Output file path (empty = stdout)
}
```

#### Usage Example

```go
logger := logger.New(&logger.Config{
    Level:        logger.LevelInfo,
    EnableColors: true,
    EnableEmojis: true,
    Component:    "pihole-api",
})

logger.Info("Pi-hole API connection established",
    slog.String("host", "192.168.1.50"),
    slog.Int("port", 80),
    slog.Bool("https", false))
```

### package `internal/cli`

Command-line interface and flag management.

#### `Flags`

CLI flag definitions.

```go
type Flags struct {
    OnlineOnly   *bool   // Show only online clients
    NoExclude    *bool   // Disable exclusions
    Pihole       *string // Pi-hole config file
    Config       *string // App config file
    NoColor      *bool   // Disable colors
    NoEmoji      *bool   // Disable emojis
    Quiet        *bool   // Suppress output
    CreateConfig *bool   // Create config file
    ShowConfig   *bool   // Show configuration
    PiholeSetup  *bool   // Pi-hole setup wizard
}
```

#### `ParseFlags`

Parses command-line flags.

```go
func ParseFlags() *Flags
```

### package `internal/network`

Network analysis and ARP table integration.

#### `GetARPTable`

Retrieves system ARP table for online status detection.

```go
func GetARPTable() (map[string]string, error)
```

**Returns:**
- `map[string]string`: IP to MAC address mapping
- `error`: ARP table access error or nil

#### `ResolveHostname`

Resolves IP address to hostname.

```go
func ResolveHostname(ip string) string
```

**Parameters:**
- `ip`: IP address to resolve

**Returns:**
- `string`: Resolved hostname or IP if resolution fails

## Error Handling

### Common Errors

```go
// Pi-hole API errors
var (
    ErrAPIConnection    = errors.New("Pi-hole API connection failed")
    ErrAPIAuth         = errors.New("Pi-hole API authentication failed")
    ErrAPITimeout      = errors.New("Pi-hole API request timeout")
    ErrAPIInvalidResp  = errors.New("Pi-hole API invalid response")
)
```

### Error Handling Example

```go
if err := client.Connect(ctx); err != nil {
    switch {
    case errors.Is(err, pihole.ErrAPIConnection):
        logger.Error("API connection issue", slog.String("error", err.Error()))
        // Handle connection-specific errors
    case errors.Is(err, pihole.ErrAPIAuth):
        logger.Error("API authentication failed", slog.String("error", err.Error()))
        // Handle authentication errors
    default:
        logger.Error("Unexpected error", slog.String("error", err.Error()))
    }
}
```

## Configuration Examples

### Complete Configuration

```go
config := &types.Config{
    OnlineOnly: false,
    NoExclude:  false,
    TestMode:   false,
    Quiet:      false,
    Pihole: types.PiholeConfig{
        Host:        "192.168.1.50",
        Port:        80,
        APIEnabled:  true,
        APIPassword: "your-api-password",
        UseHTTPS:    false,
        APITimeout:  30,
    },
    Output: types.OutputConfig{
        Colors:     true,
        Emojis:     true,
        MaxClients: 25,
        MaxDomains: 10,
    },
    Exclusions: types.ExclusionConfig{
        Networks:  []string{"172.16.0.0/12", "127.0.0.0/8"},
        IPs:       []string{"192.168.1.1"},
        Hostnames: []string{"localhost", "docker"},
    },
}
```

### Integration Example

```go
// Complete integration example
func main() {
    // Load configuration
    config := config.Load("config.json")
    
    // Create logger
    logger := logger.New(&logger.Config{
        Component: "main",
        Level:     logger.LevelInfo,
    })
    
    // Create Pi-hole client
    client := pihole.NewClient(&config.Pihole, logger)
    
    // Create analyzer
    analyzer := analyzer.NewAnalyzer(client, config, logger)
    
    // Connect and analyze
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        logger.Error("Connection failed", slog.String("error", err.Error()))
        return
    }
    
    results, err := analyzer.AnalyzeData(ctx)
    if err != nil {
        logger.Error("Analysis failed", slog.String("error", err.Error()))
        return
    }
    
    // Display results
    reporting.DisplayResults(results, config.Output)
}
```

## Testing

### Unit Testing

```go
// Test with mock data source
func TestAnalyzer(t *testing.T) {
    mockSource := &MockDataSource{
        queries: []types.PiholeRecord{
            {Client: "192.168.1.100", Domain: "example.com"},
        },
    }
    
    analyzer := analyzer.NewAnalyzer(mockSource, testConfig, testLogger)
    results, err := analyzer.AnalyzeData(context.Background())
    
    assert.NoError(t, err)
    assert.NotNil(t, results)
}
```

### Integration Testing

```go
// Test with real Pi-hole API
func TestPiholeIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    config := &pihole.Config{
        Host:        os.Getenv("PIHOLE_HOST"),
        APIPassword: os.Getenv("PIHOLE_API_PASSWORD"),
    }
    
    client := pihole.NewClient(config, testLogger)
    err := client.Connect(context.Background())
    assert.NoError(t, err)
}
```

## Performance Considerations

### API Rate Limiting

- Pi-hole API has built-in rate limiting
- Use appropriate `APITimeout` values
- Implement exponential backoff for retries

### Memory Usage

- Large query results are paginated
- Use streaming for very large datasets
- Configure appropriate `Limit` values in `QueryParams`

### Caching

- API responses can be cached for repeated analysis
- Implement TTL-based caching for real-time dashboards
- Cache ARP table lookups to reduce system calls

## Security

### API Authentication

- Store API passwords securely
- Use environment variables for sensitive data
- Enable HTTPS for production deployments
- Consider 2FA TOTP for enhanced security

### Network Security

- Restrict API access to authorized networks
- Use VPN for remote Pi-hole access
- Monitor API access logs

## Related Documentation

- **[Configuration Guide](configuration.md)** - Detailed configuration options
- **[Installation Guide](installation.md)** - Setup instructions
- **[Usage Guide](usage.md)** - Command-line usage
- **[Development Guide](development.md)** - Contributing guidelines
