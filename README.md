# Pi-hole Network Analyzer

[![CI Pipeline](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml/badge.svg)](https://github.com/GrammaTonic/pihole-network-analyzer/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A powerful, colorized DNS usage analysis tool that connects to Pi-hole servers via API to provide comprehensive network insights and beautiful terminal reports.

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

- **Go 1.24+** - For building and running the application
- **Pi-hole with API access** - Target Pi-hole server
- **Terminal with color support** (recommended)
- **Node.js** (optional) - Only needed for release automation and development workflow

### Installation

```bash
# Clone the repository
git clone https://github.com/GrammaTonic/pihole-network-analyzer.git
cd pihole-network-analyzer

# Build the binary
make build

# Or use Go directly
go build -o pihole-analyzer ./cmd/pihole-analyzer
```

### Basic Usage

```bash
# Setup Pi-hole configuration (interactive)
./pihole-analyzer --pihole-setup

# Analyze Pi-hole data
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json

# Run with test data (no Pi-hole required)
./pihole-analyzer-test

# Create default configuration file
./pihole-analyzer --create-config
```

## 📖 Documentation

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
