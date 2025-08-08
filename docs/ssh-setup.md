# SSH Setup Guide

This guide provides comprehensive instructions for setting up secure SSH connections between the Pi-hole Network Analyzer and your Pi-hole server.

## SSH Overview

The Pi-hole Network Analyzer connects to your Pi-hole server via SSH to:
- Access the Pi-hole SQLite database (`pihole-FTL.db`)
- Execute network analysis commands (ARP table queries)
- Securely transfer data for analysis

## SSH Authentication Methods

### Method 1: SSH Key Authentication (Recommended)

SSH keys provide secure, password-less authentication and are the recommended method for production use.

#### Step 1: Generate SSH Key Pair

```bash
# Generate a new SSH key specifically for Pi-hole analysis
ssh-keygen -t rsa -b 4096 -C "pihole-analyzer" -f ~/.ssh/pihole_analyzer_key

# Alternative: Use existing SSH key
ls ~/.ssh/id_rsa*
```

#### Step 2: Copy Public Key to Pi-hole Server

```bash
# Copy public key to Pi-hole server
ssh-copy-id -i ~/.ssh/pihole_analyzer_key.pub pi@192.168.1.50

# Alternative: Manual copy
cat ~/.ssh/pihole_analyzer_key.pub | ssh pi@192.168.1.50 "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

#### Step 3: Test SSH Connection

```bash
# Test SSH connection with the new key
ssh -i ~/.ssh/pihole_analyzer_key pi@192.168.1.50

# Test database access
ssh -i ~/.ssh/pihole_analyzer_key pi@192.168.1.50 "ls -la /etc/pihole/pihole-FTL.db"
```

#### Step 4: Configure Pi-hole Analyzer

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 22,
    "username": "pi",
    "keyPath": "~/.ssh/pihole_analyzer_key",
    "password": "",
    "dbPath": "/etc/pihole/pihole-FTL.db"
  }
}
```

### Method 2: Password Authentication

For development or testing environments, you can use password authentication.

⚠️ **Security Warning**: Passwords are less secure than SSH keys and should only be used for testing.

#### Configuration

```json
{
  "pihole": {
    "host": "192.168.1.50",
    "port": 22,
    "username": "pi",
    "keyPath": "",
    "password": "your-secure-password",
    "dbPath": "/etc/pihole/pihole-FTL.db"
  }
}
```

## Pi-hole Server Setup

### User Account Configuration

#### Option 1: Use Existing 'pi' User

Most Pi-hole installations use the default `pi` user:

```bash
# Test connection
ssh pi@192.168.1.50

# Verify Pi-hole database access
sudo ls -la /etc/pihole/pihole-FTL.db
```

#### Option 2: Create Dedicated User

For better security, create a dedicated user for Pi-hole analysis:

```bash
# On Pi-hole server, create new user
sudo useradd -m -s /bin/bash pihole-analyzer

# Add to required groups
sudo usermod -a -G pi,adm pihole-analyzer

# Set up sudo access for database reading
echo "pihole-analyzer ALL=(ALL) NOPASSWD: /usr/bin/sqlite3 /etc/pihole/pihole-FTL.db*" | sudo tee /etc/sudoers.d/pihole-analyzer

# Test new user
sudo su - pihole-analyzer
```

### Database Permissions

#### Default Pi-hole Database Location

```bash
# Common Pi-hole database locations
/etc/pihole/pihole-FTL.db          # Most common
/var/lib/pihole/pihole-FTL.db      # Alternative
/opt/pihole/pihole-FTL.db          # Docker installations
```

#### Set Appropriate Permissions

```bash
# Check current permissions
ls -la /etc/pihole/pihole-FTL.db

# Add read permissions for pi user (if needed)
sudo chmod 644 /etc/pihole/pihole-FTL.db

# Alternative: Add pi user to pihole group
sudo usermod -a -G pihole pi
```

### Network Commands Access

The analyzer needs to execute network commands for ARP table analysis:

```bash
# Test ARP command access
ssh pi@192.168.1.50 "arp -a"

# Test IP command access (alternative)
ssh pi@192.168.1.50 "ip neigh show"

# If commands are restricted, add to sudoers
echo "pi ALL=(ALL) NOPASSWD: /usr/sbin/arp, /sbin/ip" | sudo tee /etc/sudoers.d/pihole-analyzer
```

## SSH Security Hardening

### SSH Server Configuration

Edit `/etc/ssh/sshd_config` on the Pi-hole server:

```bash
# Disable password authentication (after setting up keys)
PasswordAuthentication no
ChallengeResponseAuthentication no

# Disable root login
PermitRootLogin no

# Limit users
AllowUsers pi pihole-analyzer

# Change default port (optional)
Port 2222

# Restart SSH service
sudo systemctl restart ssh
```

### SSH Client Configuration

Create `~/.ssh/config` on the analyzer machine:

```bash
# Pi-hole SSH configuration
Host pihole
    HostName 192.168.1.50
    Port 22
    User pi
    IdentityFile ~/.ssh/pihole_analyzer_key
    IdentitiesOnly yes
    ServerAliveInterval 60
    ServerAliveCountMax 3
    Compression yes
```

Test with simplified command:
```bash
ssh pihole "ls -la /etc/pihole/"
```

### Key Management

#### Secure Key Storage

```bash
# Set restrictive permissions on private key
chmod 600 ~/.ssh/pihole_analyzer_key

# Set appropriate permissions on public key
chmod 644 ~/.ssh/pihole_analyzer_key.pub

# Secure SSH directory
chmod 700 ~/.ssh
```

#### Key Rotation

```bash
# Generate new key pair
ssh-keygen -t rsa -b 4096 -C "pihole-analyzer-$(date +%Y%m)" -f ~/.ssh/pihole_analyzer_key_new

# Copy new public key
ssh-copy-id -i ~/.ssh/pihole_analyzer_key_new.pub pi@192.168.1.50

# Test new key
ssh -i ~/.ssh/pihole_analyzer_key_new pi@192.168.1.50

# Update configuration
# Replace keyPath in config.json

# Remove old key from server
ssh pi@192.168.1.50 "sed -i '/old-key-comment/d' ~/.ssh/authorized_keys"
```

## Docker and Container Environments

### Pi-hole in Docker

If Pi-hole runs in Docker, SSH setup differs:

#### Option 1: SSH to Docker Host

```bash
# SSH to Docker host, then access container
ssh user@docker-host
docker exec -it pihole bash
```

Configuration:
```json
{
  "pihole": {
    "host": "docker-host-ip",
    "username": "docker-user",
    "keyPath": "~/.ssh/docker_key",
    "dbPath": "/var/lib/docker/volumes/pihole_data/_data/pihole-FTL.db"
  }
}
```

#### Option 2: SSH Directly to Container

If the Pi-hole container has SSH enabled:

```bash
# Find container SSH port
docker port pihole-container 22

# Connect directly
ssh -p 2222 pi@192.168.1.50
```

### Kubernetes Environments

For Pi-hole in Kubernetes:

```bash
# Port forward to Pi-hole pod
kubectl port-forward pod/pihole-pod 2222:22

# Connect via localhost
ssh -p 2222 pi@localhost
```

## Troubleshooting SSH

### Common SSH Issues

#### Connection Refused

```bash
# Check if SSH service is running
ssh -v pi@192.168.1.50

# Test with telnet
telnet 192.168.1.50 22

# Check SSH service on Pi-hole server
sudo systemctl status ssh
sudo systemctl start ssh
```

#### Permission Denied

```bash
# Check SSH key permissions
ls -la ~/.ssh/pihole_analyzer_key

# Fix permissions
chmod 600 ~/.ssh/pihole_analyzer_key

# Test SSH key
ssh -i ~/.ssh/pihole_analyzer_key pi@192.168.1.50
```

#### Host Key Verification Failed

```bash
# Remove old host key
ssh-keygen -R 192.168.1.50

# Accept new host key
ssh pi@192.168.1.50
```

#### Database Access Denied

```bash
# Test database access manually
ssh pi@192.168.1.50 "ls -la /etc/pihole/pihole-FTL.db"

# Check if database is locked
ssh pi@192.168.1.50 "sudo lsof /etc/pihole/pihole-FTL.db"

# Try alternative database location
ssh pi@192.168.1.50 "find /var -name 'pihole-FTL.db' 2>/dev/null"
```

### Debug SSH Connection

#### Verbose SSH Debugging

```bash
# Enable verbose SSH debugging
ssh -vvv pi@192.168.1.50

# Test with specific key
ssh -vvv -i ~/.ssh/pihole_analyzer_key pi@192.168.1.50

# Check SSH agent
ssh-add -l
ssh-add ~/.ssh/pihole_analyzer_key
```

#### SSH Connection Logs

On Pi-hole server:
```bash
# View SSH logs
sudo tail -f /var/log/auth.log

# Check SSH service status
sudo systemctl status ssh --no-pager -l
```

## Network Considerations

### Firewall Configuration

#### Pi-hole Server Firewall

```bash
# Allow SSH on UFW
sudo ufw allow ssh

# Allow specific IP range
sudo ufw allow from 192.168.1.0/24 to any port 22

# Check firewall status
sudo ufw status
```

#### Router/Network Configuration

- Ensure SSH port (22 or custom) is open
- Configure port forwarding if accessing from outside network
- Consider VPN for remote access

### Network Performance

#### SSH Compression

```bash
# Enable compression for slow connections
ssh -C pi@192.168.1.50

# Configure compression in ~/.ssh/config
Host pihole
    Compression yes
    CompressionLevel 6
```

#### Connection Keep-Alive

```bash
# Configure keep-alive in ~/.ssh/config
Host pihole
    ServerAliveInterval 60
    ServerAliveCountMax 3
    TCPKeepAlive yes
```

## Testing SSH Setup

### Comprehensive Test Script

```bash
#!/bin/bash
# SSH connectivity test script

PIHOLE_HOST="192.168.1.50"
PIHOLE_USER="pi"
SSH_KEY="~/.ssh/pihole_analyzer_key"
DB_PATH="/etc/pihole/pihole-FTL.db"

echo "Testing SSH connectivity to Pi-hole..."

# Test 1: Basic SSH connection
echo "1. Testing basic SSH connection..."
if ssh -i "$SSH_KEY" -o ConnectTimeout=10 "$PIHOLE_USER@$PIHOLE_HOST" "echo 'SSH connection successful'"; then
    echo "✅ SSH connection: SUCCESS"
else
    echo "❌ SSH connection: FAILED"
    exit 1
fi

# Test 2: Database access
echo "2. Testing database access..."
if ssh -i "$SSH_KEY" "$PIHOLE_USER@$PIHOLE_HOST" "ls -la '$DB_PATH'" > /dev/null; then
    echo "✅ Database access: SUCCESS"
else
    echo "❌ Database access: FAILED"
    exit 1
fi

# Test 3: ARP command
echo "3. Testing ARP command access..."
if ssh -i "$SSH_KEY" "$PIHOLE_USER@$PIHOLE_HOST" "arp -a" > /dev/null; then
    echo "✅ ARP command: SUCCESS"
else
    echo "⚠️ ARP command: FAILED (may impact online status detection)"
fi

echo "SSH setup test completed!"
```

### Pi-hole Analyzer Test

```bash
# Test with Pi-hole analyzer
./pihole-analyzer --pihole ~/.pihole-analyzer/config.json --quiet

# Verify configuration
./pihole-analyzer --show-config
```

## Next Steps

After SSH setup is complete:

1. **[Configuration Guide](configuration.md)** - Fine-tune your settings
2. **[Usage Guide](usage.md)** - Learn analysis commands
3. **[Troubleshooting](troubleshooting.md)** - Solve common issues

---

**SSH setup complete!** Your Pi-hole Network Analyzer can now securely connect to your Pi-hole server.
