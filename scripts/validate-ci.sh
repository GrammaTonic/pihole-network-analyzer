#!/bin/bash

# Local CI/CD validation script
# This script runs the same tests that GitHub Actions will run

set -e  # Exit on any error

echo "🚀 Starting local CI/CD validation..."
echo "===================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    rm -f dns-analyzer-test
    rm -f pihole-network-analyzer
    rm -rf test_data/
    rm -rf reports/
    rm -f test-config.json
    rm -f test_output.log
}

# Set up cleanup trap
trap cleanup EXIT

echo "📦 Step 1: Module verification"
go mod download
go mod verify
print_status "Dependencies verified"

echo ""
echo "🔨 Step 2: Building application"
if go build -v -o dns-analyzer-test . >/dev/null 2>&1; then
    print_status "Build successful"
else
    print_error "Build failed"
    go build -v -o dns-analyzer-test .
    exit 1
fi

# Also build with the same name as CI uses
if go build -o pihole-network-analyzer . >/dev/null 2>&1; then
    print_status "CI-compatible build successful"
else
    print_error "CI-compatible build failed"
    exit 1
fi

echo ""
echo "🧪 Step 3: Running test suite"
if ./dns-analyzer-test --test > test_output.log 2>&1; then
    # Check if all tests passed by looking for the success message
    if grep -q "All tests passed!" test_output.log; then
        print_status "Test suite completed successfully"
        # Show test results summary
        grep "Test Results:" test_output.log || echo "Test completed"
    else
        print_error "Test suite did not complete successfully"
        echo "Test output:"
        cat test_output.log
        exit 1
    fi
else
    print_error "Test suite failed"
    echo "Last few lines of test output:"
    tail -10 test_output.log
    exit 1
fi

echo ""
echo "🚀 Step 3.5: Testing CI-exact commands"
echo "Testing the exact same commands that GitHub Actions runs..."
if ./pihole-network-analyzer --test > ci_test_output.log 2>&1; then
    if grep -q "All tests passed!" ci_test_output.log; then
        print_status "CI-exact test suite passed"
    else
        print_error "CI-exact test suite failed"
        echo "CI test output:"
        cat ci_test_output.log
        exit 1
    fi
else
    print_error "CI-exact test command failed"
    echo "CI test output:"
    cat ci_test_output.log
    exit 1
fi

echo ""
echo "⚙️  Step 4: Testing configuration system"
if ./dns-analyzer-test --create-config >/dev/null 2>&1; then
    print_status "Config creation successful"
else
    print_error "Config creation failed"
    exit 1
fi

if ./dns-analyzer-test --show-config >/dev/null 2>&1; then
    print_status "Config display successful"
else
    print_error "Config display failed"
    exit 1
fi

echo ""
echo "📝 Step 5: Code formatting check"
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    print_error "Code formatting issues found:"
    gofmt -s -l .
    exit 1
else
    print_status "Code formatting is correct"
fi

echo ""
echo "🔍 Step 6: Static analysis"
if go vet ./... 2>/dev/null; then
    print_status "Static analysis passed"
else
    print_error "Static analysis found issues"
    go vet ./...
    exit 1
fi

echo ""
echo "🎯 Step 7: Performance test"
echo "Measuring test suite execution time..."
time_output=$(/usr/bin/time -p ./dns-analyzer-test --test 2>&1 >/dev/null | tail -3)
print_status "Performance test completed"
echo "$time_output"

echo ""
echo "🏗️  Step 8: Multi-platform build test"
echo "Testing cross-compilation..."

if GOOS=linux GOARCH=amd64 go build -o dns-analyzer-linux . 2>/dev/null; then
    print_status "Linux build successful"
    rm -f dns-analyzer-linux
else
    print_warning "Linux build failed (this may be normal on some systems)"
fi

if GOOS=windows GOARCH=amd64 go build -o dns-analyzer-windows.exe . 2>/dev/null; then
    print_status "Windows build successful"  
    rm -f dns-analyzer-windows.exe
else
    print_warning "Windows build failed (this may be normal on some systems)"
fi

echo ""
echo "🎉 All local CI/CD validation tests passed!"
echo "Your code is ready for GitHub Actions!"

# Cleanup is handled by trap
rm -f test_output.log
rm -f ci_test_output.log
