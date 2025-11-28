# Deployment Guide for VPS

This guide will help you deploy the Evilginx Monitor application to a VPS (Virtual Private Server).

## Prerequisites

- A VPS running Linux (Ubuntu/Debian recommended)
- SSH access to your VPS
- Go installed on your local machine (for building)
- Basic knowledge of Linux commands

## Step 1: Build the Application for Linux

On your local Windows machine, you need to cross-compile for Linux:

```bash
# For 64-bit Linux
set GOOS=linux
set GOARCH=amd64
go build -o evilginx_monitor .

# Or use the build script (see below)
```

Alternatively, build directly on the VPS if you have Go installed there.

## Step 2: Transfer Files to VPS

### Option A: Using SCP (from Windows PowerShell or WSL)

```bash
# Transfer the binary
scp evilginx_monitor user@your-vps-ip:/home/user/

# Or if you have the source code, transfer the entire directory
scp -r . user@your-vps-ip:/home/user/evilginx_monitor/
```

### Option B: Using SFTP Client

Use tools like WinSCP, FileZilla, or VS Code's SFTP extension to transfer files.

## Step 3: SSH into Your VPS

```bash
ssh user@your-vps-ip
```

## Step 4: Set Up the Application on VPS

### Create Application Directory

```bash
sudo mkdir -p /opt/evilginx_monitor
sudo chown $USER:$USER /opt/evilginx_monitor
```

### Move Binary to Application Directory

```bash
mv evilginx_monitor /opt/evilginx_monitor/
chmod +x /opt/evilginx_monitor/evilginx_monitor
```

### Create Log Directory

```bash
sudo mkdir -p /var/log/evilginx_monitor
sudo chown $USER:$USER /var/log/evilginx_monitor
```

## Step 5: Initial Setup

Run the application once to create the config directory:

```bash
cd /opt/evilginx_monitor
./evilginx_monitor --help
```

This will create `~/.evilginx_monitor/config.json` in the user's home directory.

## Step 6: Configure the Application

Edit the configuration file:

```bash
nano ~/.evilginx_monitor/config.json
```

Or use the admin UI (recommended):

```bash
# Start with admin UI
./evilginx_monitor --admin --admin-port 8080
```

Then access `http://your-vps-ip:8080` to configure via web interface.

**Important:** Make sure to configure:
- Database file path (e.g., `/root/.evilginx/data.db`)
- Telegram/Discord/Email settings
- Enable desired notification channels

## Step 7: Create Systemd Service (Recommended)

Create a systemd service file for automatic startup and management:

```bash
sudo nano /etc/systemd/system/evilginx-monitor.service
```

Paste the following content (adjust paths as needed):

```ini
[Unit]
Description=Evilginx Monitor Service
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/opt/evilginx_monitor
ExecStart=/opt/evilginx_monitor/evilginx_monitor --admin --admin-port 8080
Restart=always
RestartSec=10
StandardOutput=append:/var/log/evilginx_monitor/output.log
StandardError=append:/var/log/evilginx_monitor/error.log

[Install]
WantedBy=multi-user.target
```

**Replace `YOUR_USERNAME` with your actual username!**

### Enable and Start the Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable evilginx-monitor

# Start the service
sudo systemctl start evilginx-monitor

# Check status
sudo systemctl status evilginx-monitor

# View logs
sudo journalctl -u evilginx-monitor -f
```

### Service Management Commands

```bash
# Start service
sudo systemctl start evilginx-monitor

# Stop service
sudo systemctl stop evilginx-monitor

# Restart service
sudo systemctl restart evilginx-monitor

# Check status
sudo systemctl status evilginx-monitor

# View logs
sudo journalctl -u evilginx-monitor -n 50
```

## Step 8: Configure Firewall

Allow the admin port through the firewall:

### UFW (Ubuntu)

```bash
sudo ufw allow 8080/tcp
sudo ufw reload
```

### firewalld (CentOS/RHEL)

```bash
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

## Step 9: Set Up Domain with Nginx (Recommended)

For production deployment, set up a domain name with SSL. See **[DOMAIN_SETUP.md](DOMAIN_SETUP.md)** for complete domain configuration guide.

### Quick Domain Setup:

1. **Configure DNS**: Point your domain to VPS IP
2. **Install Nginx**: `sudo apt install nginx`
3. **Get SSL Certificate**: `sudo certbot --nginx -d your-domain.com`
4. **Access**: `https://your-domain.com`

### Basic Reverse Proxy Setup:

### Install Nginx

```bash
sudo apt update
sudo apt install nginx
```

### Create Nginx Configuration

```bash
sudo nano /etc/nginx/sites-available/evilginx-monitor
```

Add the following (replace `your-domain.com` with your domain or use IP):

```nginx
server {
    listen 80;
    server_name your-domain.com;  # or your VPS IP

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Enable the Site

```bash
sudo ln -s /etc/nginx/sites-available/evilginx-monitor /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Set Up SSL with Let's Encrypt (Recommended)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## Step 10: Security Considerations

1. **Change Admin Port**: Use a non-standard port or restrict access
2. **Firewall Rules**: Only allow necessary ports
3. **SSL/TLS**: Use HTTPS for admin interface
4. **Authentication**: Consider adding authentication to admin UI (future enhancement)
5. **File Permissions**: Ensure config files have proper permissions

```bash
# Restrict config file permissions
chmod 600 ~/.evilginx_monitor/config.json
```

## Troubleshooting

### Check if service is running

```bash
sudo systemctl status evilginx-monitor
```

### View application logs

```bash
# Systemd logs
sudo journalctl -u evilginx-monitor -f

# Application logs
tail -f /var/log/evilginx_monitor/output.log
tail -f /var/log/evilginx_monitor/error.log
```

### Check if port is listening

```bash
sudo netstat -tlnp | grep 8080
# or
sudo ss -tlnp | grep 8080
```

### Test the application manually

```bash
cd /opt/evilginx_monitor
./evilginx_monitor --admin --admin-port 8080
```

### Check configuration

```bash
./evilginx_monitor --config
```

## Quick Reference

```bash
# Build for Linux (on Windows)
set GOOS=linux
set GOARCH=amd64
go build -o evilginx_monitor .

# Transfer to VPS
scp evilginx_monitor user@vps-ip:/home/user/

# On VPS: Install service
sudo cp evilginx-monitor.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable evilginx-monitor
sudo systemctl start evilginx-monitor

# View logs
sudo journalctl -u evilginx-monitor -f
```

## Notes

- The application creates its config in `~/.evilginx_monitor/config.json`
- Make sure the database file path is accessible by the user running the service
- If running as a service, ensure the user has read access to the evilginx database file
- The admin UI is accessible at `http://your-vps-ip:8080` (or your configured port)

