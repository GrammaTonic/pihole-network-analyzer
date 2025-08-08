# Multi-stage Dockerfile with build caching optimization
# Build stage with Go build cache
FROM golang:1.23-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations and caching
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o pihole-analyzer \
    ./cmd/pihole-analyzer

# Build test binary as well
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o pihole-analyzer-test \
    ./cmd/pihole-analyzer-test

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests and ssh for Pi-hole connections
RUN apk --no-cache add ca-certificates openssh-client

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/pihole-analyzer .
COPY --from=builder /app/pihole-analyzer-test .

# Copy test fixtures for container testing
COPY --from=builder /app/testing/fixtures ./testing/fixtures

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose any needed ports (none needed for this CLI app)
# EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./pihole-analyzer --help || exit 1

# Default command runs test mode
CMD ["./pihole-analyzer-test", "--test"]

# Labels for metadata
LABEL org.opencontainers.image.title="Pi-hole Network Analyzer"
LABEL org.opencontainers.image.description="Analyze Pi-hole DNS queries and network traffic"
LABEL org.opencontainers.image.vendor="GrammaTonic"
LABEL org.opencontainers.image.licenses="MIT"
