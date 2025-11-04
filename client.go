package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type ClientConfig struct {
	Server      string            `yaml:"server"`
	Tunnels     map[string]Tunnel `yaml:"tunnels"`
	Reconnect   bool              `yaml:"reconnect"`
	ReconnectDelay int            `yaml:"reconnect_delay"`
	LogLevel    string            `yaml:"log_level"`
}

type TunnelConfig struct {
	LocalPort  int    `yaml:"local_port"`
	RemotePort int    `yaml:"remote_port"`
	LocalHost  string `yaml:"local_host"`
}

type Client struct {
	config *ClientConfig
	conn   net.Conn
}

func startClient(server string, localPort, remotePort int) {
	client := &Client{}
	client.connectAndTunnel(server, localPort, remotePort, "127.0.0.1")
}

func startClientWithConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	config := &ClientConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

	client := &Client{config: config}
	client.startWithConfig()
}

func (c *Client) startWithConfig() {
	if len(c.config.Tunnels) == 0 {
		log.Fatal("❌ No tunnels configured")
	}

	fmt.Printf("🔗 Connecting to server: %s\n", c.config.Server)
	fmt.Printf("📝 Configured tunnels: %d\n", len(c.config.Tunnels))

	for name, tunnel := range c.config.Tunnels {
		fmt.Printf("  📍 %s: %s:%d -> :%d\n", name, tunnel.LocalHost, tunnel.LocalPort, tunnel.RemotePort)
	}

	// For simplicity, start the first tunnel
	// In a production version, you'd handle multiple tunnels concurrently
	for name, tunnel := range c.config.Tunnels {
		fmt.Printf("\n🚀 Starting tunnel: %s\n", name)
		c.connectAndTunnel(c.config.Server, tunnel.LocalPort, tunnel.RemotePort, tunnel.LocalHost)
		break
	}
}

func (c *Client) connectAndTunnel(server string, localPort, remotePort int, localHost string) {
	if localHost == "" {
		localHost = "127.0.0.1"
	}

	for {
		fmt.Printf("🔌 Connecting to %s...\n", server)
		conn, err := net.DialTimeout("tcp", server, 10*time.Second)
		if err != nil {
			log.Printf("❌ Failed to connect to server: %v", err)
			log.Println("⏳ Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		c.conn = conn
		fmt.Printf("✅ Connected to server %s\n", server)

		// Send tunnel request
		request := fmt.Sprintf("TUNNEL %d", remotePort)
		if _, err := conn.Write([]byte(request)); err != nil {
			log.Printf("❌ Failed to send tunnel request: %v", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		// Read response
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("❌ Failed to read server response: %v", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		response := string(buffer[:n])
		if strings.HasPrefix(response, "ERROR") {
			log.Printf("❌ Server error: %s", response[6:])
			conn.Close()
			return
		}

		if strings.HasPrefix(response, "OK") {
			fmt.Printf("🎯 %s\n", response[3:])
			fmt.Printf("🌐 Your local service %s:%d is now accessible via the server on port %d\n", localHost, localPort, remotePort)
			fmt.Println("📡 Tunnel is active, waiting for connections...")
		} else {
			log.Printf("⚠️ Unexpected response: %s", response)
		}

		// Handle tunnel connections
		if err := c.handleTunnel(localHost, localPort); err != nil {
			log.Printf("❌ Tunnel error: %v", err)
		}

		conn.Close()
		fmt.Println("🔌 Disconnected from server")

		// Check if reconnection is enabled
		if c.config != nil && c.config.Reconnect {
			delay := 5
			if c.config.ReconnectDelay > 0 {
				delay = c.config.ReconnectDelay
			}
			fmt.Printf("🔄 Reconnecting in %d seconds...\n", delay)
			time.Sleep(time.Duration(delay) * time.Second)
		} else {
			fmt.Println("⏳ Reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (c *Client) handleTunnel(localHost string, localPort int) error {
	buffer := make([]byte, 1024)

	for {
		// Wait for connection signal from server
		n, err := c.conn.Read(buffer)
		if err != nil {
			return fmt.Errorf("connection lost: %v", err)
		}

		signal := string(buffer[:n])
		if !strings.HasPrefix(signal, "CONN") {
			continue
		}

		// Extract port from signal
		parts := strings.Split(signal, " ")
		if len(parts) != 2 {
			continue
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		log.Printf("📞 Incoming connection on port %d", port)

		// Connect to local service
		localAddr := fmt.Sprintf("%s:%d", localHost, localPort)
		localConn, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
		if err != nil {
			log.Printf("❌ Failed to connect to local service %s: %v", localAddr, err)
			continue
		}

		log.Printf("🔗 Connected to local service %s", localAddr)

		// Relay data between local service and tunnel
		go func() {
			io.Copy(c.conn, localConn)
			localConn.Close()
		}()

		go func() {
			io.Copy(localConn, c.conn)
			c.conn.Close()
		}()
	}
}