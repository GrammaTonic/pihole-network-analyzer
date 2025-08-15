# Usage Guide

This guide covers all the ways to use the Pi-hole Network Analyzer, from basic analysis to advanced workflows with Pi-hole API connectivity.

## Basic Usage

### Quick Start Commands

```bash
# Analyze Pi-hole with default configuration
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Run with test data (no Pi-hole required)
./pihole-analyzer-test --test

# Interactive Pi-hole API setup
./pihole-analyzer --pihole-setup

# Show help
./pihole-analyzer --help
```

## Command-Line Interface

### Core Operations

| Command             | Description                           | Example                                   |
| ------------------- | ------------------------------------- | ----------------------------------------- |
| `--pihole <config>` | Analyze Pi-hole with configuration    | `--pihole ~/.pihole-analyzer/config.json` |
| `--pihole-setup`    | Interactive Pi-hole API configuration | `--pihole-setup`                          |
| `--test`            | Run with mock data (use test binary)  | `./pihole-analyzer-test --test`           |

### Configuration Management

| Command           | Description                   | Example                   |
| ----------------- | ----------------------------- | ------------------------- |
| `--config <path>` | Use custom configuration file | `--config /tmp/test.json` |
| `--create-config` | Create default configuration  | `--create-config`         |
| `--show-config`   | Display current configuration | `--show-config`           |

### Output Control

| Command      | Description                   | Example      |
| ------------ | ----------------------------- | ------------ |
| `--quiet`    | Suppress non-essential output | `--quiet`    |
| `--no-color` | Disable colored output        | `--no-color` |
| `--no-emoji` | Disable emoji symbols         | `--no-emoji` |

### Filtering Options

| Command         | Description                | Example         |
| --------------- | -------------------------- | --------------- |
| `--online-only` | Show only online clients   | `--online-only` |
| `--no-exclude`  | Disable default exclusions | `--no-exclude`  |

## Configuration Examples

### Basic Pi-hole API Configuration

```json
{
  "pihole": {
    "host": "192.168.1.100",
    "port": 80,
    "api_enabled": true,
    "api_password": "your-api-password",
    "use_https": false,
    "api_timeout": 30
  },
  "output": {
    "format": "text",
    "colors": true,
    "emoji": true
  },
  "logging": {
    "level": "info",
    "colors": true,
    "emoji": true
  }
}
```

### HTTPS Pi-hole Configuration

```json
{
  "pihole": {
    "host": "pihole.local",
    "port": 443,
    "api_enabled": true,
    "api_password": "secure-password",
    "use_https": true,
    "api_timeout": 60
  },
  "output": {
    "format": "text",
    "file": "/var/log/pihole-analysis.txt",
    "colors": false,
    "emoji": false
  },
  "logging": {
    "level": "debug",
    "file": "/var/log/pihole-analyzer.log",
    "colors": false,
    "emoji": false
  }
}
```

### 2FA Enabled Pi-hole

```json
{
  "pihole": {
    "host": "192.168.1.100",
    "port": 80,
    "api_enabled": true,
    "api_password": "your-api-password",
    "api_totp": "your-totp-secret",
    "use_https": false,
    "api_timeout": 30
  }
}
```

## Common Workflows

### Daily Analysis

```bash
# Basic network analysis
./pihole-analyzer --pihole config.json

# Focus on online devices only
./pihole-analyzer --pihole config.json --online-only

# Generate report without exclusions
./pihole-analyzer --pihole config.json --no-exclude
```

### Automated Reporting

```bash
#!/bin/bash
# daily-analysis.sh

# Configuration
CONFIG_FILE="/home/user/.pihole-analyzer/config.json"
REPORT_DIR="/var/reports/pihole"
DATE=$(date +%Y-%m-%d)

# Create reports directory
mkdir -p "$REPORT_DIR"

# Generate daily report
./pihole-analyzer \
  --pihole "$CONFIG_FILE" \
  --no-color \
  --no-emoji \
  > "$REPORT_DIR/pihole-analysis-$DATE.txt"

# Log completion
echo "$(date): Daily Pi-hole analysis completed" >> "$REPORT_DIR/analysis.log"
```

### Development Testing

```bash
# Test with mock data
./pihole-analyzer-test --test

# Test configuration validation
./pihole-analyzer --show-config --config test-config.json

# Debug mode with verbose logging
LOG_LEVEL=debug ./pihole-analyzer --pihole config.json
```

## Container Usage

### Docker Commands

```bash
# Pull latest image
docker pull ghcr.io/grammatonic/pihole-analyzer:latest

# Run with configuration file
docker run --rm \
  -v $(pwd)/config.json:/app/config.json:ro \
  ghcr.io/grammatonic/pihole-analyzer:latest \
  --config /app/config.json

# Run with output to host filesystem
docker run --rm \
  -v $(pwd)/config.json:/app/config.json:ro \
  -v $(pwd)/reports:/app/reports \
  ghcr.io/grammatonic/pihole-analyzer:latest \
  --config /app/config.json
```

### Docker Compose

```yaml
version: "3.8"

services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    container_name: pihole-analyzer

    volumes:
      - ./config.json:/app/config.json:ro
      - ./reports:/app/reports

    environment:
      - LOG_LEVEL=info
      - NO_COLOR=false
      - NO_EMOJI=false

    command: ["--config", "/app/config.json"]
```

### Container Development

```bash
# Development container with caching
RUN apk --no-cache add ca-certificates

# Copy configuration
COPY config.json /app/config.json

# Run analyzer
CMD ["./pihole-analyzer", "--config", "/app/config.json"]
```

## Advanced Usage

### Performance Optimization

```bash
# Use compression for network efficiency
./pihole-analyzer --pihole config.json --quiet

# Reduce API timeout for faster fails
# (Configure in config.json: "api_timeout": 10)

# Run in background for scheduled analysis
nohup ./pihole-analyzer --pihole config.json --quiet > analysis.log 2>&1 &
```

### Custom Exclusions

```json
{
  "exclusions": {
    "networks": ["172.17.0.0/16", "10.0.0.0/8"],
    "hostnames": ["localhost", "docker-host"],
    "clients": ["192.168.1.1"]
  }
}
```

### Output Formatting

```bash
# JSON-like structured output (future feature)
./pihole-analyzer --pihole config.json --format json

# File output with custom path
./pihole-analyzer --pihole config.json --output /tmp/report.txt

# No colors for log processing
./pihole-analyzer --pihole config.json --no-color --no-emoji
```

## Troubleshooting

### Common Issues

#### Pi-hole API Connection Failures

```bash
# Test API connectivity manually
curl "http://192.168.1.100/admin/api.php"

# Verify Pi-hole is running
ping 192.168.1.100

# Check Pi-hole logs
tail -f /var/log/pihole.log
```

#### Authentication Issues

```bash
# Verify API password
curl "http://192.168.1.100/admin/api.php?auth=your-password"

# Test with simple request
./pihole-analyzer --pihole-setup
```

#### Network Connectivity

```bash
# Test DNS resolution
nslookup pihole.local

# Check firewall settings
telnet 192.168.1.100 80

# Verify network routing
traceroute 192.168.1.100
```

### Debug Mode

```bash
# Enable verbose API debugging
export LOG_LEVEL=debug

# Run with detailed logging
./pihole-analyzer --pihole config.json

# Check structured logs
tail -f /var/log/pihole-analyzer.log | jq '.'
```

### Performance Issues

```bash
# Check API response times
time curl "http://192.168.1.100/admin/api.php"

# Monitor network usage
iftop -i eth0

# Profile application performance
go tool pprof pihole-analyzer cpu.prof
```

## Security Best Practices

### Pi-hole API Security

1. **Use strong API passwords**
2. **Enable HTTPS when possible**
3. **Restrict API access by IP**
4. **Monitor API usage logs**

### Network Security

1. **Use internal networks only**
2. **Implement firewall rules**
3. **Monitor access patterns**
4. **Regular security audits**

### Container Security

1. **Use non-root containers**
2. **Read-only filesystems**
3. **Minimal attack surface**
4. **Regular image updates**

## Integration Examples

### Cron Jobs

```bash
# /etc/crontab entry for daily analysis
0 6 * * * user /usr/local/bin/pihole-analyzer --pihole /etc/pihole-analyzer/config.json --quiet
```

### Monitoring Integration

```bash
#!/bin/bash
# monitoring-integration.sh

# Run analysis and capture metrics
RESULT=$(./pihole-analyzer --pihole config.json --quiet --online-only)
CLIENT_COUNT=$(echo "$RESULT" | grep "Total clients" | awk '{print $3}')

# Send to monitoring system
curl -X POST http://monitoring:8080/metrics \
  -d "pihole_clients_online=$CLIENT_COUNT"
```

### Log Analysis

```bash
# Parse structured logs
grep "level=ERROR" /var/log/pihole-analyzer.log | jq '.msg'

# Monitor API calls
grep "pihole-api" /var/log/pihole-analyzer.log | jq '.duration'

# Track client statistics
grep "client_stats" /var/log/pihole-analyzer.log | jq '.client_count'
```

## Related Documentation

- **[Configuration Guide](configuration.md)** - Detailed configuration options
- **[API Reference](api.md)** - Pi-hole API connectivity guide
- **[Installation Guide](installation.md)** - Setup and installation
- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions
- **[Development Guide](development.md)** - Building and extending
- **[Container Usage Guide](container-usage.md)** - Docker deployment

## FAQ

### Q: Can I analyze multiple Pi-hole instances?

A: Currently, the analyzer connects to one Pi-hole instance at a time. Configure different config files for multiple instances.

### Q: How often should I run the analyzer?

A: For daily monitoring, running once per day is sufficient. For real-time monitoring, consider running every hour.

### Q: Does the analyzer affect Pi-hole performance?

A: No, the analyzer uses read-only API calls and doesn't modify Pi-hole configuration or data.

### Q: Can I run this on Raspberry Pi?

A: Yes, ARM builds are available. Use the appropriate container image for your architecture.

### Q: How do I backup my configuration?

A: Copy your config.json file. The configuration is self-contained and portable.

This usage guide provides comprehensive coverage of the Pi-hole Network Analyzer's capabilities with API-only connectivity.
