# Enhanced Network Analysis

Pi-hole Network Analyzer now includes comprehensive network analysis capabilities that go beyond basic DNS query analysis to provide deep insights into network traffic patterns, security threats, and performance metrics.

## Overview

The Enhanced Network Analysis feature provides four main analysis components:

1. **Deep Packet Inspection (DPI)** - Analyzes network packets for protocol distribution, traffic topology, and anomalies
2. **Traffic Pattern Analysis** - Identifies bandwidth patterns, temporal trends, and client behavior
3. **Security Analysis** - Detects threats, suspicious activity, and potential security issues
4. **Performance Analysis** - Monitors network performance metrics and quality assessment

## Getting Started

### Command Line Usage

Enable all network analysis features:
```bash
./pihole-analyzer --network-analysis --pihole config.json
```

Enable specific analysis components:
```bash
# Enable only deep packet inspection
./pihole-analyzer --enable-dpi --pihole config.json

# Enable security and performance analysis
./pihole-analyzer --enable-security-analysis --enable-performance-analysis --pihole config.json

# Run with web UI for visualization
./pihole-analyzer --network-analysis --web --pihole config.json
```

### Configuration File

Add network analysis configuration to your config file:

```json
{
  "pihole": {
    "host": "192.168.1.100",
    "api_enabled": true
  },
  "network_analysis": {
    "enabled": true,
    "deep_packet_inspection": {
      "enabled": true,
      "analyze_protocols": ["DNS_UDP", "DNS_TCP"],
      "packet_sampling": 1.0,
      "time_window": "1h"
    },
    "traffic_patterns": {
      "enabled": true,
      "pattern_types": ["bandwidth", "temporal", "client"],
      "analysis_window": "2h",
      "anomaly_detection": true
    },
    "security_analysis": {
      "enabled": true,
      "threat_detection": true,
      "port_scan_detection": true,
      "dns_tunneling_detection": true
    },
    "performance": {
      "enabled": true,
      "latency_analysis": true,
      "bandwidth_analysis": true,
      "quality_thresholds": {
        "max_latency_ms": 150.0,
        "min_bandwidth_mbps": 5.0,
        "max_packet_loss_percent": 2.0
      }
    }
  }
}
```

## Deep Packet Inspection (DPI)

### Features
- **Protocol Analysis**: Identifies and categorizes network protocols (DNS_UDP, DNS_TCP, HTTP, HTTPS)
- **Packet Size Distribution**: Analyzes packet sizes and categorizes them into buckets
- **Traffic Topology**: Maps network traffic flows between source and destination IPs
- **Anomaly Detection**: Detects unusual packet patterns, volume spikes, and suspicious domains
- **DNS Tunneling Detection**: Identifies potential DNS tunneling attempts
- **Port Scanning Detection**: Detects reconnaissance activities

### Use Cases
- Network protocol optimization
- Bandwidth usage analysis
- Security threat identification
- Network troubleshooting

### Example Output
```json
{
  "packet_analysis": {
    "total_packets": 10000,
    "analyzed_packets": 10000,
    "protocol_distribution": {
      "DNS_UDP": 8500,
      "DNS_TCP": 1500
    },
    "top_source_ips": [
      {
        "ip": "192.168.1.100",
        "packet_count": 2500,
        "percentage": 25.0
      }
    ],
    "anomalies": [
      {
        "type": "volume_spike",
        "description": "Query volume spike detected",
        "severity": "MEDIUM",
        "confidence": 0.85
      }
    ]
  }
}
```

## Traffic Pattern Analysis

### Features
- **Bandwidth Patterns**: Tracks bandwidth usage over time with trend analysis
- **Temporal Patterns**: Identifies hourly, daily, and weekly traffic patterns
- **Client Behavior Analysis**: Classifies client behavior and calculates risk scores
- **Anomaly Detection**: Detects unusual traffic patterns and client behaviors
- **Trend Prediction**: Forecasts future traffic trends with confidence intervals

### Client Behavior Types
- **Browser**: Many different domains, few queries each
- **Focused**: Few domains, many queries each
- **Heavy User**: High query volume
- **Light User**: Low query volume
- **Normal**: Typical usage patterns

### Example Output
```json
{
  "traffic_patterns": {
    "bandwidth_patterns": [
      {
        "time_slot": "2024-01-01T10:00:00Z",
        "avg_bandwidth": 5.2,
        "peak_bandwidth": 8.7,
        "trend": "increasing"
      }
    ],
    "client_behavior": {
      "192.168.1.100": {
        "behavior_type": "browser",
        "activity_level": "normal",
        "risk_score": 0.1
      }
    }
  }
}
```

## Security Analysis

### Features
- **Threat Detection**: Identifies malicious domains using blacklists and patterns
- **DGA Detection**: Detects Domain Generation Algorithm patterns used by malware
- **C&C Communication**: Identifies command and control communications through timing analysis
- **Data Exfiltration**: Detects potential data exfiltration via DNS
- **Botnet Activity**: Identifies coordinated botnet activities
- **DNS Anomalies**: Detects unusual DNS query patterns

### Threat Levels
- **LOW**: No significant threats detected
- **MEDIUM**: Some suspicious activity detected
- **HIGH**: Multiple threats detected - attention required
- **CRITICAL**: Critical threats detected - immediate action required

### Example Output
```json
{
  "security_analysis": {
    "threat_level": "MEDIUM",
    "detected_threats": [
      {
        "type": "blacklisted_domain",
        "severity": "HIGH",
        "description": "Access to blacklisted domain: malicious.com",
        "confidence": 0.95
      }
    ],
    "dns_anomalies": [
      {
        "type": "unusual_query_volume",
        "domain": "suspicious-domain.com",
        "confidence": 0.78
      }
    ]
  }
}
```

## Performance Analysis

### Features
- **Latency Analysis**: Measures response times with percentile calculations (P50, P95, P99)
- **Bandwidth Analysis**: Tracks bandwidth utilization over time
- **Throughput Analysis**: Monitors queries per second and response rates
- **Packet Loss Detection**: Identifies packet loss events and patterns
- **Jitter Analysis**: Measures network stability and variance
- **Quality Assessment**: Provides overall network quality grades

### Quality Grades
- **A**: Excellent performance (90-100%)
- **B**: Good performance (80-89%)
- **C**: Fair performance (70-79%)
- **D**: Poor performance (60-69%)
- **F**: Critical performance issues (<60%)

### Example Output
```json
{
  "performance": {
    "overall_score": 85.5,
    "latency_metrics": {
      "avg_latency": 15.2,
      "p95_latency": 45.7,
      "p99_latency": 89.3
    },
    "quality_assessment": {
      "overall_grade": "B",
      "latency_grade": "A",
      "bandwidth_grade": "B"
    }
  }
}
```

## Web Interface Integration

When using the web interface (`--web` flag), the network analysis results are automatically integrated into the dashboard:

- **Traffic Visualization**: Interactive charts showing traffic patterns and anomalies
- **Security Dashboard**: Real-time threat monitoring and alerts
- **Performance Metrics**: Live performance monitoring with quality indicators
- **Network Topology**: Visual representation of network connections and flows

### Access the Web Interface
```bash
./pihole-analyzer --network-analysis --web --pihole config.json
```

Then visit `http://localhost:8080` to view the enhanced analysis dashboard.

## Performance Considerations

### Sampling Rates
For high-traffic networks, adjust the packet sampling rate to improve performance:

```json
{
  "deep_packet_inspection": {
    "packet_sampling": 0.1  // Analyze 10% of packets
  }
}
```

### Memory Usage
The network analysis features require additional memory for:
- Packet analysis buffers
- Pattern detection algorithms
- Historical data storage for trend analysis

### Processing Time
Analysis time scales with:
- Number of DNS records
- Number of network clients
- Enabled analysis components
- Sampling rates

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Reduce packet sampling rate
   - Decrease analysis window size
   - Disable unused analysis components

2. **Slow Performance**
   - Enable packet sampling for DPI
   - Reduce concurrent analysis threads
   - Use shorter time windows

3. **False Positives in Security Analysis**
   - Adjust threat detection thresholds
   - Update blacklist domains
   - Fine-tune anomaly detection sensitivity

### Debug Mode
Enable detailed logging for troubleshooting:
```bash
./pihole-analyzer --network-analysis --pihole config.json --quiet=false
```

## Advanced Configuration

### Custom Blacklists
Add custom domains to security analysis:
```json
{
  "security_analysis": {
    "blacklist_domains": [
      "malicious.com",
      "phishing.net",
      "bad-domain.org"
    ]
  }
}
```

### Performance Thresholds
Customize quality assessment thresholds:
```json
{
  "performance": {
    "quality_thresholds": {
      "max_latency_ms": 100.0,
      "min_bandwidth_mbps": 10.0,
      "max_packet_loss_percent": 1.0,
      "max_jitter_ms": 50.0
    }
  }
}
```

### Protocol Analysis
Configure which protocols to analyze:
```json
{
  "deep_packet_inspection": {
    "analyze_protocols": ["DNS_UDP", "DNS_TCP", "HTTP", "HTTPS"]
  }
}
```

## API Integration

The network analysis results are available through the existing API endpoints when web mode is enabled. Results are returned in JSON format suitable for integration with external monitoring systems.

## Future Enhancements

Planned improvements for the Enhanced Network Analysis feature:

1. **Machine Learning Integration**: Advanced ML-based anomaly detection
2. **Real-time Alerts**: Configurable alerting for security threats
3. **Export Capabilities**: Export analysis results to various formats
4. **Historical Analysis**: Long-term trend analysis and reporting
5. **Custom Dashboards**: Configurable visualization dashboards

For more information and updates, visit the project repository.