# syntax=docker/dockerfile:1
# Multi-platform Pi-hole Network Analyzer container
FROM --platform=$BUILDPLATFORM golang:1.23.12-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM
ARG VERSION=dev
ARG BUILD_TIME

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binaries for target platform with optimization and build info
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o pihole-analyzer ./cmd/pihole-analyzer

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o pihole-analyzer-test ./cmd/pihole-analyzer-test

# Production runtime stage
FROM alpine:latest AS production

# Security: Add non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Install runtime dependencies including CA certificates for HTTPS API connections
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/pihole-analyzer .
COPY --from=builder /app/pihole-analyzer-test .
COPY --from=builder /app/testing/fixtures ./testing/fixtures

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Health check using built-in validation
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./pihole-analyzer", "--help"]

# Add container labels for better metadata
LABEL org.opencontainers.image.title="Pi-hole Network Analyzer"
LABEL org.opencontainers.image.description="Analyze Pi-hole DNS queries via API"
LABEL org.opencontainers.image.vendor="GrammaTonic"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/GrammaTonic/pihole-network-analyzer"
LABEL org.opencontainers.image.documentation="https://github.com/GrammaTonic/pihole-network-analyzer/blob/main/docs"

# Default entrypoint for production
ENTRYPOINT ["./pihole-analyzer"]

# Development variant with test data
FROM production AS development
ENTRYPOINT ["./pihole-analyzer-test", "--test"]
