#!/bin/bash

# NoobTunnel Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/Ambitiousnoob/noobtunnel/main/scripts/install.sh | bash

set -e

echo "🚀 Installing NoobTunnel..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    armv7l) ARCH="armv7" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "📋 Detected OS: $OS, Architecture: $ARCH"

# Check if Go is installed
if command -v go &> /dev/null; then
    echo "✅ Go is installed, building from source..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    
    # Clone repository
    echo "📥 Cloning repository..."
    git clone https://github.com/Ambitiousnoob/noobtunnel.git
    cd noobtunnel
    
    # Build binary
    echo "🔨 Building binary..."
    go mod tidy
    go build -ldflags "-s -w" -o ntunnel
    
    # Install binary
    echo "📦 Installing binary..."
    sudo mv ntunnel /usr/local/bin/
    sudo chmod +x /usr/local/bin/ntunnel
    
    # Cleanup
    rm -rf $TEMP_DIR
    
else
    echo "❌ Go not found. Please install Go first:"
    echo "   https://golang.org/doc/install"
    echo "   Or install using your package manager:"
    echo "   - Ubuntu/Debian: sudo apt install golang-go"
    echo "   - CentOS/RHEL: sudo yum install golang"
    echo "   - macOS: brew install go"
    exit 1
fi

# Verify installation
if command -v ntunnel &> /dev/null; then
    echo "✅ NoobTunnel installed successfully!"
    echo "📖 Usage:"
    echo "   ntunnel --help"
    echo "   ntunnel --version"
    echo ""
    echo "🖥️  Server mode:"
    echo "   ntunnel --mode server --port 7000"
    echo ""
    echo "💻 Client mode:"
    echo "   ntunnel --mode client --server YOUR_VPS_IP:7000 --local-port 3000 --remote-port 80"
    echo ""
    echo "📚 Documentation: https://github.com/Ambitiousnoob/noobtunnel"
else
    echo "❌ Installation failed. Please check for errors above."
    exit 1
fi
