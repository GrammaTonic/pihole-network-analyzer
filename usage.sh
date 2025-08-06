#!/bin/bash

cat << 'EOF'
=============================================================================
                     DNS Usage Analyzer - Usage Guide
=============================================================================

OVERVIEW:
This tool analyzes DNS usage patterns from CSV files or directly from Pi-hole
servers via SSH connection. It provides comprehensive statistics per client
including query counts, domain patterns, and performance metrics.

USAGE MODES:

1. CSV FILE ANALYSIS:
   ./dns-analyzer test.csv
   go run main.go test.csv
   make run

2. PI-HOLE LIVE DATA ANALYSIS:
   
   a) First-time setup:
      go run main.go --pihole-setup
      (This creates pihole-config.json with your SSH credentials)
      
   b) Analyze Pi-hole data:
      go run main.go --pihole pihole-config.json

3. MAKEFILE COMMANDS:
   make help              # Show all available commands
   make build             # Build the application
   make run               # Run with test.csv
   make install-deps      # Install Go dependencies
   make clean             # Clean build artifacts

PI-HOLE SETUP REQUIREMENTS:
- SSH access to Pi-hole server
- Pi-hole database at /etc/pihole/pihole-FTL.db (or custom path)
- Either SSH key authentication or password
- Sudo access to read Pi-hole database

CONFIGURATION FILE (pihole-config.json):
{
  "host": "192.168.1.100",
  "port": "22",
  "username": "pi",
  "password": "your_password",
  "keyfile": "/path/to/private/key",
  "dbpath": "/etc/pihole/pihole-FTL.db"
}

FEATURES:
✓ CSV file analysis (supports large files 90MB+)
✓ Live Pi-hole database analysis via SSH
✓ Client ranking by query volume
✓ Domain usage patterns
✓ Query type distribution (A, AAAA, CNAME, etc.)
✓ Status code analysis (blocked, cached, forwarded)
✓ Hardware address mapping (Pi-hole mode)
✓ Performance metrics (reply times)
✓ Detailed report generation

OUTPUT:
- Console display with top clients and detailed analysis
- Timestamped report files (dns_usage_report_YYYYMMDD_HHMMSS.txt)
- Client statistics with hardware addresses (Pi-hole mode)

EXAMPLES:

# Analyze CSV file
./analyze.sh test.csv

# Setup Pi-hole connection
go run main.go --pihole-setup

# Analyze live Pi-hole data
go run main.go --pihole pihole-config.json

# Build and run with specific CSV
make run-with-file CSV_FILE=my-dns-logs.csv

TROUBLESHOOTING:
- Ensure Go 1.21+ is installed
- For Pi-hole: verify SSH access and database permissions
- Large CSV files may take several minutes to process
- Check network connectivity for Pi-hole connections

=============================================================================
EOF
