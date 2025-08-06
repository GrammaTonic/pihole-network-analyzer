# Pi-hole Network Analyzer

[![CI/CD Pipeline](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml/badge.svg)](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrammaTonic/pihole-network-analyzer)](https://goreportcard.com/report/github.com/GrammaTonic/pihole-network-analyzer)

A professional Go application to analyze DNS usage patterns and network traffic from Pi-hole servers. Supports both CSV log file analysis and direct Pi-hole server connections via SSH.

## 🚀 Development Workflow

This project uses **feature branch development** with automated CI/CD:

- **🌿 Feature Branches**: Tests run, builds verify, no artifacts created
- **🏗️ Main Branch**: Full builds with multi-platform binaries and releases  
- **⚡ Quick Feedback**: Fast validation for rapid development

See [FEATURE_WORKFLOW.md](FEATURE_WORKFLOW.md) for detailed workflow guide.

### Quick Start for Developers
```bash
# Test your changes before pushing
make ci-test           # Run same tests as CI
./pre-push-test.sh     # Full pre-push validation
make feature-branch    # Validate feature branch ready for push
```

## 🔮 Coming Soon: Monitoring & Docker Support

**🐳 Docker + Prometheus + Grafana Integration** is planned! 

- **Docker Containerization**: Easy deployment with Docker Compose
- **Prometheus Metrics**: Real-time DNS analytics and monitoring  
- **Grafana Dashboards**: Beautiful visualizations and alerting
- **Multi-Platform**: ARM64 support for Raspberry Pi deployments

See our comprehensive [TODO list](TODO_DOCKER_PROMETHEUS_GRAFANA.md) and [implementation roadmap](ROADMAP_DOCKER_MONITORING.md) for details.

## Features

- **CSV Analysis**: Analyzes DNS query logs from CSV files
- **Live Pi-hole Data**: Connects to Pi-hole servers via SSH to analyze real-time data
- **Comprehensive Statistics**: Provides detailed statistics per client including:
  - Total number of queries
  - Unique domains accessed
  - Average reply time (CSV mode)
  - Query type distribution (A, AAAA, CNAME, etc.)
  - Status code distribution (allowed, blocked, cached, etc.)
  - Top domains accessed per client
  - Hardware address mapping (Pi-hole mode)
- **Large File Support**: Handles large CSV files efficiently (tested with 90MB+ files)
- **Detailed Reports**: Generates comprehensive reports saved to text files
- **Smart Sorting**: Sorts clients by query volume for easy identification of heavy users

## Data Sources

### CSV Files
The application expects CSV files with the following columns:
- ID, DateTime, Domain, Query Type, Status, Client IP, Forward, Additional Info, Reply Type, Reply Time, DNSSEC, List ID, EDE

### Pi-hole Database
Connects directly to Pi-hole's SQLite database via SSH to execute the query:
```sql
SELECT
    q.timestamp,
    q.client,
    n.hwaddr,
    q.domain,
    CASE q.status
        WHEN 0 THEN 'Unknown'
        WHEN 1 THEN 'Blocked (gravity)'
        WHEN 2 THEN 'Forwarded'
        WHEN 3 THEN 'Cached'
        WHEN 4 THEN 'Blocked (regex/wildcard)'
        WHEN 5 THEN 'Blocked (exact)'
        WHEN 6 THEN 'Blocked (external, IP)'
        WHEN 7 THEN 'Blocked (external, NULL)'
        WHEN 8 THEN 'Blocked (external, NXDOMAIN)'
        WHEN 9 THEN 'Blocked (gravity, CNAME)'
        WHEN 10 THEN 'Blocked (regex/wildcard, CNAME)'
        WHEN 11 THEN 'Blocked (exact, CNAME)'
        ELSE 'Unknown'
    END AS status
FROM
    queries q
LEFT JOIN
    network n ON q.client = n.ip;
```

## Installation & Usage

### Prerequisites
- Go 1.21 or later
- For Pi-hole analysis: SSH access to Pi-hole server

### Build and Run

1. **Install dependencies:**
   ```bash
   make install-deps
   ```

2. **Build the application:**
   ```bash
   make build
   ```

### CSV Analysis

3. **Run with default CSV file (test.csv):**
   ```bash
   make run
   ```

4. **Run with a specific CSV file:**
   ```bash
   make run-with-file CSV_FILE=your-logfile.csv
   ```

### Pi-hole Live Analysis

5. **Setup Pi-hole SSH configuration (first time only):**
   ```bash
   make setup-pihole
   ```
   This will prompt for:
   - Pi-hole server IP/hostname
   - SSH port (default: 22)
   - SSH username (default: pi)
   - Authentication method (SSH key or password)
   - Pi-hole database path (default: /etc/pihole/pihole-FTL.db)

6. **Analyze live Pi-hole data:**
   ```bash
   make analyze-pihole
   ```

7. **Alternative direct usage:**
   ```bash
   # CSV analysis
   go run main.go test.csv
   
   # Pi-hole setup
   go run main.go --pihole-setup
   
   # Pi-hole analysis
   go run main.go --pihole pihole-config.json
   ```

### Available Make Commands

- `make help` - Show all available commands
- `make install-deps` - Install Go dependencies
- `make build` - Build the application
- `make run` - Build and run with test.csv
- `make analyze` - Alias for run
- `make run-with-file CSV_FILE=file.csv` - Run with specific CSV file
- `make setup-pihole` - Setup Pi-hole SSH configuration
- `make analyze-pihole` - Analyze Pi-hole live data
- `make test-pihole` - Test Pi-hole connection and analyze
- `make clean` - Clean build artifacts and reports
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make all` - Install deps, format, vet, and build

### Pi-hole Requirements

- SSH access to Pi-hole server
- Either SSH key authentication or password access
- Sudo privileges to read Pi-hole database (`/etc/pihole/pihole-FTL.db`)
- Network connectivity between analyzer and Pi-hole server

## Output

The application provides:

1. **Console Output:**
   - Summary statistics
   - Top 20 clients by query count
   - Hardware addresses when available (Pi-hole mode)
   - Detailed analysis of top 5 clients with:
     - Total queries and unique domains
     - Average reply time (CSV mode)
     - Hardware addresses (Pi-hole mode)
     - Top 10 most accessed domains
     - Query type distribution
     - Status code distribution

2. **Detailed Report File:**
   - Complete analysis of all clients saved to `dns_usage_report_YYYYMMDD_HHMMSS.txt`
   - Sorted by query volume for easy analysis

## 🎨 Colorized Output

The analyzer features **rich terminal output** with colors and emojis for enhanced readability:

### Color Scheme
- **🔵 Blue**: Processing indicators and informational messages
- **🟢 Green**: Success messages, high query counts, online clients
- **🟡 Yellow**: Section headers, moderate activity levels  
- **🔴 Red**: Warnings, offline clients, blocked content
- **🟠 Orange**: IP addresses and important identifiers
- **🟣 Cyan**: Domain names and hostnames
- **⚪ Gray**: Hardware addresses and secondary information

### Visual Elements
- **✅ Online Status**: Green checkmark for active devices
- **❌ Offline Status**: Red X for inactive devices  
- **🔄 Processing**: Blue progress indicators with emojis
- **🎯 Highlights**: Color-coded statistics and percentages
- **📊 Tables**: Properly aligned columns with colored headers

### Customization
Control the visual output with command-line flags:
```bash
# Disable colors (for scripts or non-color terminals)
go run main.go --no-color test.csv

# Disable emojis (for compatibility or preference)
go run main.go --no-emoji test.csv

# Disable both colors and emojis
go run main.go --no-color --no-emoji test.csv
```

### Examples
```bash
# Standard colorized output
🔄 Processing large file, please wait...
🔄 Processing CSV records with exclusions...
✅ ARP table refresh completed
✅ Hostname resolution completed

# Client statistics with colors
192.168.2.6 (E0:69:95:4F:...)  pi.hole              115211   322    ✅ Online
192.168.2.210                  s21-van-marloes...   114690   1301   ❌ Offline

# Top domains with highlighting  
     api.spotify.com: 17731 queries
     mobile.events.data.microsoft.com: 15749 queries
```

The colorized output makes it easier to:
- **Quickly identify** online vs offline devices
- **Spot patterns** in DNS usage with visual cues
- **Track progress** during large file processing
- **Distinguish** between different types of information

## Query Types

- A (1) - IPv4 address
- NS (2) - Name server
- CNAME (5) - Canonical name
- SOA (6) - Start of authority
- PTR (12) - Pointer
- MX (15) - Mail exchange
- TXT (16) - Text
- AAAA (28) - IPv6 address
- SRV (33) - Service

## Status Codes

- ALLOWED (1) - Query allowed
- FORWARDED (2) - Query forwarded to upstream
- CACHED (3) - Response from cache
- BLOCKED_* (4-11) - Various blocking reasons
- RETRIED (12-13) - Query retried
- NOT_BLOCKED (14) - Query not blocked
- UPSTREAM_ERROR (15) - Upstream server error
- *_CNAME (16-17) - CNAME-related statuses

## Performance

The application is optimized for large files:
- Processes records in batches
- Shows progress for large datasets
- Memory-efficient CSV parsing
- Handles malformed records gracefully

## Examples

```bash
# CSV Analysis
make run
./analyze.sh test.csv
make run-with-file CSV_FILE=dns-logs-2025-08.csv

# Pi-hole Analysis
make setup-pihole              # First-time setup
make analyze-pihole            # Analyze current Pi-hole data
go run main.go --pihole pihole-config.json

# Clean up generated files
make clean
```

## Sample Configuration File (pihole-config.json)

```json
{
  "host": "192.168.1.100",
  "port": "22",
  "username": "pi",
  "password": "",
  "keyfile": "/home/user/.ssh/id_rsa",
  "dbpath": "/etc/pihole/pihole-FTL.db"
}
```

## Sample Analysis Results

Based on the included `test.csv` file (948,160 DNS queries from 46 unique clients):

**Top 5 Clients by Query Volume:**
1. **172.20.0.8** - 237,921 queries (25.09%) - 7 unique domains
2. **172.20.0.2** - 193,514 queries (20.41%) - 6 unique domains  
3. **192.168.2.6** - 115,211 queries (12.15%) - 322 unique domains
4. **192.168.2.210** - 114,690 queries (12.10%) - 1,301 unique domains
5. **192.168.2.123** - 56,257 queries (5.93%) - 689 unique domains

**Key Insights:**
- Most active clients are in Docker networks (172.x.x.x ranges)
- Some clients show very focused usage (few unique domains, many queries)
- Mobile/application traffic visible (Microsoft telemetry, advertising SDKs)
- Mix of cached responses, forwarded queries, and blocked content
