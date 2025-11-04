package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	Version = "1.0.0"
	Banner  = `
███╗   ██╗ ██████╗  ██████╗ ██████╗ ████████╗██╗   ██╗███╗   ██╗███╗   ██╗███████╗██╗     
████╗  ██║██╔═══██╗██╔═══██╗██╔══██╗╚══██╔══╝██║   ██║████╗  ██║████╗  ██║██╔════╝██║     
██╔██╗ ██║██║   ██║██║   ██║██████╔╝   ██║   ██║   ██║██╔██╗ ██║██╔██╗ ██║█████╗  ██║     
██║╚██╗██║██║   ██║██║   ██║██╔══██╗   ██║   ██║   ██║██║╚██╗██║██║╚██╗██║██╔══╝  ██║     
██║ ╚████║╚██████╔╝╚██████╔╝██████╔╝   ██║   ╚██████╔╝██║ ╚████║██║ ╚████║███████╗███████╗
╚═╝  ╚═══╝ ╚═════╝  ╚═════╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚══════╝
                                                                                           
                        🚀 Secure Tunneling Made Easy v%s 🚀
`
)

func main() {
	var (
		mode       = flag.String("mode", "", "Mode: 'server' or 'client'")
		config     = flag.String("config", "", "Configuration file path")
		port       = flag.Int("port", 7000, "Server port (server mode only)")
		server     = flag.String("server", "", "Server address (client mode only)")
		localPort  = flag.Int("local-port", 8080, "Local port to tunnel (client mode only)")
		remotePort = flag.Int("remote-port", 0, "Remote port on server (client mode only)")
		help       = flag.Bool("help", false, "Show help")
		version    = flag.Bool("version", false, "Show version")
	)

	flag.Parse()

	if *version {
		fmt.Printf(Banner, Version)
		return
	}

	if *help || *mode == "" {
		showHelp()
		return
	}

	fmt.Printf(Banner, Version)

	switch strings.ToLower(*mode) {
	case "server":
		if *config != "" {
			startServerWithConfig(*config)
		} else {
			startServer(*port)
		}
	case "client":
		if *config != "" {
			startClientWithConfig(*config)
		} else {
			if *server == "" || *remotePort == 0 {
				fmt.Println("❌ Client mode requires --server and --remote-port parameters")
				os.Exit(1)
			}
			startClient(*server, *localPort, *remotePort)
		}
	default:
		fmt.Printf("❌ Unknown mode: %s\n", *mode)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(Banner, Version)
	fmt.Println(`
📖 USAGE:
  ntunnel --mode <server|client> [options]

🖥️  SERVER MODE:
  ntunnel --mode server --port 7000
  ntunnel --mode server --config server.yaml

💻 CLIENT MODE:
  ntunnel --mode client --server your-vps-ip:7000 --local-port 8080 --remote-port 80
  ntunnel --mode client --config client.yaml

⚙️  OPTIONS:
  --mode string        Mode: 'server' or 'client'
  --config string      Configuration file path
  --port int          Server port (default: 7000)
  --server string     Server address (client mode)
  --local-port int    Local port to tunnel (default: 8080)
  --remote-port int   Remote port on server (client mode)
  --help             Show this help
  --version          Show version

📝 EXAMPLES:
  # Start server on port 7000
  ntunnel --mode server --port 7000

  # Connect client to expose local port 3000 as remote port 80
  ntunnel --mode client --server 1.2.3.4:7000 --local-port 3000 --remote-port 80

  # Use configuration files
  ntunnel --mode server --config server.yaml
  ntunnel --mode client --config client.yaml

🔗 More info: https://github.com/Ambitiousnoob/noobtunnel
`)
}