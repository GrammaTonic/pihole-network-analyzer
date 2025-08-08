# GitHub Container Registry Implementation Guide

This document provides a comprehensive implementation plan for publishing Pi-hole Network Analyzer containers to GitHub Container Registry (GHCR).

## Overview

GitHub Container Registry provides secure, integrated container hosting for the Pi-hole Network Analyzer project. This implementation will enable users to run the analyzer without building from source, supporting multiple architectures including Raspberry Pi deployments.

## Current Container Analysis

**Existing Container Status:**
- **Image ID**: `sha256:07ae9fa1c92fe6dff54c785db858bbfe0fcf92721928c05940097cce3c236824`
- **Size**: 44.3MB (optimal for distribution)
- **Architecture**: ARM64 (perfect for Raspberry Pi)
- **Base**: Alpine Linux (secure and minimal)
- **User**: Non-root `appuser` (security best practice)
- **Health Check**: Built-in validation
- **Default Command**: Test mode execution

## Implementation Strategy

### Phase 1: Multi-Architecture Container Builds

#### 1.1 Enhanced Dockerfile Structure

```dockerfile
# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binaries for target platform
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" \
    -o pihole-analyzer ./cmd/pihole-analyzer

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" \
    -o pihole-analyzer-test ./cmd/pihole-analyzer-test

# Production runtime stage
FROM alpine:latest AS production

# Security: Add non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Install runtime dependencies
RUN apk add --no-cache ca-certificates openssh-client

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/pihole-analyzer .
COPY --from=builder /app/pihole-analyzer-test .
COPY --from=builder /app/testing/fixtures ./testing/fixtures

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./pihole-analyzer --help || exit 1

# Add container labels
LABEL org.opencontainers.image.title="Pi-hole Network Analyzer"
LABEL org.opencontainers.image.description="Analyze Pi-hole DNS queries and network traffic"
LABEL org.opencontainers.image.vendor="GrammaTonic"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/GrammaTonic/pihole-network-analyzer"

# Default to production mode
ENTRYPOINT ["./pihole-analyzer"]

# Development variant
FROM production AS development
ENTRYPOINT ["./pihole-analyzer-test", "--test"]
```

#### 1.2 Multi-Stage Build Optimization

**Benefits:**
- **Layer Caching**: Dependencies cached separately from source code
- **Multi-Architecture**: Single Dockerfile supports AMD64, ARM64, ARMv7
- **Size Optimization**: Minimal runtime image (~44MB)
- **Security**: Non-root user, minimal attack surface
- **Flexibility**: Production and development variants

### Phase 2: GitHub Actions CI/CD Integration

#### 2.1 Container Build Workflow

Create `.github/workflows/container.yml`:

```yaml
name: Container Build and Publish

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]
  schedule:
    # Rebuild weekly for security updates
    - cron: '0 2 * * 1'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: grammatonic/pihole-analyzer

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      security-events: write
    
    strategy:
      matrix:
        variant: [production, development]
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        with:
          platforms: linux/amd64,linux/arm64,linux/arm/v7

      - name: Log in to Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          flavor: |
            suffix=-${{ matrix.variant }},onlatest=false
            latest=${{ matrix.variant == 'production' }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=raw,value=latest,enable={{is_default_branch}}
            type=sha,prefix={{branch}}-

      - name: Build and push container
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64,linux/arm/v7
          target: ${{ matrix.variant }}
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha,scope=${{ matrix.variant }}
          cache-to: type=gha,mode=max,scope=${{ matrix.variant }}
          provenance: true
          sbom: true

      - name: Run security scan
        if: github.event_name != 'pull_request'
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.meta.outputs.version }}-${{ matrix.variant }}
          format: sarif
          output: trivy-results.sarif

      - name: Upload security scan results
        if: github.event_name != 'pull_request'
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif

  test-containers:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.event_name != 'pull_request'
    
    strategy:
      matrix:
        variant: [production, development]
        platform: [amd64, arm64]
    
    steps:
      - name: Test container functionality
        run: |
          docker run --rm --platform linux/${{ matrix.platform }} \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest-${{ matrix.variant }} \
            --help

      - name: Test container health
        run: |
          container_id=$(docker run -d --platform linux/${{ matrix.platform }} \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest-${{ matrix.variant }})
          
          # Wait for health check
          sleep 10
          
          # Check health status
          health=$(docker inspect --format='{{.State.Health.Status}}' $container_id)
          docker rm -f $container_id
          
          if [ "$health" != "healthy" ]; then
            echo "Container health check failed: $health"
            exit 1
          fi
```

#### 2.2 Registry Configuration

**Access Control:**
```yaml
# Repository settings
permissions:
  packages: write  # Allow publishing containers
  contents: read   # Allow reading repository
  security-events: write  # Allow security scan uploads
```

**Branch Protection:**
- Require container builds to pass before merge
- Automatic security scanning on all builds
- Multi-architecture validation

### Phase 3: Container Variants and Tagging Strategy

#### 3.1 Container Variants

**Production Container (`ghcr.io/grammatonic/pihole-analyzer:latest`)**
```bash
# Features:
- Minimal size (~44MB)
- Production-ready configuration
- Health checks enabled
- Non-root execution
- SSH client included for Pi-hole connections

# Usage:
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help
```

**Development Container (`ghcr.io/grammatonic/pihole-analyzer:latest-development`)**
```bash
# Features:
- Test utilities included
- Mock data for development
- Debug capabilities
- Development tools

# Usage:
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest-development
```

#### 3.2 Tagging Strategy

**Automated Tags:**
- `latest` - Latest stable release (production variant)
- `main` - Latest development build from main branch
- `v1.2.3` - Specific version releases
- `v1.2` - Major.minor versions
- `pr-123` - Pull request builds
- `main-sha1234` - Commit-specific builds

**Architecture-Specific Tags:**
- `latest-amd64` - Intel/AMD 64-bit
- `latest-arm64` - ARM 64-bit (Raspberry Pi 4)
- `latest-armv7` - ARM 32-bit (Raspberry Pi 3)

### Phase 4: User Experience Enhancements

#### 4.1 Docker Compose Integration

**Create `docker-compose.prod.yml`:**
```yaml
version: '3.8'

services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    container_name: pihole-analyzer
    volumes:
      - ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro
      - ./reports:/app/reports
    environment:
      - PIHOLE_CONFIG=/home/appuser/.pihole-analyzer/config.json
    networks:
      - pihole-net
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./pihole-analyzer", "--help"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

networks:
  pihole-net:
    driver: bridge
```

**Create `docker-compose.dev.yml`:**
```yaml
version: '3.8'

services:
  pihole-analyzer-dev:
    image: ghcr.io/grammatonic/pihole-analyzer:latest-development
    container_name: pihole-analyzer-dev
    volumes:
      - .:/app/source:ro
      - ~/.pihole-analyzer:/home/appuser/.pihole-analyzer
      - ./reports:/app/reports
    environment:
      - DEVELOPMENT=true
    working_dir: /app
    command: ["./pihole-analyzer-test", "--test"]
    ports:
      - "8080:8080"  # For future web interface
```

#### 4.2 Container Usage Documentation

**Quick Start Guide:**
```bash
# Basic usage
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help

# With configuration
docker run --rm \
  -v ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro \
  ghcr.io/grammatonic/pihole-analyzer:latest \
  --show-config

# Production deployment
docker-compose -f docker-compose.prod.yml up -d

# Development environment
docker-compose -f docker-compose.dev.yml up -d
```

### Phase 5: Security and Compliance

#### 5.1 Security Scanning

**Automated Security Checks:**
- **Trivy**: Vulnerability scanning on every build
- **SBOM**: Software Bill of Materials generation
- **Provenance**: Build attestation and verification
- **CodeQL**: Static analysis for security issues

**Security Reporting:**
```yaml
# Weekly security reports
- name: Security Summary
  run: |
    echo "## Security Scan Results" >> $GITHUB_STEP_SUMMARY
    echo "| Severity | Count |" >> $GITHUB_STEP_SUMMARY
    echo "|----------|-------|" >> $GITHUB_STEP_SUMMARY
    trivy image --format table --severity HIGH,CRITICAL \
      ghcr.io/grammatonic/pihole-analyzer:latest
```

#### 5.2 Supply Chain Security

**Build Attestation:**
```yaml
# Enable build provenance
provenance: true
sbom: true

# Sign containers (future enhancement)
- name: Sign container
  uses: sigstore/cosign-installer@v3
  with:
    cosign-release: 'v2.0.0'
```

### Phase 6: Performance Optimization

#### 6.1 Layer Optimization

**Multi-Stage Build Benefits:**
- **Dependency Layer**: Cached independently (go.mod/go.sum changes)
- **Build Layer**: Compiled binaries with optimization flags
- **Runtime Layer**: Minimal Alpine with only necessary components

**Cache Strategy:**
- **GitHub Actions Cache**: Build cache persistence between runs
- **Registry Cache**: Layer sharing across architectures
- **Local Cache**: Development environment optimization

#### 6.2 Build Performance Metrics

**Monitoring:**
```yaml
- name: Build Performance Report
  run: |
    echo "## Build Metrics" >> $GITHUB_STEP_SUMMARY
    echo "- Build Time: ${build_duration}s" >> $GITHUB_STEP_SUMMARY
    echo "- Image Size: $(docker image inspect --format='{{.Size}}' ${image_name} | numfmt --to=iec)" >> $GITHUB_STEP_SUMMARY
    echo "- Cache Hit Rate: ${cache_hit_rate}%" >> $GITHUB_STEP_SUMMARY
```

## Implementation Timeline

### Week 1: Core Infrastructure
- [ ] Create enhanced Dockerfile with multi-architecture support
- [ ] Set up GitHub Actions container workflow
- [ ] Configure GitHub Container Registry permissions
- [ ] Test basic container functionality

### Week 2: CI/CD Integration
- [ ] Implement automated container builds
- [ ] Add security scanning with Trivy
- [ ] Set up multi-architecture testing
- [ ] Configure cache optimization

### Week 3: User Experience
- [ ] Create Docker Compose configurations
- [ ] Add container usage documentation
- [ ] Implement health checks and monitoring
- [ ] Test deployment scenarios

### Week 4: Security and Optimization
- [ ] Enable build attestation and SBOM
- [ ] Optimize container layer caching
- [ ] Add performance monitoring
- [ ] Security compliance validation

## Success Metrics

### Performance Targets
- **Build Time**: < 5 minutes for all architectures
- **Image Size**: < 50MB for production variant
- **Cache Hit Rate**: > 80% for incremental builds
- **Security Score**: Zero high/critical vulnerabilities

### User Experience Goals
- **Simple Usage**: Single `docker run` command for basic functionality
- **Quick Start**: < 30 seconds from pull to first analysis
- **Documentation**: Complete usage examples and troubleshooting
- **Cross-Platform**: Support AMD64, ARM64, ARMv7 architectures

## Risk Assessment and Mitigation

### Potential Risks
1. **Container Size Growth**: Monitor and optimize layer efficiency
2. **Security Vulnerabilities**: Automated scanning and rapid patching
3. **Build Failures**: Comprehensive testing and rollback procedures
4. **Performance Regression**: Continuous monitoring and benchmarking

### Mitigation Strategies
- **Automated Testing**: Multi-platform validation before release
- **Security Gates**: Block releases with high-severity vulnerabilities
- **Monitoring**: Real-time alerting for build failures
- **Documentation**: Clear troubleshooting guides and support channels

## Future Enhancements

### Planned Features
1. **Multi-Registry Support**: Docker Hub, Azure Container Registry
2. **Container Signing**: Cosign integration for supply chain security
3. **OCI Compliance**: Full Open Container Initiative specification support
4. **Helm Charts**: Kubernetes deployment configurations

### Advanced Capabilities
- **Distroless Images**: Further size reduction and security hardening
- **WebAssembly**: WASM runtime support for edge deployments
- **ARM32**: Additional Raspberry Pi architecture support
- **Windows Containers**: Windows Server container variants

## Conclusion

This implementation provides a comprehensive container registry strategy that enhances the Pi-hole Network Analyzer's accessibility, security, and deployment flexibility. The multi-architecture support ensures compatibility across development laptops, cloud environments, and Raspberry Pi deployments.

The automated CI/CD pipeline with security scanning and performance monitoring ensures reliable, secure container releases while maintaining development velocity. Users benefit from simplified deployment options and consistent runtime environments.

This foundation supports future enhancements including Kubernetes deployments, advanced security features, and expanded platform support, positioning the project for long-term growth and adoption.
