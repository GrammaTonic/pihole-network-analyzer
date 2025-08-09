# Web UI Documentation

The Pi-hole Network Analyzer now includes a **Web UI Foundation** that provides a browser-based interface for Pi-hole analysis. This feature is designed as preparation for daemon mode and provides real-time access to network analysis data.

## Features

### Core Web Interface
- **Dashboard View**: Clean, responsive HTML interface displaying Pi-hole analysis results
- **Real-time Data**: Auto-refreshing dashboard with live Pi-hole statistics
- **Mobile Responsive**: Works on desktop, tablet, and mobile devices
- **Dark/Light Theme**: Clean styling optimized for readability

### API Endpoints
- **`GET /`**: Main dashboard with complete analysis view
- **`GET /api/status`**: JSON endpoint for connection status
- **`GET /api/analysis`**: JSON endpoint for complete analysis data
- **`GET /api/clients`**: JSON endpoint for client statistics (supports `?limit=N`)
- **`GET /health`**: Health check endpoint for monitoring

### Dashboard Components
- **Connection Status**: Real-time Pi-hole connectivity indicator
- **Summary Statistics**: Total queries, unique clients, active devices
- **Client Table**: Detailed client statistics with sorting and filtering
- **Auto-refresh**: 30-second automatic refresh for live monitoring

## Usage

### Command Line Options

```bash
# Enable web interface
./pihole-analyzer --web --pihole pihole-config.json

# Specify custom port and host
./pihole-analyzer --web --web-port 9090 --web-host 0.0.0.0 --pihole pihole-config.json

# Run in daemon mode (implies --web)
./pihole-analyzer --daemon --pihole pihole-config.json

# Custom configuration
./pihole-analyzer --web --web-port 8888 --web-host localhost --pihole pihole-config.json
```

### Web Flags
- `--web`: Enable web interface (starts HTTP server)
- `--web-port int`: Port for web interface (default: 8080)
- `--web-host string`: Host for web interface (default: localhost)
- `--daemon`: Run in daemon mode (implies --web)

### Configuration File

The web interface settings can be configured in the JSON configuration file:

```json
{
  "web": {
    "enabled": false,
    "port": 8080,
    "host": "localhost",
    "daemon_mode": false,
    "read_timeout_seconds": 10,
    "write_timeout_seconds": 10,
    "idle_timeout_seconds": 60
  }
}
```

## Examples

### Basic Web Interface
```bash
# Start web interface on default port 8080
./pihole-analyzer --web --pihole pi-hole-config.json

# Access at: http://localhost:8080
```

### Custom Port Configuration
```bash
# Start on custom port
./pihole-analyzer --web --web-port 9999 --pihole pi-hole-config.json

# Access at: http://localhost:9999
```

### Daemon Mode for Continuous Monitoring
```bash
# Run in daemon mode (background service)
./pihole-analyzer --daemon --pihole pi-hole-config.json

# Access at: http://localhost:8080
# Runs continuously until stopped with Ctrl+C
```

### Network Access (All Interfaces)
```bash
# Allow access from any network interface
./pihole-analyzer --web --web-host 0.0.0.0 --pihole pi-hole-config.json

# Access from any device: http://[server-ip]:8080
```

## API Usage

### Get Connection Status
```bash
curl http://localhost:8080/api/status
```

Response:
```json
{
  "connected": true,
  "last_connect": "2024-01-15T10:30:00Z",
  "response_time": 45.2,
  "metadata": {
    "data_source_type": "pihole_api",
    "test_duration_ms": "45.20"
  }
}
```

### Get Analysis Data
```bash
curl http://localhost:8080/api/analysis
```

Response:
```json
{
  "client_stats": {
    "192.168.1.100": {
      "ip": "192.168.1.100",
      "hostname": "laptop",
      "query_count": 150,
      "domain_count": 45,
      "is_online": true
    }
  },
  "total_queries": 500,
  "unique_clients": 8,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get Client List
```bash
# Get all clients
curl http://localhost:8080/api/clients

# Get limited results
curl http://localhost:8080/api/clients?limit=10
```

### Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "connected": true,
  "server_info": "Pi-hole Network Analyzer Web UI (Port 8080)"
}
```

## Architecture

### Components
- **Web Server**: HTTP server with structured logging and middleware
- **Data Source Adapter**: Caching layer between web interface and Pi-hole API
- **Template Engine**: HTML templates with embedded CSS for responsive design
- **API Layer**: RESTful endpoints for programmatic access

### Caching
- **Smart Caching**: 30-second cache TTL for analysis results
- **Cache Management**: Automatic refresh and manual cache invalidation
- **Performance**: Optimized for frequent dashboard refreshes

### Logging
All web interactions are logged with structured logging:
- HTTP request/response logging with duration tracking
- Error tracking and debugging information
- Connection status monitoring
- Cache performance metrics

### Security
- **Local Access**: Default configuration only allows localhost access
- **Request Validation**: Input validation on all API endpoints
- **Error Handling**: Graceful error handling with appropriate HTTP status codes
- **Resource Limits**: Configurable timeouts and request limits

## Development

### Testing
The web interface includes comprehensive unit tests:
```bash
# Run web package tests
go test ./internal/web/... -v

# Run all tests including integration
make test
```

### Adding Custom Features
The web interface is designed to be extensible:
- Add new API endpoints in `internal/web/server.go`
- Extend data sources through the `DataSourceProvider` interface
- Customize templates in `internal/web/templates/`
- Add structured logging for new features

### Performance Monitoring
Monitor web interface performance:
- HTTP request duration logging
- Cache hit/miss ratios
- Connection health metrics
- Memory usage tracking

## Troubleshooting

### Common Issues

**Web interface not accessible**
- Check that the port is not in use: `netstat -ln | grep :8080`
- Verify firewall settings allow the specified port
- Ensure Pi-hole configuration is valid

**Slow dashboard loading**
- Check Pi-hole API response times
- Monitor cache effectiveness with logging
- Verify network connectivity to Pi-hole

**Connection errors**
- Verify Pi-hole API password and settings
- Check Pi-hole service status
- Review structured logs for error details

### Debug Mode
Enable verbose logging for troubleshooting:
```bash
./pihole-analyzer --web --pihole config.json --no-quiet
```

### Log Analysis
Web interface logs include:
- HTTP request patterns and response times
- Cache performance and hit rates
- Pi-hole API connectivity status
- Error tracking with stack traces

## Future Enhancements

The Web UI Foundation is designed to support upcoming features:
- **Real-time Dashboards**: WebSocket support for live updates
- **User Authentication**: Multi-user access controls
- **Advanced Filtering**: Complex query and time-based filtering
- **Export Capabilities**: JSON/CSV data export
- **Alert System**: Configurable network monitoring alerts
- **Plugin Architecture**: Extensible analysis modules

This foundation provides the groundwork for a full-featured web-based Pi-hole analysis platform while maintaining the simplicity and performance of the current CLI tool.