# Domain Setup Guide

This guide will help you set up your Evilginx Monitor with a custom domain name instead of using IP:port.

## Prerequisites

- A VPS with the application already deployed
- A domain name (e.g., `monitor.example.com`)
- DNS access to configure A record
- Root or sudo access on your VPS

## Step 1: Configure DNS

Point your domain to your VPS IP address:

1. Log into your domain registrar's DNS management panel
2. Add an **A Record**:
   - **Name/Host**: `monitor` (or `@` for root domain)
   - **Type**: `A`
   - **Value/IP**: Your VPS IP address
   - **TTL**: 3600 (or default)

**Example:**
- Domain: `example.com`
- Subdomain: `monitor.example.com`
- A Record: `monitor` → `123.45.67.89`

### Verify DNS

Wait a few minutes, then verify DNS propagation:

```bash
# Check if DNS is resolving
nslookup monitor.example.com
# or
dig monitor.example.com
```

## Step 2: Install Nginx

```bash
sudo apt update
sudo apt install nginx -y
sudo systemctl enable nginx
sudo systemctl start nginx
```

## Step 3: Configure Nginx

### Create Nginx Configuration File

```bash
sudo nano /etc/nginx/sites-available/evilginx-monitor
```

### Configuration for HTTP (Port 80)

Paste the following configuration (replace `monitor.example.com` with your domain):

```nginx
server {
    listen 80;
    server_name monitor.example.com;  # Replace with your domain

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy settings
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        
        # WebSocket support (if needed in future)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        
        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffer settings
        proxy_buffering off;
        proxy_cache_bypass $http_upgrade;
    }

    # Logging
    access_log /var/log/nginx/evilginx-monitor-access.log;
    error_log /var/log/nginx/evilginx-monitor-error.log;
}
```

### Enable the Site

```bash
# Create symbolic link
sudo ln -s /etc/nginx/sites-available/evilginx-monitor /etc/nginx/sites-enabled/

# Remove default site (optional)
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

## Step 4: Set Up SSL with Let's Encrypt

### Install Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
```

### Obtain SSL Certificate

```bash
# Replace with your domain
sudo certbot --nginx -d monitor.example.com
```

Certbot will:
1. Automatically configure SSL
2. Set up automatic renewal
3. Redirect HTTP to HTTPS

### Verify Auto-Renewal

```bash
# Test renewal
sudo certbot renew --dry-run

# Check renewal status
sudo systemctl status certbot.timer
```

## Step 5: Update Firewall

```bash
# Allow HTTP and HTTPS
sudo ufw allow 'Nginx Full'
# or specifically:
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Reload firewall
sudo ufw reload
```

## Step 6: Final Nginx Configuration (After SSL)

After running certbot, your configuration will be automatically updated. The final config should look like:

```nginx
server {
    listen 80;
    server_name monitor.example.com;
    return 301 https://$server_name$request_uri;  # Redirect HTTP to HTTPS
}

server {
    listen 443 ssl http2;
    server_name monitor.example.com;

    ssl_certificate /etc/letsencrypt/live/monitor.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/monitor.example.com/privkey.pem;
    
    # SSL Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        proxy_buffering off;
        proxy_cache_bypass $http_upgrade;
    }

    access_log /var/log/nginx/evilginx-monitor-access.log;
    error_log /var/log/nginx/evilginx-monitor-error.log;
}
```

## Step 7: Update Systemd Service (Optional)

You can update the service to bind only to localhost since Nginx will handle external access:

The service file is already correct - it runs on localhost:8080, which is what we want.

## Step 8: Access Your Application

Now you can access your admin panel at:
- **HTTPS**: `https://monitor.example.com`
- HTTP will automatically redirect to HTTPS

## Troubleshooting

### Domain Not Resolving

```bash
# Check DNS
nslookup monitor.example.com
dig monitor.example.com

# Check if Nginx is listening
sudo netstat -tlnp | grep nginx
```

### SSL Certificate Issues

```bash
# Check certificate status
sudo certbot certificates

# Renew manually if needed
sudo certbot renew

# Check Nginx SSL configuration
sudo nginx -t
```

### Nginx Not Starting

```bash
# Check Nginx status
sudo systemctl status nginx

# Check error logs
sudo tail -f /var/log/nginx/error.log

# Test configuration
sudo nginx -t
```

### Application Not Accessible

```bash
# Check if application is running
sudo systemctl status evilginx-monitor

# Check if port 8080 is listening
sudo netstat -tlnp | grep 8080

# Test local connection
curl http://localhost:8080
```

### Check Nginx Logs

```bash
# Access logs
sudo tail -f /var/log/nginx/evilginx-monitor-access.log

# Error logs
sudo tail -f /var/log/nginx/evilginx-monitor-error.log
```

## Security Recommendations

1. **Firewall**: Only allow ports 80, 443, and SSH (22)
   ```bash
   sudo ufw allow 22/tcp
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. **Fail2Ban**: Install to protect against brute force attacks
   ```bash
   sudo apt install fail2ban -y
   ```

3. **Rate Limiting**: Add to Nginx config to prevent abuse
   ```nginx
   limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
   
   location / {
       limit_req zone=api_limit burst=20 nodelay;
       # ... rest of config
   }
   ```

4. **IP Whitelist** (Optional): Restrict access to specific IPs
   ```nginx
   location / {
       allow 1.2.3.4;  # Your IP
       deny all;
       # ... proxy settings
   }
   ```

## Multiple Domains

If you want to use multiple domains or subdomains:

```nginx
server {
    listen 443 ssl http2;
    server_name monitor.example.com monitor2.example.com;
    
    # ... rest of config
}
```

Or create separate server blocks for each domain.

## Custom Port (Alternative)

If you prefer to keep the application on a custom port but still use a domain:

```nginx
location / {
    proxy_pass http://localhost:9000;  # Your custom port
    # ... rest of config
}
```

Then update the systemd service file to use port 9000.

## Summary

After completing these steps:
- ✅ Domain points to your VPS
- ✅ Nginx handles reverse proxy
- ✅ SSL certificate installed (HTTPS)
- ✅ HTTP redirects to HTTPS
- ✅ Access via `https://monitor.example.com`

Your application is now accessible via a secure domain name! 🎉

