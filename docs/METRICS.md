# Pi-hole Network Analyzer - Prometheus Metrics

This document describes the Prometheus metrics endpoints and monitoring capabilities of the Pi-hole Network Analyzer.

## Overview

The Pi-hole Network Analyzer includes built-in Prometheus metrics collection and HTTP endpoints for monitoring. This provides comprehensive insights into DNS analysis performance, Pi-hole data patterns, and system health.

## Configuration

Metrics functionality is configured in the application's JSON configuration file under the `metrics` section:

```json
{
  "metrics": {
    "enabled": true,
    "port": "9090",
    "host": "localhost",
    "enable_endpoint": true,
    "collect_metrics": true,
    "enable_detailed_metrics": true
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable/disable the entire metrics system |
| `port` | string | `"9090"` | HTTP port for the metrics server |
| `host` | string | `"localhost"` | Host interface to bind the metrics server |
| `enable_endpoint` | boolean | `true` | Whether to start the HTTP metrics endpoint |
| `collect_metrics` | boolean | `true` | Whether to collect metrics during analysis |
| `enable_detailed_metrics` | boolean | `true` | Enable collection of detailed domain and client metrics |

## HTTP Endpoints

When the metrics endpoint is enabled, the following HTTP endpoints are available:

### `/metrics`
- **Purpose**: Prometheus metrics endpoint
- **Format**: Prometheus text exposition format
- **Method**: GET
- **Usage**: Configure Prometheus to scrape this endpoint

### `/health`
- **Purpose**: Health check endpoint
- **Format**: JSON
- **Method**: GET
- **Response**: `{"status":"healthy","component":"metrics-server"}`

### `/`
- **Purpose**: Information page with available endpoints
- **Format**: HTML
- **Method**: GET
- **Usage**: Human-readable endpoint listing

## Available Metrics

### Query Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `pihole_analyzer_total_queries` | Counter | Total number of DNS queries processed | - |
| `pihole_analyzer_queries_per_second` | Gauge | Current rate of queries per second | - |
| `pihole_analyzer_query_response_time_seconds` | Histogram | Time taken to process DNS queries | - |
| `pihole_analyzer_queries_by_type_total` | Counter | Queries by DNS record type | `query_type` |
| `pihole_analyzer_queries_by_status_total` | Counter | Queries by response status | `status` |

### Client Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `pihole_analyzer_active_clients` | Gauge | Number of active clients in the network | - |
| `pihole_analyzer_unique_clients` | Gauge | Number of unique clients in current analysis | - |
| `pihole_analyzer_top_clients_queries` | Gauge | Query count for top clients | `client_ip`, `hostname` |

### Domain Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `pihole_analyzer_top_domains_total` | Counter | Total queries for top domains | `domain` |
| `pihole_analyzer_blocked_domains_total` | Counter | Total blocked domain queries | - |
| `pihole_analyzer_allowed_domains_total` | Counter | Total allowed domain queries | - |

### Performance Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `pihole_analyzer_analysis_process_time_seconds` | Histogram | Time taken for complete analysis processing | - |
| `pihole_analyzer_api_call_time_seconds` | Histogram | Time taken for Pi-hole API calls | - |
| `pihole_analyzer_errors_total` | Counter | Total errors by error type | `error_type` |

### System Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `pihole_analyzer_last_analysis_timestamp` | Gauge | Unix timestamp of last successful analysis | - |
| `pihole_analyzer_analysis_duration_seconds` | Histogram | Duration of complete analysis process | - |
| `pihole_analyzer_data_source_health` | Gauge | Health status of data source (1=healthy, 0=unhealthy) | - |

## Prometheus Configuration

To scrape metrics from the Pi-hole Network Analyzer, add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'pihole-analyzer'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s
    metrics_path: /metrics
```

## Example Queries

### Basic Monitoring

```promql
# Total queries processed
pihole_analyzer_total_queries

# Active clients
pihole_analyzer_active_clients

# Data source health
pihole_analyzer_data_source_health

# Query rate
rate(pihole_analyzer_total_queries[5m])
```

### Performance Monitoring

```promql
# Average analysis duration
rate(pihole_analyzer_analysis_duration_seconds_sum[5m]) / 
rate(pihole_analyzer_analysis_duration_seconds_count[5m])

# API call latency percentiles
histogram_quantile(0.95, rate(pihole_analyzer_api_call_time_seconds_bucket[5m]))
histogram_quantile(0.50, rate(pihole_analyzer_api_call_time_seconds_bucket[5m]))

# Error rate
rate(pihole_analyzer_errors_total[5m])
```

### DNS Analysis

```promql
# Top query types
topk(5, increase(pihole_analyzer_queries_by_type_total[1h]))

# Block rate percentage
(
  rate(pihole_analyzer_blocked_domains_total[5m]) / 
  (rate(pihole_analyzer_blocked_domains_total[5m]) + rate(pihole_analyzer_allowed_domains_total[5m]))
) * 100

# Top domains by query count
topk(10, increase(pihole_analyzer_top_domains_total[1h]))
```

## Grafana Dashboard

### Recommended Panels

1. **Overview**
   - Total Queries (Single Stat)
   - Active Clients (Single Stat)
   - Data Source Health (Single Stat)
   - Block Rate Percentage (Single Stat)

2. **Performance**
   - Analysis Duration (Histogram panel)
   - API Call Latency (Histogram panel)
   - Error Rate (Graph panel)
   - Queries Per Second (Graph panel)

3. **DNS Analysis**
   - Top Query Types (Pie chart)
   - Top Domains (Table)
   - Top Clients (Table)
   - Blocked vs Allowed (Graph panel)

### Sample Dashboard JSON

A complete Grafana dashboard configuration is available in the `docs/grafana/` directory.

## Alerting Rules

### Recommended Prometheus Alerts

```yaml
groups:
  - name: pihole-analyzer
    rules:
      - alert: PiholeAnalyzerDown
        expr: up{job="pihole-analyzer"} == 0
        for: 2m
        annotations:
          summary: "Pi-hole Analyzer is down"
          
      - alert: DataSourceUnhealthy
        expr: pihole_analyzer_data_source_health == 0
        for: 1m
        annotations:
          summary: "Pi-hole data source is unhealthy"
          
      - alert: HighErrorRate
        expr: rate(pihole_analyzer_errors_total[5m]) > 0.1
        for: 2m
        annotations:
          summary: "High error rate in Pi-hole Analyzer"
          
      - alert: SlowAnalysis
        expr: histogram_quantile(0.95, rate(pihole_analyzer_analysis_duration_seconds_bucket[5m])) > 60
        for: 5m
        annotations:
          summary: "Pi-hole analysis taking too long"
```

## Security Considerations

- The metrics endpoint runs on HTTP by default
- Consider using a reverse proxy with HTTPS for production
- Limit access to the metrics port using firewall rules
- The endpoint does not require authentication by default

## Troubleshooting

### Common Issues

1. **Metrics endpoint not accessible**
   - Check `enable_endpoint` is `true` in configuration
   - Verify port is not blocked by firewall
   - Ensure application has permission to bind to the port

2. **No metrics data**
   - Check `collect_metrics` is `true` in configuration
   - Verify analysis is running successfully
   - Check application logs for errors

3. **High memory usage**
   - Consider disabling `enable_detailed_metrics` for high-volume environments
   - Reduce metrics retention in Prometheus
   - Monitor the number of unique label values

### Debug Commands

```bash
# Test metrics endpoint
curl http://localhost:9090/metrics

# Test health endpoint
curl http://localhost:9090/health

# Check specific metric
curl -s http://localhost:9090/metrics | grep pihole_analyzer_total_queries
```

## Development

### Adding New Metrics

1. Define the metric in `internal/metrics/metrics.go`
2. Add the metric to the collector struct
3. Register the metric in the `New()` function
4. Create methods to update the metric
5. Call the methods from the appropriate places in the application
6. Add tests in `internal/metrics/metrics_test.go`
7. Update this documentation

### Testing

```bash
# Run unit tests
go test ./internal/metrics

# Run integration tests
go test ./tests/integration

# Run all tests
go test ./...
```