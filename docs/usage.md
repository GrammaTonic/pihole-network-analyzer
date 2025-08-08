# Usage Guide

This guide covers all the ways to use the Pi-hole Network Analyzer, from basic analysis to advanced workflows.

## Basic Usage

### Quick Start Commands

```bash
# Analyze Pi-hole with default configuration
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Run with test data (no Pi-hole required)
./pihole-analyzer --test

# Interactive setup
./pihole-analyzer --pihole-setup

# Show help
./pihole-analyzer --help
```

## Command-Line Interface

### Core Operations

| Command | Description | Example |
|---------|-------------|---------|
| `--pihole <config>` | Analyze Pi-hole with configuration | `--pihole ~/.pihole-analyzer/config.json` |
| `--pihole-setup` | Interactive Pi-hole configuration | `--pihole-setup` |
| `--test` | Run with mock data | `--test` |

### Configuration Management

| Command | Description | Example |
|---------|-------------|---------|
| `--config <path>` | Use custom configuration file | `--config /tmp/test.json` |
| `--create-config` | Create default configuration | `--create-config` |
| `--show-config` | Display current configuration | `--show-config` |

### Output Control

| Command | Description | Example |
|---------|-------------|---------|
| `--quiet` | Suppress non-essential output | `--quiet` |
| `--no-color` | Disable colored output | `--no-color` |
| `--no-emoji` | Disable emoji symbols | `--no-emoji` |

### Filtering Options

| Command | Description | Example |
|---------|-------------|---------|
| `--online-only` | Show only online clients | `--online-only` |
| `--no-exclude` | Disable default exclusions | `--no-exclude` |

## Usage Scenarios

### Scenario 1: Daily Network Monitoring

```bash
# Standard analysis with full output
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Focus on active devices only
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --online-only

# Quiet output for logging
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet >> daily-report.log
```

### Scenario 2: Troubleshooting Network Issues

```bash
# Show all devices (including excluded ones)
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --no-exclude

# Text-only output for analysis
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --no-color --no-emoji

# Detailed output with all information
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json > network-analysis.txt
```

### Scenario 3: Automated Reporting

```bash
#!/bin/bash
# Daily report script

DATE=$(date +%Y%m%d)
REPORT_DIR="$HOME/pihole-reports"
mkdir -p "$REPORT_DIR"

# Generate daily report
./pihole-analyzer \
  --pihole ~/.pihole-analyzer/config.json \
  --quiet \
  --no-color \
  > "$REPORT_DIR/pihole-report-$DATE.txt"

echo "Report saved to $REPORT_DIR/pihole-report-$DATE.txt"
```

### Scenario 4: Development and Testing

```bash
# Test with mock data
./pihole-analyzer --test

# Test configuration without connecting to Pi-hole
./pihole-analyzer --show-config

# Test custom configuration
./pihole-analyzer --config test-config.json --test
```

## Understanding the Output

### Summary Section

```
================================================================================
DNS Query Analysis Summary
================================================================================
Total Clients: 12 | Online: 8 | Total Queries: 1,247
```

- **Total Clients**: Number of unique devices found
- **Online**: Devices currently connected (have MAC addresses in ARP table)
- **Total Queries**: Sum of all DNS queries in the analysis period

### Client Statistics Table

```
Client Statistics
-----------------------------------------------------------------------------------------------------------
IP Address       Hostname           MAC Address        Status     Queries    Unique       Avg RT   Top Domain Query %
192.168.1.100   johns-macbook      aa:bb:cc:dd:ee:ff  ✅ Online  234        89          0.002s   google.com      24%
192.168.1.101   smart-tv           11:22:33:44:55:66  ✅ Online  156        23          0.001s   netflix.com     67%
192.168.1.102   iphone-sarah       77:88:99:aa:bb:cc  ❌ Offline 89         45          N/A      facebook.com    31%
```

#### Column Descriptions

| Column | Description |
|--------|-------------|
| **IP Address** | Client's IP address |
| **Hostname** | Device hostname (if available) |
| **MAC Address** | Hardware address (truncated for display) |
| **Status** | Online/Offline based on ARP table |
| **Queries** | Total DNS queries from this client |
| **Unique** | Number of unique domains queried |
| **Avg RT** | Average response time (if available) |
| **Top Domain** | Most frequently queried domain |
| **Query %** | Percentage of queries to top domain |

### Status Indicators

| Symbol | Meaning |
|--------|---------|
| ✅ Online | Device is currently connected (MAC in ARP table) |
| ❌ Offline | Device not found in current ARP table |
| 🟡 Unknown | Status could not be determined |

## Advanced Usage

### Custom Configuration Files

```bash
# Use different configurations for different networks
./pihole-analyzer --config ~/.pihole-analyzer/home.json      # Home network
./pihole-analyzer --config ~/.pihole-analyzer/office.json    # Office network
./pihole-analyzer --config ~/.pihole-analyzer/lab.json       # Lab environment
```

### Combining Options

```bash
# Comprehensive analysis with custom settings
./pihole-analyzer \
  --pihole ~/.pihole-analyzer/config.json \
  --online-only \
  --quiet \
  --no-emoji > analysis.txt

# Development testing with custom config
./pihole-analyzer \
  --config test-config.json \
  --test \
  --no-color
```

### Output Redirection

```bash
# Save full output to file
./pihole-analyzer --pihole config.json > full-report.txt

# Save only errors to separate file
./pihole-analyzer --pihole config.json 2> errors.log

# Append to existing log file
./pihole-analyzer --pihole config.json --quiet >> daily.log

# Split output and errors
./pihole-analyzer --pihole config.json > output.txt 2> errors.txt
```

## Integration Examples

### Bash Scripts

```bash
#!/bin/bash
# Network monitoring script

CONFIG="$HOME/.pihole-analyzer/config.json"
ALERT_THRESHOLD=100

# Run analysis and capture output
OUTPUT=$(./pihole-analyzer --pihole "$CONFIG" --quiet --no-color)

# Extract total queries
TOTAL_QUERIES=$(echo "$OUTPUT" | grep "Total Queries:" | awk '{print $4}')

# Alert if queries exceed threshold
if [ "$TOTAL_QUERIES" -gt "$ALERT_THRESHOLD" ]; then
    echo "High DNS query volume detected: $TOTAL_QUERIES queries"
    # Send notification (email, Slack, etc.)
fi
```

### Cron Jobs

```bash
# Add to crontab (crontab -e)

# Daily report at 6 AM
0 6 * * * /usr/local/bin/pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet >> ~/logs/daily-pihole.log

# Hourly monitoring
0 * * * * /usr/local/bin/pihole-analyzer --pihole ~/.pihole-analyzer/config.json --online-only --quiet --no-color > /tmp/pihole-status.txt

# Weekly comprehensive report
0 7 * * 1 /usr/local/bin/pihole-analyzer --pihole ~/.pihole-analyzer/config.json --no-exclude > ~/reports/weekly-$(date +\%Y\%m\%d).txt
```

### Docker Integration

```dockerfile
# Dockerfile for containerized analysis
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o pihole-analyzer ./cmd/pihole-analyzer

FROM alpine:latest
RUN apk --no-cache add ca-certificates openssh-client
WORKDIR /root/
COPY --from=builder /app/pihole-analyzer .
COPY config.json .pihole-analyzer/

CMD ["./pihole-analyzer", "--pihole", ".pihole-analyzer/config.json"]
```

### Python Integration

```python
#!/usr/bin/env python3
import subprocess
import json
import sys

def run_pihole_analysis(config_path):
    """Run Pi-hole analysis and return parsed output."""
    try:
        result = subprocess.run([
            './pihole-analyzer',
            '--pihole', config_path,
            '--quiet',
            '--no-color'
        ], capture_output=True, text=True, check=True)
        
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running analysis: {e}", file=sys.stderr)
        return None

def main():
    config_path = "~/.pihole-analyzer/config.json"
    output = run_pihole_analysis(config_path)
    
    if output:
        print("Analysis completed successfully")
        # Process output as needed
    else:
        print("Analysis failed")

if __name__ == "__main__":
    main()
```

## Performance Considerations

### Large Pi-hole Databases

```bash
# For databases with millions of queries, use quiet mode
./pihole-analyzer --pihole config.json --quiet

# Limit client display for better performance
# Configure maxClients in config.json
{
  "output": {
    "maxClients": 25
  }
}
```

### Network Optimization

```bash
# Use SSH compression for slow connections
# Add to SSH config (~/.ssh/config):
Host pihole-server
    Compression yes
    CompressionLevel 6
```

### Memory Usage

```bash
# Monitor memory usage during large analyses
/usr/bin/time -v ./pihole-analyzer --pihole config.json

# Limit Go memory if needed
export GOMEMLIMIT=512MiB
./pihole-analyzer --pihole config.json
```

## Troubleshooting Usage

### Common Issues

#### SSH Connection Failures

```bash
# Test SSH connectivity manually
ssh -i ~/.ssh/id_rsa pi@192.168.1.50

# Verify Pi-hole database exists
ssh pi@192.168.1.50 "ls -la /etc/pihole/pihole-FTL.db"
```

#### Permission Issues

```bash
# Check file permissions
ls -la ~/.pihole-analyzer/config.json

# Fix permissions if needed
chmod 600 ~/.pihole-analyzer/config.json
```

#### Configuration Problems

```bash
# Validate configuration
./pihole-analyzer --show-config

# Test with minimal configuration
./pihole-analyzer --test
```

### Debug Mode

```bash
# Enable verbose SSH debugging
export SSH_DEBUG=1
./pihole-analyzer --pihole config.json

# Check Go environment
go env
./pihole-analyzer --help
```

## Best Practices

### Security

1. **Use SSH keys** instead of passwords
2. **Restrict SSH permissions** on Pi-hole server
3. **Store configuration securely** (appropriate file permissions)
4. **Regular key rotation** for production environments

### Performance

1. **Use quiet mode** for automated scripts
2. **Limit output** with maxClients setting
3. **Schedule analysis** during low-traffic periods
4. **Monitor resource usage** on Pi-hole server

### Reliability

1. **Test configuration** before automation
2. **Handle errors gracefully** in scripts
3. **Log output** for troubleshooting
4. **Monitor SSH connectivity** health

## Next Steps

- **[SSH Setup Guide](ssh-setup.md)** - Configure secure connections
- **[Development Guide](development.md)** - Extend functionality
- **[Troubleshooting](troubleshooting.md)** - Solve common problems
- **[API Reference](api.md)** - Internal package documentation

---

**Ready to analyze!** Your Pi-hole Network Analyzer usage guide is complete.
