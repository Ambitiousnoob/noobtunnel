#!/bin/bash

# NoobTunnel Uninstallation Script

echo "🗑️  Uninstalling NoobTunnel..."

# Remove binary
if [ -f "/usr/local/bin/ntunnel" ]; then
    sudo rm -f /usr/local/bin/ntunnel
    echo "✅ Removed binary: /usr/local/bin/ntunnel"
fi

# Remove systemd service if exists
if [ -f "/etc/systemd/system/ntunnel-server.service" ]; then
    sudo systemctl stop ntunnel-server 2>/dev/null || true
    sudo systemctl disable ntunnel-server 2>/dev/null || true
    sudo rm -f /etc/systemd/system/ntunnel-server.service
    echo "✅ Removed systemd service: ntunnel-server"
fi

if [ -f "/etc/systemd/system/ntunnel-client.service" ]; then
    sudo systemctl stop ntunnel-client 2>/dev/null || true
    sudo systemctl disable ntunnel-client 2>/dev/null || true
    sudo rm -f /etc/systemd/system/ntunnel-client.service
    echo "✅ Removed systemd service: ntunnel-client"
fi

# Reload systemd
sudo systemctl daemon-reload 2>/dev/null || true

echo "✅ NoobTunnel uninstalled successfully!"
