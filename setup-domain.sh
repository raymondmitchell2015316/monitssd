#!/bin/bash

# Domain Setup Script for Evilginx Monitor
# Usage: ./setup-domain.sh your-domain.com

set -e

if [ -z "$1" ]; then
    echo "Usage: ./setup-domain.sh your-domain.com"
    exit 1
fi

DOMAIN=$1
NGINX_CONFIG="/etc/nginx/sites-available/evilginx-monitor"

echo "🚀 Setting up domain: $DOMAIN"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Step 1: Install Nginx if not installed
if ! command -v nginx &> /dev/null; then
    echo "📦 Installing Nginx..."
    apt update
    apt install -y nginx
    systemctl enable nginx
    systemctl start nginx
    echo "✅ Nginx installed"
else
    echo "✅ Nginx already installed"
fi

# Step 2: Create Nginx configuration
echo "📝 Creating Nginx configuration..."
cat > $NGINX_CONFIG <<EOF
# HTTP to HTTPS Redirect
server {
    listen 80;
    server_name $DOMAIN;
    return 301 https://\$server_name\$request_uri;
}

# HTTPS Server
server {
    listen 443 ssl http2;
    server_name $DOMAIN;

    # SSL will be configured by certbot
    # ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        proxy_buffering off;
        proxy_cache_bypass \$http_upgrade;
    }

    access_log /var/log/nginx/evilginx-monitor-access.log;
    error_log /var/log/nginx/evilginx-monitor-error.log;
}
EOF

# Step 3: Enable site
echo "🔗 Enabling Nginx site..."
ln -sf $NGINX_CONFIG /etc/nginx/sites-enabled/evilginx-monitor

# Remove default site if exists
if [ -f /etc/nginx/sites-enabled/default ]; then
    rm /etc/nginx/sites-enabled/default
fi

# Test configuration
echo "🧪 Testing Nginx configuration..."
nginx -t

# Reload Nginx
systemctl reload nginx
echo "✅ Nginx configured"

# Step 4: Install Certbot if not installed
if ! command -v certbot &> /dev/null; then
    echo "📦 Installing Certbot..."
    apt install -y certbot python3-certbot-nginx
    echo "✅ Certbot installed"
else
    echo "✅ Certbot already installed"
fi

# Step 5: Configure firewall
echo "🔥 Configuring firewall..."
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full' 2>/dev/null || ufw allow 80/tcp && ufw allow 443/tcp
    echo "✅ Firewall configured"
else
    echo "⚠️  UFW not found, please configure firewall manually"
fi

# Step 6: Obtain SSL certificate
echo ""
echo "🔐 Obtaining SSL certificate..."
echo "⚠️  Make sure your domain DNS is pointing to this server!"
read -p "Press Enter to continue or Ctrl+C to cancel..."
echo ""

certbot --nginx -d $DOMAIN --non-interactive --agree-tos --register-unsafely-without-email || {
    echo "❌ SSL certificate setup failed"
    echo "Please run manually: sudo certbot --nginx -d $DOMAIN"
    exit 1
}

# Step 7: Test auto-renewal
echo "🔄 Testing certificate auto-renewal..."
certbot renew --dry-run

echo ""
echo "✅ Domain setup complete!"
echo ""
echo "🌐 Access your admin panel at: https://$DOMAIN"
echo ""
echo "📋 Useful commands:"
echo "   Check Nginx status: sudo systemctl status nginx"
echo "   Check SSL cert: sudo certbot certificates"
echo "   View logs: sudo tail -f /var/log/nginx/evilginx-monitor-access.log"
echo ""

