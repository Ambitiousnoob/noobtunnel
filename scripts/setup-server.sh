#!/bin/bash

# NoobTunnel Server Setup Script for VPS
# This script sets up NoobTunnel as a systemd service

set -e

echo "🚀 Setting up NoobTunnel Server..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Check if ntunnel is installed
if ! command -v ntunnel &> /dev/null; then
    echo "❌ ntunnel not found. Please install it first:"
    echo "   curl -sSL https://raw.githubusercontent.com/Ambitiousnoob/noobtunnel/main/scripts/install.sh | bash"
    exit 1
fi

# Create ntunnel user
if ! id "ntunnel" &>/dev/null; then
    useradd -r -s /bin/false ntunnel
    echo "✅ Created ntunnel user"
fi

# Create configuration directory
mkdir -p /etc/ntunnel
chown ntunnel:ntunnel /etc/ntunnel
echo "✅ Created configuration directory: /etc/ntunnel"

# Create default server configuration
cat > /etc/ntunnel/server.yaml << EOF
# NoobTunnel Server Configuration
port: 7000
max_connections: 100
rate_limit: 60
timeout_minutes: 30
allowed_ports:
  - 80
  - 8080
  - 3000
  - 3001
  - 8000
  - 8001
  - 9000
  - 5000
  - 4000
banned_ips: []
log_level: info
security:
  enabled: true
  max_connections_per_ip: 5
  rate_limit_per_ip: 30
EOF

chown ntunnel:ntunnel /etc/ntunnel/server.yaml
echo "✅ Created server configuration: /etc/ntunnel/server.yaml"

# Create systemd service
cat > /etc/systemd/system/ntunnel-server.service << EOF
[Unit]
Description=NoobTunnel Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=ntunnel
Group=ntunnel
ExecStart=/usr/local/bin/ntunnel --mode server --config /etc/ntunnel/server.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ntunnel-server

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/tmp

[Install]
WantedBy=multi-user.target
EOF

echo "✅ Created systemd service: /etc/systemd/system/ntunnel-server.service"

# Reload systemd and enable service
systemctl daemon-reload
systemctl enable ntunnel-server
echo "✅ Enabled ntunnel-server service"

# Configure firewall (if ufw is available)
if command -v ufw &> /dev/null; then
    echo "🔥 Configuring firewall..."
    ufw allow 7000/tcp comment "NoobTunnel Server"
    ufw allow 80/tcp comment "NoobTunnel HTTP"
    ufw allow 8080/tcp comment "NoobTunnel HTTP Alt"
    ufw allow 3000:9000/tcp comment "NoobTunnel Dev Ports"
    echo "✅ Firewall rules added"
fi

# Start the service
systemctl start ntunnel-server
echo "✅ Started ntunnel-server service"

# Show status
echo ""
echo "🎉 NoobTunnel Server setup completed!"
echo ""
echo "📊 Service Status:"
systemctl status ntunnel-server --no-pager -l
echo ""
echo "📝 Configuration file: /etc/ntunnel/server.yaml"
echo "📋 Service logs: journalctl -u ntunnel-server -f"
echo ""
echo "🔗 Your server is running on port 7000"
echo "💻 Clients can connect using:"
echo "   ntunnel --mode client --server $(hostname -I | awk '{print $1}'):7000 --local-port 3000 --remote-port 80"
echo ""
echo "⚙️  Service management:"
echo "   sudo systemctl start ntunnel-server"
echo "   sudo systemctl stop ntunnel-server"
echo "   sudo systemctl restart ntunnel-server"
echo "   sudo systemctl status ntunnel-server"
