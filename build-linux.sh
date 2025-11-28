#!/bin/bash
echo "Building Evilginx Monitor for Linux..."
export GOOS=linux
export GOARCH=amd64
go build -o evilginx_monitor .
echo "Build complete! Binary: evilginx_monitor"
echo "Transfer this file to your VPS using SCP or SFTP"

