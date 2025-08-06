#!/bin/bash

# DNS Usage Analyzer Runner Script
# This script ensures Go is in PATH and runs the DNS analyzer

# Add common Go installation paths to PATH
export PATH="/usr/local/go/bin:/opt/homebrew/bin:$PATH"

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Navigate to script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Build if binary doesn't exist or source is newer
if [ ! -f dns-analyzer ] || [ main.go -nt dns-analyzer ]; then
    echo "Building DNS analyzer..."
    go build -o dns-analyzer main.go
    if [ $? -ne 0 ]; then
        echo "Error: Failed to build application"
        exit 1
    fi
fi

# Run the analyzer
if [ $# -eq 0 ]; then
    # No arguments provided, use default CSV file
    if [ -f test.csv ]; then
        ./dns-analyzer test.csv
    else
        echo "Usage: $0 [csv_file]"
        echo "No CSV file specified and test.csv not found"
        exit 1
    fi
else
    # Use provided CSV file
    ./dns-analyzer "$1"
fi
