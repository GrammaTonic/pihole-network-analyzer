# API Documentation - Enhanced Dashboard Endpoints

## Overview

The Enhanced Dashboard introduces comprehensive REST API endpoints for chart data and real-time monitoring. All endpoints return JSON data optimized for Chart.js and D3.js consumption.

## Base URL

```
http://localhost:8080/api/charts/
```

## Authentication

Currently no authentication is required. For production deployments, consider implementing:
- Reverse proxy with authentication (nginx, Apache)
- API key validation
- Network-level access control

## Common Parameters

### Query Parameters
- `limit` (integer): Maximum number of items to return (default: 10, max: 100)
- `hours` (integer): Time window in hours (default: 24, max: 168)
- `granularity` (string): Time granularity for timeline data (`1m`, `5m`, `15m`, `1h`, `6h`, `24h`)

### Response Format
All endpoints return JSON in the following structure:
```json
{
  "labels": ["Label1", "Label2"],
  "datasets": [{
    "label": "Dataset Name",
    "data": [10, 20, 30],
    "backgroundColor": ["#color1", "#color2"],
    "borderColor": ["#color1", "#color2"]
  }],
  "options": {},
  "metadata": {}
}
```

## Endpoints

### 1. Timeline Chart Data

**Endpoint**: `GET /api/charts/timeline`

**Description**: Returns time-series data for DNS query patterns over time.

**Parameters**:
- `hours` (optional): Time window in hours (default: 24)
- `granularity` (optional): Time bucket size (default: "1h")

**Example Request**:
```bash
curl "http://localhost:8080/api/charts/timeline?hours=12&granularity=30m"
```

**Example Response**:
```json
{
  "labels": ["10:00", "10:30", "11:00", "11:30"],
  "datasets": [{
    "label": "DNS Queries",
    "data": [45, 52, 38, 61],
    "borderColor": "#3b82f6",
    "backgroundColor": "rgba(59, 130, 246, 0.1)",
    "fill": true,
    "tension": 0.4
  }],
  "options": {
    "scales": {
      "y": {
        "beginAtZero": true,
        "title": { "display": true, "text": "Queries" }
      }
    }
  },
  "metadata": {
    "total_queries": 196,
    "time_window": "12h",
    "granularity": "30m"
  }
}
```

### 2. Client Distribution Chart

**Endpoint**: `GET /api/charts/client-distribution`

**Description**: Returns client device query distribution data for pie chart visualization.

**Parameters**:
- `limit` (optional): Maximum number of clients to include (default: 10)

**Example Request**:
```bash
curl "http://localhost:8080/api/charts/client-distribution?limit=5"
```

**Example Response**:
```json
{
  "labels": ["Desktop-001", "Phone-Android", "Laptop-Mac", "IoT-Device", "Tablet-iPad"],
  "datasets": [{
    "label": "Query Distribution",
    "data": [45, 23, 18, 10, 4],
    "backgroundColor": [
      "#ef4444", "#f97316", "#eab308", "#22c55e", "#3b82f6"
    ],
    "borderWidth": 2
  }],
  "options": {
    "responsive": true,
    "plugins": {
      "legend": { "position": "bottom" },
      "tooltip": {
        "callbacks": {
          "label": "function(context) { return context.label + ': ' + context.parsed + '%'; }"
        }
      }
    }
  },
  "metadata": {
    "total_clients": 5,
    "total_queries": 100,
    "percentage_covered": 100
  }
}
```

### 3. Domain Statistics Chart

**Endpoint**: `GET /api/charts/domain-stats`

**Description**: Returns domain query statistics for horizontal bar chart visualization.

**Parameters**:
- `limit` (optional): Maximum number of domains (default: 10)
- `category` (optional): Filter by category ("all", "allowed", "blocked")

**Example Request**:
```bash
curl "http://localhost:8080/api/charts/domain-stats?limit=8&category=all"
```

**Example Response**:
```json
{
  "labels": ["google.com", "youtube.com", "github.com", "stackoverflow.com", "reddit.com"],
  "datasets": [{
    "label": "Top Domains",
    "data": [145, 89, 76, 43, 32],
    "backgroundColor": "#22c55e",
    "borderColor": "#16a34a",
    "borderWidth": 1
  }, {
    "label": "Blocked Domains", 
    "data": [0, 0, 0, 0, 0],
    "backgroundColor": "#ef4444",
    "borderColor": "#dc2626",
    "borderWidth": 1
  }],
  "options": {
    "indexAxis": "y",
    "responsive": true,
    "scales": {
      "x": {
        "beginAtZero": true,
        "title": { "display": true, "text": "Query Count" }
      }
    }
  },
  "metadata": {
    "top_domains_count": 5,
    "blocked_domains_count": 3,
    "total_queries": 385
  }
}
```

### 4. Performance Metrics

**Endpoint**: `GET /api/charts/performance`

**Description**: Returns performance metrics for gauge and metric visualization.

**Example Request**:
```bash
curl "http://localhost:8080/api/charts/performance"
```

**Example Response**:
```json
{
  "labels": ["Response Time", "Queries/Sec", "Cache Hit Rate", "Block Rate"],
  "datasets": [{
    "label": "Performance Metrics",
    "data": [45.6, 12.4, 78.3, 15.2],
    "backgroundColor": ["#3b82f6", "#10b981", "#f59e0b", "#ef4444"],
    "borderWidth": 0
  }],
  "options": {
    "plugins": {
      "legend": { "display": false },
      "tooltip": {
        "callbacks": {
          "label": "function(context) { return context.label + ': ' + context.parsed + (context.dataIndex < 2 ? (context.dataIndex === 0 ? 'ms' : '/s') : '%'); }"
        }
      }
    }
  },
  "metadata": {
    "average_response_time": 45.6,
    "queries_per_second": 12.4,
    "cache_hit_rate": 78.3,
    "blocked_percentage": 15.2,
    "last_updated": "2025-08-10T20:24:30Z"
  }
}
```

### 5. Network Topology Data

**Endpoint**: `GET /api/charts/network-topology`

**Description**: Returns network topology data for D3.js force-directed graph visualization.

**Example Request**:
```bash
curl "http://localhost:8080/api/charts/network-topology"
```

**Example Response**:
```json
{
  "nodes": [
    {
      "id": "pihole",
      "label": "Pi-hole Server", 
      "type": "server",
      "size": 30,
      "color": "#3b82f6"
    },
    {
      "id": "router",
      "label": "Router",
      "type": "router", 
      "size": 25,
      "color": "#10b981"
    },
    {
      "id": "desktop1",
      "label": "Desktop-001",
      "type": "client",
      "size": 20,
      "color": "#f59e0b"
    }
  ],
  "links": [
    {
      "source": "router",
      "target": "pihole",
      "value": 10,
      "type": "dns"
    },
    {
      "source": "router", 
      "target": "desktop1",
      "value": 8,
      "type": "network"
    }
  ],
  "metadata": {
    "node_count": 3,
    "link_count": 2,
    "last_updated": "2025-08-10T20:24:30Z"
  }
}
```

## WebSocket API

### Connection

**Endpoint**: `ws://localhost:8080/ws`

**Description**: WebSocket endpoint for real-time dashboard updates.

### Message Format

**Client → Server (Subscribe)**:
```json
{
  "type": "subscribe",
  "channel": "dashboard"
}
```

**Server → Client (Update)**:
```json
{
  "type": "update",
  "channel": "dashboard", 
  "timestamp": "2025-08-10T20:24:30Z",
  "data": {
    "total_queries": 1500,
    "active_clients": 8,
    "blocked_queries": 150,
    "response_time": 42.3
  }
}
```

### Available Channels
- `dashboard`: General dashboard updates
- `timeline`: Timeline chart updates
- `clients`: Client distribution updates
- `domains`: Domain statistics updates
- `performance`: Performance metrics updates

## Error Handling

### HTTP Status Codes
- `200 OK`: Success
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Endpoint not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Error Response Format
```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Parameter 'hours' must be between 1 and 168",
    "details": {
      "parameter": "hours",
      "provided": "200",
      "max_allowed": "168"
    }
  }
}
```

## Rate Limiting

### Limits
- **Chart APIs**: 60 requests per minute per IP address
- **WebSocket**: 1 concurrent connection per client
- **Bulk Operations**: Automatic request batching

### Rate Limit Headers
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1691686170
```

## Caching

### Client-Side Caching
- **Timeline Data**: 30 seconds
- **Client Distribution**: 60 seconds  
- **Domain Stats**: 60 seconds
- **Performance Metrics**: 15 seconds
- **Network Topology**: 300 seconds

### Cache Headers
```
Cache-Control: public, max-age=30
ETag: "abc123def456"
Last-Modified: Sat, 10 Aug 2025 18:24:30 GMT
```

## Development

### Adding New Endpoints

1. **Define Handler Function**:
```go
func (h *ChartAPIHandler) HandleNewChart(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

2. **Register Route**:
```go
r.HandleFunc("/api/charts/new-chart", handler.HandleNewChart).Methods("GET")
```

3. **Add Tests**:
```go
func TestChartAPIHandler_HandleNewChart(t *testing.T) {
    // Test implementation
}
```

### Testing

```bash
# Run API tests
go test ./internal/web -run TestChartAPIHandler -v

# Test with curl
curl -v "http://localhost:8080/api/charts/timeline"

# Load testing
ab -n 1000 -c 10 "http://localhost:8080/api/charts/timeline"
```

## Examples

### JavaScript Integration

```javascript
// Fetch timeline data
async function fetchTimelineData() {
    const response = await fetch('/api/charts/timeline?hours=6');
    const data = await response.json();
    
    // Use with Chart.js
    new Chart(ctx, {
        type: 'line',
        data: data,
        options: data.options
    });
}

// WebSocket real-time updates
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = () => {
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'dashboard'
    }));
};

ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    updateDashboard(update.data);
};
```

### Python Integration

```python
import requests
import json

# Fetch performance metrics
response = requests.get('http://localhost:8080/api/charts/performance')
data = response.json()

print(f"Response Time: {data['metadata']['average_response_time']}ms")
print(f"Cache Hit Rate: {data['metadata']['cache_hit_rate']}%")
```

### curl Examples

```bash
# Get last 6 hours of timeline data with 15-minute granularity
curl "http://localhost:8080/api/charts/timeline?hours=6&granularity=15m"

# Get top 5 clients
curl "http://localhost:8080/api/charts/client-distribution?limit=5"

# Get only blocked domains
curl "http://localhost:8080/api/charts/domain-stats?category=blocked&limit=15"

# Get current performance metrics
curl "http://localhost:8080/api/charts/performance"

# Get network topology
curl "http://localhost:8080/api/charts/network-topology"
```

---

*For more information, see the [Enhanced Dashboard User Guide](ENHANCED_DASHBOARD_USER_GUIDE.md).*
