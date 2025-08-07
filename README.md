# Pi-hole Network Analyzer

[![CI/CD Pipeline](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml/badge.svg)](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrammaTonic/pihole-network-analyzer)](https://goreportcard.com/report/github.com/GrammaTonic/pihole-network-analyzer)

A professional Go application to analyze DNS usage patterns and network traffic from Pi-hole servers. Features **rich colorized terminal output** with smart domain highlighting, visual progress indicators, and comprehensive analytics. Supports both CSV log file analysis and direct Pi-hole server connections via SSH.

## ✨ What's New: Colorized Output!

Transform your DNS analysis with beautiful, informative terminal output:

- **🎨 Color-Coded Analytics**: Instantly spot patterns with intelligent color schemes
- **📊 Visual Progress**: Real-time progress indicators and status updates  
- **🔍 Smart Highlighting**: Automatically categorizes domains (ads/tracking in red, services in green)
- **⚡ Enhanced UX**: Online/offline indicators, emoji status, and organized data display

Perfect for both interactive analysis and automated reporting!

## 🚀 Quick Start

Get started with beautiful DNS analysis in seconds:

```bash
# 1. Clone and build
git clone https://github.com/GrammaTonic/OMG.git
cd OMG
make build

# 2. Run with sample data (experience the colorized output!)
make run

# 3. Or analyze your own CSV file
go run main.go your-dns-logs.csv

# 4. For scripts/automation (text-only mode)
go run main.go --quiet --no-color your-logs.csv
```

**See it in action**: The analyzer will display beautiful progress indicators, color-coded statistics, and smart domain categorization right in your terminal!

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

### 🎨 **Rich Colorized Output** 
- **Beautiful Terminal Display**: Color-coded statistics, progress indicators, and status messages
- **Smart Domain Highlighting**: Automatically highlights tracking/ads (red), major services (green), development sites (cyan)
- **Visual Status Indicators**: Online/offline clients with emoji indicators (✅/❌)
- **Customizable**: Disable colors (`--no-color`) or emojis (`--no-emoji`) for compatibility

### 📊 **Comprehensive Analysis**
- **CSV Analysis**: Analyzes DNS query logs from CSV files with intelligent parsing
- **Live Pi-hole Data**: Connects to Pi-hole servers via SSH to analyze real-time data
- **Detailed Statistics**: Provides per-client analytics including:
  - Total number of queries with color-coded volume indicators
  - Unique domains accessed with smart categorization
  - Average reply time analysis (CSV mode)
  - Query type distribution (A, AAAA, CNAME, etc.)
  - Status code distribution (allowed, blocked, cached, etc.)
  - Top domains accessed per client with intelligent highlighting
  - Hardware address mapping and hostname resolution (Pi-hole mode)

### 🚀 **Performance & Reliability**
- **Large File Support**: Handles massive CSV files efficiently (tested with 90MB+ files)
- **Progress Tracking**: Real-time progress indicators for long-running operations
- **Robust Processing**: Graceful handling of malformed records and network issues
- **Memory Efficient**: Optimized for processing large datasets without memory bloat

### 📝 **Reporting & Export**
- **Detailed Reports**: Generates comprehensive reports saved to timestamped text files
- **Smart Sorting**: Sorts clients by query volume for easy identification of heavy users
- **Multiple Output Formats**: Console display with optional file export

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
- **Go 1.21+** - For building and running the application
- **Terminal with color support** - For the best visual experience (optional)
- **SSH access** - Only required for Pi-hole live analysis

### Quick Installation

```bash
# Method 1: Direct build (recommended)
git clone https://github.com/GrammaTonic/OMG.git
cd OMG
make build

# Method 2: Using Go
go install github.com/GrammaTonic/OMG@latest
```

### 🎨 CSV Analysis (with colorized output)

**Experience the rich terminal interface:**
**Experience the rich terminal interface:**

```bash
# Analyze the included sample data (948,160 DNS queries!)
make run

# Or analyze your own CSV file with full colorized output
go run main.go your-dns-logs.csv

# Quiet mode for automation/scripts
go run main.go --quiet your-logs.csv
```

**What you'll see:**
- 🔄 Real-time progress indicators during file processing
- 📊 Color-coded client statistics and query distributions  
- 🎯 Smart domain highlighting (ads in red, services in green)
- ✅/❌ Visual online/offline status for each device
- 📈 Beautiful formatted tables and charts

### 🔗 Pi-hole Live Analysis

**Connect directly to your Pi-hole server:**
**Connect directly to your Pi-hole server:**

```bash
# 1. First-time setup (interactive configuration)
make setup-pihole

# 2. Analyze live Pi-hole data with colorized output
make analyze-pihole

# 3. Alternative direct usage
go run main.go --pihole pihole-config.json
```

**Live analysis features:**
- Real-time data from your Pi-hole's SQLite database
- Hardware address mapping and hostname resolution
- All the colorized output benefits for live data
- Secure SSH connection with key or password authentication

### 🛠️ Available Make Commands

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
- **🔵 Blue**: Processing indicators, informational messages, and private IP addresses
- **🟢 Green**: Success messages, high query counts, online clients, and major service domains
- **🟡 Yellow**: Section headers, moderate activity levels, and public IP addresses  
- **🔴 Red**: Warnings, offline clients, blocked content, and tracking/ads domains
- **🟠 Orange**: Public IP addresses and important identifiers
- **🟣 Cyan**: Hostnames and development-related domains (GitHub, StackOverflow)
- **⚪ Gray**: Hardware addresses, low activity, and secondary information

### Smart Domain Highlighting
The analyzer intelligently colorizes domain names based on their purpose:
- **🟢 Major Services**: `google.com`, `microsoft.com`, `apple.com` (Green)
- **🔴 Tracking/Ads**: `doubleclick.net`, `telemetry.microsoft.com`, `ads.*` (Red)
- **🟣 Development**: `github.com`, `stackoverflow.com`, `npm.*` (Cyan)
- **Default**: Other domains remain uncolored for clarity

### Visual Elements
- **✅ Online Status**: Green checkmark for active devices
- **❌ Offline Status**: Red X for inactive devices  
- **🔄 Processing**: Blue progress indicators with emojis
- **🎯 Highlights**: Color-coded statistics and percentages
- **📊 Tables**: Properly aligned columns with colored headers

### Customization
Control the visual experience with command-line flags:
```bash
# Disable colors (for scripts, CI/CD, or monochrome terminals)
go run main.go --no-color test.csv

# Disable emojis (for terminal compatibility)
go run main.go --no-emoji test.csv

# Text-only mode (disable both colors and emojis)
go run main.go --no-color --no-emoji test.csv

# Quiet mode (minimal output for scripts)
go run main.go --quiet test.csv
```

### CI/CD & Automation Friendly
The analyzer automatically detects CI environments and non-interactive terminals, switching to text-only mode for compatibility with:
- **GitHub Actions** and other CI/CD systems
- **Automated scripts** and cron jobs
- **Log files** and text processing pipelines
- **Terminal multiplexers** without color support

### Examples
Experience the rich terminal output:
```bash
# Colorized analysis with progress indicators
🔄 Processing large file, please wait...
🔄 Processing CSV records with exclusions...
Excluded: IP 192.168.2.6 is in exclusion list
Excluded: IP 172.20.0.8 is in excluded network 172.16.0.0/12  
✅ ARP table refresh completed (5 entries found)
✅ Hostname resolution completed

# Client statistics with visual indicators and colors
IP Address          Hostname             Queries  Domains  Status    
192.168.2.6         pi.hole              115211   322      ✅ Online
192.168.2.210       s21-van-marloes...   114690   1301     ❌ Offline
192.168.2.123       samsung-galaxy...    56257    689      ✅ Online

# Top domains with intelligent highlighting and categories
📊 Top Domains Accessed:
     google.com: 1731 queries           # 🟢 Green (major service)
     tracking.doubleclick.net: 1249     # 🔴 Red (ads/tracking)  
     github.com: 892 queries            # 🟣 Cyan (development)
     api.spotify.com: 733 queries       # ⚪ No color (regular)
     telemetry.microsoft.com: 421       # 🔴 Red (telemetry)

# Query type distribution with visual formatting
📈 Query Type Distribution:
     A (IPv4):     847234 (89.4%) ████████████████████
     AAAA (IPv6):   95421 (10.1%) ██
     CNAME:          3891 (0.4%)  ▌
     PTR:            1614 (0.2%)  ▌
```

The colorized output transforms DNS analysis by making it easier to:
- **🔍 Instantly identify** online vs offline devices with visual indicators
- **📊 Spot usage patterns** with color-coded statistics and progress bars
- **🚨 Detect suspicious activity** with red highlighting for ads/tracking domains
- **⚡ Track progress** during large file processing with real-time indicators
- **🎯 Focus attention** on important information with strategic color usage
- **🔧 Debug issues** with clear error messages and status indicators

Perfect for both **interactive troubleshooting** and **automated monitoring** workflows!

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

## 🎮 Usage Examples

### Basic Analysis
```bash
# Quick start with beautiful colorized output
make run                                    # Use included test.csv
go run main.go large-dns-logs.csv         # Analyze your own file
```

### Advanced Usage  
```bash
# Automation-friendly (no colors, minimal output)
go run main.go --quiet --no-color logs.csv

# Terminal compatibility mode
go run main.go --no-emoji logs.csv

# Pi-hole live analysis
make setup-pihole && make analyze-pihole

# Clean up generated reports
make clean
```

### Output Examples

**🎨 Interactive Mode (default):**
```
🔄 Processing large file, please wait...
📊 DNS Usage Analysis Results
════════════════════════════════

📈 Summary Statistics:
Total Records: 948,160 queries
Unique Clients: 46 devices  
Time Range: 2025-08-01 to 2025-08-07

🏆 Top 5 Clients by Activity:
┌─────────────────┬──────────────────┬─────────┬─────────┬────────┐
│ IP Address      │ Hostname         │ Queries │ Domains │ Status │  
├─────────────────┼──────────────────┼─────────┼─────────┼────────┤
│ 192.168.2.6     │ pi.hole          │ 115,211 │ 322     │ ✅ Online│
│ 192.168.2.210   │ samsung-galaxy   │ 114,690 │ 1,301   │ ❌ Offline│  
│ 192.168.2.123   │ iphone-12        │ 56,257  │ 689     │ ✅ Online│
└─────────────────┴──────────────────┴─────────┴─────────┴────────┘
```

**⚙️ Automation Mode (`--quiet --no-color`):**
```
Total Records: 948160
Unique Clients: 46
Top Client: 192.168.2.6 (115211 queries)
Report saved: dns_usage_report_20250807_142856.txt
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
