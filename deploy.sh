#!/bin/bash

# Deployment script for Evilginx Monitor
# Usage: ./deploy.sh user@vps-ip

if [ -z "$1" ]; then
    echo "Usage: ./deploy.sh user@vps-ip"
    exit 1
fi

VPS=$1
APP_NAME="evilginx_monitor"
APP_DIR="/opt/$APP_NAME"
SERVICE_FILE="evilginx-monitor.service"

echo "Building for Linux..."
export GOOS=linux
export GOARCH=amd64
go build -o $APP_NAME .

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Transferring files to VPS..."
scp $APP_NAME $VPS:/tmp/
scp $SERVICE_FILE $VPS:/tmp/

echo "Setting up on VPS..."
ssh $VPS << 'ENDSSH'
    sudo mkdir -p /opt/evilginx_monitor
    sudo mkdir -p /var/log/evilginx_monitor
    sudo mv /tmp/evilginx_monitor /opt/evilginx_monitor/
    sudo chmod +x /opt/evilginx_monitor/evilginx_monitor
    sudo chown $USER:$USER /opt/evilginx_monitor
    sudo chown $USER:$USER /var/log/evilginx_monitor
    
    # Update service file with current user
    sed "s/YOUR_USERNAME/$USER/g" /tmp/evilginx-monitor.service > /tmp/evilginx-monitor.service.tmp
    sudo mv /tmp/evilginx-monitor.service.tmp /etc/systemd/system/evilginx-monitor.service
    
    # Reload and start service
    sudo systemctl daemon-reload
    sudo systemctl enable evilginx-monitor
    sudo systemctl restart evilginx-monitor
    
    echo "Deployment complete!"
    echo "Check status with: sudo systemctl status evilginx-monitor"
    echo "View logs with: sudo journalctl -u evilginx-monitor -f"
ENDSSH

echo "Deployment finished!"
echo "Access admin UI at: http://$(echo $VPS | cut -d@ -f2):8080"

