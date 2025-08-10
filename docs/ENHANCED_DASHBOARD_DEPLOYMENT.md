# Enhanced Dashboard Deployment Guide

## Overview

This guide provides step-by-step deployment instructions for the Enhanced Dashboard feature in both development and production environments.

## Prerequisites

### Development Environment
- Go 1.21+ installed
- Docker and Docker Compose
- Make build system
- Node.js (for release automation only)

### Production Environment
- Docker runtime environment
- Reverse proxy (nginx/Apache) recommended
- SSL/TLS certificates for HTTPS
- Network access to Pi-hole API

## Quick Start Deployment

### 1. Development Deployment

```bash
# Clone repository
git clone https://github.com/your-org/pihole-network-analyzer
cd pihole-network-analyzer

# Build enhanced dashboard
make build-all

# Start with mock data (no Pi-hole required)
./pihole-analyzer-test --web --test

# Access dashboard
open http://localhost:8080/dashboard/enhanced
```

### 2. Production Deployment

```bash
# Pull production image
docker pull ghcr.io/your-org/pihole-network-analyzer:latest

# Create config file
cp config-example-with-alerts.json config.json
# Edit config.json with your Pi-hole settings

# Run container
docker run -d \
  --name pihole-analyzer \
  -p 8080:8080 \
  -v $(pwd)/config.json:/app/config.json:ro \
  ghcr.io/your-org/pihole-network-analyzer:latest \
  --web --config /app/config.json
```

## Configuration

### Minimal Configuration

Create `config.json`:
```json
{
  "pihole": {
    "host": "192.168.1.100",
    "api_key": "your-api-key-here",
    "timeout": "30s"
  },
  "web": {
    "host": "0.0.0.0",
    "port": 8080,
    "enable_dashboard": true,
    "enable_enhanced_dashboard": true
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

### Enhanced Configuration

For full feature set:
```json
{
  "pihole": {
    "host": "192.168.1.100",
    "api_key": "your-api-key-here",
    "timeout": "30s",
    "verify_ssl": true
  },
  "web": {
    "host": "0.0.0.0",
    "port": 8080,
    "enable_dashboard": true,
    "enable_enhanced_dashboard": true,
    "websocket_enabled": true,
    "cors_enabled": true,
    "rate_limit": {
      "requests_per_minute": 60,
      "burst_size": 10
    }
  },
  "metrics": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 9090,
    "path": "/metrics"
  },
  "ml": {
    "enabled": true,
    "anomaly_detection": {
      "enabled": true,
      "sensitivity": 0.75,
      "window_size": "1h"
    },
    "trend_analysis": {
      "enabled": true,
      "prediction_window": "24h"
    }
  },
  "alerts": {
    "enabled": true,
    "rules": [
      {
        "name": "high_blocked_queries",
        "condition": "blocked_percentage > 30",
        "severity": "warning",
        "notifications": ["slack", "email"]
      }
    ],
    "notifications": {
      "slack": {
        "webhook_url": "https://hooks.slack.com/...",
        "channel": "#network-alerts"
      },
      "email": {
        "smtp_host": "smtp.gmail.com",
        "smtp_port": 587,
        "to": ["admin@example.com"]
      }
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  }
}
```

## Deployment Options

### Option 1: Docker Compose (Recommended)

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  pihole-analyzer:
    image: ghcr.io/your-org/pihole-network-analyzer:latest
    container_name: pihole-analyzer
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./config.json:/app/config.json:ro
      - ./data:/app/data
    command: [
      "--web",
      "--metrics", 
      "--daemon",
      "--config", "/app/config.json"
    ]
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    environment:
      - TZ=America/New_York

  # Optional: Add reverse proxy
  nginx:
    image: nginx:alpine
    container_name: pihole-analyzer-proxy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - pihole-analyzer
```

Deploy:
```bash
docker-compose up -d
docker-compose logs -f pihole-analyzer
```

### Option 2: Kubernetes Deployment

Create `k8s-deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pihole-analyzer
  labels:
    app: pihole-analyzer
spec:
  replicas: 2
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
        image: ghcr.io/your-org/pihole-network-analyzer:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        command: [
          "--web",
          "--metrics",
          "--daemon", 
          "--config", "/app/config.json"
        ]
        volumeMounts:
        - name: config
          mountPath: /app/config.json
          subPath: config.json
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: pihole-analyzer-config

---
apiVersion: v1
kind: Service
metadata:
  name: pihole-analyzer-service
spec:
  selector:
    app: pihole-analyzer
  ports:
  - name: web
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pihole-analyzer-config
data:
  config.json: |
    {
      "pihole": {
        "host": "pihole.local",
        "api_key": "your-api-key"
      },
      "web": {
        "host": "0.0.0.0",
        "port": 8080,
        "enable_enhanced_dashboard": true
      }
    }
```

Deploy:
```bash
kubectl apply -f k8s-deployment.yaml
kubectl get pods -l app=pihole-analyzer
kubectl port-forward service/pihole-analyzer-service 8080:8080
```

### Option 3: Systemd Service

Create `/etc/systemd/system/pihole-analyzer.service`:
```ini
[Unit]
Description=Pi-hole Network Analyzer with Enhanced Dashboard
After=network.target
Wants=network.target

[Service]
Type=exec
User=pihole-analyzer
Group=pihole-analyzer
WorkingDirectory=/opt/pihole-analyzer
ExecStart=/opt/pihole-analyzer/pihole-analyzer --web --metrics --daemon --config /opt/pihole-analyzer/config.json
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=pihole-analyzer

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/pihole-analyzer/data

[Install]
WantedBy=multi-user.target
```

Install and start:
```bash
# Create user and directories
sudo useradd --system --shell /bin/false pihole-analyzer
sudo mkdir -p /opt/pihole-analyzer/{data,logs}
sudo chown -R pihole-analyzer:pihole-analyzer /opt/pihole-analyzer

# Copy binary and config
sudo cp pihole-analyzer /opt/pihole-analyzer/
sudo cp config.json /opt/pihole-analyzer/
sudo chown pihole-analyzer:pihole-analyzer /opt/pihole-analyzer/*

# Install and start service
sudo systemctl daemon-reload
sudo systemctl enable pihole-analyzer
sudo systemctl start pihole-analyzer
sudo systemctl status pihole-analyzer
```

## Reverse Proxy Configuration

### Nginx Configuration

Create `/etc/nginx/sites-available/pihole-analyzer`:
```nginx
upstream pihole_analyzer {
    server 127.0.0.1:8080;
    keepalive 32;
}

upstream pihole_analyzer_metrics {
    server 127.0.0.1:9090;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name analyzer.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name analyzer.yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/ssl/certs/analyzer.yourdomain.com.crt;
    ssl_certificate_key /etc/ssl/private/analyzer.yourdomain.com.key;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # Modern SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/json;

    # Main application
    location / {
        proxy_pass http://pihole_analyzer;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://pihole_analyzer;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }

    # Metrics endpoint (optional, for monitoring)
    location /metrics {
        proxy_pass http://pihole_analyzer_metrics;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Restrict access to monitoring systems
        allow 192.168.1.0/24;
        allow 10.0.0.0/8;
        deny all;
    }

    # Static assets caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        proxy_pass http://pihole_analyzer;
        proxy_set_header Host $host;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/pihole-analyzer /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Apache Configuration

Create `/etc/apache2/sites-available/pihole-analyzer.conf`:
```apache
<VirtualHost *:80>
    ServerName analyzer.yourdomain.com
    Redirect permanent / https://analyzer.yourdomain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName analyzer.yourdomain.com
    DocumentRoot /var/www/html

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/ssl/certs/analyzer.yourdomain.com.crt
    SSLCertificateKeyFile /etc/ssl/private/analyzer.yourdomain.com.key
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384

    # Security Headers
    Header always set Strict-Transport-Security "max-age=63072000"
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff

    # Proxy Configuration
    ProxyPreserveHost On
    ProxyRequests Off

    # Main application
    ProxyPass / http://127.0.0.1:8080/
    ProxyPassReverse / http://127.0.0.1:8080/

    # WebSocket support
    ProxyPass /ws ws://127.0.0.1:8080/ws
    ProxyPassReverse /ws ws://127.0.0.1:8080/ws

    # Metrics (restricted)
    <Location "/metrics">
        ProxyPass http://127.0.0.1:9090/metrics
        ProxyPassReverse http://127.0.0.1:9090/metrics
        Require ip 192.168.1
        Require ip 10
    </Location>

    # Logging
    ErrorLog ${APACHE_LOG_DIR}/pihole-analyzer_error.log
    CustomLog ${APACHE_LOG_DIR}/pihole-analyzer_access.log combined
</VirtualHost>
```

Enable modules and site:
```bash
sudo a2enmod ssl proxy proxy_http proxy_wstunnel headers
sudo a2ensite pihole-analyzer.conf
sudo systemctl reload apache2
```

## Monitoring and Health Checks

### Health Check Endpoint

The application provides a health check endpoint:
```bash
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "timestamp": "2025-08-10T20:24:30Z",
  "uptime": "2h34m12s",
  "components": {
    "pihole_api": "healthy",
    "web_server": "healthy",
    "websocket": "healthy",
    "metrics": "healthy"
  }
}
```

### Prometheus Monitoring

Add to Prometheus configuration (`prometheus.yml`):
```yaml
scrape_configs:
  - job_name: 'pihole-analyzer'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s
    metrics_path: /metrics
```

Key metrics to monitor:
- `pihole_analyzer_queries_total`: Total DNS queries processed
- `pihole_analyzer_response_time_seconds`: API response times
- `pihole_analyzer_active_connections`: Active WebSocket connections
- `pihole_analyzer_cache_hit_ratio`: Cache hit percentage

### Log Monitoring

Configure log aggregation (example with Fluentd):
```yaml
<source>
  @type tail
  path /var/log/pihole-analyzer/app.log
  pos_file /var/log/fluentd/pihole-analyzer.log.pos
  tag pihole.analyzer
  format json
  time_key timestamp
  time_format %Y-%m-%dT%H:%M:%S.%LZ
</source>

<match pihole.analyzer>
  @type elasticsearch
  host elasticsearch.local
  port 9200
  index_name pihole-analyzer
  type_name logs
</match>
```

## Troubleshooting

### Common Issues

**1. Cannot Connect to Pi-hole API**
```bash
# Check Pi-hole connectivity
curl -H "X-Pi-hole-API: your-api-key" "http://192.168.1.100/admin/api.php?summaryRaw"

# Check config
./pihole-analyzer --test-config config.json

# Enable debug logging
./pihole-analyzer --web --config config.json --log-level debug
```

**2. WebSocket Connection Fails**
```bash
# Test WebSocket directly
wscat -c ws://localhost:8080/ws

# Check reverse proxy WebSocket support
curl -H "Upgrade: websocket" -H "Connection: upgrade" http://localhost:8080/ws
```

**3. High Memory Usage**
```bash
# Monitor resource usage
docker stats pihole-analyzer

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap
```

**4. Slow Dashboard Loading**
```bash
# Enable caching in reverse proxy
# Check network latency to Pi-hole
ping 192.168.1.100

# Optimize query intervals
# Reduce ML processing frequency in config
```

### Debug Commands

```bash
# Test configuration
./pihole-analyzer --test-config config.json

# Validate Pi-hole connection
./pihole-analyzer --validate-pihole config.json

# Test with mock data
./pihole-analyzer-test --web --test

# Check build info
./pihole-analyzer --version

# Health check
curl -f http://localhost:8080/health || echo "Health check failed"
```

### Log Analysis

```bash
# Follow application logs
docker logs -f pihole-analyzer

# Filter error logs
docker logs pihole-analyzer 2>&1 | grep ERROR

# Analyze access patterns
tail -f /var/log/nginx/access.log | grep "dashboard"

# Monitor WebSocket connections
docker logs pihole-analyzer 2>&1 | grep "websocket"
```

## Security Considerations

### Network Security
- Run behind reverse proxy with SSL/TLS
- Restrict API access to trusted networks
- Use strong API keys and rotate regularly
- Enable rate limiting and DDoS protection

### Container Security
- Use non-root user in containers
- Enable read-only root filesystem where possible
- Limit container capabilities
- Scan images for vulnerabilities

### Application Security
- Enable CSRF protection in production
- Validate all input parameters
- Use secure WebSocket connections (WSS)
- Implement proper session management

### Example Secure Configuration
```json
{
  "web": {
    "host": "127.0.0.1",
    "port": 8080,
    "enable_enhanced_dashboard": true,
    "security": {
      "csrf_protection": true,
      "secure_cookies": true,
      "rate_limiting": {
        "enabled": true,
        "requests_per_minute": 60
      }
    }
  },
  "logging": {
    "level": "info",
    "audit_enabled": true,
    "sensitive_data_masking": true
  }
}
```

## Performance Tuning

### Application Tuning
```json
{
  "web": {
    "cache": {
      "enabled": true,
      "ttl": "30s",
      "max_size": "100MB"
    },
    "websocket": {
      "buffer_size": 1024,
      "max_connections": 100,
      "ping_interval": "30s"
    }
  },
  "pihole": {
    "connection_pool_size": 10,
    "timeout": "15s",
    "retry_attempts": 3
  }
}
```

### Database Optimization
```bash
# If using persistent storage
docker run -v analyzer-data:/app/data pihole-analyzer

# Optimize query intervals
# Reduce unnecessary API calls
# Implement intelligent caching strategies
```

---

*For additional help, see the [Enhanced Dashboard User Guide](ENHANCED_DASHBOARD_USER_GUIDE.md) and [API Documentation](ENHANCED_DASHBOARD_API.md).*
