# GitHub Workflows Documentation

This directory contains GitHub Actions workflows for the DNS Analyzer project.

## Available Workflows

### 1. `ci.yml` - Full CI/CD Pipeline
**Comprehensive workflow with advanced features:**
- **Test Job**: Runs test suite, code formatting checks, and static analysis
- **Build Job**: Multi-platform builds (Linux, Windows, macOS) for both amd64 and arm64
- **Release Job**: Automatically creates releases with binaries for tagged commits
- **Security Job**: Runs security scans and vulnerability checks

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` branch

### 2. `simple-ci.yml` - Simplified Test and Build
**Streamlined workflow focused on essential checks:**
- Tests and builds the application
- Validates configuration system
- Checks code formatting and runs static analysis
- Multi-platform builds on main branch pushes
- Uploads build artifacts for 30 days

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

### 3. `test.yml` - Focused Test Suite Validation
**Dedicated testing workflow:**
- Comprehensive test suite execution (9 test scenarios)
- Configuration system testing
- CSV analysis functionality validation
- Pi-hole integration testing
- Exclusion logic verification
- Performance testing
- Test data integrity validation

**Triggers:**
- Push/PR with changes to Go files, go.mod, or go.sum
- Changes to the test workflow itself

## Workflow Features

### Test Coverage
All workflows include:
- ✅ **9 comprehensive test scenarios** covering all functionality
- ✅ **Configuration system testing** (create, load, display)
- ✅ **Mock data generation and validation**
- ✅ **CSV and Pi-hole analysis testing**
- ✅ **Exclusion logic verification**
- ✅ **Online/offline mode testing**

### Build Features
- **Multi-platform support**: Linux, Windows, macOS
- **Multi-architecture**: amd64 and arm64
- **Artifact retention**: 30 days for build artifacts
- **Go module caching**: Speeds up subsequent runs

### Code Quality
- **Formatting checks**: `gofmt` validation
- **Static analysis**: `go vet` for potential issues
- **Security scanning**: Gosec and govulncheck (in full CI)
- **Dependency verification**: `go mod verify`

## Local Testing

Before pushing, you can run the same tests locally:

```bash
# Run the test suite
./dns-analyzer --test

# Test configuration system
./dns-analyzer --create-config
./dns-analyzer --show-config

# Test build process
go build -v -o dns-analyzer .

# Check code formatting
gofmt -s -l .

# Run static analysis
go vet ./...
```

## Workflow Status

When you push code or create a pull request, you can view the workflow status:

1. Go to your repository on GitHub
2. Click on the "Actions" tab
3. View the status of running/completed workflows
4. Click on individual workflow runs for detailed logs

## Build Artifacts

Successful builds on the `main` branch will generate downloadable artifacts:
- `dns-analyzer-linux-amd64`
- `dns-analyzer-linux-arm64`
- `dns-analyzer-windows-amd64.exe`
- `dns-analyzer-windows-arm64.exe`
- `dns-analyzer-darwin-amd64`
- `dns-analyzer-darwin-arm64`

These can be downloaded from the "Actions" tab → "Build" workflow → "Artifacts" section.

## Recommended Workflow

For most development, `simple-ci.yml` provides the best balance of thorough testing and reasonable execution time. Use `ci.yml` for production releases where you need the full security scanning and release automation.

The `test.yml` workflow is perfect for validating that changes don't break the test suite without running the full build matrix.
