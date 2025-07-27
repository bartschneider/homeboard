# Installation Guide - E-Paper Homelab Dashboard

## Prerequisites

### Hardware Requirements
- Jailbroken Amazon Kindle (Paperwhite, Oasis, or similar)
- Homelab server (Raspberry Pi, NAS, or any Linux server)
- Local network connection for both devices

### Software Requirements
- **Server**: Go 1.21+, Python 3.7+
- **Kindle**: KUAL (Kindle Unified Application Launcher)
- **Optional**: psutil, pytz Python packages for enhanced widgets

## Step 1: Server Setup

### 1.1 Download and Build
```bash
# Clone the repository
git clone <repository-url>
cd homeboard

# Install Go dependencies
go mod tidy

# Build the server
go build -o bin/homeboard cmd/server/main.go
```

### 1.2 Optional Python Dependencies
```bash
# For enhanced widget functionality
pip3 install psutil pytz requests

# Verify installation
python3 -c "import psutil, pytz; print('Dependencies installed successfully')"
```

### 1.3 Configure the Dashboard
```bash
# Copy and edit configuration
cp config.json config.json.backup
nano config.json
```

Key configuration options:
- `server_port`: Choose an available port (default: 8081)
- `refresh_interval`: How often the dashboard refreshes (default: 15 minutes)
- `widgets`: Enable/disable widgets and configure parameters

### 1.4 Test the Server
```bash
# Start the server
./bin/homeboard -config config.json -verbose

# In another terminal, test the endpoints
curl http://localhost:8081/health
curl http://localhost:8081/api/config

# View dashboard in browser
open http://localhost:8081
```

## Step 2: Kindle Setup

### 2.1 Prerequisites Check
Ensure your Kindle is:
- Jailbroken (check with `uname -a` in terminal)
- Has KUAL installed (visible in main menu)
- Connected to the same network as your server

### 2.2 Install KUAL Extension
```bash
# On your computer, prepare the extension
cp -r kindle-extension/ dashboard-launcher/

# Edit the configuration
nano dashboard-launcher/dashboard-launcher.sh
# Set SERVER_IP to your homelab server's IP address
# Set SERVER_PORT to match your server configuration
```

### 2.3 Copy to Kindle
```bash
# Mount Kindle via USB or copy over network
# Copy extension to Kindle extensions directory
cp -r dashboard-launcher/ /mnt/us/extensions/

# Ensure scripts are executable
chmod +x /mnt/us/extensions/dashboard-launcher/*.sh
```

### 2.4 Launch Dashboard
1. Restart your Kindle or refresh KUAL
2. Open KUAL from the main menu
3. Find "E-Paper Dashboard" in the list
4. Tap "Launch Dashboard"
5. The browser should open in fullscreen mode

## Step 3: Validation and Testing

### 3.1 Server Validation
```bash
# Check server logs
./bin/homeboard -verbose

# Monitor widget execution
tail -f /var/log/dashboard.log

# Test individual widgets
python3 widgets/clock.py '{"format": "%H:%M", "timezone": "Local"}'
python3 widgets/system.py '{"show_cpu": true}'
```

### 3.2 Kindle Validation
```bash
# Check Kindle logs (if SSH access available)
tail -f /mnt/us/dashboard.log

# Test browser manually
/usr/bin/webreader http://YOUR_SERVER_IP:8081

# Verify screensaver is disabled
lipc-get-prop com.lab126.powerd preventScreenSaver
```

### 3.3 Network Troubleshooting
```bash
# From Kindle (if SSH available)
ping YOUR_SERVER_IP
curl -I http://YOUR_SERVER_IP:8081/health

# From server
netstat -ln | grep :8081
journalctl -f
```

## Step 4: Production Deployment

### 4.1 Systemd Service (Optional)
```bash
# Create systemd service
sudo nano /etc/systemd/system/homeboard.service
```

```ini
[Unit]
Description=E-Paper Homelab Dashboard
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/homeboard
ExecStart=/home/pi/homeboard/bin/homeboard -config config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl enable homeboard
sudo systemctl start homeboard
sudo systemctl status homeboard
```

### 4.2 Firewall Configuration
```bash
# Allow dashboard port (adjust for your firewall)
sudo ufw allow 8081/tcp

# Or for iptables
sudo iptables -A INPUT -p tcp --dport 8081 -j ACCEPT
```

### 4.3 Backup Configuration
```bash
# Create backup script
cat > backup-config.sh << 'EOF'
#!/bin/bash
cp config.json "config-backup-$(date +%Y%m%d).json"
tar -czf "dashboard-backup-$(date +%Y%m%d).tar.gz" config.json widgets/ kindle-extension/
EOF

chmod +x backup-config.sh
```

## Troubleshooting Guide

### Common Issues

#### Server Won't Start
```bash
# Check port availability
sudo netstat -tlnp | grep :8081

# Check permissions
ls -la config.json bin/homeboard

# Test with different port
./bin/homeboard -config config.json -verbose
```

#### Widgets Not Working
```bash
# Test widgets individually
python3 widgets/clock.py '{}'
python3 widgets/system.py '{"show_cpu": true}'

# Check Python path
which python3

# Install missing dependencies
pip3 install psutil pytz
```

#### Kindle Browser Issues
```bash
# Check browser executable
ls -la /usr/bin/webreader /usr/bin/browser

# Clear browser cache
rm -rf /var/local/browser/cache/*

# Test manual launch
/usr/bin/webreader http://SERVER_IP:8081 &
```

#### Network Connectivity
```bash
# Test from server
curl -I http://localhost:8081/health

# Test from another device
curl -I http://SERVER_IP:8081/health

# Check firewall
sudo ufw status
sudo iptables -L
```

### Log Files and Debugging

#### Server Logs
```bash
# Application logs
./bin/homeboard -verbose

# System logs
journalctl -u homeboard -f

# Access logs (if configured)
tail -f /var/log/homeboard/access.log
```

#### Kindle Logs
```bash
# Dashboard specific logs
tail -f /mnt/us/dashboard.log

# System logs (if available)
dmesg | tail
```

#### Widget Debugging
```bash
# Test with parameters
python3 widgets/clock.py '{"format": "%Y-%m-%d %H:%M:%S", "timezone": "UTC"}'

# Check widget timeout
timeout 30s python3 widgets/system.py '{}'

# Debug widget execution
strace -e trace=write python3 widgets/clock.py '{}'
```

### Performance Optimization

#### Server Performance
```bash
# Monitor resource usage
htop
iotop
netstat -i

# Optimize Go binary
go build -ldflags="-s -w" -o bin/homeboard cmd/server/main.go

# Use production settings
export GOMAXPROCS=1  # For single-core systems
```

#### Widget Performance
```bash
# Profile widget execution time
time python3 widgets/system.py '{}'

# Reduce widget timeout for faster responses
# Edit config.json: "timeout": 5
```

#### Kindle Performance
```bash
# Disable unnecessary services
lipc-set-prop com.lab126.powerd preventScreenSaver 1
lipc-set-prop com.lab126.powerd screenSaverTimeout 3600000

# Clear browser memory
pkill webreader; sleep 1; /usr/bin/webreader http://SERVER_IP:8081
```

## Next Steps

After successful installation:

1. **Create Custom Widgets**: Follow the widget development guide
2. **Configure Auto-Start**: Set up the server to start on boot
3. **Add Monitoring**: Monitor dashboard health and performance
4. **Backup Strategy**: Regular configuration and widget backups
5. **Security**: Consider VPN or firewall rules for external access

## Support

- Check the troubleshooting section above
- Review widget logs and server output
- Test components individually (server, widgets, Kindle)
- Verify network connectivity between all components