# Container Registry & Deployment Strategy

This guide covers container deployment, registry management, and production deployment strategies for the Pi-hole Network Analyzer.

## Table of Contents
- [Container Architecture](#container-architecture)
- [Registry Configuration](#registry-configuration)
- [Deployment Strategies](#deployment-strategies)
- [Security & Performance](#security--performance)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Container Architecture

### Multi-Architecture Support
The Pi-hole Network Analyzer supports multiple CPU architectures:
- **AMD64**: Standard x86_64 systems
- **ARM64**: Apple Silicon, ARM64 servers
- **ARMv7**: Raspberry Pi 3/4, ARM-based devices

### Container Variants

#### Production Container
- **Registry**: `ghcr.io/grammatonic/pihole-analyzer:latest`
- **Size**: ~44MB
- **Purpose**: Production deployments
- **Features**: Optimized runtime, security hardening

#### Development Container
- **Registry**: `ghcr.io/grammatonic/pihole-analyzer:latest-development`
- **Size**: ~45MB
- **Purpose**: Development and testing
- **Features**: Test utilities, debug tools, mock data

### Dockerfile Architecture

#### Phase 1: Multi-Stage Production Dockerfile

```dockerfile
# Multi-platform build arguments
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Build stage with Go toolchain
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

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
RUN apk add --no-cache ca-certificates

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
LABEL org.opencontainers.image.description="Analyze Pi-hole DNS queries via API"
LABEL org.opencontainers.image.vendor="GrammaTonic"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/GrammaTonic/pihole-network-analyzer"

# Default to production mode
ENTRYPOINT ["./pihole-analyzer"]

# Development variant
FROM production AS development
ENTRYPOINT ["./pihole-analyzer-test", "--test"]
```

#### Build Optimization Features

**Benefits:**
- **Layer Caching**: Dependencies cached separately from source code
- **Multi-Architecture**: Single Dockerfile supports AMD64, ARM64, ARMv7
- **Size Optimization**: Minimal runtime image (~44MB)
- **Security**: Non-root user, minimal attack surface
- **Flexibility**: Production and development variants

## Registry Configuration

### GitHub Container Registry (GHCR)

#### Phase 2: GitHub Actions CI/CD Integration

Create `.github/workflows/container.yml`:

```yaml
name: Container Build and Publish

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

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
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        platforms: linux/amd64,linux/arm64,linux/arm/v7
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
```

#### Registry Authentication

```bash
# Authenticate with GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin

# Pull latest image
docker pull ghcr.io/grammatonic/pihole-analyzer:latest

# Run with Pi-hole API configuration
docker run --rm -v $(pwd)/config.json:/app/config.json \
  ghcr.io/grammatonic/pihole-analyzer:latest \
  --config /app/config.json
```

### Alternative Registry Support

#### Docker Hub
```bash
# Build and tag for Docker Hub
docker build -t username/pihole-analyzer:latest .
docker push username/pihole-analyzer:latest
```

#### AWS ECR
```bash
# Authenticate with ECR
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-west-2.amazonaws.com

# Build and push
docker build -t 123456789012.dkr.ecr.us-west-2.amazonaws.com/pihole-analyzer:latest .
docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/pihole-analyzer:latest
```

## Deployment Strategies

### Phase 3: Production Deployment

#### Docker Compose Production

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    container_name: pihole-analyzer-prod
    restart: unless-stopped
    
    volumes:
      - ./config.json:/app/config.json:ro
      - ./reports:/app/reports
      - ./logs:/app/logs
    
    environment:
      - PIHOLE_CONFIG_FILE=/app/config.json
      - LOG_LEVEL=info
      - NO_COLOR=false
      - NO_EMOJI=false
    
    healthcheck:
      test: ["CMD", "./pihole-analyzer", "--help"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    
    security_opt:
      - no-new-privileges:true
    
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
    
    networks:
      - pihole-analyzer-net

  # Optional: Log aggregation
  loki:
    image: grafana/loki:latest
    container_name: pihole-analyzer-loki
    restart: unless-stopped
    
    volumes:
      - ./loki-config.yml:/etc/loki/local-config.yml:ro
      - loki-data:/loki
    
    networks:
      - pihole-analyzer-net

volumes:
  loki-data:

networks:
  pihole-analyzer-net:
    driver: bridge
```

#### Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pihole-analyzer
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pihole-analyzer
  template:
    metadata:
      labels:
        app: pihole-analyzer
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      
      containers:
      - name: pihole-analyzer
        image: ghcr.io/grammatonic/pihole-analyzer:latest
        imagePullPolicy: Always
        
        args:
          - "--config"
          - "/app/config/config.json"
          - "--no-color"
        
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: reports
          mountPath: /app/reports
        
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        
        livenessProbe:
          exec:
            command:
            - ./pihole-analyzer
            - --help
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
      
      volumes:
      - name: config
        configMap:
          name: pihole-analyzer-config
      - name: reports
        persistentVolumeClaim:
          claimName: pihole-analyzer-reports

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pihole-analyzer-config
  namespace: monitoring
data:
  config.json: |
    {
      "pihole": {
        "host": "192.168.1.100",
        "port": 80,
        "api_enabled": true,
        "api_password": "your-api-password",
        "use_https": false,
        "api_timeout": 30
      },
      "output": {
        "format": "text",
        "file": "/app/reports/analysis.txt",
        "colors": false,
        "emoji": false
      },
      "logging": {
        "level": "info",
        "file": "/app/reports/pihole-analyzer.log",
        "colors": false,
        "emoji": false
      }
    }

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pihole-analyzer-reports
  namespace: monitoring
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

### Phase 4: Development Environment

#### Docker Compose Development

Create `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  pihole-analyzer-dev:
    build:
      context: .
      dockerfile: Dockerfile
      target: development
    
    container_name: pihole-analyzer-dev
    
    volumes:
      - .:/app/src:ro
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
      - ./testing/fixtures:/app/testing/fixtures
      - ./config.json:/app/config.json:ro
    
    environment:
      - CGO_ENABLED=0
      - GOCACHE=/root/.cache/go-build
      - GOMODCACHE=/go/pkg/mod
      - TEST_MODE=true
    
    working_dir: /app/src
    
    command: ["./pihole-analyzer-test", "--test", "--config", "/app/config.json"]
    
    networks:
      - dev-network

volumes:
  go-mod-cache:
  go-build-cache:

networks:
  dev-network:
    driver: bridge
```

#### Development Workflow

```bash
# Start development environment
make docker-dev

# Run tests in container
docker exec -it pihole-analyzer-dev make test

# Build and test
docker exec -it pihole-analyzer-dev make fast-build
docker exec -it pihole-analyzer-dev ./pihole-analyzer-test --test

# Interactive shell for debugging
docker exec -it pihole-analyzer-dev sh
```

## Security & Performance

### Security Hardening

#### Container Security
```bash
# Scan for vulnerabilities
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image ghcr.io/grammatonic/pihole-analyzer:latest

# Security audit
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  docker/dockle ghcr.io/grammatonic/pihole-analyzer:latest
```

#### Runtime Security
```yaml
# Security context in docker-compose
security_opt:
  - no-new-privileges:true
read_only: true
tmpfs:
  - /tmp:noexec,nosuid,size=100m
```

### Performance Optimization

#### Build Cache Strategy
```bash
# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1

# Use cache mounts for Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
```

#### Resource Management
```yaml
# Resource limits in production
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

## Monitoring & Maintenance

### Phase 5: Health Monitoring

#### Container Health Checks
```bash
# Built-in health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./pihole-analyzer --help || exit 1

# Manual health check
docker inspect --format='{{.State.Health.Status}}' pihole-analyzer-prod
```

#### Log Management
```bash
# View container logs
docker logs pihole-analyzer-prod --follow --tail 100

# Structured log analysis
docker logs pihole-analyzer-prod 2>&1 | jq '.level'

# Log rotation configuration
docker run --log-driver json-file --log-opt max-size=10m --log-opt max-file=3 \
  ghcr.io/grammatonic/pihole-analyzer:latest
```

### Monitoring Integration

#### Prometheus Metrics (Future)
```yaml
# Add metrics endpoint to container
- name: metrics
  containerPort: 8080
  protocol: TCP

# Prometheus scrape config
- job_name: 'pihole-analyzer'
  static_configs:
    - targets: ['pihole-analyzer:8080']
```

#### Grafana Dashboard (Future)
```bash
# Import dashboard for Pi-hole analysis metrics
curl -X POST \
  http://grafana:3000/api/dashboards/db \
  -H 'Content-Type: application/json' \
  -d @dashboards/pihole-analyzer.json
```

### Automated Updates

#### Watchtower Integration
```yaml
# Auto-update containers
services:
  watchtower:
    image: containrrr/watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WATCHTOWER_POLL_INTERVAL=3600
      - WATCHTOWER_CLEANUP=true
    command: pihole-analyzer-prod
```

## Troubleshooting

### Common Container Issues

#### Image Pull Failures
```bash
# Check registry authentication
docker login ghcr.io

# Manual image pull
docker pull ghcr.io/grammatonic/pihole-analyzer:latest

# Verify image integrity
docker image inspect ghcr.io/grammatonic/pihole-analyzer:latest
```

#### Container Startup Failures
```bash
# Check container logs for errors
docker logs pihole-analyzer-prod --details

# Inspect container configuration
docker inspect pihole-analyzer-prod

# Test with minimal configuration
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help
```

#### Permission Issues
```bash
# Check file ownership
ls -la config.json reports/

# Fix permissions if needed
sudo chown 1001:1001 config.json
sudo chmod 644 config.json
```

### Performance Troubleshooting

#### Resource Usage
```bash
# Monitor container resources
docker stats pihole-analyzer-prod

# Check system resources
docker system df
docker system events --filter container=pihole-analyzer-prod
```

#### Network Connectivity
```bash
# Test Pi-hole API connectivity from container
docker exec -it pihole-analyzer-prod \
  wget -q --spider http://192.168.1.100/admin/api.php

# Debug network issues
docker exec -it pihole-analyzer-prod \
  ./pihole-analyzer --config /app/config.json --debug
```

### Registry Issues

#### Push/Pull Failures
```bash
# Check authentication
docker login ghcr.io --username YOUR_USERNAME

# Verify repository permissions
gh api repos/OWNER/REPO/actions/permissions

# Test with different tag
docker tag pihole-analyzer:latest ghcr.io/grammatonic/pihole-analyzer:test
docker push ghcr.io/grammatonic/pihole-analyzer:test
```

#### Build Failures
```bash
# Check GitHub Actions logs
gh run list --repo OWNER/REPO
gh run view RUN_ID --repo OWNER/REPO

# Local build test
docker build --platform linux/amd64 -t test .
docker build --platform linux/arm64 -t test .
```

## Best Practices Summary

1. **Multi-Architecture**: Always build for AMD64, ARM64, and ARMv7
2. **Security**: Use non-root users, read-only filesystems, minimal attack surface
3. **Performance**: Implement layer caching, resource limits, health checks
4. **Monitoring**: Enable structured logging, health checks, metrics collection
5. **Automation**: Use CI/CD for builds, Watchtower for updates
6. **Testing**: Test in containers before production deployment
7. **Documentation**: Maintain deployment guides and troubleshooting steps

This container registry strategy provides a robust foundation for deploying the Pi-hole Network Analyzer in production environments with proper security, monitoring, and maintenance procedures.
