# Integration Testing Quick Reference

## 🚀 Quick Start

### Run Integration Tests Locally
```bash
# Test specific functionality
./scripts/integration-test.sh csv-analysis      # CSV processing tests
./scripts/integration-test.sh pihole-db        # Pi-hole database tests  
./scripts/integration-test.sh colorized-output # Color output tests
./scripts/integration-test.sh all-features     # Comprehensive test suite

# Run CI validation tests
./scripts/ci-test.sh
```

### Check Test Status
```bash
# View last test results
echo $?  # 0 = success, 1 = failure

# Run with verbose output (remove >/dev/null)
./integration-test.sh colorized-output
```

## 📊 Test Scenarios

| Scenario | Purpose | Duration | Tests |
|----------|---------|----------|-------|
| `csv-analysis` | CSV file processing | ~3-5 min | 3 tests |
| `pihole-db` | Pi-hole database ops | ~2-4 min | 2 tests |
| `colorized-output` | Color functionality | ~8 sec | 3 test suites |
| `all-features` | End-to-end testing | ~15 min | 13+ tests |

## 🔧 CI/CD Pipeline Jobs

### GitHub Actions Workflow
```yaml
jobs:
  test:                          # Basic unit tests + integration tests
  validate-integration-tests:    # Integration framework validation  
  integration-test:              # Matrix testing (4 scenarios × 2 Go versions)
  security:                      # Security scans (depends on integration tests)
  build-check:                   # Feature branch build verification
  build:                         # Production builds (main branch only)
```

### Matrix Testing
- **Go Versions**: 1.22.x, 1.23.x
- **Scenarios**: csv-analysis, pihole-db, colorized-output, all-features
- **Platforms**: Ubuntu Latest
- **Total Jobs**: 8 integration test jobs + validation

## 🎯 What Gets Tested

### Application Features
- ✅ CSV analysis with exclusions
- ✅ Pi-hole database processing  
- ✅ ARP table functionality
- ✅ Hostname resolution
- ✅ Network exclusion logic
- ✅ Colorized output (4 scenarios)
- ✅ Online/offline status detection
- ✅ Domain highlighting
- ✅ Table formatting

### Technical Validation  
- ✅ Cross-platform builds
- ✅ Race condition detection
- ✅ Performance benchmarks
- ✅ Memory usage
- ✅ Configuration system
- ✅ Error handling

## 📈 Performance Benchmarks

### Color Functions
```
BenchmarkColorFunctions/Red              77M ops    15.04 ns/op
BenchmarkColorFunctions/BoldGreen        77M ops    15.09 ns/op  
BenchmarkColorFunctions/HighlightDomain  38M ops    29.47 ns/op
BenchmarkColorFunctions/ColoredQueryCount 16M ops   74.45 ns/op
```

### Expected Results
- **Main Test Suite**: 13/13 tests passing
- **Unit Tests**: 25+ tests passing
- **Build Time**: <5 minutes total
- **Memory Usage**: <100MB during tests

## 🐛 Troubleshooting

### Common Issues
1. **Color tests failing**: Expected in CI - colors disabled in non-terminal
2. **Timeout issues**: macOS doesn't have `timeout` - script handles gracefully
3. **Build failures**: Check Go version and dependencies

### Debug Commands
```bash
# Check Go environment
go version
go env

# Validate test environment  
./integration-test.sh --help

# Run specific test manually
go test -v -run="TestColorFunctions" 

# Check build with race detection
go build -race -o test-binary .
```

### CI/CD Debugging
```bash
# Simulate CI environment locally
CI=true ./integration-test.sh all-features

# Check pipeline configuration
cat .github/workflows/ci.yml

# Validate integration script
chmod +x ./integration-test.sh
./integration-test.sh colorized-output
```

## 📝 Adding New Tests

### Add to Integration Script
```bash
# Edit integration-test.sh
# Add new function:
test_my_feature() {
    print_status $YELLOW "🧪 Testing My Feature"
    run_test_with_timeout "My Test" "my-test-command" 120
    print_status $GREEN "✅ My Feature tests completed"
}

# Add to main() case statement:
"my-feature")
    test_my_feature
    ;;
```

### Add to CI Pipeline
```yaml
# Edit .github/workflows/ci.yml
# Add to matrix strategy:
test-scenario: ['csv-analysis', 'pihole-db', 'colorized-output', 'all-features', 'my-feature']
```

## 🎉 Success Criteria

### Local Development
- All scenarios should complete without errors
- Performance benchmarks within expected ranges
- Cross-platform builds successful

### CI/CD Pipeline  
- All matrix jobs passing
- Security scans clean
- Build artifacts generated (main branch)
- No performance regressions

The integration testing framework provides comprehensive coverage while being resilient to environment-specific issues.
