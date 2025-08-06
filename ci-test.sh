#!/bin/bash

# Simple CI test script that mimics GitHub Actions behavior
set -e

echo "🔄 Running CI simulation test..."

# Test 1: Basic build
echo "Test 1: Building application..."
go build -o test-binary . >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Test 2: Quick test execution (with timeout simulation)
echo "Test 2: Running tests with timeout protection..."
(
    timeout_duration=60  # 60 seconds timeout
    (
        sleep $timeout_duration
        echo "⚠️ Test timed out after ${timeout_duration} seconds"
        pkill -f test-binary
    ) &
    timeout_pid=$!
    
    ./test-binary --test >/dev/null 2>&1
    test_result=$?
    
    # Kill the timeout process
    kill $timeout_pid 2>/dev/null || true
    
    if [ $test_result -eq 0 ]; then
        echo "✅ Tests completed successfully"
    else
        echo "❌ Tests failed"
        exit 1
    fi
)

# Test 3: Configuration test
echo "Test 3: Testing configuration system..."
./test-binary --create-config >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Configuration system working"
else
    echo "❌ Configuration system failed"
    exit 1
fi

# Cleanup
rm -f test-binary

echo "🎉 All CI simulation tests passed!"
echo "The application should work fine in GitHub Actions."
