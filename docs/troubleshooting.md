# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Pi-hole Network Analyzer.

## Quick Diagnosis

### 1. Basic Health Check

```bash
# Test if the binary works
./pihole-analyzer --help

# Test with mock data (no Pi-hole required)
./pihole-analyzer --test

# Verify configuration
./pihole-analyzer --show-config
```

### 2. Common Issues Quick Reference

| Issue | Quick Fix |
|-------|-----------|
| Command not found | `make build` or check PATH |
| SSH connection refused | Verify Pi-hole IP and SSH service |
| Permission denied | Check SSH key permissions: `chmod 600 ~/.ssh/id_rsa` |
| Database not found | Verify Pi-hole database path |
| No color output | Check terminal capability or use `--no-color` |

## SSH Connection Issues

### Connection Refused

**Symptoms:**
```
Error: Failed to connect to Pi-hole server: ssh: connect to host 192.168.1.50 port 22: connection refused
```

**Diagnosis:**
```bash
# Test basic connectivity
ping 192.168.1.50

# Test SSH port
telnet 192.168.1.50 22
# or
nmap -p 22 192.168.1.50

# Check if SSH service is running on Pi-hole
ssh user@192.168.1.50 "sudo systemctl status ssh"
```

**Solutions:**

1. **Start SSH service:**
   ```bash
   # On Pi-hole server
   sudo systemctl start ssh
   sudo systemctl enable ssh
   ```

2. **Check firewall:**
   ```bash
   # On Pi-hole server
   sudo ufw status
   sudo ufw allow ssh
   ```

3. **Verify SSH configuration:**
   ```bash
   # On Pi-hole server
   sudo nano /etc/ssh/sshd_config
   # Ensure: Port 22 (or your custom port)
   # Restart: sudo systemctl restart ssh
   ```

### Permission Denied (Public Key)

**Symptoms:**
```
Error: Failed to connect to Pi-hole server: ssh: handshake failed: ssh: unable to authenticate
```

**Diagnosis:**
```bash
# Test SSH key manually
ssh -i ~/.ssh/id_rsa pi@192.168.1.50

# Check SSH key permissions
ls -la ~/.ssh/id_rsa*

# Verbose SSH debugging
ssh -vvv -i ~/.ssh/id_rsa pi@192.168.1.50
```

**Solutions:**

1. **Fix key permissions:**
   ```bash
   chmod 600 ~/.ssh/id_rsa
   chmod 644 ~/.ssh/id_rsa.pub
   chmod 700 ~/.ssh
   ```

2. **Copy public key to Pi-hole:**
   ```bash
   ssh-copy-id -i ~/.ssh/id_rsa.pub pi@192.168.1.50
   ```

3. **Verify authorized_keys:**
   ```bash
   # On Pi-hole server
   cat ~/.ssh/authorized_keys
   chmod 600 ~/.ssh/authorized_keys
   chmod 700 ~/.ssh
   ```

### Host Key Verification Failed

**Symptoms:**
```
Host key verification failed
```

**Solutions:**
```bash
# Remove old host key
ssh-keygen -R 192.168.1.50

# Accept new host key
ssh pi@192.168.1.50
```

## Database Access Issues

### Database Not Found

**Symptoms:**
```
Error: Failed to query Pi-hole database: no such file or directory
```

**Diagnosis:**
```bash
# Find Pi-hole database
ssh pi@192.168.1.50 "find /var -name 'pihole-FTL.db' 2>/dev/null"
ssh pi@192.168.1.50 "find /etc -name 'pihole-FTL.db' 2>/dev/null"
ssh pi@192.168.1.50 "find /opt -name 'pihole-FTL.db' 2>/dev/null"

# Check common locations
ssh pi@192.168.1.50 "ls -la /etc/pihole/pihole-FTL.db"
ssh pi@192.168.1.50 "ls -la /var/lib/pihole/pihole-FTL.db"
```

**Solutions:**

1. **Update database path in config:**
   ```json
   {
     "pihole": {
       "dbPath": "/var/lib/pihole/pihole-FTL.db"
     }
   }
   ```

2. **Check Pi-hole installation:**
   ```bash
   # On Pi-hole server
   pihole status
   pihole -v
   ```

### Database Permission Denied

**Symptoms:**
```
Error: Failed to query Pi-hole database: permission denied
```

**Diagnosis:**
```bash
# Check database permissions
ssh pi@192.168.1.50 "ls -la /etc/pihole/pihole-FTL.db"

# Test database access
ssh pi@192.168.1.50 "sqlite3 /etc/pihole/pihole-FTL.db '.tables'"
```

**Solutions:**

1. **Add user to pihole group:**
   ```bash
   # On Pi-hole server
   sudo usermod -a -G pihole pi
   ```

2. **Use sudo for database access:**
   ```bash
   # On Pi-hole server
   echo "pi ALL=(ALL) NOPASSWD: /usr/bin/sqlite3 /etc/pihole/pihole-FTL.db*" | sudo tee /etc/sudoers.d/pihole-analyzer
   ```

3. **Fix database permissions:**
   ```bash
   # On Pi-hole server (be careful!)
   sudo chmod 644 /etc/pihole/pihole-FTL.db
   ```

### Database Locked

**Symptoms:**
```
Error: Failed to query Pi-hole database: database is locked
```

**Diagnosis:**
```bash
# Check what's using the database
ssh pi@192.168.1.50 "sudo lsof /etc/pihole/pihole-FTL.db"

# Check Pi-hole FTL status
ssh pi@192.168.1.50 "sudo systemctl status pihole-FTL"
```

**Solutions:**

1. **Wait and retry:**
   ```bash
   # Database locks are usually temporary
   sleep 30
   ./pihole-analyzer --pihole config.json
   ```

2. **Restart Pi-hole FTL (careful!):**
   ```bash
   # On Pi-hole server
   sudo systemctl restart pihole-FTL
   ```

## Configuration Issues

### Config File Not Found

**Symptoms:**
```
Config file not found at ~/.pihole-analyzer/config.json, using defaults
Error: Pi-hole configuration required
```

**Solutions:**

1. **Create default configuration:**
   ```bash
   ./pihole-analyzer --create-config
   ```

2. **Use custom config path:**
   ```bash
   ./pihole-analyzer --config /path/to/config.json --pihole
   ```

3. **Run interactive setup:**
   ```bash
   ./pihole-analyzer --pihole-setup
   ```

### Invalid JSON Configuration

**Symptoms:**
```
Error: Failed to parse configuration file: invalid character
```

**Diagnosis:**
```bash
# Validate JSON syntax
cat ~/.pihole-analyzer/config.json | python -m json.tool

# Or use online JSON validator
```

**Solutions:**

1. **Fix JSON syntax:**
   - Check for missing commas
   - Verify quote marks
   - Ensure proper nesting

2. **Recreate configuration:**
   ```bash
   mv ~/.pihole-analyzer/config.json ~/.pihole-analyzer/config.json.backup
   ./pihole-analyzer --create-config
   ```

## Network Analysis Issues

### ARP Table Empty

**Symptoms:**
```
Warning: No ARP entries found, all clients will show as offline
```

**Diagnosis:**
```bash
# Test ARP command access
ssh pi@192.168.1.50 "arp -a"

# Alternative IP command
ssh pi@192.168.1.50 "ip neigh show"

# Check network interface
ssh pi@192.168.1.50 "ip addr show"
```

**Solutions:**

1. **Use alternative network command:**
   ```bash
   # Add to sudoers if needed
   echo "pi ALL=(ALL) NOPASSWD: /sbin/ip" | sudo tee /etc/sudoers.d/pihole-network
   ```

2. **Generate ARP traffic:**
   ```bash
   # On Pi-hole server
   ping -c 1 192.168.1.1  # Gateway
   ping -c 1 192.168.1.100  # Known device
   ```

### Network Detection Failures

**Symptoms:**
```
Warning: Failed to determine network status for client 192.168.1.100
```

**Solutions:**

1. **Check network interface:**
   ```bash
   ssh pi@192.168.1.50 "ip route show default"
   ```

2. **Update network exclusions:**
   ```json
   {
     "exclusions": {
       "networks": [
         "192.168.1.0/24",  // Your actual network
         "172.16.0.0/12"
       ]
     }
   }
   ```

## Performance Issues

### Slow Analysis

**Symptoms:**
- Analysis takes a very long time
- High memory usage
- System becomes unresponsive

**Diagnosis:**
```bash
# Check database size
ssh pi@192.168.1.50 "ls -lh /etc/pihole/pihole-FTL.db"

# Monitor memory usage
/usr/bin/time -v ./pihole-analyzer --pihole config.json

# Check Pi-hole server load
ssh pi@192.168.1.50 "top -n 1"
```

**Solutions:**

1. **Use quiet mode:**
   ```bash
   ./pihole-analyzer --pihole config.json --quiet
   ```

2. **Limit client display:**
   ```json
   {
     "output": {
       "maxClients": 25
     }
   }
   ```

3. **Optimize database query:**
   ```bash
   # On Pi-hole server, optimize database
   ssh pi@192.168.1.50 "sqlite3 /etc/pihole/pihole-FTL.db 'VACUUM;'"
   ```

### Memory Issues

**Symptoms:**
```
fatal error: out of memory
```

**Solutions:**

1. **Limit Go memory:**
   ```bash
   export GOMEMLIMIT=512MiB
   ./pihole-analyzer --pihole config.json
   ```

2. **Increase system swap:**
   ```bash
   # Add swap space if needed
   sudo fallocate -l 1G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

## Output and Display Issues

### No Color Output

**Symptoms:**
- Output appears without colors
- All text is plain/monochrome

**Diagnosis:**
```bash
# Check terminal capabilities
echo $TERM
tput colors

# Test color manually
echo -e "\033[31mRed Text\033[0m"
```

**Solutions:**

1. **Force color mode:**
   ```bash
   FORCE_COLOR=1 ./pihole-analyzer --pihole config.json
   ```

2. **Use color-capable terminal:**
   - iTerm2 (macOS)
   - Windows Terminal (Windows)
   - GNOME Terminal (Linux)

3. **Disable colors if not needed:**
   ```bash
   ./pihole-analyzer --pihole config.json --no-color
   ```

### Emoji Display Issues

**Symptoms:**
- Boxes or question marks instead of emojis
- Garbled characters

**Solutions:**

1. **Disable emoji:**
   ```bash
   ./pihole-analyzer --pihole config.json --no-emoji
   ```

2. **Install emoji fonts:**
   ```bash
   # Ubuntu/Debian
   sudo apt install fonts-noto-color-emoji
   
   # macOS (usually built-in)
   # Windows: Install emoji support
   ```

### Output Formatting Issues

**Symptoms:**
- Misaligned columns
- Text wrapping problems

**Solutions:**

1. **Increase terminal width:**
   ```bash
   # Resize terminal window or
   stty cols 120
   ```

2. **Use quiet mode:**
   ```bash
   ./pihole-analyzer --pihole config.json --quiet
   ```

## Docker-Specific Issues

### Pi-hole in Docker

**Symptoms:**
```
Error: Cannot connect to Pi-hole in Docker container
```

**Solutions:**

1. **SSH to Docker host:**
   ```json
   {
     "pihole": {
       "host": "docker-host-ip",
       "dbPath": "/var/lib/docker/volumes/pihole_data/_data/pihole-FTL.db"
     }
   }
   ```

2. **Use docker exec:**
   ```bash
   # Access database directly
   docker exec pihole sqlite3 /etc/pihole/pihole-FTL.db ".tables"
   ```

### Container Networking

**Diagnosis:**
```bash
# Check container network
docker inspect pihole | grep IPAddress

# Test connectivity
docker exec pihole ping host.docker.internal
```

## Platform-Specific Issues

### macOS Issues

**SSH Agent Problems:**
```bash
# Start SSH agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa
```

**Permission Issues:**
```bash
# Fix macOS permission errors
sudo chown -R $(whoami) ~/.pihole-analyzer/
```

### Windows Issues

**Path Issues:**
```powershell
# Use full paths
.\pihole-analyzer.exe --config C:\Users\username\.pihole-analyzer\config.json
```

**SSH Issues:**
```powershell
# Use Windows SSH client
ssh -i C:\Users\username\.ssh\id_rsa pi@192.168.1.50
```

### Linux Issues

**Package Dependencies:**
```bash
# Install missing dependencies
sudo apt update
sudo apt install openssh-client sqlite3
```

## Error Code Reference

| Exit Code | Meaning | Solution |
|-----------|---------|----------|
| 0 | Success | Normal operation |
| 1 | Configuration error | Check config file and paths |
| 2 | SSH connection failed | Verify SSH setup |
| 3 | Database access failed | Check Pi-hole database permissions |
| 4 | Analysis failed | Check data integrity |
| 5 | Output error | Check file permissions |

## Getting Help

### Diagnostic Information

When reporting issues, include:

```bash
# System information
uname -a
go version

# Configuration (remove sensitive info)
./pihole-analyzer --show-config

# Test results
./pihole-analyzer --test

# SSH test
ssh -v pi@your-pihole-ip
```

### Log Collection

```bash
# Enable debug mode
export PIHOLE_ANALYZER_DEBUG=1
./pihole-analyzer --pihole config.json 2>&1 | tee debug.log

# SSH debugging
export SSH_DEBUG=1
./pihole-analyzer --pihole config.json 2>&1 | tee ssh-debug.log
```

### Community Support

- **GitHub Issues**: [Create an Issue](https://github.com/GrammaTonic/pihole-network-analyzer/issues)
- **Documentation**: [Full Documentation](README.md)
- **Discussions**: [GitHub Discussions](https://github.com/GrammaTonic/pihole-network-analyzer/discussions)

## Prevention

### Regular Maintenance

```bash
# Keep SSH keys updated
ssh-keygen -t rsa -b 4096 -f ~/.ssh/pihole_new_key

# Test connections regularly
./pihole-analyzer --pihole config.json --quiet

# Monitor Pi-hole health
ssh pi@192.168.1.50 "pihole status"
```

### Monitoring Setup

```bash
#!/bin/bash
# Health check script
if ! ./pihole-analyzer --test --quiet > /dev/null; then
    echo "Pi-hole analyzer test failed"
    exit 1
fi

if ! ./pihole-analyzer --pihole config.json --quiet > /dev/null; then
    echo "Pi-hole connection failed"
    exit 1
fi

echo "All checks passed"
```

---

**Need more help?** Check the [Development Guide](development.md) for advanced troubleshooting or create an issue on GitHub.
