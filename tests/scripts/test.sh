#!/bin/bash

# DNS Analyzer Test Helper Script
# Provides quick access to common test scenarios

set -e

echo "DNS Analyzer Test Helper"
echo "========================"

# Build the application
echo "Building DNS analyzer..."
go build -o dns-analyzer *.go
echo "✓ Build complete"

# Function to run a test scenario
run_test() {
    echo
    echo "Running: $1"
    echo "Command: $2"
    echo "---"
    eval $2
}

case "${1:-help}" in
    "all")
        echo
        echo "Running all test scenarios..."
        run_test "Complete Test Suite" "./dns-analyzer --test"
        ;;
    
    "csv")
        run_test "CSV Analysis (Mock Data)" "./dns-analyzer --test-mode test_data/mock_dns_data.csv"
        ;;
    
    "csv-no-exclude")
        run_test "CSV Analysis - No Exclusions" "./dns-analyzer --test-mode --no-exclude test_data/mock_dns_data.csv"
        ;;
    
    "csv-online")
        run_test "CSV Analysis - Online Only" "./dns-analyzer --test-mode --online-only test_data/mock_dns_data.csv"
        ;;
    
    "pihole")
        run_test "Pi-hole Analysis (Mock Data)" "./dns-analyzer --test-mode --pihole test_data/mock_pihole_config.json"
        ;;
    
    "pihole-no-exclude")
        run_test "Pi-hole Analysis - No Exclusions" "./dns-analyzer --test-mode --no-exclude --pihole test_data/mock_pihole_config.json"
        ;;
    
    "pihole-online")
        run_test "Pi-hole Analysis - Online Only" "./dns-analyzer --test-mode --online-only --pihole test_data/mock_pihole_config.json"
        ;;
    
    "setup")
        echo
        echo "Setting up test environment..."
        ./dns-analyzer --test-mode test_data/mock_dns_data.csv > /dev/null 2>&1 || true
        echo "✓ Test data created in test_data/ directory"
        echo "✓ Ready for testing!"
        ;;
    
    "clean")
        echo
        echo "Cleaning up test files..."
        rm -rf test_data/
        rm -f dns_usage_report_*.txt
        echo "✓ Test files cleaned up"
        ;;
    
    "benchmark")
        echo
        echo "Running performance benchmarks..."
        echo "Testing CSV analysis performance..."
        time ./dns-analyzer --test-mode test_data/mock_dns_data.csv > /dev/null
        echo
        echo "Testing Pi-hole analysis performance..."
        time ./dns-analyzer --test-mode --pihole test_data/mock_pihole_config.json > /dev/null
        ;;
    
    "help"|*)
        echo
        echo "Available test scenarios:"
        echo "  all              Run complete test suite"
        echo "  csv              Test CSV analysis with mock data"
        echo "  csv-no-exclude   Test CSV analysis without exclusions"
        echo "  csv-online       Test CSV analysis with online-only filter"
        echo "  pihole           Test Pi-hole analysis with mock data"
        echo "  pihole-no-exclude Test Pi-hole analysis without exclusions"
        echo "  pihole-online    Test Pi-hole analysis with online-only filter"
        echo "  setup            Create test environment"
        echo "  clean            Remove test files"
        echo "  benchmark        Run performance tests"
        echo "  help             Show this help message"
        echo
        echo "Examples:"
        echo "  ./test.sh all"
        echo "  ./test.sh csv"
        echo "  ./test.sh pihole-online"
        echo "  ./test.sh benchmark"
        ;;
esac

echo
echo "Done!"
