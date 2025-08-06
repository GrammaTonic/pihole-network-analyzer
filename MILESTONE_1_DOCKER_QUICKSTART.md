# 🐳 Milestone 1: Basic Docker Support - Quick Start

This document provides the essential files and changes needed to implement basic Docker containerization for the Pi-hole Network Analyzer.

## Files to Create

### 1. `Dockerfile`
```dockerfile
# Multi-stage build for optimal image size
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pihole-network-analyzer .

# Final minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates openssh-client
WORKDIR /root/

COPY --from=builder /app/pihole-network-analyzer .

# Create non-root user for security
RUN adduser -D -s /bin/sh pihole
USER pihole
WORKDIR /home/pihole

COPY --from=builder /app/pihole-network-analyzer .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./pihole-network-analyzer --health || exit 1

EXPOSE 8080

CMD ["./pihole-network-analyzer", "--server"]
```

### 2. `docker-compose.yml`
```yaml
version: '3.8'

services:
  pihole-analyzer:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PIHOLE_HOST=${PIHOLE_HOST:-192.168.1.100}
      - PIHOLE_USERNAME=${PIHOLE_USERNAME:-pi}
      - PIHOLE_DB_PATH=${PIHOLE_DB_PATH:-/etc/pihole/pihole-FTL.db}
      - SERVER_PORT=8080
      - LOG_LEVEL=info
    volumes:
      - ${SSH_KEY_PATH:-~/.ssh/id_rsa}:/home/pihole/.ssh/id_rsa:ro
      - ./config:/home/pihole/config
      - ./reports:/home/pihole/reports
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "./pihole-network-analyzer", "--health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Future: Prometheus and Grafana will be added here
  # prometheus:
  #   image: prom/prometheus:latest
  #   ports:
  #     - "9090:9090"
  
  # grafana:
  #   image: grafana/grafana:latest
  #   ports:
  #     - "3000:3000"
```

### 3. `.dockerignore`
```
# Build artifacts
pihole-network-analyzer*
dns_usage_report_*.txt
*.db

# Development files
.git/
.github/
*.md
!README.md
.gitignore

# Test files
test_data/
*_test.go

# IDE files
.vscode/
.idea/
*.swp

# OS files
.DS_Store
Thumbs.db
```

### 4. `docker/entrypoint.sh`
```bash
#!/bin/sh
set -e

# Wait for SSH key if mounted
if [ -f "/home/pihole/.ssh/id_rsa" ]; then
    chmod 600 /home/pihole/.ssh/id_rsa
    echo "✅ SSH key loaded"
fi

# Create config directory if needed
mkdir -p /home/pihole/config /home/pihole/reports

# Start the application
echo "🚀 Starting Pi-hole Network Analyzer..."
exec "$@"
```

## Code Changes Required

### 1. Add Server Mode Flag
Add to `main.go`:
```go
var (
    // ... existing flags ...
    serverFlag = flag.Bool("server", false, "Run in server mode with HTTP endpoints")
    serverPortFlag = flag.String("port", "8080", "Server port for HTTP endpoints")
    healthFlag = flag.Bool("health", false, "Run health check and exit")
)
```

### 2. Environment Variable Support
Add to `config.go`:
```go
func LoadConfigFromEnv() *Config {
    config := DefaultConfig()
    
    if host := os.Getenv("PIHOLE_HOST"); host != "" {
        config.Pihole.Host = host
    }
    if username := os.Getenv("PIHOLE_USERNAME"); username != "" {
        config.Pihole.Username = username
    }
    if dbPath := os.Getenv("PIHOLE_DB_PATH"); dbPath != "" {
        config.Pihole.DBPath = dbPath
    }
    
    return config
}
```

### 3. Basic HTTP Server
Add new file `server.go`:
```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "log"
)

type HealthResponse struct {
    Status  string `json:"status"`
    Version string `json:"version"`
    Uptime  string `json:"uptime"`
}

func startServer(port string) {
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/", indexHandler)
    
    fmt.Printf("🚀 Server starting on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{
        Status:  "healthy",
        Version: "v1.0.0", // TODO: Use build version
        Uptime:  "running",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Pi-hole Network Analyzer - Server Mode\n")
    fmt.Fprintf(w, "Health: /health\n")
    fmt.Fprintf(w, "Metrics: /metrics (coming soon)\n")
}
```

## Quick Implementation Steps

### Step 1: Add Basic Server Support
```bash
# 1. Add server mode to main.go
# 2. Create server.go with HTTP endpoints
# 3. Add environment variable support
# 4. Test locally: go run . --server
```

### Step 2: Create Docker Files
```bash
# 1. Create Dockerfile
# 2. Create docker-compose.yml  
# 3. Create .dockerignore
# 4. Test build: docker build -t pihole-analyzer .
```

### Step 3: Test Container
```bash
# 1. Build image
docker build -t pihole-analyzer .

# 2. Run container
docker run -p 8080:8080 pihole-analyzer

# 3. Test endpoints
curl http://localhost:8080/health

# 4. Test with docker-compose
docker-compose up
```

## Expected Results

After implementing Milestone 1:

✅ **Application runs in Docker container**  
✅ **Health check endpoint working**  
✅ **Environment variable configuration**  
✅ **Docker Compose orchestration**  
✅ **Ready for Prometheus metrics (Milestone 2)**

## Next Steps

Once Milestone 1 is complete:
1. **Milestone 2**: Add Prometheus metrics endpoint
2. **Milestone 3**: Create Grafana dashboard
3. **Milestone 4**: Production deployment features

---

**Estimated Time**: 1-2 days for basic Docker support
**Complexity**: Low - mostly configuration and basic HTTP server
