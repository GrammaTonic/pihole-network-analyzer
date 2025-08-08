# Installation Guide

This guide provides detailed instructions for installing and setting up the Pi-hole Network Analyzer.

## System Requirements

### Minimum Requirements
- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.23.12 or later
- **Memory**: 256MB RAM
- **Storage**: 50MB free space

### Network Requirements
- Pi-hole API access enabled
- Network connectivity to Pi-hole server
- Terminal with color support (recommended)

## Installation Methods

### Method 1: Build from Source (Recommended)

#### 1. Clone Repository

```bash
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
```

#### 2. Build Application

```bash
# Build production binary
make build

# Or build both production and test binaries
make build-all

# Or use Go directly
go build -o pihole-analyzer ./cmd/pihole-analyzer
go build -o pihole-analyzer-test ./cmd/pihole-analyzer-test
```

#### 3. Install (Optional)

```bash
# Install to system PATH
sudo cp pihole-analyzer /usr/local/bin/
sudo chmod +x /usr/local/bin/pihole-analyzer

# Or add to your PATH
export PATH=$PATH:$(pwd)
```

### Method 2: Using Docker

#### Pull from Registry

```bash
# Pull latest image
docker pull ghcr.io/grammatonic/pihole-analyzer:latest

# Run with configuration
docker run --rm -v $(pwd)/config.json:/config.json \
  ghcr.io/grammatonic/pihole-analyzer:latest --pihole /config.json
```

#### Build from Source

```bash
# Build development image
make docker-build-dev

# Build production image
make docker-build-prod

# Build multi-architecture images
make docker-build-multi
```

### Method 3: Download Release Binary

```bash
# Download latest release (replace with actual version)
wget https://github.com/GrammaTonic/pihole-network-analyzer/releases/download/v1.0.0/pihole-analyzer-linux-amd64
chmod +x pihole-analyzer-linux-amd64
mv pihole-analyzer-linux-amd64 pihole-analyzer
```

## Initial Setup

### 1. Create Configuration Directory

```bash
mkdir -p ~/.pihole-analyzer
```

### 2. Create Configuration File

```bash
# Generate default configuration
./pihole-analyzer --create-config

# Or create manually
cat > ~/.pihole-analyzer/config.json << 'EOF'
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 80,
    "apiEnabled": true,
    "apiPassword": "your-api-password",
    "useHTTPS": false
  },
  "output": {
    "colors": true,
    "emojis": true,
    "maxClients": 25
  },
  "exclusions": {
    "networks": ["172.16.0.0/12", "127.0.0.0/8"],
    "hostnames": ["localhost", "docker"]
  }
}
EOF
```

### 3. Configure Pi-hole API Access

#### Enable Pi-hole API

1. Access Pi-hole admin interface: `http://your-pihole-ip/admin`
2. Go to **Settings > API / Web interface**
3. Ensure **"API"** is enabled
4. Copy the **API token** (click "Show API token")

#### Update Configuration

```bash
# Edit configuration file
nano ~/.pihole-analyzer/config.json

# Update with your Pi-hole details:
{
  "pihole": {
    "host": "YOUR_PIHOLE_IP",
    "apiPassword": "YOUR_API_TOKEN"
  }
}
```

### 4. Test Installation

```bash
# Test with mock data
./pihole-analyzer-test

# Test Pi-hole connection
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Show configuration
./pihole-analyzer --show-config
```

## Platform-Specific Instructions

### Linux (Ubuntu/Debian)

```bash
# Install dependencies
sudo apt update
sudo apt install git build-essential

# Install Go 1.23+
wget https://go.dev/dl/go1.23.12.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.12.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Build and install
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
make build
sudo cp pihole-analyzer /usr/local/bin/
```

### macOS

```bash
# Install dependencies using Homebrew
brew install git go

# Clone and build
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
make build

# Add to PATH (add to ~/.zshrc or ~/.bash_profile)
export PATH=$PATH:$(pwd)
```

### Windows

#### Using WSL (Recommended)

```bash
# Install WSL and Ubuntu
wsl --install

# Follow Linux instructions in WSL
```

#### Native Windows

```powershell
# Install Go from https://golang.org/dl/
# Download source and build

git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer
go build -o pihole-analyzer.exe ./cmd/pihole-analyzer
```

## Docker Setup

### Development Environment

```bash
# Start development environment
make docker-dev

# Access container shell
make docker-shell

# View logs
make docker-logs
```

### Production Deployment

```bash
# Create production configuration
mkdir -p ./config
cp ~/.pihole-analyzer/config.json ./config/

# Start production container
docker-compose -f docker-compose.prod.yml up -d

# Or use make target
make docker-prod
```

### Container Configuration

```yaml
# docker-compose.yml example
version: '3.8'
services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    volumes:
      - ./config/config.json:/config.json:ro
      - ./reports:/reports
    command: ["--pihole", "/config.json"]
    restart: unless-stopped
```

## Development Setup

### Prerequisites for Development

```bash
# Additional development tools
make dev-setup

# Install testing dependencies
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Build Targets

```bash
make help                 # Show all available targets
make build               # Build production binary
make build-test          # Build test binary
make fast-build          # Optimized incremental build
make docker-build        # Build Docker image
make test               # Run tests
make fmt                # Format code
make vet                # Run go vet
make clean              # Clean build artifacts
```

## Verification

### Test Basic Functionality

```bash
# Test with mock data (no Pi-hole required)
./pihole-analyzer-test
# Expected: Colorized output with mock client statistics

# Test configuration validation
./pihole-analyzer --show-config
# Expected: Display current configuration

# Test Pi-hole connectivity
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet
# Expected: Connect to Pi-hole and analyze data
```

### Performance Testing

```bash
# Run benchmarks
make benchmark

# Analyze binary size
make analyze-size

# Test build performance
make cache-info
make fast-build
```

## Troubleshooting Installation

### Go Version Issues

```bash
# Check Go version
go version
# Should show: go version go1.23.12 or later

# Update Go if needed
# Download from: https://golang.org/dl/
```

### Build Failures

```bash
# Clean and rebuild
make clean
make build

# Check for missing dependencies
go mod tidy
go mod download

# Verbose build
go build -v -o pihole-analyzer ./cmd/pihole-analyzer
```

### API Connection Issues

```bash
# Test Pi-hole API manually
curl "http://your-pihole-ip/admin/api.php?summary"

# Check network connectivity
ping your-pihole-ip
telnet your-pihole-ip 80

# Verify Pi-hole API is enabled
# Access: http://your-pihole-ip/admin/settings.php?tab=api
```

### Permission Issues

```bash
# Fix binary permissions
chmod +x pihole-analyzer

# Fix configuration permissions
chmod 600 ~/.pihole-analyzer/config.json

# Check network access
# Ensure firewall allows connections to Pi-hole
```

### Container Issues

```bash
# Check container logs
docker logs pihole-analyzer

# Test container connectivity
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help

# Rebuild container
make docker-clean
make docker-build
```

## Next Steps

After successful installation:

1. **[Configuration Guide](configuration.md)** - Detailed configuration options
2. **[Usage Guide](usage.md)** - Command-line usage and examples  
3. **[API Setup](api.md)** - Pi-hole API integration details
4. **[Development Guide](development.md)** - Contributing and development

## Support

- 📋 **Issues**: [GitHub Issues](https://github.com/GrammaTonic/pihole-network-analyzer/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/GrammaTonic/pihole-network-analyzer/discussions)
- 📖 **Documentation**: [Full Documentation](../README.md)
