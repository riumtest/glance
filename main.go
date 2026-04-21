package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/glanceapp/glance/internal/config"
	"github.com/glanceapp/glance/internal/server"
)

const (
	defaultConfigPath = "glance.yml"
	version           = "0.1.0"
)

func main() {
	var (
		configPath  = flag.String("config", defaultConfigPath, "Path to the configuration file")
		showVersion = flag.Bool("version", false, "Print version information and exit")
		validate    = flag.Bool("validate", false, "Validate the configuration file and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("glance version %s\n", version)
		os.Exit(0)
	}

	// Load configuration from the specified file
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// If validate flag is set, just check config and exit
	if *validate {
		log.Println("Configuration is valid.")
		os.Exit(0)
	}

	// Initialize and start the HTTP server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	log.Printf("Starting glance v%s on %s:%d", version, cfg.Server.Host, cfg.Server.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
