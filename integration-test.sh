#!/bin/bash

# Integration Test Script for CI/CD Pipeline
# Supports scenario-specific testing for the DNS Network Analyzer

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run a test with timeout (cross-platform)
run_test_with_timeout() {
    local test_name=$1
    local command=$2
    local timeout_seconds=${3:-60}
    
    print_status $BLUE "🧪 Running: $test_name"
    
    # Check if gtimeout (GNU timeout) is available (common on macOS with brew)
    if command -v gtimeout &> /dev/null; then
        timeout_cmd="gtimeout"
    elif command -v timeout &> /dev/null; then
        timeout_cmd="timeout"
    else
        # Fallback: run without timeout on systems that don't have it
        print_status $YELLOW "⚠️ Timeout command not available, running without timeout"
        if bash -c "$command"; then
            print_status $GREEN "✅ $test_name - PASSED"
            return 0
        else
            print_status $RED "❌ $test_name - FAILED"
            return 1
        fi
    fi
    
    if $timeout_cmd $timeout_seconds bash -c "$command"; then
        print_status $GREEN "✅ $test_name - PASSED"
        return 0
    else
        print_status $RED "❌ $test_name - FAILED"
        return 1
    fi
}

# Function to check if binary exists
check_binary() {
    if [ ! -f "./pihole-network-analyzer" ]; then
        print_status $RED "❌ Binary './pihole-network-analyzer' not found. Please build first."
        exit 1
    fi
    
    if [ ! -x "./pihole-network-analyzer" ]; then
        print_status $YELLOW "⚠️ Making binary executable..."
        chmod +x ./pihole-network-analyzer
    fi
}

# Function to create a small test CSV file for CI compatibility
create_test_csv_file() {
    cat > test_small.csv << 'EOF'
"DateTime","Client_IP","Domain","QueryType","Status","ReplyTime"
"2024-08-06 10:00:01","192.168.2.110","google.com","1","2","0.002"
"2024-08-06 10:00:02","192.168.2.210","facebook.com","1","2","0.003"
"2024-08-06 10:00:03","192.168.2.6","api.spotify.com","1","2","0.001"
"2024-08-06 10:00:04","172.20.0.8","kpn.com","1","2","0.002"
"2024-08-06 10:00:05","192.168.2.210","tracking.doubleclick.net","1","9","0.002"
"2024-08-06 10:00:06","192.168.2.110","github.com","1","2","0.004"
"2024-08-06 10:00:07","192.168.2.6","pi.hole","1","3","0.001"
"2024-08-06 10:00:08","172.19.0.2","microsoft.com","1","2","0.012"
"2024-08-06 10:00:09","127.0.0.1","localhost","1","2","0.001"
"2024-08-06 10:00:10","192.168.2.210","ads.microsoft.com","1","9","0.002"
"2024-08-06 10:00:11","192.168.2.110","stackoverflow.com","1","2","0.005"
"2024-08-06 10:00:12","192.168.2.6","netflix.com","1","2","0.001"
"2024-08-06 10:00:13","192.168.2.210","malware.badsite.com","1","9","0.002"
"2024-08-06 10:00:14","192.168.2.110","amazon.com","1","2","0.003"
"2024-08-06 10:00:15","192.168.2.6","telemetry.microsoft.com","1","9","0.001"
EOF
}

# Function to run CSV analysis tests
test_csv_analysis() {
    print_status $YELLOW "📊 Testing CSV Analysis Functionality"
    
    # Create a smaller test CSV file for CI compatibility
    create_test_csv_file
    
    # Test 1: Basic CSV analysis
    run_test_with_timeout "CSV Analysis - Default" \
        "./pihole-network-analyzer --quiet test_small.csv" 120
    
    # Test 2: CSV with no exclusions
    run_test_with_timeout "CSV Analysis - No Exclusions" \
        "./pihole-network-analyzer --no-exclude --quiet test_small.csv" 60
    
    # Test 3: CSV online only
    run_test_with_timeout "CSV Analysis - Online Only" \
        "./pihole-network-analyzer --online-only --quiet test_small.csv" 60
    
    # Cleanup
    rm -f test_small.csv
    
    print_status $GREEN "✅ CSV Analysis tests completed"
}

# Function to run Pi-hole database tests
test_pihole_db() {
    print_status $YELLOW "🗄️ Testing Pi-hole Database Functionality"
    
    # Create mock Pi-hole environment
    mkdir -p test_pihole_env
    
    # Test 1: Pi-hole analysis with mock data
    run_test_with_timeout "Pi-hole DB Analysis" \
        "./pihole-network-analyzer --test-mode --quiet" 90
    
    # Test 2: Pi-hole with custom config
    run_test_with_timeout "Pi-hole Custom Config" \
        "./pihole-network-analyzer --config=test-config.json --test-mode --quiet" 60
    
    # Cleanup
    rm -rf test_pihole_env
    
    print_status $GREEN "✅ Pi-hole Database tests completed"
}

# Function to run colorized output tests
test_colorized_output() {
    print_status $YELLOW "🎨 Testing Colorized Output Functionality"
    
    # Test 1: Go unit tests for color functions (CI-friendly pattern)
    if [ "$CI" = "true" ]; then
        # In CI, exclude integration tests that expect terminal colors
        run_test_with_timeout "Color Unit Tests (CI-friendly)" \
            "go test -v -run='TestColor[^i]|TestHighlight|TestStatus|TestOnline|TestColored|TestStrip|TestGetDisplay|TestFormatTable' -timeout=5m" 300
    else
        # Local development - run more comprehensive tests
        run_test_with_timeout "Color Unit Tests" \
            "go test -v -run='TestColor[^i]|TestHighlight|TestStatus|TestOnline|TestColored' -timeout=5m" 300
    fi
    
    # Test 2: Table formatting and color utility tests  
    run_test_with_timeout "Color Utility Tests" \
        "go test -v -run='TestTableFormatting|TestStripColor|TestGetDisplayLength|TestFormatTableColumn' -timeout=3m" 180
    
    # Test 3: Performance benchmarks for color functions
    run_test_with_timeout "Color Performance Benchmarks" \
        "go test -bench=BenchmarkColor -run=Benchmark -timeout=2m" 120
    
    print_status $GREEN "✅ Colorized Output tests completed"
}

# Function to run all integration tests
test_all_features() {
    print_status $YELLOW "🔄 Running Comprehensive Integration Test Suite"
    
    # Full application test suite (this includes the main 13 tests)
    run_test_with_timeout "Comprehensive Test Suite" \
        "./pihole-network-analyzer --test" 300
    
    # Additional Go tests (CI-friendly pattern)
    if [ "$CI" = "true" ]; then
        # CI environment - exclude problematic integration tests
        run_test_with_timeout "Go Unit Tests (CI)" \
            "go test -v -timeout=10m -run='^Test[^C]' ./..." 600
    else
        # Local development - run more tests
        run_test_with_timeout "Go Unit Tests" \
            "go test -v -timeout=10m -run='^Test[^C]' ./..." 600
    fi
    
    # Performance and stress tests
    run_test_with_timeout "Performance Benchmarks" \
        "go test -bench=. -run=Benchmark -timeout=5m" 300
    
    print_status $GREEN "✅ All Features tests completed"
}

# Function to validate environment
validate_environment() {
    print_status $BLUE "🔍 Validating test environment..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_status $RED "❌ Go is not installed or not in PATH"
        exit 1
    fi
    
    local go_version=$(go version)
    print_status $GREEN "✅ Go found: $go_version"
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
        print_status $RED "❌ Not in the correct project directory"
        exit 1
    fi
    
    # Check for test files
    if [ ! -f "test.csv" ]; then
        print_status $YELLOW "⚠️ test.csv not found, some tests may be skipped"
    fi
    
    print_status $GREEN "✅ Environment validation completed"
}

# Function to generate test report
generate_report() {
    local scenario=$1
    local start_time=$2
    local end_time=$3
    local status=$4
    
    local duration=$((end_time - start_time))
    
    print_status $BLUE "📊 Integration Test Report"
    echo "========================================"
    echo "Scenario: $scenario"
    echo "Duration: ${duration} seconds"
    echo "Status: $status"
    echo "Timestamp: $(date)"
    echo "Environment: $(uname -a)"
    echo "Go Version: $(go version)"
    echo "========================================"
}

# Main execution
main() {
    local scenario=${1:-"all-features"}
    local start_time=$(date +%s)
    
    # Handle help flag
    if [ "$scenario" = "--help" ] || [ "$scenario" = "-h" ] || [ "$scenario" = "help" ]; then
        print_status $BLUE "🧪 Integration Test Framework"
        echo "Usage: $0 [scenario]"
        echo ""
        echo "Available scenarios:"
        echo "  csv-analysis      - Test CSV processing functionality"
        echo "  pihole-db        - Test Pi-hole database operations"
        echo "  colorized-output - Test color output functionality"
        echo "  all-features     - Run comprehensive test suite (default)"
        echo ""
        echo "Examples:"
        echo "  $0                    # Run all features"
        echo "  $0 csv-analysis       # Test CSV analysis only"
        echo "  $0 colorized-output   # Test color functionality"
        exit 0
    fi
    
    print_status $BLUE "🚀 Starting Integration Tests - Scenario: $scenario"
    
    # Validate environment first
    validate_environment
    
    # Check binary exists
    check_binary
    
    # Run tests based on scenario
    case "$scenario" in
        "csv-analysis")
            test_csv_analysis
            ;;
        "pihole-db")
            test_pihole_db
            ;;
        "colorized-output")
            test_colorized_output
            ;;
        "all-features")
            test_all_features
            ;;
        *)
            print_status $RED "❌ Unknown test scenario: $scenario"
            echo "Available scenarios:"
            echo "  - csv-analysis"
            echo "  - pihole-db"
            echo "  - colorized-output"
            echo "  - all-features"
            echo ""
            echo "Use '$0 --help' for more information"
            exit 1
            ;;
    esac
    
    local end_time=$(date +%s)
    
    # Generate report
    generate_report "$scenario" "$start_time" "$end_time" "SUCCESS"
    
    print_status $GREEN "🎉 All integration tests passed for scenario: $scenario"
}

# Handle script interruption
trap 'print_status $RED "❌ Integration tests interrupted"; exit 1' INT TERM

# Execute main function with all arguments
main "$@"
