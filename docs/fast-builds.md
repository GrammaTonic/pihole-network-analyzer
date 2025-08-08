# Fast Builds with Caching

This document describes the comprehensive build optimization and caching strategies implemented in the Pi-hole Network Analyzer project.

## Overview

The project implements multi-layer caching to significantly reduce build times across different environments:

- **CI/CD Pipeline**: GitHub Actions with advanced caching strategies
- **Local Development**: Enhanced Makefile with build timing and cache management
- **Containerized Builds**: Docker multi-stage builds with layer caching
- **Development Environment**: Docker Compose with persistent Go caches

## Performance Improvements

Expected build time improvements:

- **Cold builds**: 20-30% faster due to optimized flags and parallel compilation
- **Warm builds**: 60-80% faster due to comprehensive caching
- **CI builds**: 50-70% faster due to cache restoration between runs
- **Docker builds**: 40-60% faster due to multi-stage builds and layer caching

## Caching Strategies

### 1. GitHub Actions CI/CD

Located in `.github/workflows/ci.yml`:

#### Multi-Layer Caching
```yaml
# Go modules cache (restored across runs)
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-mod-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-mod-

# Build cache (speeds up compilation)
- uses: actions/cache@v4
  with:
    path: ~/.cache/go-build
    key: ${{ runner.os }}-go-build-${{ github.sha }}
    restore-keys: |
      ${{ runner.os }}-go-build-
```

#### Binary Artifact Sharing
```yaml
# Share binaries between jobs to avoid rebuilding
- uses: actions/upload-artifact@v4
  with:
    name: pihole-analyzer-${{ matrix.os }}
    path: pihole-analyzer*
```

#### Parallel Builds
- Cross-platform builds run in parallel
- Independent test and build jobs
- Conditional execution based on changes

### 2. Local Development (Makefile)

Enhanced targets for optimal local development:

#### Fast Build Target
```make
fast-build: ## Fast incremental build with timing
	@echo "⚡ Fast build starting..."
	@start_time=$$(date +%s); \
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PATH); \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Fast build completed in $${duration}s"
```

#### Cache Management
```make
cache-warm: ## Warm up build caches for faster builds
	@./scripts/cache-warm.sh

cache-info: ## Show build cache information
	@echo "📊 Build Cache Information:"
	@echo "GOCACHE: $$(go env GOCACHE)"
	@echo "GOMODCACHE: $$(go env GOMODCACHE)"
```

#### Development Setup
```make
dev-setup: cache-warm ## One-time development environment setup
	@echo "🔧 Setting up development environment..."
	@go mod download
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Development setup complete"
```

### 3. Docker Containerization

#### Multi-Stage Dockerfile
```dockerfile
# Dependencies stage (cached separately)
FROM golang:1.24.5-alpine AS deps
COPY go.mod go.sum ./
RUN go mod download

# Build stage (uses cached dependencies)
FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pihole-analyzer cmd/pihole-analyzer/main.go

# Runtime stage (minimal final image)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/pihole-analyzer .
```

#### Docker Compose Development
```yaml
# Development environment with persistent caches
services:
  pihole-analyzer-dev:
    build: .
    volumes:
      - .:/app
      - go-cache:/go/pkg/mod
      - build-cache:/root/.cache/go-build
```

## Usage Guide

### First-Time Setup

1. **Warm up caches** (recommended for new environments):
   ```bash
   make cache-warm
   ```

2. **Development setup** (installs tools and prepares environment):
   ```bash
   make dev-setup
   ```

### Daily Development Workflow

1. **Fast incremental builds**:
   ```bash
   make fast-build
   ```

2. **Watch mode** (auto-rebuild on changes):
   ```bash
   make watch
   ```

3. **Cached testing**:
   ```bash
   make test-cached
   ```

### Docker Development

1. **Build development environment**:
   ```bash
   make docker-dev
   ```

2. **Run tests in container**:
   ```bash
   make docker-test
   ```

3. **Production Docker build**:
   ```bash
   make docker-build
   ```

### CI/CD Integration

The enhanced GitHub Actions workflow automatically:

- Restores caches from previous runs
- Builds in parallel across platforms
- Shares artifacts between jobs
- Reports build timing and cache hit rates

## Cache Management

### View Cache Information
```bash
make cache-info
```

Example output:
```
📊 Build Cache Information:
GOCACHE: /Users/user/Library/Caches/go-build
GOMODCACHE: /Users/user/go/pkg/mod

Cache sizes:
  Build cache: 45M
  Module cache: 128M
```

### Clean Caches
```bash
make cache-clean
```

### Warm Caches
```bash
make cache-warm
```

## Performance Monitoring

### Build Timing
All build targets include timing information:
```
⚡ Fast build starting...
✅ Fast build completed in 3s
```

### Benchmark Testing
```bash
make benchmark
```

### Binary Size Analysis
```bash
make analyze-size
```

## Optimization Features

### 1. Incremental Builds
- Only rebuilds changed packages
- Preserves build cache between runs
- Optimized dependency order

### 2. Parallel Compilation
- Multi-core builds with `-p` flag
- Parallel cross-platform builds in CI
- Concurrent test execution

### 3. Build Flags Optimization
```go
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
BUILD_FLAGS := -trimpath -ldflags="$(LDFLAGS)"
```

### 4. Smart Cache Invalidation
- Go modules cached by `go.sum` hash
- Build cache includes commit SHA
- Graceful fallback to partial cache hits

## Troubleshooting

### Cache Issues
If builds are unexpectedly slow:

1. **Check cache status**:
   ```bash
   make cache-info
   ```

2. **Clean and rebuild caches**:
   ```bash
   make cache-clean
   make cache-warm
   ```

3. **Verify Docker layer caching**:
   ```bash
   docker system df
   ```

### CI/CD Issues
If CI builds are slow:

1. Check GitHub Actions cache hit rates in build logs
2. Verify cache restore keys are working
3. Check for cache size limits (GitHub Actions: 10GB)

### Development Setup Issues
If initial setup is slow:

1. Run `make dev-setup` for complete environment preparation
2. Use `make cache-warm` to pre-populate caches
3. Check network connectivity for module downloads

## Best Practices

### Local Development
- Use `make fast-build` for quick iterations
- Use `make watch` for continuous development
- Warm caches after major dependency changes
- Use `make dev-setup` for new development environments

### CI/CD
- Leverage parallel builds for multi-platform support
- Monitor cache hit rates to optimize cache keys
- Use artifact sharing to avoid redundant builds
- Include build timing in logs for performance tracking

### Docker
- Use multi-stage builds for optimal layer caching
- Mount Go caches as volumes in development
- Use BuildKit for advanced caching features
- Keep dependencies and application code in separate stages

## Implementation Files

- **CI Pipeline**: `.github/workflows/ci.yml`
- **Local Builds**: `Makefile`
- **Docker**: `Dockerfile`, `docker-compose.yml`
- **Cache Warming**: `scripts/cache-warm.sh`
- **Documentation**: `docs/fast-builds.md`

## Future Enhancements

### Planned Optimizations
1. **Remote Build Cache**: Implement shared cache for teams
2. **Build Monitoring**: Add metrics and dashboards
3. **Incremental Testing**: Only run tests for changed packages
4. **Cross-Compilation Cache**: Optimize multi-platform builds

### Performance Targets
- Sub-5-second incremental builds
- Sub-30-second cold builds
- 90%+ cache hit rate in CI
- 50%+ reduction in CI build times

This comprehensive caching strategy ensures fast, efficient builds across all development workflows while maintaining reliability and consistency.
