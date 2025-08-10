# Alert System Documentation

The Pi-hole Network Analyzer includes a comprehensive alert system that monitors your network for anomalies and sends notifications when issues are detected.

## Overview

The alert system provides:
- **Real-time monitoring** of DNS query patterns and network behavior
- **ML-powered anomaly detection** with automatic alert generation
- **Configurable alert rules** for custom thresholds and conditions
- **Multiple notification channels** including Slack, email, and structured logging
- **Alert management** with history, resolution, and status tracking

## Architecture

The alert system consists of several key components:

### Core Components

- **Alert Manager**: Central orchestrator that manages alerts, rules, and notifications
- **Alert Evaluator**: Processes conditions and evaluates data against configurable rules
- **Storage System**: Handles alert persistence with memory and file-based options
- **Notification Handlers**: Sends alerts through various channels (log, Slack, email)
- **ML Integration**: Automatically creates alerts from ML anomaly detection results

### Data Flow

1. **Data Processing**: Analysis results and ML findings are fed into the alert manager
2. **Rule Evaluation**: Configured rules are evaluated against the incoming data
3. **Alert Generation**: Alerts are created when conditions are met
4. **Notification**: Alerts are sent through configured notification channels
5. **Storage**: Alerts are stored for history and management
6. **Resolution**: Alerts can be resolved manually or automatically

## Configuration

### Basic Configuration

Add the alert configuration to your main config file:

```json
{
  "alerts": {
    "enabled": true,
    "default_severity": "warning",
    "default_cooldown": "5m",
    "max_active_alerts": 100,
    "alert_retention": "7d",
    "storage": {
      "type": "memory",
      "max_size": 1000,
      "retention": "24h"
    },
    "notifications": {
      "slack": {
        "enabled": false,
        "webhook_url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
        "channel": "#alerts",
        "username": "PiHole-Analyzer",
        "icon_emoji": ":warning:",
        "timeout": "30s"
      },
      "email": {
        "enabled": false,
        "smtp_host": "smtp.gmail.com",
        "smtp_port": 587,
        "username": "your-email@gmail.com",
        "password": "your-app-password",
        "from": "pihole-alerts@yourdomain.com",
        "recipients": ["admin@yourdomain.com"],
        "use_tls": true,
        "timeout": "30s"
      }
    }
  }
}
```

### Storage Options

#### Memory Storage (Default)
- Fast, in-memory storage
- No persistence across restarts
- Configurable size limits and retention

```json
"storage": {
  "type": "memory",
  "max_size": 1000,
  "retention": "24h"
}
```

#### File Storage
- Persistent storage to disk
- Survives application restarts
- JSON-based format

```json
"storage": {
  "type": "file",
  "path": "/var/lib/pihole-analyzer/alerts.json",
  "max_size": 5000,
  "retention": "7d"
}
```

## Default Alert Rules

The system comes with predefined alert rules:

### 1. High Anomaly Score Detection
Triggers when ML anomaly score exceeds threshold
- **Condition**: `anomaly_score > 0.8`
- **Severity**: Warning
- **Cooldown**: 5 minutes

### 2. Critical ML Anomalies
Automatic alerts from ML engine for critical anomalies
- **Condition**: `anomaly_severity == "critical"`
- **Severity**: Critical
- **Cooldown**: 2 minutes

### 3. DNS Query Volume Spike
Detects unusual increases in DNS query volume
- **Condition**: `query_count > 1000` in 5-minute window
- **Severity**: Warning
- **Cooldown**: 10 minutes

### 4. Unusual Domain Activity
Identifies suspicious domain patterns
- **Condition**: `anomaly_type == "unusual_domain"`
- **Severity**: Warning
- **Cooldown**: 15 minutes

### 5. Pi-hole Connectivity Issues
Monitors Pi-hole API connectivity
- **Condition**: `connection_status == "disconnected"`
- **Severity**: Error
- **Cooldown**: 1 minute

## Custom Alert Rules

You can define custom alert rules in the configuration:

```json
{
  "alerts": {
    "rules": [
      {
        "id": "custom-high-volume",
        "name": "High Volume Alert",
        "description": "Alert when query volume is very high",
        "enabled": true,
        "type": "threshold",
        "severity": "critical",
        "conditions": [
          {
            "field": "total_queries",
            "operator": "gt",
            "value": 5000,
            "time_window": "1h"
          }
        ],
        "cooldown_period": "30m",
        "channels": ["log", "slack", "email"],
        "tags": ["custom", "volume"]
      }
    ]
  }
}
```

### Alert Rule Fields

- **id**: Unique identifier for the rule
- **name**: Human-readable name
- **description**: Detailed description of what the rule detects
- **enabled**: Whether the rule is active
- **type**: Type of alert (`threshold`, `anomaly`, `security`, etc.)
- **severity**: Alert severity (`info`, `warning`, `error`, `critical`)
- **conditions**: Array of conditions that must be met
- **cooldown_period**: Minimum time between alerts for this rule
- **channels**: Notification channels to use
- **tags**: Tags for organization and filtering

### Condition Operators

- **Comparison**: `gt`, `gte`, `lt`, `lte`, `eq`, `ne`
- **String**: `contains`, `not_contains`, `regex`
- **Array**: `in`, `not_in`

### Available Fields

Common fields available for rule conditions:

- **total_queries**: Total number of DNS queries
- **unique_clients**: Number of unique client IPs
- **anomaly_score**: ML anomaly score (0-1)
- **anomaly_severity**: ML anomaly severity
- **anomaly_type**: Type of ML anomaly detected
- **average_response_time**: Average DNS response time
- **health_score**: Overall network health score
- **connection_status**: Pi-hole connectivity status

## Notification Channels

### Log Notifications

Always enabled, provides structured logging of alerts:

```
time=2025-01-10T15:30:45Z level=WARN msg="🚨 ALERT: High Volume Alert - Query volume exceeded threshold" 
  component=alert-manager alert_id=alert_12345 type=threshold severity=warning 
  source=rule:custom-high-volume client_ip=192.168.1.100
```

### Slack Notifications

Rich formatted messages with alert details:

1. **Configure Slack Webhook**:
   - Go to https://api.slack.com/apps
   - Create a new app and add incoming webhooks
   - Copy the webhook URL to your configuration

2. **Enable in Configuration**:
```json
"slack": {
  "enabled": true,
  "webhook_url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
  "channel": "#alerts",
  "username": "PiHole-Analyzer",
  "icon_emoji": ":warning:"
}
```

### Email Notifications

SMTP-based email alerts with detailed information:

1. **Configure SMTP Settings**:
```json
"email": {
  "enabled": true,
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from": "pihole-alerts@yourdomain.com",
  "recipients": ["admin@yourdomain.com", "ops@yourdomain.com"],
  "use_tls": true
}
```

2. **Gmail Configuration**:
   - Enable 2FA on your Google account
   - Generate an App Password for the analyzer
   - Use the App Password in the configuration

## Alert Management

### Viewing Active Alerts

Query active alerts through the API or logs:

```bash
# Check alert status in logs
grep "Active alerts:" /var/log/pihole-analyzer.log

# View alert history
grep "ALERT:" /var/log/pihole-analyzer.log | tail -20
```

### Resolving Alerts

Alerts can be resolved automatically or manually. The system tracks:
- Alert creation time
- Resolution time
- Alert duration
- Notification status

### Alert History

All alerts are stored according to the retention policy:
- **Memory storage**: Retained until restart or retention period
- **File storage**: Persisted to disk with configurable retention

## Monitoring and Troubleshooting

### Health Monitoring

Monitor the alert system health:

```bash
# Check alert manager status
grep "alert-manager" /var/log/pihole-analyzer.log

# Monitor notification delivery
grep "notification" /var/log/pihole-analyzer.log

# Check rule evaluation
grep "rule" /var/log/pihole-analyzer.log
```

### Common Issues

#### Slack Notifications Not Working
- Verify webhook URL is correct
- Check network connectivity to Slack
- Ensure webhook permissions are set correctly

#### Email Notifications Failing
- Verify SMTP settings (host, port, credentials)
- Check firewall rules for SMTP ports
- Ensure TLS/SSL settings match your provider

#### Too Many Alerts
- Adjust rule thresholds
- Increase cooldown periods
- Review anomaly detection sensitivity

#### Missing Alerts
- Check if rules are enabled
- Verify rule conditions are correctly configured
- Review ML confidence thresholds

### Performance Tuning

For high-volume environments:

```json
{
  "alerts": {
    "performance": {
      "max_concurrent_notifications": 10,
      "notification_timeout": "30s",
      "batch_size": 20,
      "evaluation_interval": "30s"
    },
    "storage": {
      "type": "file",
      "max_size": 10000,
      "retention": "30d"
    }
  }
}
```

## Security Considerations

### Webhook Security
- Use HTTPS webhooks only
- Keep webhook URLs confidential
- Rotate webhook URLs periodically

### Email Security
- Use App Passwords instead of account passwords
- Enable TLS for SMTP connections
- Restrict recipient lists to authorized personnel

### Alert Data
- Be mindful of sensitive data in alert messages
- Configure log rotation for alert logs
- Secure alert storage files appropriately

## Integration Examples

### Integration with Monitoring Systems

The alert system can integrate with external monitoring:

```bash
# Export alerts to Prometheus metrics
curl http://localhost:9090/metrics | grep pihole_alerts

# Send alerts to external systems via webhooks
# Configure custom notification handlers
```

### Automation Scripts

Example script to process alerts:

```bash
#!/bin/bash
# Process new alerts from log file
tail -f /var/log/pihole-analyzer.log | grep "🚨 ALERT:" | while read line; do
  # Extract alert data
  alert_id=$(echo "$line" | grep -o 'alert_id=[^ ]*' | cut -d= -f2)
  severity=$(echo "$line" | grep -o 'severity=[^ ]*' | cut -d= -f2)
  
  # Take action based on severity
  case "$severity" in
    "critical")
      # Send SMS, page on-call engineer
      echo "Critical alert: $alert_id"
      ;;
    "error")
      # Send urgent notification
      echo "Error alert: $alert_id"
      ;;
  esac
done
```

## Best Practices

1. **Start with Default Rules**: Begin with the provided default rules and adjust as needed
2. **Tune Sensitivity**: Adjust ML confidence thresholds to reduce false positives
3. **Use Appropriate Channels**: Match notification channels to alert severity
4. **Monitor Alert Volume**: Track alert frequency to avoid notification fatigue
5. **Regular Review**: Periodically review and update alert rules
6. **Test Notifications**: Regularly test notification channels to ensure they work
7. **Secure Configuration**: Protect configuration files containing credentials
8. **Document Custom Rules**: Maintain documentation for custom alert rules

## API Reference

The alert system exposes management functions through the application API:

- **Get Active Alerts**: Retrieve currently active alerts
- **Get Alert History**: Query historical alert data
- **Update Rules**: Modify alert rules dynamically
- **Test Notifications**: Verify notification channel configuration
- **Alert Status**: Check alert system health and statistics

For detailed API documentation, see the main application API reference.