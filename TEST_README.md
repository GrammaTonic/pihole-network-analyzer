# DNS Analyzer Test Suite

This test suite provides comprehensive offline testing capabilities for the DNS analyzer, allowing you to develop and iterate without requiring live Pi-hole servers or real network dependencies.

## Quick Start

### Run All Tests
```bash
./dns-analyzer --test
```

### Test Mode (Interactive Development)
```bash
# Test CSV analysis with mock data
./dns-analyzer --test-mode test_data/mock_dns_data.csv

# Test Pi-hole analysis with mock data
./dns-analyzer --test-mode --pihole test_data/mock_pihole_config.json

# Test with different flags
./dns-analyzer --test-mode --online-only --no-exclude test_data/mock_dns_data.csv
```

## Test Components

### 1. Mock Data (`test_data.go`)
- **Mock DNS Records**: 20 realistic DNS query records with various clients, domains, and query types
- **Mock Pi-hole Records**: Simulated Pi-hole database entries with hardware addresses and status codes
- **Mock ARP Table**: 5 online devices with MAC addresses for testing network status
- **Mock Hostnames**: Hostname resolution data for testing DNS lookups

### 2. Test Runner (`test_runner.go`)
Comprehensive test scenarios covering:

- **CSV Analysis Tests**:
  - Default behavior (with exclusions)
  - No exclusions mode
  - Online-only filtering

- **Pi-hole Analysis Tests**:
  - Default behavior (with exclusions)
  - No exclusions mode
  - Online-only filtering

- **Network Functionality Tests**:
  - ARP table functionality
  - Hostname resolution
  - Exclusion logic validation

### 3. Mock Network Data

#### Mock Clients:
- `192.168.2.110` - mac.home (Online)
- `192.168.2.210` - s21-van-marloes.home (Online)
- `192.168.2.6` - pi.hole (Online, excluded by default)
- `192.168.2.202` - maximus.home (Online)
- `192.168.2.119` - alexa.home (Online)
- `192.168.2.123` - grammatonic.home (Offline)
- `192.168.2.156` - syam-s-a33.home (Offline)
- `172.20.0.8` - Docker container (Offline, excluded by default)
- `172.19.0.2` - Docker container (Offline, excluded by default)

#### Mock Domains:
- Normal traffic: google.com, youtube.com, github.com, netflix.com, amazon.com
- Blocked content: tracking.doubleclick.net, malware.badsite.com, telemetry.microsoft.com
- System traffic: api.spotify.com, pi.hole, docker.internal

## Test Scenarios

### 1. Default Exclusions Test
Tests that Docker networks (172.16.0.0/12) and Pi-hole host are properly excluded:
```bash
./dns-analyzer --test-mode test_data/mock_dns_data.csv
```
Expected: 5 clients (excluding Docker containers and Pi-hole)

### 2. No Exclusions Test
Tests analysis with all clients included:
```bash
./dns-analyzer --test-mode --no-exclude test_data/mock_dns_data.csv
```
Expected: 9 clients (including Docker containers and Pi-hole)

### 3. Online Only Test
Tests filtering to show only currently online devices:
```bash
./dns-analyzer --test-mode --online-only test_data/mock_dns_data.csv
```
Expected: 4 online clients (5 online devices minus Pi-hole which is excluded)

### 4. Pi-hole Database Test
Tests direct Pi-hole database analysis:
```bash
./dns-analyzer --test-mode --pihole test_data/mock_pihole_config.json
```
Expected: Realistic Pi-hole data with hardware addresses and status codes

## Development Workflow

### 1. Feature Development
Use test mode to develop new features without live network dependencies:
```bash
# Edit code
./dns-analyzer --test-mode [flags] test_data/mock_dns_data.csv

# Quick iteration without network delays
# All ARP, DNS, and SSH operations are mocked
```

### 2. Regression Testing
Run the full test suite before commits:
```bash
./dns-analyzer --test
```

### 3. Flag Testing
Test new command-line flags with controlled data:
```bash
./dns-analyzer --test-mode --your-new-flag test_data/mock_dns_data.csv
```

## Mock Data Structure

### CSV Data Format
```csv
ID,DateTime,Domain,Type,Status,Client,Forward,AdditionalInfo,ReplyType,ReplyTime,DNSSEC,ListID,EDE
1,2024-08-06 10:00:01,google.com,1,2,192.168.2.110,,,,0.003245,0,0,0
```

### Pi-hole Database Schema
- `queries` table: DNS query logs with timestamps and status codes
- `network` table: Device information with hardware addresses
- `network_addresses` table: IP to device mapping

### ARP Table Simulation
Mock ARP entries simulate `arp -a` output for testing network status detection.

### Hostname Resolution Simulation
Mock DNS lookups return realistic hostnames for known IP addresses.

## Extending Tests

### Adding New Test Scenarios
Edit `test_runner.go` and add to the `testScenarios` slice:
```go
{"New Test Name", testNewFunctionality},
```

### Adding More Mock Data
Edit `test_data.go` and extend the mock data structures:
```go
// Add more DNS records
mock.DNSRecords = append(mock.DNSRecords, DNSRecord{...})

// Add more ARP entries
mock.ARPEntries["new.ip"] = &ARPEntry{...}
```

### Custom Mock Scenarios
Create specialized mock data for specific test cases:
```go
func CreateCustomMockData() *MockData {
    // Your custom scenario
}
```

## Benefits

1. **Fast Iteration**: No network delays or SSH connections
2. **Predictable Results**: Controlled data ensures consistent test outcomes
3. **Offline Development**: Work without Pi-hole server access
4. **Comprehensive Coverage**: Tests all major functionality paths
5. **Easy Debugging**: Known data makes issue identification easier

## Integration with CI/CD

The test suite is designed to run in automated environments:
```bash
# In your CI pipeline
go build -o dns-analyzer *.go
./dns-analyzer --test
```

All tests use temporary files and clean up automatically, making them suitable for automated testing environments.
