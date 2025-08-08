# Configuration Guide

This guide covers all configuration options for the Pi-hole Network Analyzer, from basic setup to advanced customization.

## Configuration Overview

The Pi-hole Network Analyzer uses JSON configuration files to manage:
- Pi-hole server connection settings
- Network exclusion rules
- Output formatting preferences
- SSH authentication options

## Default Configuration Location

```bash
# Default configuration file
~/.pihole-analyzer/config.json

# Create default configuration
./pihole-analyzer --create-config

# Use custom configuration location
./pihole-analyzer --config /path/to/custom/config.json
```

## Configuration File Structure

### Complete Configuration Example

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 22,
    "username": "pi",
    "keyPath": "~/.ssh/id_rsa",
    "password": "",
    "dbPath": "/etc/pihole/pihole-FTL.db",
    "timeout": 30
  },
  "exclusions": {
    "networks": [
      "172.16.0.0/12",
      "10.0.0.0/8",
      "127.0.0.0/8",
      "169.254.0.0/16"
    ],
    "hostnames": [
      "docker",
      "localhost",
      "pi.hole",
      "gateway"
    ],
    "ips": [
      "192.168.1.1",
      "192.168.1.2"
    ]
  },
  "output": {
    "colors": true,
    "emoji": true,
    "saveReports": true,
    "reportPath": "~/pihole-reports/",
    "maxClients": 50,
    "sortBy": "queries"
  }
}
```

## Configuration Sections

### Pi-hole Server Configuration

```json
{
  "pihole": {
    "host": "192.168.1.50",        // Pi-hole server IP or hostname
    "port": 22,                     // SSH port (default: 22)
    "username": "pi",               // SSH username
    "keyPath": "~/.ssh/id_rsa",    // SSH private key path
    "password": "",                 // SSH password (if not using keys)
    "dbPath": "/etc/pihole/pihole-FTL.db",  // Pi-hole database path
    "timeout": 30                   // SSH connection timeout (seconds)
  }
}
```

#### Pi-hole Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `host` | string | **required** | Pi-hole server IP address or hostname |
| `port` | int | `22` | SSH port number |
| `username` | string | `"pi"` | SSH username |
| `keyPath` | string | `"~/.ssh/id_rsa"` | Path to SSH private key |
| `password` | string | `""` | SSH password (alternative to key) |
| `dbPath` | string | `"/etc/pihole/pihole-FTL.db"` | Pi-hole database path |
| `timeout` | int | `30` | SSH connection timeout in seconds |

### Network Exclusions

```json
{
  "exclusions": {
    "networks": [
      "172.16.0.0/12",    // Docker networks
      "10.0.0.0/8",       // Private networks
      "127.0.0.0/8"       // Loopback
    ],
    "hostnames": [
      "docker",           // Docker containers
      "localhost",        // Local machine
      "pi.hole"          // Pi-hole itself
    ],
    "ips": [
      "192.168.1.1",     // Router/gateway
      "192.168.1.50"     // Pi-hole server
    ]
  }
}
```

#### Exclusion Types

| Type | Format | Example | Purpose |
|------|--------|---------|---------|
| `networks` | CIDR notation | `"192.168.0.0/16"` | Exclude entire network ranges |
| `hostnames` | String match | `"docker"` | Exclude by hostname pattern |
| `ips` | IP address | `"192.168.1.1"` | Exclude specific IP addresses |

### Output Configuration

```json
{
  "output": {
    "colors": true,              // Enable colored output
    "emoji": true,               // Enable emoji in output
    "saveReports": true,         // Save reports to files
    "reportPath": "~/reports/",  // Report save directory
    "maxClients": 50,           // Maximum clients to display
    "sortBy": "queries"         // Sort criteria
  }
}
```

#### Output Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `colors` | bool | `true` | Enable terminal colors |
| `emoji` | bool | `true` | Enable emoji symbols |
| `saveReports` | bool | `false` | Save reports to files |
| `reportPath` | string | `"~/pihole-reports/"` | Directory for saved reports |
| `maxClients` | int | `50` | Maximum clients to display |
| `sortBy` | string | `"queries"` | Sort order: `queries`, `unique`, `hostname`, `ip` |

## Configuration Methods

### Method 1: Interactive Setup

```bash
# Run interactive configuration wizard
./pihole-analyzer --pihole-setup
```

The wizard will prompt for:
1. Pi-hole server details
2. SSH authentication preferences
3. Network exclusions
4. Output preferences

### Method 2: Manual Configuration

```bash
# Create default config file
./pihole-analyzer --create-config

# Edit with your preferred editor
nano ~/.pihole-analyzer/config.json
```

### Method 3: Environment Variables

Override configuration with environment variables:

```bash
# Pi-hole server settings
export PIHOLE_HOST="192.168.1.50"
export PIHOLE_USERNAME="pi"
export PIHOLE_KEY_PATH="~/.ssh/id_rsa"

# Run analyzer
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json
```

## Advanced Configuration

### Multiple Pi-hole Servers

Create separate configuration files for each Pi-hole:

```bash
# Primary Pi-hole
./pihole-analyzer --config ~/.pihole-analyzer/primary.json

# Secondary Pi-hole  
./pihole-analyzer --config ~/.pihole-analyzer/secondary.json
```

### SSH Key Authentication

#### Generate SSH Keys

```bash
# Generate new SSH key pair
ssh-keygen -t rsa -b 4096 -f ~/.ssh/pihole_analyzer_key

# Copy public key to Pi-hole server
ssh-copy-id -i ~/.ssh/pihole_analyzer_key.pub pi@192.168.1.50
```

#### Configure Key Path

```json
{
  "pihole": {
    "keyPath": "~/.ssh/pihole_analyzer_key",
    "password": ""
  }
}
```

### Password Authentication

```json
{
  "pihole": {
    "keyPath": "",
    "password": "your-ssh-password"
  }
}
```

⚠️ **Security Note**: SSH keys are more secure than passwords. Use passwords only for testing.

### Custom Database Paths

Different Pi-hole installations may use different database paths:

```json
{
  "pihole": {
    "dbPath": "/var/lib/pihole/pihole-FTL.db"     // Alternative location
  }
}
```

Common Pi-hole database locations:
- `/etc/pihole/pihole-FTL.db` (default)
- `/var/lib/pihole/pihole-FTL.db`
- `/opt/pihole/pihole-FTL.db`

## Network Exclusion Patterns

### Docker Networks

```json
{
  "exclusions": {
    "networks": [
      "172.16.0.0/12",    // Default Docker bridge
      "172.17.0.0/16",    // Docker bridge
      "172.18.0.0/16",    // Docker compose networks
      "172.19.0.0/16",
      "172.20.0.0/16"
    ]
  }
}
```

### Common Infrastructure

```json
{
  "exclusions": {
    "hostnames": [
      "router",
      "gateway", 
      "modem",
      "switch",
      "access-point",
      "repeater"
    ],
    "ips": [
      "192.168.1.1",      // Common gateway
      "192.168.0.1",      // Alternative gateway
      "10.0.0.1"          // Enterprise gateway
    ]
  }
}
```

### IoT Device Patterns

```json
{
  "exclusions": {
    "hostnames": [
      "chromecast",
      "alexa",
      "ring-",
      "nest-",
      "philips-hue"
    ]
  }
}
```

## Configuration Validation

### Validate Configuration

```bash
# Show current configuration
./pihole-analyzer --show-config

# Test Pi-hole connection
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet
```

### Common Configuration Errors

#### SSH Connection Issues

```json
// ❌ Incorrect
{
  "pihole": {
    "host": "pihole.local",     // May not resolve
    "keyPath": "~/ssh/key"      // Wrong path
  }
}

// ✅ Correct
{
  "pihole": {
    "host": "192.168.1.50",     // Use IP address
    "keyPath": "~/.ssh/id_rsa"  // Correct path
  }
}
```

#### Invalid Network Ranges

```json
// ❌ Incorrect
{
  "exclusions": {
    "networks": [
      "192.168.1.0",           // Missing CIDR
      "10.0.0.0/33"           // Invalid CIDR
    ]
  }
}

// ✅ Correct
{
  "exclusions": {
    "networks": [
      "192.168.1.0/24",        // Valid CIDR
      "10.0.0.0/8"            // Valid CIDR
    ]
  }
}
```

## Configuration Templates

### Home Network Template

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "username": "pi",
    "keyPath": "~/.ssh/id_rsa"
  },
  "exclusions": {
    "networks": ["127.0.0.0/8"],
    "hostnames": ["router", "pi.hole"],
    "ips": ["192.168.1.1", "192.168.1.50"]
  },
  "output": {
    "colors": true,
    "emoji": true,
    "maxClients": 25
  }
}
```

### Enterprise Network Template

```json
{
  "pihole": {
    "host": "10.0.0.50",
    "username": "pihole-admin",
    "keyPath": "~/.ssh/enterprise_key"
  },
  "exclusions": {
    "networks": [
      "10.0.0.0/8",
      "172.16.0.0/12",
      "127.0.0.0/8"
    ],
    "hostnames": [
      "server-",
      "workstation-",
      "printer-"
    ]
  },
  "output": {
    "colors": false,
    "emoji": false,
    "saveReports": true,
    "maxClients": 100
  }
}
```

### Development Template

```json
{
  "pihole": {
    "host": "localhost",
    "port": 2222,
    "username": "dev",
    "keyPath": "~/.ssh/dev_key"
  },
  "exclusions": {
    "networks": ["172.16.0.0/12"],
    "hostnames": ["docker", "localhost"]
  },
  "output": {
    "colors": true,
    "emoji": true,
    "saveReports": false
  }
}
```

## Troubleshooting Configuration

### Debug Configuration Loading

```bash
# Verbose configuration display
./pihole-analyzer --show-config --no-color

# Test with different config file
./pihole-analyzer --config /tmp/test-config.json --show-config
```

### Common Solutions

1. **File permissions**: `chmod 600 ~/.pihole-analyzer/config.json`
2. **Path expansion**: Use absolute paths for SSH keys
3. **JSON syntax**: Validate JSON with online tools
4. **SSH connectivity**: Test SSH manually first

## Next Steps

- **[SSH Setup Guide](ssh-setup.md)** - Configure SSH connectivity
- **[Usage Guide](usage.md)** - Learn command-line options
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions

---

**Configuration complete!** Your Pi-hole Network Analyzer is ready to use.
