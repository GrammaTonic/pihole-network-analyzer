# Configuration Guide

Complete guide to configuring the Pi-hole Network Analyzer for API-based connectivity.

## Overview

The Pi-hole Network Analyzer uses JSON configuration files to manage:
- Pi-hole API connection settings
- Authentication credentials
- Output formatting preferences
- Network exclusion rules
- Logging configuration

## Quick Start Configuration

### Basic Pi-hole API Configuration

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 80,
    "apiEnabled": true,
    "apiPassword": "your-api-password",
    "useHTTPS": false,
    "apiTimeout": 30
  },
  "output": {
    "colors": true,
    "emojis": true,
    "maxClients": 25,
    "maxDomains": 10
  },
  "exclusions": {
    "networks": ["172.16.0.0/12", "127.0.0.0/8"],
    "ips": ["192.168.1.1"],
    "hostnames": ["localhost", "docker"]
  }
}
```

## Configuration Structure

### Pi-hole Connection Settings

```json
{
  "pihole": {
    "host": "192.168.1.50",          // Pi-hole server IP/hostname
    "port": 80,                      // Pi-hole web interface port (80/443)
    "apiEnabled": true,              // Enable API access
    "apiPassword": "your-password",  // Pi-hole API password
    "apiTOTP": "",                   // 2FA TOTP secret (optional)
    "useHTTPS": false,               // Force HTTPS connection
    "apiTimeout": 30                 // API request timeout (seconds)
  }
}
```

#### Pi-hole Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | - | Pi-hole server hostname/IP |
| `port` | int | `80` | Pi-hole web interface port |
| `apiEnabled` | bool | `true` | Enable Pi-hole API access |
| `apiPassword` | string | - | Pi-hole API password |
| `apiTOTP` | string | `""` | 2FA TOTP secret |
| `useHTTPS` | bool | `false` | Force HTTPS connections |
| `apiTimeout` | int | `30` | API timeout in seconds |

### Output Configuration

Controls display formatting and report generation.

```json
{
  "output": {
    "colors": true,                  // Enable colored output
    "emojis": true,                  // Enable emoji indicators
    "verbose": false,                // Enable verbose output
    "format": "table",               // Output format: table, json
    "maxClients": 25,                // Max clients to display
    "maxDomains": 10,                // Max domains per client
    "saveReports": false,            // Save reports to files
    "reportDir": "./reports"         // Report output directory
  }
}
```

### Exclusion Rules

Configure which networks, IPs, and hostnames to exclude from analysis.

```json
{
  "exclusions": {
    "networks": [                    // CIDR networks to exclude
      "172.16.0.0/12",              // Docker networks
      "127.0.0.0/8",                // Loopback
      "169.254.0.0/16"              // Link-local
    ],
    "ips": [                         // Specific IPs to exclude
      "192.168.1.1",               // Router
      "192.168.1.50"               // Pi-hole itself
    ],
    "hostnames": [                   // Hostnames to exclude
      "localhost",
      "docker",
      "router"
    ],
    "enableDocker": true             // Auto-exclude Docker networks
  }
}
```

### Logging Configuration

```json
{
  "logging": {
    "level": "INFO",                 // Log level: DEBUG, INFO, WARN, ERROR
    "enableColors": true,            // Colored log output
    "enableEmojis": true,            // Emoji in logs
    "outputFile": "",                // Log file path (empty = stdout)
    "component": "main"              // Component identifier
  }
}
```

## Environment Variables

Override configuration values using environment variables:

```bash
# Pi-hole connection
export PIHOLE_HOST="192.168.1.50"
export PIHOLE_PORT="80"
export PIHOLE_API_PASSWORD="your-password"
export PIHOLE_USE_HTTPS="false"

# Output preferences
export PIHOLE_NO_COLOR="false"
export PIHOLE_NO_EMOJI="false"
export PIHOLE_VERBOSE="false"

# Exclusions
export PIHOLE_NO_EXCLUDE="false"
```

## Configuration Examples

### Basic Home Setup

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "apiPassword": "home-network-password"
  },
  "output": {
    "maxClients": 15
  }
}
```

### Enterprise Environment

```json
{
  "pihole": {
    "host": "pihole.company.com",
    "port": 443,
    "useHTTPS": true,
    "apiPassword": "enterprise-password",
    "apiTOTP": "TOTP_SECRET_HERE"
  },
  "output": {
    "colors": false,
    "emojis": false,
    "saveReports": true,
    "reportDir": "/var/log/pihole-analyzer"
  },
  "exclusions": {
    "networks": [
      "10.0.0.0/8",
      "172.16.0.0/12",
      "192.168.0.0/16"
    ]
  }
}
```

### Development/Testing

```json
{
  "pihole": {
    "host": "dev-pihole.local",
    "apiPassword": "dev-password"
  },
  "output": {
    "verbose": true,
    "maxClients": 50
  },
  "logging": {
    "level": "DEBUG",
    "outputFile": "./debug.log"
  }
}
```

## Pi-hole API Setup

### 1. Enable Pi-hole API

Access Pi-hole admin interface:
```
http://your-pihole-ip/admin
```

### 2. Generate API Password

1. Go to **Settings > API / Web interface**
2. Click **"Show API token"**
3. Copy the generated token
4. Use this as your `apiPassword` in configuration

### 3. Optional: Enable 2FA

For enhanced security, enable TOTP 2FA:
1. Go to **Settings > API / Web interface**
2. Enable **"Two-factor authentication"**
3. Scan QR code with authenticator app
4. Use the TOTP secret in `apiTOTP` configuration

## Configuration Validation

The analyzer validates configuration on startup:

```bash
# Validate current configuration
./pihole-analyzer --show-config

# Test Pi-hole API connectivity
./pihole-analyzer --pihole config.json --quiet
```

## Troubleshooting

### Common Configuration Issues

#### API Connection Failed

**Problem**: Cannot connect to Pi-hole API
```json
{
  "pihole": {
    "host": "wrong-ip",
    "apiPassword": "wrong-password"
  }
}
```

**Solution**: Verify correct IP and API password
```json
{
  "pihole": {
    "host": "192.168.1.50",
    "apiPassword": "correct-api-password"
  }
}
```

### Configuration File Locations

1. **Command line**: `--config path/to/config.json`
2. **Default location**: `~/.pihole-analyzer/config.json`
3. **Environment**: `PIHOLE_CONFIG_PATH`

## Best Practices

1. **Use HTTPS**: Enable `useHTTPS` for production
2. **Path expansion**: Use absolute paths for report directories
3. **Security**: Store API passwords securely
4. **API connectivity**: Test Pi-hole API access first

## Related Documentation

- **[Installation Guide](installation.md)** - Setup and installation
- **[API Reference](api.md)** - Pi-hole API integration details
- **[Usage Guide](usage.md)** - Command-line usage
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
