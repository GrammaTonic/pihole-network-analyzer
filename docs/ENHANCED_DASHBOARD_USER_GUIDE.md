# Enhanced Dashboard User Guide

## 🎯 Overview

The Enhanced Dashboard transforms the Pi-hole Network Analyzer into a powerful web-based monitoring platform with interactive charts, real-time updates, and comprehensive network insights.

## 🚀 Quick Start

### Starting the Enhanced Dashboard

```bash
# With real Pi-hole data
./pihole-analyzer --web --config your-pihole-config.json

# With demo data (no Pi-hole required)
./pihole-analyzer --web

# Access the enhanced dashboard
open http://localhost:8080/dashboard/enhanced
```

### Key Features

- **📊 Interactive Charts**: Timeline, client distribution, domain statistics, performance metrics, network topology
- **⚡ Real-time Updates**: WebSocket-powered live data refreshing
- **📱 Responsive Design**: Works seamlessly on desktop, tablet, and mobile devices
- **🎨 Professional UI**: Modern design with Chart.js and D3.js visualizations

## 📈 Dashboard Components

### 1. Query Timeline Chart
- **Purpose**: Visualize DNS query patterns over time
- **Technology**: Chart.js line chart with real-time updates
- **Customization**: Adjustable time windows (1h, 6h, 24h, 7d)
- **API Endpoint**: `/api/charts/timeline?hours=24&granularity=1h`

### 2. Client Distribution Chart
- **Purpose**: Show which devices generate the most DNS queries
- **Technology**: Chart.js pie chart with hover interactions
- **Features**: Device identification, query count ranking
- **API Endpoint**: `/api/charts/client-distribution`

### 3. Domain Statistics Chart
- **Purpose**: Display top domains and blocked content
- **Technology**: Chart.js horizontal bar chart
- **Categories**: Top domains, blocked domains, categorized analysis
- **API Endpoint**: `/api/charts/domain-stats?limit=10`

### 4. Performance Metrics Dashboard
- **Purpose**: Monitor Pi-hole and network performance
- **Metrics**: Response time, queries per second, cache hit rate, block percentage
- **Alerts**: Visual indicators for performance issues
- **API Endpoint**: `/api/charts/performance`

### 5. Network Topology Visualization
- **Purpose**: Interactive network topology mapping
- **Technology**: D3.js force-directed graph
- **Features**: Node relationships, traffic flow visualization, interactive exploration
- **API Endpoint**: `/api/charts/network-topology`

## 🔄 Real-time Features

### WebSocket Integration
The enhanced dashboard uses WebSocket connections for live updates:

```javascript
// Automatic connection to WebSocket endpoint
const ws = new WebSocket(`ws://${window.location.host}/ws`);

// Subscribe to dashboard updates
ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'dashboard'
}));
```

### Update Frequency
- **Dashboard Summary**: Every 30 seconds
- **Chart Data**: Every 60 seconds  
- **Performance Metrics**: Every 15 seconds
- **Network Topology**: Every 5 minutes

## 🎛️ Configuration

### Web Server Configuration
```json
{
  "web": {
    "enabled": true,
    "host": "localhost",
    "port": 8080,
    "daemon_mode": true,
    "read_timeout": 30,
    "write_timeout": 30,
    "idle_timeout": 120
  }
}
```

### Chart Customization
The dashboard supports extensive customization through URL parameters:

```bash
# Timeline with custom window
/dashboard/enhanced?timeline_hours=12&timeline_granularity=30m

# Client chart with custom limit
/dashboard/enhanced?client_limit=15

# Domain stats with category filter
/dashboard/enhanced?domain_category=blocked&domain_limit=20
```

## 🛠️ Technical Details

### Browser Requirements
- **Modern Browsers**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- **JavaScript**: ES6+ support required
- **WebSocket**: Native WebSocket support required

### Performance Optimization
- **Lazy Loading**: Charts load progressively for faster initial page load
- **Data Caching**: 30-second client-side cache for API responses
- **Efficient Updates**: Only modified chart data is refreshed
- **Memory Management**: Automatic cleanup of old chart data

### API Rate Limiting
- **Chart APIs**: 60 requests per minute per IP
- **WebSocket**: 1 connection per client
- **Bulk Operations**: Automatic batching for efficiency

## 🔧 Customization Guide

### Adding Custom Charts

1. **Create API Endpoint**:
```go
func (h *ChartAPIHandler) HandleCustomChart(w http.ResponseWriter, r *http.Request) {
    // Your custom chart logic
    data := &ChartData{
        Labels: []string{"Label1", "Label2"},
        Datasets: []ChartDataset{{
            Label: "Custom Data",
            Data: []interface{}{10, 20},
        }},
    }
    json.NewEncoder(w).Encode(data)
}
```

2. **Add Route**:
```go
r.HandleFunc("/api/charts/custom", handler.HandleCustomChart).Methods("GET")
```

3. **Update Dashboard Template**:
```html
<div class="chart-container">
    <canvas id="customChart"></canvas>
</div>
```

### Custom Styling
Override CSS variables for theme customization:
```css
:root {
    --primary-color: #your-color;
    --chart-background: #your-background;
    --text-color: #your-text-color;
}
```

## 🚨 Troubleshooting

### Common Issues

#### Dashboard Not Loading
```bash
# Check web server status
curl http://localhost:8080/health

# Verify configuration
./pihole-analyzer --config your-config.json --validate
```

#### Charts Not Updating
1. **Check WebSocket Connection**: Open browser developer tools → Network tab
2. **Verify API Endpoints**: Test `/api/charts/timeline` directly
3. **Check Logs**: Review application logs for errors

#### Performance Issues
1. **Reduce Update Frequency**: Modify WebSocket update intervals
2. **Limit Chart Data**: Use smaller time windows or data limits
3. **Check System Resources**: Monitor CPU and memory usage

### Debug Mode
Enable debug logging for troubleshooting:
```bash
./pihole-analyzer --web --config config.json --log-level DEBUG
```

## 📱 Mobile Experience

### Responsive Design Features
- **Touch-friendly**: Optimized touch targets for mobile interaction
- **Swipe Navigation**: Swipe between chart views on mobile devices
- **Adaptive Layout**: Charts resize automatically for different screen sizes
- **Offline Indication**: Visual feedback when connection is lost

### Mobile-specific Optimizations
- **Reduced Data**: Smaller datasets for mobile to improve performance
- **Simplified Charts**: Less complex visualizations on small screens
- **Battery Efficiency**: Reduced update frequency on mobile devices

## 🔐 Security Considerations

### Access Control
- **Host Binding**: Configure `web.host` to restrict access
- **Reverse Proxy**: Use nginx/Apache for HTTPS and authentication
- **Firewall**: Limit access to trusted networks only

### Data Privacy
- **Local Processing**: All data remains on your local network
- **No External Calls**: Dashboard works completely offline
- **Encrypted Storage**: Configuration files can be encrypted

## 📊 Analytics and Monitoring

### Built-in Metrics
The enhanced dashboard provides comprehensive metrics:
- **Response Times**: Track dashboard performance
- **User Activity**: Monitor dashboard usage patterns  
- **Data Freshness**: Verify real-time update effectiveness
- **Error Rates**: Track API and WebSocket error rates

### Integration with Monitoring Systems
```bash
# Prometheus metrics endpoint
curl http://localhost:8080/metrics

# Health check endpoint
curl http://localhost:8080/health
```

## 🆕 What's New in Enhanced Dashboard

### vs. Basic Dashboard
| Feature | Basic Dashboard | Enhanced Dashboard |
|---------|----------------|-------------------|
| Chart Types | 2 (table, basic pie) | 5 (timeline, distribution, domains, performance, topology) |
| Real-time Updates | Manual refresh | Automatic WebSocket updates |
| Interactivity | Static | Fully interactive with hover, zoom, click |
| Mobile Support | Limited | Fully responsive |
| API Access | None | Complete REST API |
| Customization | Minimal | Extensive configuration options |

### Migration from Basic Dashboard
No migration required! The enhanced dashboard is accessible at `/dashboard/enhanced` while the basic dashboard remains at `/dashboard`.

## 🎓 Best Practices

### Deployment
1. **Use Daemon Mode**: Enable `daemon_mode` for production deployments
2. **Configure Timeouts**: Set appropriate timeout values for your environment
3. **Monitor Resources**: Track memory and CPU usage in production
4. **Enable Logging**: Use structured logging for troubleshooting

### Performance
1. **Limit Historical Data**: Use reasonable time windows for charts
2. **Cache Configuration**: Enable caching for static assets
3. **Network Optimization**: Use compression and efficient API design
4. **Regular Maintenance**: Restart services periodically to clear caches

---

## 🔗 Related Documentation

- [Installation Guide](installation.md)
- [Configuration Reference](configuration.md) 
- [API Documentation](api.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Development Guide](development.md)

---

*For additional support, please refer to the [troubleshooting guide](troubleshooting.md) or open an issue on GitHub.*
