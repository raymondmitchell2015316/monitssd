@echo off
echo Building Evilginx Monitor for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o evilginx_monitor .
echo Build complete! Binary: evilginx_monitor
echo Transfer this file to your VPS using SCP or SFTP

