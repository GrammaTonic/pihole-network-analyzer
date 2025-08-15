# Container Usage Guide

This guide covers using Pi-hole Network Analyzer with Docker containers, including both local development and production deployment scenarios.

## Quick Start

### Using Published Containers (Recommended)

```bash
# Pull and run the latest version
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest --help

# Run with your Pi-hole configuration
docker run --rm \
  -v ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro \
  ghcr.io/grammatonic/pihole-analyzer:latest \
  --show-config
```

### Using Docker Compose (Production)

```bash
# Start production environment
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop services
docker-compose -f docker-compose.prod.yml down
```

## Available Container Images

### Production Containers

| Image | Description | Size | Architectures |
|-------|-------------|------|---------------|
| `ghcr.io/grammatonic/pihole-analyzer:latest` | Latest stable release | ~44MB | amd64, arm64, armv7 |
| `ghcr.io/grammatonic/pihole-analyzer:v1.x.x` | Specific version | ~44MB | amd64, arm64, armv7 |
| `ghcr.io/grammatonic/pihole-analyzer:main` | Latest development | ~44MB | amd64, arm64, armv7 |

### Development Containers

| Image | Description | Size | Use Case |
|-------|-------------|------|----------|
| `ghcr.io/grammatonic/pihole-analyzer:latest-development` | Development variant | ~45MB | Testing, debugging |
| `ghcr.io/grammatonic/pihole-analyzer:main-development` | Development build | ~45MB | Latest features |

## Container Variants

### Production Variant (Default)

```dockerfile
# Optimized for production use
ENTRYPOINT ["./pihole-analyzer"]
```

**Features:**
- Minimal size (~44MB)
- Non-root user execution
- Health checks enabled
- Pi-hole API client for data access
- Production-ready logging

**Usage:**
```bash
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest [options]
```

### Development Variant

```dockerfile
# Includes test utilities and debugging tools
ENTRYPOINT ["./pihole-analyzer-test", "--test"]
```

**Features:**
- Test utilities included
- Mock data for development
- Debug capabilities
- Extended logging

**Usage:**
```bash
docker run --rm ghcr.io/grammatonic/pihole-analyzer:latest-development
```

## Configuration

### Volume Mounts

```bash
# Configuration directory
-v ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro

# Reports output
-v ./reports:/app/reports

# Logs directory
-v ./logs:/app/logs
```

### Environment Variables

<<<<<<< HEAD
| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PIHOLE_CONFIG` | Configuration file path | `/home/appuser/.pihole-analyzer/config.json` | `/config/pihole.json` |
| `LOG_LEVEL` | Logging level | `info` | `debug`, `info`, `warn`, `error` |
| `OUTPUT_FORMAT` | Output format | `table` | `json`, `csv`, `table` |
| `DEVELOPMENT` | Development mode | `false` | `true` |
=======
Complete environment variable support for container-first deployments:

| Category | Variable | Description | Default | Example |
|----------|----------|-------------|---------|---------|
| **Pi-hole** | `PIHOLE_HOST` | Pi-hole server IP/hostname | `pi.hole` | `192.168.1.100` |
| | `PIHOLE_PORT` | Pi-hole server port | `80` | `8080` |
| | `PIHOLE_API_PASSWORD` | Pi-hole API token | *(required)* | `your-api-token` |
| | `PIHOLE_USE_HTTPS` | Use HTTPS for API | `false` | `true` |
| | `PIHOLE_API_TIMEOUT` | API timeout seconds | `30` | `60` |
| **Web** | `WEB_ENABLED` | Enable web dashboard | `false` | `true` |
| | `WEB_HOST` | Web server bind address | `localhost` | `0.0.0.0` |
| | `WEB_PORT` | Web server port | `8080` | `3000` |
| | `WEB_DAEMON_MODE` | Run as background service | `false` | `true` |
| **Logging** | `LOG_LEVEL` | Logging level | `info` | `debug`, `warn`, `error` |
| | `LOG_ENABLE_COLORS` | Colorized log output | `true` | `false` |
| | `LOG_ENABLE_EMOJIS` | Emoji in logs | `true` | `false` |
| **Analysis** | `ANALYSIS_ONLINE_ONLY` | Show only online devices | `false` | `true` |
| **Metrics** | `METRICS_ENABLED` | Enable Prometheus metrics | `true` | `false` |
| | `METRICS_HOST` | Metrics server bind address | `localhost` | `0.0.0.0` |
| | `METRICS_PORT` | Metrics server port | `9090` | `9091` |
| **Runtime** | `GOMEMLIMIT` | Go memory limit | `128MiB` | `256MiB` |
| | `GOMAXPROCS` | Go max processors | `2` | `4` |

### Configuration Priority

Environment variables follow this precedence order:
1. **CLI flags** (highest priority)
2. **Environment variables**
3. **Configuration file** 
4. **Defaults** (lowest priority)
>>>>>>> main

## Docker Compose Configurations

### Basic Development Setup

```yaml
version: '3.8'
services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest-development
    volumes:
      - ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro
      - ./reports:/app/reports
    environment:
      - LOG_LEVEL=debug
```

### Production Deployment

```yaml
version: '3.8'
services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    volumes:
      - ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro
      - ./reports:/app/reports
    environment:
      - LOG_LEVEL=info
      - OUTPUT_FORMAT=json
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./pihole-analyzer", "--help"]
      interval: 30s
      timeout: 3s
      retries: 3
```

### Scheduled Analysis

```yaml
version: '3.8'
services:
  pihole-analyzer-scheduler:
    image: ghcr.io/grammatonic/pihole-analyzer:latest
    volumes:
      - ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro
      - ./reports:/app/reports
    environment:
      - SCHEDULE_ENABLED=true
      - SCHEDULE_INTERVAL=3600  # Every hour
    command: ["./pihole-analyzer", "--schedule"]
    restart: unless-stopped
```

## Makefile Targets

### Container Building

```bash
# Build local Docker image
make docker-build

# Build development variant
make docker-build-dev

# Build production variant
make docker-build-prod

# Build multi-architecture images
make docker-build-multi
```

### Container Management

```bash
# Start development environment
make docker-dev

# Start production environment
make docker-prod

# Run tests in container
make docker-test

# View container logs
make docker-logs

# Access container shell
make docker-shell

# Clean up containers
make docker-clean
```

### Container Registry

```bash
# Push to registry (requires authentication)
make docker-push
```

## Multi-Architecture Support

### Supported Platforms

- **linux/amd64** - Intel/AMD 64-bit (development machines, cloud)
- **linux/arm64** - ARM 64-bit (Raspberry Pi 4, Apple Silicon)
- **linux/arm/v7** - ARM 32-bit (Raspberry Pi 3)

### Platform-Specific Usage

```bash
# Force specific platform
docker run --platform linux/arm64 \
  ghcr.io/grammatonic/pihole-analyzer:latest --help

# Multi-platform build
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  --target production \
  -t pihole-analyzer:multi .
```

## Security Features

### Container Security

- **Non-root execution** - Runs as `appuser` (UID 1001)
- **Minimal attack surface** - Alpine Linux base with minimal packages
- **Read-only filesystem** - Configuration mounted read-only
- **Health checks** - Built-in health monitoring
- **No privileged access** - Standard user permissions

### Security Scanning

All published containers undergo:
- **Vulnerability scanning** with Trivy
- **Supply chain verification** with SBOM
- **Build attestation** with provenance
- **Regular security updates** via automated rebuilds

### Security Best Practices

```bash
# Use read-only configuration
-v ~/.pihole-analyzer:/home/appuser/.pihole-analyzer:ro

# Limit container resources
--memory=512m --cpus=0.5

# Use specific version tags
ghcr.io/grammatonic/pihole-analyzer:v1.2.3

# Enable security scanning
docker scan ghcr.io/grammatonic/pihole-analyzer:latest
```

## Performance Optimization

### Container Caching

```bash
# Enable BuildKit for better caching
export DOCKER_BUILDKIT=1

# Use cache mounts for development
docker run --rm \
  -v go-cache:/go/pkg/mod \
  -v build-cache:/home/appuser/.cache/go-build \
  ghcr.io/grammatonic/pihole-analyzer:latest-development
```

### Resource Limits

```yaml
# Docker Compose resource limits
services:
  pihole-analyzer:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

## Troubleshooting

### Common Issues

**Container won't start:**
```bash
# Check container logs
docker logs pihole-analyzer

# Verify health status
docker inspect --format='{{.State.Health.Status}}' pihole-analyzer
```

**Permission issues:**
```bash
# Ensure correct ownership
sudo chown -R 1001:1001 ~/.pihole-analyzer

# Check mount permissions
ls -la ~/.pihole-analyzer
```

**Network connectivity:**
```bash
# Test Pi-hole API connection from container
docker exec -it pihole-analyzer curl http://192.168.1.100/admin/api.php

# Verify network access
docker exec -it pihole-analyzer ping pihole-server
```

### Debug Mode

```bash
# Run with debug logging
docker run --rm \
  -e LOG_LEVEL=debug \
  ghcr.io/grammatonic/pihole-analyzer:latest-development

# Interactive debugging
docker run --rm -it \
  ghcr.io/grammatonic/pihole-analyzer:latest-development sh
```

## Registry Authentication

### GitHub Container Registry

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull private images
docker pull ghcr.io/grammatonic/pihole-analyzer:main

# Push custom builds
docker tag pihole-analyzer:custom ghcr.io/grammatonic/pihole-analyzer:custom
docker push ghcr.io/grammatonic/pihole-analyzer:custom
```

## Integration Examples

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pihole-analyzer
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
      containers:
      - name: pihole-analyzer
        image: ghcr.io/grammatonic/pihole-analyzer:latest
        env:
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /home/appuser/.pihole-analyzer
          readOnly: true
        - name: reports
          mountPath: /app/reports
      volumes:
      - name: config
        configMap:
          name: pihole-config
      - name: reports
        persistentVolumeClaim:
          claimName: pihole-reports
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Run analysis in container
  run: |
    docker run --rm \
      -v ${{ github.workspace }}/config:/home/appuser/.pihole-analyzer:ro \
      -v ${{ github.workspace }}/reports:/app/reports \
      ghcr.io/grammatonic/pihole-analyzer:latest \
      --output-format json > analysis.json
```

This guide provides comprehensive coverage of container usage patterns, from simple local development to complex production deployments. The containerized approach simplifies deployment while maintaining security and performance.
