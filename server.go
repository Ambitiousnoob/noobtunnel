package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
	"os"
)

type ServerConfig struct {
	Port           int               `yaml:"port"`
	MaxConnections int               `yaml:"max_connections"`
	RateLimit      int               `yaml:"rate_limit"` // requests per minute
	TimeoutMinutes int               `yaml:"timeout_minutes"`
	AllowedPorts   []int             `yaml:"allowed_ports"`
	BannedIPs      []string          `yaml:"banned_ips"`
	LogLevel       string            `yaml:"log_level"`
	Security       ServerSecurity    `yaml:"security"`
}

type ServerSecurity struct {
	Enabled            bool `yaml:"enabled"`
	MaxConnectionsPerIP int `yaml:"max_connections_per_ip"`
	RateLimitPerIP     int `yaml:"rate_limit_per_ip"`
}

type Server struct {
	config        *ServerConfig
	listener      net.Listener
	tunnels       map[int]*Tunnel
	tunnelsMutex  sync.RWMutex
	connections   map[string]int
	connMutex     sync.RWMutex
	rateLimiter   map[string]*RateLimit
	rateMutex     sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

type Tunnel struct {
	Port       int
	Conn       net.Conn
	Listener   net.Listener
	ClientAddr string
	CreatedAt  time.Time
	ctx        context.Context
	cancel     context.CancelFunc
}

type RateLimit struct {
	Requests  int
	ResetTime time.Time
}

func startServer(port int) {
	config := &ServerConfig{
		Port:           port,
		MaxConnections: 100,
		RateLimit:      60, // 60 requests per minute
		TimeoutMinutes: 30,
		AllowedPorts:   []int{80, 8080, 3000, 3001, 8000, 8001, 9000},
		LogLevel:       "info",
		Security: ServerSecurity{
			Enabled:            true,
			MaxConnectionsPerIP: 5,
			RateLimitPerIP:     30,
		},
	}

	s := NewServer(config)
	s.Start()
}

func startServerWithConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	config := &ServerConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

	s := NewServer(config)
	s.Start()
}

func NewServer(config *ServerConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		config:      config,
		tunnels:     make(map[int]*Tunnel),
		connections: make(map[string]int),
		rateLimiter: make(map[string]*RateLimit),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
	s.listener = listener

	fmt.Printf("🚀 NoobTunnel Server started on port %d\n", s.config.Port)
	fmt.Printf("🔒 Security: %v | Max Connections: %d | Rate Limit: %d/min\n", 
		s.config.Security.Enabled, s.config.MaxConnections, s.config.RateLimit)
	fmt.Printf("🎯 Allowed Ports: %v\n", s.config.AllowedPorts)
	fmt.Println("📡 Waiting for client connections...")

	// Start cleanup routine
	go s.cleanupRoutine()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("⚠️ Failed to accept connection: %v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	host, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		log.Printf("⚠️ Invalid client address: %s", clientAddr)
		return
	}

	// Security checks
	if !s.checkSecurity(host) {
		log.Printf("🚫 Connection rejected from %s (security check failed)", host)
		return
	}

	// Rate limiting
	if !s.checkRateLimit(host) {
		log.Printf("⏰ Connection rate limited from %s", host)
		return
	}

	s.connMutex.Lock()
	s.connections[host]++
	connCount := s.connections[host]
	s.connMutex.Unlock()

	log.Printf("✅ New connection from %s (total: %d)", clientAddr, connCount)

	// Read tunnel request
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("⚠️ Failed to read from client %s: %v", clientAddr, err)
		return
	}

	request := string(buffer[:n])
	var remotePort int
	if _, err := fmt.Sscanf(request, "TUNNEL %d", &remotePort); err != nil {
		log.Printf("⚠️ Invalid tunnel request from %s: %s", clientAddr, request)
		return
	}

	// Check if port is allowed
	if !s.isPortAllowed(remotePort) {
		response := fmt.Sprintf("ERROR Port %d not allowed", remotePort)
		conn.Write([]byte(response))
		log.Printf("🚫 Port %d not allowed for %s", remotePort, clientAddr)
		return
	}

	// Check if port is already in use
	s.tunnelsMutex.Lock()
	if existingTunnel, exists := s.tunnels[remotePort]; exists {
		s.tunnelsMutex.Unlock()
		response := fmt.Sprintf("ERROR Port %d already in use by %s", remotePort, existingTunnel.ClientAddr)
		conn.Write([]byte(response))
		log.Printf("⚠️ Port %d already in use, request from %s denied", remotePort, clientAddr)
		return
	}
	s.tunnelsMutex.Unlock()

	// Create tunnel
	tunnel, err := s.createTunnel(remotePort, conn, clientAddr)
	if err != nil {
		response := fmt.Sprintf("ERROR %s", err.Error())
		conn.Write([]byte(response))
		log.Printf("❌ Failed to create tunnel for %s: %v", clientAddr, err)
		return
	}

	response := fmt.Sprintf("OK Tunnel established on port %d", remotePort)
	conn.Write([]byte(response))
	log.Printf("🎯 Tunnel created: %s -> port %d", clientAddr, remotePort)

	// Keep connection alive and handle tunnel
	s.handleTunnel(tunnel)
}

func (s *Server) createTunnel(port int, clientConn net.Conn, clientAddr string) (*Tunnel, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	ctx, cancel := context.WithCancel(s.ctx)
	tunnel := &Tunnel{
		Port:       port,
		Conn:       clientConn,
		Listener:   listener,
		ClientAddr: clientAddr,
		CreatedAt:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	s.tunnelsMutex.Lock()
	s.tunnels[port] = tunnel
	s.tunnelsMutex.Unlock()

	// Start accepting connections for this tunnel
	go s.acceptTunnelConnections(tunnel)

	return tunnel, nil
}

func (s *Server) acceptTunnelConnections(tunnel *Tunnel) {
	for {
		select {
		case <-tunnel.ctx.Done():
			return
		default:
		}

		conn, err := tunnel.Listener.Accept()
		if err != nil {
			select {
			case <-tunnel.ctx.Done():
				return
			default:
				log.Printf("⚠️ Failed to accept connection on tunnel port %d: %v", tunnel.Port, err)
				continue
			}
		}

		go s.handleTunnelConnection(tunnel, conn)
	}
}

func (s *Server) handleTunnelConnection(tunnel *Tunnel, publicConn net.Conn) {
	defer publicConn.Close()

	// Send connection signal to client
	signal := fmt.Sprintf("CONN %d", tunnel.Port)
	if _, err := tunnel.Conn.Write([]byte(signal)); err != nil {
		log.Printf("⚠️ Failed to signal client for port %d: %v", tunnel.Port, err)
		return
	}

	// Relay data between public connection and tunnel connection
	go func() {
		io.Copy(tunnel.Conn, publicConn)
		publicConn.Close()
	}()

	io.Copy(publicConn, tunnel.Conn)
}

func (s *Server) handleTunnel(tunnel *Tunnel) {
	defer s.cleanupTunnel(tunnel)

	// Keep connection alive
	buffer := make([]byte, 1024)
	for {
		select {
		case <-tunnel.ctx.Done():
			return
		default:
		}

		tunnel.Conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.TimeoutMinutes) * time.Minute))
		_, err := tunnel.Conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("⏰ Tunnel timeout for port %d from %s", tunnel.Port, tunnel.ClientAddr)
			} else {
				log.Printf("🔌 Client disconnected from port %d (%s)", tunnel.Port, tunnel.ClientAddr)
			}
			return
		}
	}
}

func (s *Server) cleanupTunnel(tunnel *Tunnel) {
	tunnel.cancel()
	tunnel.Listener.Close()
	tunnel.Conn.Close()

	s.tunnelsMutex.Lock()
	delete(s.tunnels, tunnel.Port)
	s.tunnelsMutex.Unlock()

	// Update connection count
	host, _, _ := net.SplitHostPort(tunnel.ClientAddr)
	s.connMutex.Lock()
	if s.connections[host] > 0 {
		s.connections[host]--
	}
	if s.connections[host] == 0 {
		delete(s.connections, host)
	}
	s.connMutex.Unlock()

	log.Printf("🧹 Tunnel cleaned up: port %d from %s", tunnel.Port, tunnel.ClientAddr)
}

func (s *Server) checkSecurity(host string) bool {
	if !s.config.Security.Enabled {
		return true
	}

	// Check banned IPs
	for _, bannedIP := range s.config.BannedIPs {
		if host == bannedIP {
			return false
		}
	}

	// Check max connections per IP
	s.connMutex.RLock()
	connCount := s.connections[host]
	s.connMutex.RUnlock()

	if connCount >= s.config.Security.MaxConnectionsPerIP {
		return false
	}

	return true
}

func (s *Server) checkRateLimit(host string) bool {
	if !s.config.Security.Enabled {
		return true
	}

	s.rateMutex.Lock()
	defer s.rateMutex.Unlock()

	rate, exists := s.rateLimiter[host]
	now := time.Now()

	if !exists || now.After(rate.ResetTime) {
		s.rateLimiter[host] = &RateLimit{
			Requests:  1,
			ResetTime: now.Add(time.Minute),
		}
		return true
	}

	if rate.Requests >= s.config.Security.RateLimitPerIP {
		return false
	}

	rate.Requests++
	return true
}

func (s *Server) isPortAllowed(port int) bool {
	if len(s.config.AllowedPorts) == 0 {
		return true
	}

	for _, allowedPort := range s.config.AllowedPorts {
		if port == allowedPort {
			return true
		}
	}
	return false
}

func (s *Server) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.rateMutex.Lock()
			now := time.Now()
			for host, rate := range s.rateLimiter {
				if now.After(rate.ResetTime.Add(5 * time.Minute)) {
					delete(s.rateLimiter, host)
				}
			}
			s.rateMutex.Unlock()

			// Log active tunnels
			s.tunnelsMutex.RLock()
			activeTunnels := len(s.tunnels)
			s.tunnelsMutex.RUnlock()

			s.connMutex.RLock()
			activeIPs := len(s.connections)
			s.connMutex.RUnlock()

			log.Printf("📊 Status: %d active tunnels, %d unique IPs connected", activeTunnels, activeIPs)
		}
	}
}