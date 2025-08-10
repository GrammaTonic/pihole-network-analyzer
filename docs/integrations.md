# Integration Ecosystem Documentation

The Pi-hole Network Analyzer provides comprehensive integration capabilities with popular monitoring platforms including Grafana, Prometheus, and Loki. This document provides setup guides, configuration examples, and troubleshooting information for the complete monitoring ecosystem.

## Overview

The integration ecosystem enables you to:

- **📊 Visualize** Pi-hole data through rich Grafana dashboards
- **📈 Monitor** network metrics with Prometheus 
- **📝 Aggregate** structured logs with Loki
- **🔔 Alert** on network anomalies and issues
- **🔧 Automate** dashboard and alert provisioning

## Quick Start

### 1. Enable Integrations

```json
{
  "integrations": {
    "enabled": true,
    "grafana": {
      "enabled": true,
      "url": "http://grafana:3000",
      "api_key": "your-grafana-api-key"
    },
    "loki": {
      "enabled": true,
      "url": "http://loki:3100"
    },
    "prometheus": {
      "enabled": true,
      "push_gateway": {
        "enabled": true,
        "url": "http://prometheus-pushgateway:9091"
      }
    }
  }
}
```

### 2. Run with Integrations

```bash
# Start with integrations enabled
./pihole-analyzer --config config-with-integrations.json --pihole pihole-config.json

# Run in daemon mode for continuous monitoring
./pihole-analyzer --daemon --config config-with-integrations.json
```

## Grafana Integration

### Configuration

```json
{
  "integrations": {
    "grafana": {
      "enabled": true,
      "url": "http://grafana:3000",
      "api_key": "glsa_...",
      "organization": "Main Org.",
      
      "data_source": {
        "create_if_not_exists": true,
        "name": "pihole-prometheus",
        "type": "prometheus",
        "url": "http://prometheus:9090",
        "access": "proxy"
      },
      
      "dashboards": {
        "auto_provision": true,
        "folder_name": "Pi-hole",
        "overwrite_existing": true,
        "tags": ["pihole", "dns", "network"]
      },
      
      "alerts": {
        "enabled": true,
        "notification_channels": ["slack", "email"],
        "default_severity": "warning"
      },
      
      "timeout_seconds": 30,
      "verify_tls": true,
      "retry_count": 3
    }
  }
}
```

### Features

- **🎯 Automatic Data Source Setup**: Creates Prometheus data source if it doesn't exist
- **📊 Dashboard Provisioning**: Auto-creates comprehensive Pi-hole dashboard
- **🔔 Alert Management**: Configurable alerts for network anomalies
- **📁 Folder Organization**: Organizes dashboards in dedicated folders

### Dashboard Panels

The auto-provisioned dashboard includes:

1. **DNS Queries Overview** - Total queries, blocked queries, query types
2. **Client Activity** - Active clients, top clients by query count
3. **Domain Statistics** - Top queried domains, blocked domains
4. **Response Times** - DNS response time distribution and trends
5. **Network Health** - Query success rate, error rates
6. **Trends Analysis** - Historical patterns and anomaly detection

### API Key Setup

1. **Create API Key in Grafana**:
   - Go to Configuration → API Keys
   - Create new key with Admin or Editor role
   - Copy the generated key

2. **Configure in Pi-hole Analyzer**:
   ```json
   {
     "grafana": {
       "api_key": "glsa_your_api_key_here"
     }
   }
   ```

## Prometheus Integration

### Configuration

```json
{
  "integrations": {
    "prometheus": {
      "enabled": true,
      
      "push_gateway": {
        "enabled": true,
        "url": "http://prometheus-pushgateway:9091",
        "job_name": "pihole-analyzer",
        "username": "",
        "password": "",
        "push_interval": "30s"
      },
      
      "external_labels": {
        "service": "pihole-analyzer",
        "instance": "home-network",
        "environment": "production"
      },
      
      "custom_metrics": {
        "enabled": true,
        "prefix": "pihole_"
      }
    }
  }
}
```

### Metrics Exported

#### DNS Query Metrics
- `pihole_queries_total` - Total DNS queries processed
- `pihole_queries_blocked_total` - Total blocked queries
- `pihole_queries_by_type` - Queries by type (A, AAAA, PTR, etc.)
- `pihole_response_time_seconds` - DNS response time distribution

#### Client Metrics
- `pihole_clients_active` - Number of active clients
- `pihole_clients_total` - Total unique clients seen
- `pihole_client_queries_total` - Queries per client

#### Domain Metrics
- `pihole_domains_queried_total` - Unique domains queried
- `pihole_top_domains` - Most queried domains
- `pihole_blocked_domains_total` - Blocked domain attempts

#### System Metrics
- `pihole_analysis_duration_seconds` - Time taken for analysis
- `pihole_analysis_timestamp` - Last analysis timestamp

### Push Gateway Setup

1. **Deploy Prometheus Push Gateway**:
   ```bash
   docker run -d -p 9091:9091 prom/pushgateway
   ```

2. **Configure Prometheus** to scrape Push Gateway:
   ```yaml
   scrape_configs:
     - job_name: 'pushgateway'
       static_configs:
         - targets: ['pushgateway:9091']
   ```

## Loki Integration

### Configuration

```json
{
  "integrations": {
    "loki": {
      "enabled": true,
      "url": "http://loki:3100",
      "username": "",
      "password": "",
      "tenant_id": "",
      
      "batch_size": 100,
      "batch_timeout": "10s",
      "max_retries": 3,
      "retry_delay": "1s",
      
      "labels": {
        "job": "pihole-analyzer",
        "service": "dns-analysis",
        "environment": "production"
      },
      
      "dynamic_labels": {
        "level": true,
        "component": true,
        "client_ip": false,
        "domain": false
      }
    }
  }
}
```

### Log Types

#### Analysis Logs
- **Query Analysis**: Detailed breakdown of DNS queries
- **Client Analysis**: Client behavior and patterns
- **Domain Analysis**: Domain statistics and trends
- **Performance**: Analysis timing and resource usage

#### Alert Logs
- **Anomalies**: Unusual network patterns detected
- **Thresholds**: Metrics exceeding configured limits
- **Errors**: Analysis errors and warnings
- **Health**: System health status

#### Integration Logs
- **Grafana**: Dashboard provisioning and API calls
- **Prometheus**: Metric pushing and registration
- **Loki**: Log forwarding and batching

### Log Format

Logs are structured in JSON format with consistent fields:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Analysis completed successfully",
  "component": "analyzer",
  "labels": {
    "queries_processed": "5432",
    "clients_found": "18",
    "analysis_duration": "12.5s"
  }
}
```

## Complete Setup Guide

### 1. Monitoring Stack Deployment

#### Docker Compose Example

```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    
  pushgateway:
    image: prom/pushgateway:latest
    ports:
      - "9091:9091"
    
  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki-data:/loki
    
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana

volumes:
  prometheus-data:
  loki-data:
  grafana-data:
```

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'pushgateway'
    static_configs:
      - targets: ['pushgateway:9091']
    scrape_interval: 30s
    metrics_path: /metrics
```

### 2. Pi-hole Analyzer Configuration

```json
{
  "pihole": {
    "host": "192.168.1.100",
    "port": 80,
    "api_key": "your-pihole-api-key",
    "use_tls": false
  },
  
  "integrations": {
    "enabled": true,
    
    "grafana": {
      "enabled": true,
      "url": "http://localhost:3000",
      "api_key": "glsa_your_grafana_api_key",
      "data_source": {
        "create_if_not_exists": true,
        "name": "pihole-prometheus",
        "type": "prometheus",
        "url": "http://localhost:9090"
      },
      "dashboards": {
        "auto_provision": true,
        "folder_name": "Pi-hole Network Analysis"
      }
    },
    
    "loki": {
      "enabled": true,
      "url": "http://localhost:3100",
      "batch_size": 100,
      "batch_timeout": "10s",
      "labels": {
        "job": "pihole-analyzer",
        "instance": "home-network"
      }
    },
    
    "prometheus": {
      "enabled": true,
      "push_gateway": {
        "enabled": true,
        "url": "http://localhost:9091"
      },
      "external_labels": {
        "service": "pihole-analyzer",
        "environment": "home"
      }
    }
  },
  
  "analysis": {
    "mode": "comprehensive",
    "exclude_networks": ["127.0.0.0/8", "169.254.0.0/16"]
  },
  
  "logging": {
    "level": "info",
    "enable_colors": true,
    "enable_emojis": true
  }
}
```

### 3. Running the Integration

```bash
# Start monitoring stack
docker-compose up -d

# Wait for services to be ready
sleep 30

# Create Grafana API key (manual step)
# - Access Grafana at http://localhost:3000 (admin/admin)
# - Go to Configuration → API Keys
# - Create new key with Editor role
# - Copy the key to your config

# Run Pi-hole analyzer with integrations
./pihole-analyzer --config integration-config.json --pihole pihole-config.json

# Or run in daemon mode for continuous monitoring
./pihole-analyzer --daemon --config integration-config.json
```

### 4. Verification

#### Check Prometheus Metrics
```bash
# List all pihole metrics
curl http://localhost:9090/api/v1/label/__name__/values | jq '.data[] | select(startswith("pihole_"))'

# Query specific metric
curl "http://localhost:9090/api/v1/query?query=pihole_queries_total"
```

#### Check Loki Logs
```bash
# Query recent logs
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={job="pihole-analyzer"}' \
  --data-urlencode 'start=1h' | jq '.data.result'
```

#### Check Grafana Dashboard
- Access Grafana at http://localhost:3000
- Navigate to Dashboards → Pi-hole Network Analysis
- Verify data is being populated

## Troubleshooting

### Common Issues

#### 1. Grafana API Key Issues
```
Error: failed to connect to Grafana: unauthorized
```

**Solution**:
- Verify API key is correct and not expired
- Ensure API key has sufficient permissions (Editor or Admin)
- Check Grafana URL accessibility

#### 2. Prometheus Push Gateway Issues
```
Error: failed to push metrics to Prometheus: connection refused
```

**Solution**:
- Verify Push Gateway is running and accessible
- Check URL and port configuration
- Ensure no firewall blocking connections

#### 3. Loki Connection Issues
```
Error: failed to send logs to Loki: 500 Internal Server Error
```

**Solution**:
- Check Loki service status and logs
- Verify Loki URL and configuration
- Check log batch size and timeout settings

#### 4. Dashboard Not Auto-Creating
```
Warning: dashboard auto-provisioning failed
```

**Solution**:
- Verify Grafana API key has dashboard creation permissions
- Check folder permissions in Grafana
- Review dashboard JSON format and validation

### Health Checks

```bash
# Check integration health
./pihole-analyzer --config integration-config.json --show-config

# Test connections only
curl -X POST http://localhost:8080/api/integrations/test

# View integration status
curl http://localhost:8080/api/integrations/status
```

### Debugging

Enable debug logging for detailed troubleshooting:

```json
{
  "logging": {
    "level": "debug",
    "enable_colors": true,
    "enable_emojis": true
  }
}
```

## Advanced Configuration

### Custom Metrics

Add custom metrics for specific use cases:

```json
{
  "prometheus": {
    "custom_metrics": {
      "enabled": true,
      "metrics": [
        {
          "name": "pihole_custom_blocked_ratio",
          "type": "gauge",
          "help": "Custom blocked query ratio",
          "query": "blocked_queries / total_queries"
        }
      ]
    }
  }
}
```

### Alert Rules

Configure custom Grafana alerts:

```json
{
  "grafana": {
    "alerts": {
      "enabled": true,
      "rules": [
        {
          "name": "High Block Rate",
          "query": "pihole_queries_blocked_total / pihole_queries_total",
          "condition": "> 0.8",
          "duration": "5m",
          "severity": "warning"
        }
      ]
    }
  }
}
```

### Multi-Tenant Loki

For organizations using multi-tenant Loki:

```json
{
  "loki": {
    "tenant_id": "pihole-team",
    "username": "tenant-user",
    "password": "tenant-password"
  }
}
```

## Best Practices

### 1. Security
- Use HTTPS for all external connections
- Rotate API keys regularly
- Implement proper authentication for monitoring services
- Use environment variables for sensitive configuration

### 2. Performance
- Adjust batch sizes based on log volume
- Monitor Push Gateway memory usage
- Use appropriate retention policies for metrics and logs
- Consider data aggregation for high-volume environments

### 3. Monitoring
- Set up alerts for integration failures
- Monitor integration service health
- Track metric delivery success rates
- Log integration errors for debugging

### 4. Maintenance
- Regularly update monitoring stack components
- Clean up old dashboards and unused metrics
- Review and optimize queries and alerts
- Backup Grafana dashboards and configurations

## Migration Guide

### From Manual Setup

If you're currently using manual Grafana/Prometheus setup:

1. **Export existing dashboards** from Grafana
2. **Configure integrations** in Pi-hole Analyzer
3. **Enable auto-provisioning** to recreate dashboards
4. **Verify data continuity** after migration
5. **Update alert rules** to use new metric names

### Version Compatibility

| Component | Minimum Version | Recommended | Notes |
|-----------|----------------|-------------|-------|
| Grafana | 8.0 | Latest | API v4+ required |
| Prometheus | 2.20 | Latest | Push Gateway support |
| Loki | 2.0 | Latest | LogQL v2 features |
| Docker | 20.10 | Latest | Compose v3.8+ |

## Support

For integration-specific issues:

1. **Check logs** with debug level enabled
2. **Test connections** using health check endpoints
3. **Verify configurations** against examples in this guide
4. **Review service logs** for the monitoring stack components
5. **Open issues** on GitHub with detailed error messages and configurations

## Examples Repository

Complete working examples are available in the `examples/integrations/` directory:

- `docker-compose.monitoring.yml` - Complete monitoring stack
- `config.integration.json` - Full integration configuration
- `grafana-dashboards/` - Dashboard JSON exports
- `prometheus-rules/` - Alert rule examples
- `loki-queries/` - Common LogQL queries

---

**Happy Monitoring!** 🎉📊

The integration ecosystem provides powerful insights into your network's DNS activity. With proper setup, you'll have comprehensive visibility into Pi-hole performance, network patterns, and potential security issues.