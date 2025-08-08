# Troubleshooting Guide

This guide helps resolve common issues with the Pi-hole Network Analyzer and Pi-hole API connectivity.

## Quick Diagnosis

### Common Error Patterns

| Error Message | Likely Cause | Quick Fix |
|---------------|--------------|-----------|
| API connection refused | Pi-hole API disabled or wrong port | Verify Pi-hole API settings |
| Authentication failed | Wrong API password | Check Pi-hole admin password |
| Network timeout | Network connectivity issue | Test network connectivity |
| Permission denied | TOTP/2FA issue | Verify TOTP configuration |

## Pi-hole API Connection Issues

### API Connection Refused

**Error Example:**
```
Error: Failed to connect to Pi-hole API: connection refused at 192.168.1.50:80
```

**Diagnosis Steps:**
```bash
# Test Pi-hole web interface
curl http://192.168.1.50/admin/

# Test basic API endpoint
curl http://192.168.1.50/admin/api.php

# Check Pi-hole service status
systemctl status pihole-FTL
```

**Solutions:**

1. **Verify Pi-hole is running:**
   ```bash
   sudo systemctl start pihole-FTL
   sudo systemctl enable pihole-FTL
   ```

2. **Check firewall settings:**
   ```bash
   sudo ufw allow 80
   sudo ufw allow 443  # for HTTPS
   ```

3. **Verify Pi-hole API is enabled:**
   - Access Pi-hole admin interface: `http://192.168.1.50/admin/`
   - Go to Settings → API/Web interface
   - Ensure "API" is enabled

### Authentication Failed

**Error Example:**
```
Error: Failed to authenticate with Pi-hole API: invalid password
```

**Diagnosis Steps:**
```bash
# Test with your current password
curl "http://192.168.1.50/admin/api.php?auth=your-password"

# Check if 2FA is enabled
curl "http://192.168.1.50/admin/api.php?auth=your-password&totp=123456"
```

**Solutions:**

1. **Reset Pi-hole admin password:**
   ```bash
   pihole -a -p
   ```

2. **Verify password in configuration:**
   ```json
   {
     "pihole": {
       "api_password": "correct-password-here"
     }
   }
   ```

3. **For 2FA/TOTP enabled Pi-hole:**
   ```json
   {
     "pihole": {
       "api_password": "your-password",
       "api_totp": "your-totp-secret-key"
     }
   }
   ```

### Network Connectivity Issues

**Error Example:**
```
Error: Network timeout connecting to Pi-hole API
```

**Diagnosis Steps:**
```bash
# Test basic connectivity
ping 192.168.1.50

# Test port connectivity
telnet 192.168.1.50 80

# Check DNS resolution
nslookup pihole.local
```

**Solutions:**

1. **Verify network routing:**
   ```bash
   traceroute 192.168.1.50
   ```

2. **Check Pi-hole network settings:**
   ```bash
   # On Pi-hole server
   ip addr show
   sudo netstat -tlnp | grep :80
   ```

3. **Update configuration with correct IP:**
   ```json
   {
     "pihole": {
       "host": "correct-ip-address",
       "port": 80
     }
   }
   ```

## Configuration Issues

### Invalid Configuration File

**Error Example:**
```
Error: Failed to parse configuration file: invalid JSON
```

**Solutions:**

1. **Validate JSON syntax:**
   ```bash
   # Check JSON validity
   jq '.' config.json

   # Pretty print to find errors
   cat config.json | python -m json.tool
   ```

2. **Create fresh configuration:**
   ```bash
   ./pihole-analyzer --create-config
   ```

3. **Use configuration wizard:**
   ```bash
   ./pihole-analyzer --pihole-setup
   ```

### Missing Configuration Values

**Error Example:**
```
Error: Pi-hole host not specified in configuration
```

**Solutions:**

1. **Verify required fields:**
   ```json
   {
     "pihole": {
       "host": "192.168.1.50",
       "port": 80,
       "api_enabled": true,
       "api_password": "required-password"
     }
   }
   ```

2. **Check configuration:**
   ```bash
   ./pihole-analyzer --show-config
   ```

## Performance Issues

### Slow API Responses

**Symptoms:**
- Long delays during analysis
- Timeout errors
- High network usage

**Diagnosis:**
```bash
# Test API response time
time curl "http://192.168.1.50/admin/api.php"

# Monitor network usage
iftop -i eth0
```

**Solutions:**

1. **Increase API timeout:**
   ```json
   {
     "pihole": {
       "api_timeout": 60
     }
   }
   ```

2. **Use Pi-hole during low usage:**
   ```bash
   # Run during off-peak hours
   0 3 * * * /usr/local/bin/pihole-analyzer --pihole config.json
   ```

3. **Optimize Pi-hole database:**
   ```bash
   # On Pi-hole server
   sudo service pihole-FTL stop
   sudo sqlite3 /etc/pihole/pihole-FTL.db "VACUUM;"
   sudo service pihole-FTL start
   ```

### High Memory Usage

**Symptoms:**
- System slowdown during analysis
- Out of memory errors

**Solutions:**

1. **Limit query results:**
   ```bash
   # Process only online clients
   ./pihole-analyzer --pihole config.json --online-only
   ```

2. **Run with resource limits:**
   ```bash
   # Limit memory usage
   ulimit -v 1048576  # 1GB limit
   ./pihole-analyzer --pihole config.json
   ```

## Container Issues

### Container Build Failures

**Error Example:**
```
Error: Docker build failed
```

**Solutions:**

1. **Clean Docker cache:**
   ```bash
   docker system prune -a
   ```

2. **Build with verbose output:**
   ```bash
   docker build --no-cache --progress=plain .
   ```

3. **Use pre-built image:**
   ```bash
   docker pull ghcr.io/grammatonic/pihole-analyzer:latest
   ```

### Container Runtime Issues

**Error Example:**
```
Error: Container exits immediately
```

**Solutions:**

1. **Check container logs:**
   ```bash
   docker logs pihole-analyzer
   ```

2. **Run interactively:**
   ```bash
   docker run -it --rm ghcr.io/grammatonic/pihole-analyzer:latest sh
   ```

3. **Verify volume mounts:**
   ```bash
   docker run --rm \
     -v $(pwd)/config.json:/app/config.json:ro \
     ghcr.io/grammatonic/pihole-analyzer:latest \
     --show-config
   ```

### Permission Issues

**Error Example:**
```
Error: Permission denied accessing configuration file
```

**Solutions:**

1. **Fix file permissions:**
   ```bash
   chmod 644 config.json
   chown 1001:1001 config.json  # Container user
   ```

2. **Use proper volume mounts:**
   ```bash
   docker run --rm \
     -v $(pwd)/config.json:/app/config.json:ro \
     ghcr.io/grammatonic/pihole-analyzer:latest
   ```

## Application Issues

### Binary Not Found

**Error Example:**
```
bash: ./pihole-analyzer: No such file or directory
```

**Solutions:**

1. **Build the application:**
   ```bash
   make build
   ```

2. **Check binary permissions:**
   ```bash
   chmod +x pihole-analyzer
   ```

3. **Use correct binary:**
   ```bash
   # Production binary
   ./pihole-analyzer --pihole config.json
   
   # Test binary with mock data
   ./pihole-analyzer-test --test
   ```

### Segmentation Faults

**Error Example:**
```
Segmentation fault (core dumped)
```

**Solutions:**

1. **Check Go version:**
   ```bash
   go version  # Requires Go 1.23+
   ```

2. **Rebuild with debug info:**
   ```bash
   go build -race -o pihole-analyzer-debug ./cmd/pihole-analyzer
   ```

3. **Run with debugging:**
   ```bash
   gdb ./pihole-analyzer
   ```

## Logging and Debug

### Enable Debug Logging

```bash
# Environment variable
export LOG_LEVEL=debug

# Configuration file
{
  "logging": {
    "level": "debug",
    "file": "/tmp/pihole-analyzer-debug.log"
  }
}
```

### Log Analysis

```bash
# View structured logs
tail -f /var/log/pihole-analyzer.log | jq '.'

# Filter error logs
grep "level=ERROR" /var/log/pihole-analyzer.log

# Analyze API calls
grep "pihole-api" /var/log/pihole-analyzer.log | jq '.duration'
```

### Performance Profiling

```bash
# CPU profiling
go build -o pihole-analyzer-profile ./cmd/pihole-analyzer
./pihole-analyzer-profile --pihole config.json --cpuprofile cpu.prof

# Memory profiling
go tool pprof pihole-analyzer-profile mem.prof
```

## Getting Help

### Information Gathering

Before seeking help, gather this information:

1. **Version information:**
   ```bash
   ./pihole-analyzer --version
   go version
   ```

2. **Configuration (sanitized):**
   ```bash
   ./pihole-analyzer --show-config | sed 's/password.*/password: ***REDACTED***/'
   ```

3. **Error logs:**
   ```bash
   tail -n 50 /var/log/pihole-analyzer.log
   ```

4. **System information:**
   ```bash
   uname -a
   docker version  # if using containers
   ```

### Support Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and community support
- **Documentation**: Comprehensive guides and references
- **Container Registry**: Pre-built images and deployment guides

### Common Debugging Commands

```bash
# Test Pi-hole API connectivity
curl "http://192.168.1.50/admin/api.php"

# Validate configuration
./pihole-analyzer --show-config

# Run with verbose logging
LOG_LEVEL=debug ./pihole-analyzer --pihole config.json

# Test with mock data
./pihole-analyzer-test --test

# Container debugging
docker run -it --rm ghcr.io/grammatonic/pihole-analyzer:latest sh
```

This troubleshooting guide covers the most common issues with API-only Pi-hole connectivity and provides systematic approaches to diagnosis and resolution.
