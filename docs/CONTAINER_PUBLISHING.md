# Container Publishing Guide

This guide covers the comprehensive container publishing automation for the Pi-hole Network Analyzer, including GitHub Container Registry (GHCR) integration, multi-architecture builds, and development workflows.

## Overview

The Pi-hole Network Analyzer provides a complete container publishing pipeline with:

- **Multi-Architecture Support**: linux/amd64, linux/arm64, linux/arm/v7
- **GitHub Container Registry Integration**: Automated publishing to ghcr.io
- **Semantic Versioning**: Automatic version tagging based on semantic-release
- **Development & Production Variants**: Separate workflows for different environments
- **Advanced Automation**: Comprehensive scripts and Makefile integration

## Container Publishing Targets

### Automated Publishing (CI/CD)

The repository automatically publishes containers on releases via GitHub Actions:

```yaml
# .github/workflows/release.yml
- Builds multi-architecture images
- Publishes to ghcr.io/grammatonic/pihole-network-analyzer
- Tags with semantic versions (v1.2.3, v1.2, v1, latest)
- Includes build metadata and labels
```

### Manual Publishing Commands

#### Quick Publishing
```bash
# Push production containers to GHCR
make docker-push-ghcr

# Push development containers
make container-push-dev

# Push production containers  
make container-push-prod

# Push all container variants
make container-push-all
```

#### Container Management
```bash
# List available container targets
make container-list

# Show container information
make container-info

# Login to GitHub Container Registry
make container-login
```

## Advanced Container Script

The `scripts/container-push.sh` script provides comprehensive container publishing automation:

### Features

- **Multi-Architecture Builds**: Supports AMD64, ARM64, and ARMv7
- **Semantic Version Tagging**: Automatic version detection and tagging
- **Authentication Handling**: GHCR authentication with fallback strategies
- **Development Variants**: Separate development and production builds
- **Error Handling**: Comprehensive error checking and recovery
- **Prerequisites Checking**: Validates required tools and dependencies

### Usage Examples

```bash
# Basic production push
./scripts/container-push.sh

# Development variant push
./scripts/container-push.sh --dev

# Specific version push
./scripts/container-push.sh --version v1.2.3

# Dry run (show what would be built)
./scripts/container-push.sh --dry-run

# Multiple tags
./scripts/container-push.sh --tag latest --tag stable

# Custom registry
./scripts/container-push.sh --registry custom.registry.io

# Enable debug output
./scripts/container-push.sh --debug
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--dev` | Build development variant | Production |
| `--version VERSION` | Specific version to tag | Auto-detect |
| `--tag TAG` | Additional tag to apply | semantic tags |
| `--registry REGISTRY` | Container registry | ghcr.io |
| `--no-push` | Build only, don't push | Push enabled |
| `--dry-run` | Show commands without executing | Execute |
| `--debug` | Enable debug output | Normal output |
| `--platforms PLATFORMS` | Target platforms | linux/amd64,linux/arm64,linux/arm/v7 |

## Container Architecture

### Multi-Stage Dockerfile

The Dockerfile uses multi-stage builds for optimized production containers:

```dockerfile
# Development stage (full toolchain)
FROM golang:1.23-alpine AS development

# Production stage (minimal runtime)
FROM alpine:3.20 AS production
```

### Build Variants

#### Production Container
- **Base**: Alpine Linux 3.20
- **Size**: Minimal (~15MB)
- **User**: Non-root (`analyzer:1001`)
- **Health Check**: Built-in health monitoring
- **Security**: Minimal attack surface

#### Development Container  
- **Base**: Golang 1.23 Alpine
- **Tools**: Full Go toolchain, debugging tools
- **Volumes**: Source code mounting for live development
- **Features**: Hot reload, comprehensive logging

### Container Metadata

All containers include comprehensive metadata:

```dockerfile
LABEL org.opencontainers.image.title="Pi-hole Network Analyzer"
LABEL org.opencontainers.image.description="DNS analysis tool for Pi-hole"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/GrammaTonic/pihole-network-analyzer"
LABEL org.opencontainers.image.licenses="MIT"
```

## GitHub Container Registry Integration

### Automatic Publishing

Containers are automatically published on releases:

1. **Semantic Release**: Creates new version tag
2. **Docker Build**: Multi-architecture container build  
3. **Registry Push**: Publishes to ghcr.io
4. **Tagging Strategy**: Semantic version tags (v1.2.3, v1.2, v1, latest)

### Manual Publishing

#### Authentication

Login to GitHub Container Registry:

```bash
# Using GitHub CLI (recommended)
gh auth token | docker login ghcr.io -u USERNAME --password-stdin

# Using Personal Access Token
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Using Makefile helper
make container-login
```

#### Publishing Workflow

```bash
# 1. Ensure clean working directory
git status

# 2. Login to registry
make container-login

# 3. Build and push containers
make docker-push-ghcr

# 4. Verify publication
make container-info
```

### Registry Configuration

The container registry is configured for:

- **Registry**: ghcr.io (GitHub Container Registry)
- **Repository**: grammatonic/pihole-network-analyzer
- **Visibility**: Public (configurable)
- **Retention**: Configurable via GitHub settings

## Development Workflows

### Local Development

```bash
# Build development container
docker build --target development -t pihole-analyzer:dev .

# Run with source mounting
docker run -v $(pwd):/app pihole-analyzer:dev

# Use docker-compose for development
docker-compose -f docker-compose.dev.yml up
```

### Testing Containers

```bash
# Build test containers
make docker-build

# Run container tests
make docker-test

# Integration testing
docker run --rm pihole-analyzer:latest --version
```

### Container Debugging

```bash
# Run interactive shell in container
docker run -it --entrypoint sh pihole-analyzer:latest

# Debug with development container
docker run -it pihole-analyzer:dev sh

# Check container metadata
docker inspect pihole-analyzer:latest
```

## CI/CD Integration

### GitHub Actions Workflow

The release workflow includes comprehensive container publishing:

```yaml
docker-publish:
  name: Publish Docker Images  
  needs: release
  runs-on: ubuntu-latest
  if: needs.release.outputs.released == 'true'
  permissions:
    contents: read
    packages: write  # Required for GHCR
  
  steps:
  - name: Set up Docker Buildx
    uses: docker/setup-buildx-action@v3
    
  - name: Log in to GitHub Container Registry
    uses: docker/login-action@v3
    with:
      registry: ghcr.io
      username: ${{ github.actor }}
      password: ${{ secrets.GITHUB_TOKEN }}
      
  - name: Build and push Docker image
    uses: docker/build-push-action@v5
    with:
      platforms: linux/amd64,linux/arm64,linux/arm/v7
      push: true
      cache-from: type=gha
      cache-to: type=gha,mode=max
```

### Build Optimization

- **Build Cache**: GitHub Actions cache for faster builds
- **Multi-Stage**: Optimized layer caching
- **Parallel Builds**: Multi-architecture parallel building
- **Cache Persistence**: Cross-workflow cache reuse

## Security Considerations

### Container Security

- **Non-Root User**: Containers run as `analyzer:1001`
- **Minimal Base**: Alpine Linux for reduced attack surface
- **No Privileged Access**: Standard user permissions
- **Read-Only Filesystem**: Immutable container layers

### Registry Security

- **Authentication**: GitHub token-based authentication
- **Permissions**: Scoped registry access
- **Vulnerability Scanning**: Automated security scanning
- **Access Control**: Repository-based access control

## Troubleshooting

### Common Issues

#### Authentication Errors
```bash
# Check GitHub CLI authentication
gh auth status

# Verify token permissions
gh auth token

# Re-authenticate if needed
gh auth login
```

#### Build Failures
```bash
# Check Docker daemon
docker version

# Verify buildx installation
docker buildx version

# Clean build cache
docker builder prune
```

#### Registry Issues
```bash
# Check registry connectivity
docker pull ghcr.io/grammatonic/pihole-network-analyzer:latest

# Verify repository permissions
gh api repos/GrammaTonic/pihole-network-analyzer

# Check container metadata
make container-info
```

### Debug Commands

```bash
# Enable debug logging
export DEBUG=1
./scripts/container-push.sh --debug

# Dry run to see commands
./scripts/container-push.sh --dry-run

# Check container layers
docker history pihole-analyzer:latest

# Inspect build context
make container-info
```

## Integration Examples

### Monitoring Stack

```yaml
# docker-compose.monitoring.yml
version: '3.8'
services:
  pihole-analyzer:
    image: ghcr.io/grammatonic/pihole-network-analyzer:latest
    environment:
      - WEB_ENABLED=true
      - METRICS_ENABLED=true
    ports:
      - "8080:8080"
      - "9090:9090"
```

### Production Deployment

```bash
# Pull latest production image
docker pull ghcr.io/grammatonic/pihole-network-analyzer:latest

# Run with production configuration
docker run -d \
  --name pihole-analyzer \
  -p 8080:8080 \
  -v /path/to/config:/config \
  ghcr.io/grammatonic/pihole-network-analyzer:latest \
  --config /config/pihole.json --web --daemon
```

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
        image: ghcr.io/grammatonic/pihole-network-analyzer:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
```

## Best Practices

### Version Management

1. **Semantic Versioning**: Use conventional commits for automatic versioning
2. **Tag Strategy**: Apply multiple tags (specific version, major.minor, major, latest)
3. **Version Consistency**: Ensure version alignment across binaries and containers
4. **Release Notes**: Include container information in release documentation

### Build Optimization

1. **Layer Caching**: Optimize Dockerfile for effective layer caching
2. **Multi-Stage**: Use multi-stage builds for minimal production images
3. **Build Context**: Minimize build context with .dockerignore
4. **Parallel Builds**: Leverage multi-architecture parallel building

### Security Best Practices

1. **Minimal Images**: Use minimal base images (Alpine)
2. **Non-Root**: Always run as non-root user
3. **Secrets Management**: Never embed secrets in images
4. **Regular Updates**: Keep base images updated

### Development Workflow

1. **Local Testing**: Test containers locally before pushing
2. **Development Variants**: Use development containers for active development
3. **Automated Testing**: Include container tests in CI/CD
4. **Documentation**: Maintain up-to-date container documentation

## Additional Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Multi-Architecture Builds](https://docs.docker.com/build/building/multi-platform/)
- [Docker Buildx](https://docs.docker.com/buildx/)

---

This container publishing system provides a comprehensive, automated, and secure solution for distributing the Pi-hole Network Analyzer across multiple architectures and deployment scenarios.
