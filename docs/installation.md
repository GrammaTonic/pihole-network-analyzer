# Installation Guide

This guide provides detailed instructions for installing and setting up the Pi-hole Network Analyzer.

## System Requirements

### Minimum Requirements
- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.23 or later
- **Memory**: 256MB RAM
- **Storage**: 50MB free space

### Network Requirements
- Pi-hole API access enabled
- Network connectivity to Pi-hole server
- Terminal with color support (recommended)

## Installation Methods

### Method 1: Build from Source (Recommended)

#### Prerequisites
```bash
# Install Go 1.23+
# Visit: https://golang.org/dl/

# Verify Go installation
go version
```

#### Clone and Build
```bash
# Clone the repository
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer

# Build using Makefile
make build

# Or build directly with Go
go build -o pihole-analyzer ./cmd/pihole-analyzer

# Verify installation
./pihole-analyzer --help
```

### Method 2: Go Install (Direct)

```bash
# Install directly from source
go install github.com/GrammaTonic/pihole-network-analyzer/cmd/pihole-analyzer@latest

# Binary will be in $GOPATH/bin or $GOBIN
pihole-analyzer --help
```

### Method 3: Manual Binary Setup

```bash
# Download binary (when releases are available)
# wget https://github.com/GrammaTonic/pihole-network-analyzer/releases/latest/download/pihole-analyzer-linux-amd64
# chmod +x pihole-analyzer-linux-amd64
# sudo mv pihole-analyzer-linux-amd64 /usr/local/bin/pihole-analyzer
```

## Platform-Specific Instructions

### Linux (Ubuntu/Debian)

```bash
# Install Go
sudo apt update
sudo apt install golang-go

# Install build tools
sudo apt install make git

# Clone and build
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
make build
```

### macOS

```bash
# Install Go using Homebrew
brew install go

# Install build tools (usually pre-installed)
xcode-select --install

# Clone and build
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
make build
```

### Windows

```powershell
# Install Go from https://golang.org/dl/
# Install Git from https://git-scm.com/

# Clone and build
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
go build -o pihole-analyzer.exe ./cmd/pihole-analyzer
```

## Post-Installation Setup

### 1. Verify Installation

```bash
# Check binary works
./pihole-analyzer --help

# Run test mode to verify functionality
./pihole-analyzer --test
```

### 2. Create Configuration Directory

```bash
# Create default configuration
./pihole-analyzer --create-config

# Verify configuration was created
ls -la ~/.pihole-analyzer/
```

### 3. SSH Key Setup (Optional but Recommended)

```bash
# Generate SSH key if you don't have one
ssh-keygen -t rsa -b 4096 -C "pihole-analyzer"

# Copy public key to Pi-hole server
ssh-copy-id pi@your-pihole-ip

# Test SSH connection
ssh pi@your-pihole-ip "ls -la /etc/pihole/"
```

## Configuration

### Initial Pi-hole Setup

```bash
# Interactive setup wizard
./pihole-analyzer --pihole-setup
```

The setup wizard will prompt for:
- Pi-hole server IP address
- SSH username
- SSH authentication method (key or password)
- Pi-hole database path
- Network exclusions

### Manual Configuration

Edit `~/.pihole-analyzer/config.json`:

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 22,
    "username": "pi",
    "keyPath": "~/.ssh/id_rsa",
    "password": "",
    "dbPath": "/etc/pihole/pihole-FTL.db"
  },
  "exclusions": {
    "networks": [
      "172.16.0.0/12",
      "10.0.0.0/8",
      "127.0.0.0/8"
    ],
    "hostnames": [
      "docker",
      "localhost",
      "pi.hole"
    ],
    "ips": [
      "192.168.1.1"
    ]
  },
  "output": {
    "colors": true,
    "emoji": true,
    "saveReports": true,
    "reportPath": "~/pihole-reports/"
  }
}
```

## Verification

### Test Installation

```bash
# Test with mock data
./pihole-analyzer --test

# Test configuration
./pihole-analyzer --show-config

# Test Pi-hole connection (if configured)
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet
```

### Expected Output

A successful test should show:
```
🧪 Running Test Mode
Using mock Pi-hole database
✅ Test mode analysis completed

================================================================================
DNS Query Analysis Summary
================================================================================
Total Clients: 9 | Online: 0 | Total Queries: 20
...
```

## Troubleshooting Installation

### Common Issues

#### Go Version Issues
```bash
# Check Go version
go version

# Update Go if version < 1.23
# Visit: https://golang.org/dl/
```

#### Permission Issues
```bash
# Make binary executable
chmod +x pihole-analyzer

# Check file permissions
ls -la pihole-analyzer
```

#### SSH Connection Issues
```bash
# Test SSH manually
ssh pi@your-pihole-ip

# Check SSH key permissions
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
```

#### Path Issues
```bash
# Add to PATH (optional)
echo 'export PATH=$PATH:$(pwd)' >> ~/.bashrc
source ~/.bashrc
```

### Getting Help

If you encounter issues during installation:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review [GitHub Issues](https://github.com/GrammaTonic/pihole-network-analyzer/issues)
3. Create a new issue with:
   - Operating system and version
   - Go version (`go version`)
   - Error messages
   - Steps to reproduce

## Next Steps

After successful installation:

1. **[Configuration Guide](configuration.md)** - Detailed configuration options
2. **[SSH Setup Guide](ssh-setup.md)** - Pi-hole SSH connection setup
3. **[Usage Guide](usage.md)** - Learn how to use all features
4. **[Development Guide](development.md)** - If you want to contribute

## Performance Optimization

### For Large Pi-hole Databases

```bash
# Increase timeout for large databases
export PIHOLE_ANALYZER_TIMEOUT=300

# Use quiet mode for better performance
./pihole-analyzer --pihole config.json --quiet
```

### Memory Optimization

```bash
# Limit Go memory usage if needed
export GOGC=100
export GOMEMLIMIT=512MiB
```

---

**Installation complete!** You're ready to start analyzing your Pi-hole network data.
