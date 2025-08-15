# DHCP Server Feature

The Pi-hole Network Analyzer now includes a comprehensive DHCP server with dynamic IP allocation and lease management capabilities.

## Features

### Core DHCP Functionality
- **Dynamic IP Allocation**: Automatic IP address assignment from configurable pools
- **Lease Management**: Full lease lifecycle with renewals, releases, and expiration
- **Static Reservations**: MAC-based IP reservations for specific devices
- **Multiple Storage Backends**: Memory, file, and database storage options
- **Security Features**: Client filtering, rate limiting, and device fingerprinting

### Web Interface
- **Real-time Dashboard**: View DHCP server status and statistics
- **Lease Management**: Monitor active leases and their details
- **Reservation Management**: Create and manage static IP reservations
- **Auto-refresh**: Live updates every 30 seconds

### Configuration Options
- **Network Interface**: Specify which interface to bind to
- **IP Pool Configuration**: Define start/end IP ranges and subnet settings
- **Lease Times**: Configurable lease duration and renewal settings
- **DNS Integration**: Automatic DNS server configuration
- **Performance Tuning**: Worker pools, timeouts, and buffer sizes

## Command Line Usage

### Basic Usage
```bash
# Enable DHCP server with default settings
./pihole-analyzer --dhcp --web

# Specify network interface
./pihole-analyzer --dhcp --dhcp-interface eth0 --web

# Configure IP pool range
./pihole-analyzer --dhcp --dhcp-pool-start 192.168.1.100 --dhcp-pool-end 192.168.1.200 --web

# Use custom configuration file
./pihole-analyzer --dhcp --dhcp-config dhcp-config.json --web
```

### Command Line Flags
- `--dhcp`: Enable DHCP server
- `--dhcp-interface <name>`: Network interface (e.g., eth0, wlan0)
- `--dhcp-pool-start <ip>`: Pool start IP address
- `--dhcp-pool-end <ip>`: Pool end IP address
- `--dhcp-config <file>`: Path to DHCP configuration file

## Configuration File

### Basic DHCP Configuration
```json
{
  "dhcp": {
    "enabled": true,
    "interface": "eth0",
    "listen_address": "0.0.0.0",
    "port": 67,
    "lease_time": "24h",
    "max_lease_time": "72h",
    "pool": {
      "start_ip": "192.168.1.100",
      "end_ip": "192.168.1.200",
      "subnet": "192.168.1.0/24",
      "gateway": "192.168.1.1",
      "dns_servers": ["192.168.1.1", "8.8.8.8"],
      "exclude": ["192.168.1.150", "192.168.1.151"]
    },
    "options": {
      "router": "192.168.1.1",
      "domain_name": "local",
      "domain_name_server": ["192.168.1.1", "8.8.8.8"],
      "mtu": 1500
    }
  }
}
```

### Advanced Configuration
```json
{
  "dhcp": {
    "enabled": true,
    "interface": "eth0",
    "listen_address": "0.0.0.0",
    "port": 67,
    "lease_time": "24h",
    "reservations": [
      {
        "mac": "00:11:22:33:44:55",
        "ip": "192.168.1.50",
        "hostname": "server",
        "description": "Main server",
        "enabled": true
      }
    ],
    "storage": {
      "type": "memory",
      "sync_interval": "5m",
      "max_leases": 10000
    },
    "security": {
      "enable_rate_limit": true,
      "max_requests_per_ip": 100,
      "log_all_requests": false,
      "enable_fingerprinting": true
    },
    "performance": {
      "max_connections": 1000,
      "read_timeout": "30s",
      "write_timeout": "30s",
      "worker_pool_size": 10
    }
  }
}
```

## Web Interface

### Accessing the DHCP Dashboard
1. Start the server with web interface: `./pihole-analyzer --dhcp --web`
2. Open your browser to `http://localhost:8080/dhcp`
3. View real-time DHCP server status, active leases, and reservations

### Dashboard Features
- **Server Status**: Running state, interface, pool utilization
- **Active Leases**: Current IP assignments with expiration times
- **IP Reservations**: Static MAC-to-IP mappings
- **Statistics**: Request counts, success rates, and client activity

### API Endpoints
- `GET /api/dhcp/status` - Server status and statistics
- `GET /api/dhcp/leases` - List all DHCP leases
- `GET /api/dhcp/reservations` - List IP reservations
- `POST /api/dhcp/reservation/` - Create new reservation
- `DELETE /api/dhcp/reservation/{mac}` - Delete reservation

## Architecture

### Components
- **DHCPServer**: Main server interface and lifecycle management
- **LeaseManager**: IP allocation, lease tracking, and renewals
- **PacketHandler**: DHCP protocol message processing
- **Storage**: Persistent lease and reservation storage
- **Networking**: UDP socket management and packet I/O
- **Security**: Client validation, rate limiting, and auditing

### Storage Backends
- **Memory**: Fast in-memory storage (default)
- **File**: JSON file-based persistence
- **Database**: SQLite database storage

### Security Features
- **Client Filtering**: Allow/block lists by MAC address
- **Rate Limiting**: Prevent DHCP request flooding
- **Device Fingerprinting**: Automatic device type detection
- **Request Logging**: Comprehensive audit trails

## Integration

### With Pi-hole
The DHCP server integrates seamlessly with Pi-hole analysis:
- Combined network device discovery
- Correlated DNS query analysis
- Unified client tracking
- Comprehensive network insights

### With Web Dashboard
- Real-time lease monitoring
- Interactive management interface
- WebSocket live updates
- Mobile-friendly responsive design

## Troubleshooting

### Common Issues
1. **Port 67 already in use**: Stop existing DHCP services or use different port
2. **Permission denied**: Run with sudo for privileged port binding
3. **Interface not found**: Verify network interface name with `ip link show`
4. **Pool exhaustion**: Increase IP range or check for lease cleanup

### Logging
Enable detailed logging to troubleshoot issues:
```json
{
  "logging": {
    "level": "DEBUG",
    "enable_colors": true
  },
  "dhcp": {
    "security": {
      "log_all_requests": true
    }
  }
}
```

### Testing
Test DHCP functionality without affecting production:
```bash
# Use test mode with mock data
./pihole-analyzer-test --dhcp --test --web

# Use non-standard port for testing
./pihole-analyzer --dhcp --dhcp-config test-dhcp.json --web
```

## Performance Considerations

### Tuning Options
- **Worker Pool Size**: Adjust based on expected concurrent clients
- **Buffer Sizes**: Increase for high-traffic networks
- **Cleanup Intervals**: Balance storage efficiency with performance
- **Storage Backend**: Memory for speed, database for persistence

### Monitoring
- Monitor pool utilization to prevent exhaustion
- Track request rates for capacity planning
- Review lease statistics for optimization opportunities
- Use metrics integration for long-term analysis

## Future Enhancements

### Planned Features
- DHCPv6 support for IPv6 networks
- Clustered DHCP for high availability
- Advanced traffic shaping integration
- Enhanced device classification
- Automated network discovery