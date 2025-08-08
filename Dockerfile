# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23.12-alpine AS builder

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

# Build binaries for target platform with optimization
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
LABEL org.opencontainers.image.documentation="https://github.com/GrammaTonic/pihole-network-analyzer/blob/main/docs"

# Default to production mode
ENTRYPOINT ["./pihole-analyzer"]

# Development variant
FROM production AS development
ENTRYPOINT ["./pihole-analyzer-test", "--test"]
