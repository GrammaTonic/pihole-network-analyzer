# Pi-hole Network Analyzer

[![CI Pipeline](https://github.com/GrammaTonic/pihole-network-analyzer/actions## 📖 Documentation

- **[Installation Guide](docs/installation.md)** - Detailed setup instructions
- **[Configuration Guide](docs/configuration.md)** - Configuration options and examples  
- **[Usage Guide](docs/usage.md)** - Command-line options and workflows
- **[Container Publishing Guide](docs/CONTAINER_PUBLISHING.md)** - Docker containers and GHCR publishing
- **[API Setup](docs/api.md)** - Pi-hole API connection configuration
- **[Development Guide](docs/development.md)** - Building, testing, and contributing
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutionsws/ci.yml/badge.svg)](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

# Pi-hole Network Analyzer

🎉 **Production Ready v1.0** 🎉

A Go-based DNS analysis tool that connects to Pi-hole via API to generate colorized terminal reports with network insights, featuring web UI, metrics collection, and daemon mode.

## 🚀 Features

- **🔗 Pi-hole API Integration**: Direct connection to Pi-hole via REST API
- **🎨 Beautiful Terminal Output**: Colorized reports with emoji support
- **📊 Comprehensive Analysis**: Client statistics, domain queries, network status
- **🌐 Network Detection**: ARP table integration for online/offline status
- **⚙️ Flexible Configuration**: JSON-based configuration with exclusion rules
- **🧪 Testing Support**: Built-in mock data for development and CI
- **🔒 Security**: Session-based API authentication with 2FA support
- **📈 Performance**: Optimized for large Pi-hole datasets

## 📋 Quick Start

### Prerequisites

- **Docker** (recommended) - For the easiest setup and deployment
- **Terminal with color support** (recommended)
- **Pi-hole with API access** - Target Pi-hole server

### 🐳 Docker Installation (Recommended)

The fastest way to get started is using Docker:

```bash
# Pull the latest image
docker pull ghcr.io/grammatonic/pihole-network-analyzer:latest

# Run with environment variables (no config file needed)
docker run --rm \
  -e PIHOLE_HOST=192.168.1.100 \
  -e PIHOLE_API_PASSWORD=your-api-token \
  ghcr.io/grammatonic/pihole-network-analyzer:latest

# Run with configuration file
docker run --rm \
  -v ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro \
  ghcr.io/grammatonic/pihole-network-analyzer:latest \
  --config /home/appuser/.pihole-analyzer/config.json

# Start web UI with environment variables
docker run -d -p 8080:8080 \
  -e PIHOLE_HOST=192.168.1.100 \
  -e PIHOLE_API_PASSWORD=your-api-token \
  -e WEB_ENABLED=true \
  -e WEB_DAEMON_MODE=true \
  --name pihole-analyzer \
  ghcr.io/grammatonic/pihole-network-analyzer:latest
```

### Docker Compose (Production Ready)

```bash
# Create docker-compose.yml with your Pi-hole details
curl -O https://raw.githubusercontent.com/GrammaTonic/pihole-network-analyzer/main/docker-compose.yml

# Configure via environment variables
export PIHOLE_HOST=192.168.1.100
export PIHOLE_API_PASSWORD=your-api-token

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f
```

### 🔧 Alternative: Build from Source

For development or custom builds:

```bash
# Requires Go 1.24+ and Node.js (dev tools only)
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer

# Build the binary
make build

# Or use Go directly
go build -o pihole-analyzer ./cmd/pihole-analyzer
```

### Configuration Options

**Environment Variables** (Docker-friendly):
```bash
# Required
PIHOLE_HOST=192.168.1.100
PIHOLE_API_PASSWORD=your-api-token

# Optional
LOG_LEVEL=info
WEB_ENABLED=true
WEB_PORT=8080
METRICS_ENABLED=true
ANALYSIS_ONLINE_ONLY=false
```

**Configuration File** (traditional):
```bash
# Create default config
docker run --rm -v $(pwd):/output \
  ghcr.io/grammatonic/pihole-network-analyzer:latest \
  --create-config

# Edit the generated config.json file
```

### Basic Usage

```bash
# Docker with environment variables
docker run --rm \
  -e PIHOLE_HOST=192.168.1.100 \
  -e PIHOLE_API_PASSWORD=your-token \
  ghcr.io/grammatonic/pihole-network-analyzer:latest

# Binary with config file (if built from source)
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Test with mock data (no Pi-hole required)
docker run --rm ghcr.io/grammatonic/pihole-network-analyzer:latest-development
```

## � Container Support

Pi-hole Network Analyzer is available as multi-architecture Docker containers published to GitHub Container Registry (GHCR).

### Quick Container Usage

```bash
# Pull the latest version
docker pull ghcr.io/grammatonic/pihole-network-analyzer:latest

# Run with configuration file
docker run --rm -v $(pwd)/config.json:/config.json \
  ghcr.io/grammatonic/pihole-network-analyzer:latest \
  --config /config.json

# Run web UI mode
docker run -d -p 8080:8080 -v $(pwd)/config.json:/config.json \
  ghcr.io/grammatonic/pihole-network-analyzer:latest \
  --config /config.json --web --daemon
```

### Container Publishing Automation

For developers and advanced users:

```bash
# Build and push production containers
make docker-push-ghcr

# Build development containers
make container-push-dev

# List container information
make container-info

# See all container commands
make container-list
```

**Supported Architectures**: linux/amd64, linux/arm64, linux/arm/v7  
**Registry**: [ghcr.io/grammatonic/pihole-network-analyzer](https://github.com/GrammaTonic/pihole-network-analyzer/pkgs/container/pihole-network-analyzer)

For detailed container usage, see the **[Container Publishing Guide](docs/CONTAINER_PUBLISHING.md)**

## �📖 Documentation

- **[Installation Guide](docs/installation.md)** - Detailed setup instructions
- **[Configuration Guide](docs/configuration.md)** - Configuration options and examples  
- **[Usage Guide](docs/usage.md)** - Command-line options and workflows
- **[API Setup](docs/api.md)** - Pi-hole API connection configuration
- **[Development Guide](docs/development.md)** - Building, testing, and contributing
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

## 🎯 Example Output

```
================================================================================
DNS Query Analysis Summary  
================================================================================
Total Clients: 12 | Online: 8 | Total Queries: 1,247

Client Statistics
-----------------------------------------------------------------------------------------------------------
IP Address       Hostname           MAC Address        Status     Queries    Unique       Avg RT   Top Domain Query %
192.168.1.100   johns-macbook      aa:bb:cc:dd:ee:ff  ✅ Online  234        89          0.002s   google.com      24%
192.168.1.101   smart-tv           11:22:33:44:55:66  ✅ Online  156        23          0.001s   netflix.com     67%
192.168.1.102   iphone-sarah       77:88:99:aa:bb:cc  ❌ Offline 89         45          N/A      facebook.com    31%
```

## ⚙️ Configuration

Create a configuration file to customize behavior:

```bash
./pihole-analyzer --create-config
```

Example configuration:

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 80,
    "apiEnabled": true,
    "apiPassword": "your-api-password",
    "useHTTPS": false
  },
  "exclusions": {
    "networks": ["172.16.0.0/12", "127.0.0.0/8"],
    "hostnames": ["docker", "localhost"],
    "ips": ["192.168.1.1"]
  },
  "output": {
    "colors": true,
    "emoji": true,
    "saveReports": true
  }
}
```

## 🔧 Command Line Options

```bash
# Core Operations
--pihole <config>     # Analyze Pi-hole with configuration file
--pihole-setup        # Interactive Pi-hole configuration setup
--test               # Run with mock data for testing

# Configuration
--config <path>      # Custom configuration file path
--create-config      # Create default configuration file
--show-config        # Display current configuration

# Output Control  
--quiet              # Suppress non-essential output
--no-color           # Disable colored output
--no-emoji           # Disable emoji in output

# Filtering
--online-only        # Show only online clients
--no-exclude         # Disable default exclusions
```

## 🏗️ Architecture

The project follows the Standard Go Project Layout:

```
├── cmd/pihole-analyzer/     # Main application entry point
├── internal/                # Private application packages
│   ├── analyzer/           # Pi-hole data analysis engine
│   ├── cli/                # Command-line interface
│   ├── colors/             # Terminal colorization
│   ├── config/             # Configuration management
│   ├── interfaces/         # Data source abstraction
│   ├── logger/             # Structured logging
│   ├── network/            # Network analysis & ARP
│   ├── pihole/             # Pi-hole API client
│   ├── reporting/          # Output formatting
│   └── types/              # Core data structures
├── docs/                   # Documentation
├── scripts/                # Build and automation
└── test_data/              # Mock data for testing
```

## 🧪 Development

```bash
# Run tests
make test

# Run integration tests
make ci-test

# Build for development
make build

# Clean build artifacts
make clean
```

## 🤝 Contributing

We follow a structured development workflow with semantic versioning and conventional commits. Please see our [Quick Start Workflow Guide](docs/QUICK_START_WORKFLOW.md) for details.

### Development Process

1. **Setup**: Install dependencies with `make release-setup`
2. **Feature Branch**: Create from main: `git checkout -b feat/amazing-feature`
3. **Conventional Commits**: Use `make commit` for interactive commit creation
4. **Testing**: Run `make ci-test` before pushing
5. **Pull Request**: Create PR to main branch

### Commit Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
feat(component): add new feature
fix(api): resolve bug in authentication
docs: update installation guide
```

### Documentation

- 📚 [Branching Strategy](docs/BRANCHING_STRATEGY.md) - Detailed workflow guide
- 🚀 [Quick Start Workflow](docs/QUICK_START_WORKFLOW.md) - Getting started
- 🔧 [Development](docs/development.md) - Technical development guide

## 📦 Releases

Releases are automated using semantic versioning:

- **Patch** (x.y.Z): Bug fixes (`fix:` commits)
- **Minor** (x.Y.z): New features (`feat:` commits)  
- **Major** (X.y.z): Breaking changes (`BREAKING CHANGE:` footer)

View releases: [GitHub Releases](https://github.com/GrammaTonic/pihole-network-analyzer/releases)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Pi-hole](https://pi-hole.net/) - Network-wide ad blocking
- [Pi-hole API](https://docs.pi-hole.net/api/) - Official Pi-hole API
- [SQLite](https://www.sqlite.org/) - Database support

## 📞 Support

- 📋 **Issues**: [GitHub Issues](https://github.com/GrammaTonic/pihole-network-analyzer/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/GrammaTonic/pihole-network-analyzer/discussions)
- 📖 **Documentation**: [Full Documentation](docs/)

---

**Made with ❤️ for the Pi-hole community**
