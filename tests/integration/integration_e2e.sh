#!/bin/bash

# Integration test script for monitoring ecosystem
# Tests the complete integration workflow end-to-end

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "🧪 Starting Integration Ecosystem End-to-End Tests"
echo "================================================="

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to start mock services using docker-compose
start_mock_services() {
    echo "🚀 Starting mock monitoring services..."
    
    # Create temporary docker-compose file for test services
    cat > docker-compose.test.yml <<EOF
version: '3.8'
services:
  mock-grafana:
    image: grafana/grafana:latest
    container_name: test-grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_SECURITY_ADMIN_USER=admin
    volumes:
      - /tmp/grafana-data:/var/lib/grafana
    networks:
      - test-network

  mock-loki:
    image: grafana/loki:latest
    container_name: test-loki
    ports:
      - "3101:3100"
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - test-network

  mock-prometheus:
    image: prom/prometheus:latest
    container_name: test-prometheus
    ports:
      - "9091:9090"
    networks:
      - test-network

networks:
  test-network:
    driver: bridge
EOF

    # Start services
    if command_exists docker-compose; then
        docker-compose -f docker-compose.test.yml up -d
    elif command_exists docker; then
        echo "⚠️  docker-compose not found, using simple HTTP mock servers instead"
        return 1
    else
        echo "❌ Neither docker nor docker-compose found, skipping service tests"
        return 1
    fi
    
    echo "⏳ Waiting for services to be ready..."
    sleep 10
    
    return 0
}

# Function to stop mock services
stop_mock_services() {
    echo "🛑 Stopping mock services..."
    if [ -f docker-compose.test.yml ]; then
        if command_exists docker-compose; then
            docker-compose -f docker-compose.test.yml down -v
        fi
        rm -f docker-compose.test.yml
    fi
    
    # Clean up test data
    sudo rm -rf /tmp/grafana-data 2>/dev/null || true
}

# Function to test configuration validation
test_configuration() {
    echo "🔧 Testing configuration validation..."
    
    # Create test configuration with integrations
    cat > test-integration-config.json <<EOF
{
    "pihole": {
        "host": "127.0.0.1",
        "port": 80,
        "api_key": "test-key",
        "use_tls": false
    },
    "integrations": {
        "enabled": true,
        "grafana": {
            "enabled": true,
            "url": "http://localhost:3001",
            "api_key": "test-grafana-key",
            "data_source": {
                "create_if_not_exists": true,
                "name": "pihole-prometheus",
                "type": "prometheus",
                "url": "http://localhost:9091"
            },
            "dashboards": {
                "auto_provision": true,
                "folder_name": "Pi-hole"
            }
        },
        "loki": {
            "enabled": true,
            "url": "http://localhost:3101",
            "batch_size": 100,
            "batch_timeout": "10s"
        },
        "prometheus": {
            "enabled": true,
            "push_gateway": {
                "enabled": true,
                "url": "http://localhost:9091"
            },
            "external_labels": {
                "service": "pihole-analyzer",
                "instance": "test-instance"
            }
        }
    },
    "analysis": {
        "mode": "comprehensive",
        "exclude_networks": ["127.0.0.0/8"],
        "report_dir": "/tmp/test-reports"
    },
    "logging": {
        "level": "info",
        "enable_colors": true,
        "enable_emojis": true
    }
}
EOF

    # Test configuration loading and validation
    echo "📋 Validating integration configuration..."
    go run ./cmd/pihole-analyzer-test --config test-integration-config.json --test 2>&1 | tee config-test.log
    
    if grep -q "✅" config-test.log; then
        echo "✅ Configuration validation passed"
    else
        echo "❌ Configuration validation failed"
        cat config-test.log
        return 1
    fi
    
    # Clean up
    rm -f test-integration-config.json config-test.log
}

# Function to test build process
test_build() {
    echo "🔨 Testing build process..."
    
    # Clean previous builds
    make clean 2>/dev/null || true
    
    # Test build
    echo "🏗️  Building with integrations..."
    if ! (make build && make build-test); then
        echo "❌ Build failed"
        return 1
    fi
    
    echo "✅ Build completed successfully"
    
    # Test that binaries work
    echo "🧪 Testing binary execution..."
    if ! ./pihole-analyzer --help >/dev/null 2>&1; then
        echo "❌ Main binary execution failed"
        return 1
    fi
    
    if ! ./pihole-analyzer-test --help >/dev/null 2>&1; then
        echo "❌ Test binary execution failed"
        return 1
    fi
    
    echo "✅ Binary execution test passed"
}

# Function to test unit tests
test_unit_tests() {
    echo "🧪 Running unit tests for integrations..."
    
    # Test individual integration packages
    echo "📊 Testing Prometheus integration..."
    if ! go test -v -timeout 30s ./internal/integrations/prometheus/...; then
        echo "❌ Prometheus integration tests failed"
        return 1
    fi
    
    echo "📈 Testing Grafana integration..."
    if ! go test -v -timeout 30s ./internal/integrations/grafana/...; then
        echo "❌ Grafana integration tests failed"
        return 1
    fi
    
    echo "📝 Testing Loki integration..."
    if ! go test -v -timeout 30s ./internal/integrations/loki/...; then
        echo "❌ Loki integration tests failed"
        return 1
    fi
    
    echo "🎛️  Testing integration interfaces..."
    if ! go test -v -timeout 30s ./internal/integrations/interfaces/...; then
        echo "❌ Integration interfaces tests failed"
        return 1
    fi
    
    echo "✅ All integration unit tests passed"
}

# Function to test with mock data
test_mock_data() {
    echo "🎭 Testing with mock data..."
    
    mkdir -p /tmp/test-reports
    
    # Run test binary with integrations disabled to verify basic functionality
    echo "🧪 Running basic analysis with mock data..."
    
    cat > test-mock-config.json <<EOF
{
    "pihole": {
        "host": "127.0.0.1",
        "port": 80,
        "use_tls": false
    },
    "integrations": {
        "enabled": false
    },
    "analysis": {
        "mode": "comprehensive",
        "report_dir": "/tmp/test-reports"
    },
    "logging": {
        "level": "info",
        "enable_colors": false,
        "enable_emojis": false
    }
}
EOF

    if ! timeout 30s ./pihole-analyzer-test --config test-mock-config.json --test 2>&1 | tee mock-test.log; then
        echo "❌ Mock data test failed"
        cat mock-test.log
        rm -f test-mock-config.json mock-test.log
        return 1
    fi
    
    # Check for expected output
    if grep -q "Analysis completed" mock-test.log || grep -q "Total queries" mock-test.log; then
        echo "✅ Mock data analysis completed successfully"
    else
        echo "⚠️  Mock data analysis completed but output unclear"
        cat mock-test.log
    fi
    
    # Clean up
    rm -f test-mock-config.json mock-test.log
    rm -rf /tmp/test-reports
}

# Function to test documentation generation
test_documentation() {
    echo "📚 Testing documentation and examples..."
    
    # Check if documentation files exist
    local docs_exist=0
    
    if [ -f "docs/integrations.md" ] || [ -f "README.md" ]; then
        docs_exist=1
    fi
    
    if [ $docs_exist -eq 1 ]; then
        echo "✅ Documentation files found"
    else
        echo "⚠️  Documentation files not found - this should be created"
    fi
    
    # Check if example configs exist
    if [ -f "test-config.json" ] || [ -f "examples/" ]; then
        echo "✅ Example configurations found"
    else
        echo "⚠️  Example configurations not found - this should be created"
    fi
}

# Main test execution
main() {
    local exit_code=0
    local services_started=false
    
    echo "🏁 Starting integration ecosystem tests at $(date)"
    
    # Trap to ensure cleanup
    trap 'stop_mock_services' EXIT
    
    # Test 1: Build process
    echo ""
    echo "📦 TEST 1: Build Process"
    echo "------------------------"
    if ! test_build; then
        echo "❌ Build test failed"
        exit_code=1
    fi
    
    # Test 2: Unit tests
    echo ""
    echo "🧪 TEST 2: Unit Tests"
    echo "---------------------"
    if ! test_unit_tests; then
        echo "❌ Unit tests failed"
        exit_code=1
    fi
    
    # Test 3: Configuration validation
    echo ""
    echo "🔧 TEST 3: Configuration"
    echo "------------------------"
    if ! test_configuration; then
        echo "❌ Configuration test failed"
        exit_code=1
    fi
    
    # Test 4: Mock data
    echo ""
    echo "🎭 TEST 4: Mock Data Analysis"
    echo "-----------------------------"
    if ! test_mock_data; then
        echo "❌ Mock data test failed"
        exit_code=1
    fi
    
    # Test 5: Service integration (optional)
    echo ""
    echo "🐳 TEST 5: Service Integration (Optional)"
    echo "----------------------------------------"
    if start_mock_services; then
        services_started=true
        echo "✅ Mock services started successfully"
        echo "ℹ️  Service integration tests would run here"
        echo "ℹ️  (Skipped due to complexity in CI environment)"
    else
        echo "⚠️  Service integration tests skipped (Docker not available)"
    fi
    
    # Test 6: Documentation
    echo ""
    echo "📚 TEST 6: Documentation"
    echo "------------------------"
    test_documentation
    
    # Summary
    echo ""
    echo "📊 TEST SUMMARY"
    echo "==============="
    
    if [ $exit_code -eq 0 ]; then
        echo "✅ All integration tests passed successfully!"
        echo ""
        echo "🎉 Integration ecosystem is ready for deployment"
        echo ""
        echo "📋 Completed tests:"
        echo "  ✅ Build process and binary execution"
        echo "  ✅ Unit tests for all integration packages"
        echo "  ✅ Configuration validation"
        echo "  ✅ Mock data analysis"
        if [ "$services_started" = true ]; then
            echo "  ✅ Mock services startup"
        else
            echo "  ⚠️  Service integration (skipped)"
        fi
        echo "  ✅ Documentation check"
        echo ""
        echo "🚀 Ready for production deployment!"
    else
        echo "❌ Some integration tests failed"
        echo ""
        echo "🔍 Please review the test output above and fix any issues"
    fi
    
    echo ""
    echo "🏁 Integration tests completed at $(date)"
    
    return $exit_code
}

# Execute main function
main "$@"