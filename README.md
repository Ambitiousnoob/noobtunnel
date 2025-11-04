# NoobTunnel 🚀

**A secure, lightweight, and user-friendly tunneling tool for exposing private services through a public VPS.**

NoobTunnel allows you to easily expose your local services (web servers, APIs, databases, etc.) running behind NAT, firewalls, or private networks to the internet via your public VPS. Think of it as your personal ngrok alternative!

## ✨ Features

- 🔒 **Secure by Design** - Built-in rate limiting, connection limits, and IP-based restrictions
- 🚀 **Easy to Use** - Simple command-line interface with sensible defaults
- ⚡ **Fast & Lightweight** - Written in Go for optimal performance
- 🛡️ **VPS Protection** - Advanced security features to protect your server from abuse
- 🔧 **Configurable** - YAML configuration support for advanced setups
- 🔄 **Auto-Reconnect** - Automatic reconnection with exponential backoff
- 📊 **Monitoring** - Built-in logging and connection monitoring
- 🎯 **Port Management** - Whitelist allowed ports for enhanced security

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Ambitiousnoob/noobtunnel.git
cd noobtunnel

# Build the binary
go build -o ntunnel

# Make it executable (Linux/Mac)
chmod +x ntunnel

# Optional: Move to PATH for global access
sudo mv ntunnel /usr/local/bin/
```

### Server Setup (Your Public VPS)

```bash
# Start the server on default port 7000
./ntunnel --mode server --port 7000

# Or with custom port
./ntunnel --mode server --port 8080
```

### Client Setup (Your Local Machine)

```bash
# Expose local port 3000 as public port 80
./ntunnel --mode client --server YOUR_VPS_IP:7000 --local-port 3000 --remote-port 80

# Expose local port 8080 as public port 8080
./ntunnel --mode client --server YOUR_VPS_IP:7000 --local-port 8080 --remote-port 8080
```

## 📖 Usage

### Basic Commands

```bash
# Show help
./ntunnel --help

# Show version
./ntunnel --version

# Server mode
./ntunnel --mode server --port 7000

# Client mode
./ntunnel --mode client --server 1.2.3.4:7000 --local-port 3000 --remote-port 80
```

### Configuration Files

For advanced setups, you can use YAML configuration files:

#### Server Configuration (`server.yaml`)

```yaml
port: 7000
max_connections: 100
rate_limit: 60  # requests per minute
timeout_minutes: 30
allowed_ports:
  - 80
  - 8080
  - 3000
  - 3001
  - 8000
  - 8001
  - 9000
banned_ips:
  - 192.168.1.100
log_level: info
security:
  enabled: true
  max_connections_per_ip: 5
  rate_limit_per_ip: 30
```

#### Client Configuration (`client.yaml`)

```yaml
server: "your-vps-ip:7000"
reconnect: true
reconnect_delay: 5
log_level: info
tunnels:
  webapp:
    local_port: 3000
    remote_port: 80
    local_host: "127.0.0.1"
  api:
    local_port: 8080
    remote_port: 8080
    local_host: "127.0.0.1"
```

Run with configuration:

```bash
# Server with config
./ntunnel --mode server --config server.yaml

# Client with config  
./ntunnel --mode client --config client.yaml
```

## 🛡️ Security Features

### Server Protection
- **Rate Limiting**: Prevents spam connections (configurable per IP)
- **Connection Limits**: Maximum connections per IP address
- **Port Whitelisting**: Only allow specific ports to be tunneled
- **IP Banning**: Block specific IP addresses
- **Connection Timeouts**: Automatic cleanup of idle connections
- **Resource Monitoring**: Track active tunnels and connections

### Built-in Safeguards
- No authentication required (passwordless)
- Automatic connection cleanup
- Memory-efficient connection handling
- Protection against port conflicts
- Graceful error handling and recovery

## 🌟 Use Cases

- **Web Development**: Share your local development server with clients
- **API Testing**: Expose APIs for webhook testing
- **Home Lab**: Access home services from anywhere
- **IoT Projects**: Connect IoT devices behind NAT
- **Game Servers**: Host game servers from home
- **File Sharing**: Temporary file sharing solutions
- **Remote Access**: Access internal services securely

## 📊 Example Scenarios

### Scenario 1: Web Development
```bash
# Local React app on port 3000, expose as port 80
ntunnel --mode client --server vps.example.com:7000 --local-port 3000 --remote-port 80
# Now accessible at: http://vps.example.com
```

### Scenario 2: API Development
```bash
# Local API on port 8080, expose as port 8080
ntunnel --mode client --server vps.example.com:7000 --local-port 8080 --remote-port 8080
# Now accessible at: http://vps.example.com:8080
```

### Scenario 3: Home Lab Access
```bash
# Local home assistant on port 8123, expose as port 8123
ntunnel --mode client --server vps.example.com:7000 --local-port 8123 --remote-port 8123
# Now accessible at: http://vps.example.com:8123
```

## 🔧 Advanced Configuration

### Environment Variables
```bash
export NTUNNEL_SERVER="vps.example.com:7000"
export NTUNNEL_LOCAL_PORT="3000"
export NTUNNEL_REMOTE_PORT="80"
```

### Systemd Service (Linux)

Create `/etc/systemd/system/ntunnel-client.service`:

```ini
[Unit]
Description=NoobTunnel Client
After=network.target

[Service]
Type=simple
User=tunnel
ExecStart=/usr/local/bin/ntunnel --mode client --config /etc/ntunnel/client.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable ntunnel-client
sudo systemctl start ntunnel-client
```

## 🐛 Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if server is running
   - Verify firewall settings
   - Ensure correct port and IP

2. **Port Already in Use**
   - Choose a different remote port
   - Check what's using the port: `netstat -tulpn | grep :PORT`

3. **Rate Limited**
   - Wait for rate limit reset
   - Adjust server configuration if you control it

4. **Local Service Unreachable**
   - Verify local service is running
   - Check local firewall settings
   - Test local connection: `telnet 127.0.0.1 PORT`

### Debug Mode
```bash
# Enable verbose logging
./ntunnel --mode client --server vps:7000 --local-port 3000 --remote-port 80 -v
```

## 📈 Performance

- **Lightweight**: ~10MB memory footprint
- **Fast**: Native Go performance
- **Scalable**: Handles hundreds of concurrent connections
- **Efficient**: Minimal CPU usage

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⭐ Support

If you find NoobTunnel helpful, please consider giving it a star! ⭐

## 📞 Contact

- GitHub: [@Ambitiousnoob](https://github.com/Ambitiousnoob)
- Issues: [Report bugs or request features](https://github.com/Ambitiousnoob/noobtunnel/issues)

---

**Made with ❤️ for developers who need simple, secure tunneling solutions.**