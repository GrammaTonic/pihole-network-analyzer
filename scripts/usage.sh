#!/bin/bash

cat << 'EOF'
=============================================================================
                     DNS Usage Analyzer - Usage Guide
=============================================================================

OVERVIEW:
This tool analyzes DNS usage patterns from CSV files or directly from Pi-hole
servers via API connection. It provides comprehensive statistics per client
including query counts, domain patterns, and performance metrics.

USAGE MODES:

1. CSV FILE ANALYSIS:
   ./dns-analyzer test.csv
   ./dns-analyzer --online-only test.csv
   ./dns-analyzer --no-exclude test.csv

2. PI-HOLE LIVE DATA ANALYSIS:
   
   a) First-time setup:
      ./dns-analyzer --pihole-setup
      (This creates pihole-config.json with your Pi-hole API credentials)
      
   b) Analyze Pi-hole data:
      ./dns-analyzer --pihole pihole-config.json
      ./dns-analyzer --online-only --pihole pihole-config.json

3. TEST SUITE AND DEVELOPMENT:
   ./dns-analyzer --test                    # Run complete test suite
   ./dns-analyzer --test-mode [flags] file  # Use mock data for development
   ./test.sh all                           # Run all test scenarios
   ./test.sh csv                           # Test CSV analysis
   ./test.sh pihole                        # Test Pi-hole analysis

COMMAND LINE FLAGS:
   --online-only     Show only clients currently online (with MAC addresses in ARP)
   --no-exclude      Disable default exclusions (Docker networks, Pi-hole host)
   --test-mode       Enable test mode (uses mock data instead of real network)
   --test            Run automated test suite
   --pihole-setup    Setup Pi-hole configuration interactively

PI-HOLE SETUP REQUIREMENTS:
- Pi-hole API access enabled
- Pi-hole API password or TOTP setup
- Either password authentication or 2FA
- Network access to Pi-hole server

CONFIGURATION FILE (pihole-config.json):
{
  "host": "192.168.2.6",
  "port": "80",
  "apiEnabled": true,
  "apiPassword": "your-api-password",
  "useHTTPS": false
}

Note: Configuration supports 2FA TOTP authentication (recommended for security).
API password can be generated in Pi-hole Settings > API / Web interface.

EXCLUSION CONFIGURATION:
By default, the analyzer excludes:
- Docker networks (172.16.0.0/12)
- Loopback addresses (127.0.0.0/8) 
- Pi-hole host itself
- Use --no-exclude flag to disable all exclusions

FEATURES:
✓ CSV file analysis (supports large files 90MB+)
✓ Live Pi-hole database analysis via API
✓ Client ranking by query volume
✓ Domain usage patterns
✓ Query type distribution (A, AAAA, CNAME, etc.)
✓ Status code analysis (blocked, cached, forwarded)
✓ Hardware address mapping (Pi-hole mode)
✓ ARP table checking for online/offline status
✓ Hostname resolution for friendly device names
✓ Performance metrics (reply times)
✓ Detailed report generation
✓ Configurable exclusions and filtering
✓ Comprehensive test suite for offline development

OUTPUT:
- Console display with top clients and detailed analysis
- Timestamped report files (dns_usage_report_YYYYMMDD_HHMMSS.txt)
- Client statistics with hardware addresses and online status
- ARP table status (online/offline indicators)
- Hostname resolution results

EXAMPLES:

# Basic CSV analysis
./dns-analyzer test.csv

# CSV analysis without exclusions (show Docker containers, etc.)
./dns-analyzer --no-exclude test.csv

# Show only currently online devices
./dns-analyzer --online-only test.csv

# Combined flags
./dns-analyzer --online-only --no-exclude test.csv

# Setup Pi-hole connection
./dns-analyzer --pihole-setup

# Analyze live Pi-hole data
./dns-analyzer --pihole pihole-config.json

# Pi-hole analysis with filtering
./dns-analyzer --online-only --pihole pihole-config.json

# Development and testing
./dns-analyzer --test                              # Run test suite
./dns-analyzer --test-mode test_data/mock_dns_data.csv  # Use mock data
./test.sh csv                                      # Quick CSV test
./test.sh pihole-online                           # Quick Pi-hole test
./test.sh all                                     # All test scenarios

DEVELOPMENT WORKFLOW:
1. Use test mode for feature development:
   ./dns-analyzer --test-mode --your-new-flag test_data/mock_dns_data.csv

2. Run tests before committing:
   ./dns-analyzer --test

3. Quick scenario testing:
   ./test.sh csv-online    # Test online-only filtering
   ./test.sh pihole        # Test Pi-hole functionality

TROUBLESHOOTING:
- Ensure Go 1.23+ is installed
- For Pi-hole: verify API access and permissions
- Large datasets may take several minutes to process
- Check network connectivity for Pi-hole connections
- Use --test-mode for development without network dependencies
- API authentication is recommended for security
- ARP table functionality requires appropriate network permissions
- Hostname resolution may be slow on some networks

TEST SUITE:
The analyzer includes a comprehensive test suite for offline development:
- 9 automated test scenarios covering all functionality
- Mock data simulating real network environments
- Performance benchmarking capabilities
- No network dependencies required for testing

For detailed test documentation, see docs/TEST_README.md

=============================================================================
EOF
