# Quick Start - VPS Deployment

## Fastest Way to Deploy

### 1. Build for Linux (on Windows)

```powershell
# Run the build script
.\build-linux.bat

# Or manually:
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o evilginx_monitor .
```

### 2. Transfer to VPS

```powershell
# Using SCP (replace with your VPS details)
scp evilginx_monitor user@your-vps-ip:/home/user/
scp evilginx-monitor.service user@your-vps-ip:/home/user/
```

### 3. On Your VPS - Quick Setup

```bash
# SSH into your VPS
ssh user@your-vps-ip

# Create directories
sudo mkdir -p /opt/evilginx_monitor /var/log/evilginx_monitor

# Move files
sudo mv ~/evilginx_monitor /opt/evilginx_monitor/
sudo chmod +x /opt/evilginx_monitor/evilginx_monitor

# Edit service file - REPLACE YOUR_USERNAME with your actual username
sudo nano ~/evilginx-monitor.service
# Change YOUR_USERNAME to your actual username, then:
sudo mv ~/evilginx-monitor.service /etc/systemd/system/

# Start service
sudo systemctl daemon-reload
sudo systemctl enable evilginx-monitor
sudo systemctl start evilginx-monitor

# Check status
sudo systemctl status evilginx-monitor

# Allow firewall
sudo ufw allow 8080/tcp
```

### 4. Access Admin UI

**Option A: Direct IP Access**
Open in browser: `http://your-vps-ip:8080`

**Option B: Use Domain (Recommended)**
See **[DOMAIN_SETUP.md](DOMAIN_SETUP.md)** for setting up a domain with SSL.
After setup, access at: `https://monitor.example.com`

### 5. Configure

Use the web interface to:
- Set database file path (e.g., `/root/.evilginx/data.db`)
- Configure Telegram/Discord/Email
- Start monitoring

## Common Commands

```bash
# View logs
sudo journalctl -u evilginx-monitor -f

# Restart service
sudo systemctl restart evilginx-monitor

# Stop service
sudo systemctl stop evilginx-monitor

# Check if running
sudo systemctl status evilginx-monitor
```

## Troubleshooting

**Service won't start?**
- Check logs: `sudo journalctl -u evilginx-monitor -n 50`
- Verify user in service file matches your username
- Check file permissions: `ls -la /opt/evilginx_monitor/`

**Can't access web UI?**
- Check firewall: `sudo ufw status`
- Verify port is listening: `sudo netstat -tlnp | grep 8080`
- Check service is running: `sudo systemctl status evilginx-monitor`

**Database path issues?**
- Ensure the path exists and is readable
- If using `/root/.evilginx/data.db`, you may need to run as root or adjust permissions

