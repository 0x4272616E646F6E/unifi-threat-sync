package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/http"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/sync"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/unifi"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Command-line flags
	configPath := flag.String("config", "/config/config.yaml", "Path to configuration file")
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.Parse()

	// Print version and exit
	if *versionFlag {
		fmt.Printf("UniFi Threat Sync %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("UniFi Threat Sync %s starting...\n", Version)
	fmt.Printf("UniFi Controller: %s\n", cfg.UniFi.URL)
	fmt.Printf("Sync Interval: %s\n", cfg.Sync.Interval)
	fmt.Printf("Enabled Feeds: %d\n", cfg.Feeds.EnabledCount())

	// Create UniFi client
	unifiClient, err := unifi.NewClient(cfg.UniFi)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating UniFi client: %v\n", err)
		os.Exit(1)
	}

	// Test UniFi connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := unifiClient.Login(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Failed to connect to UniFi controller: %v\n", err)
		os.Exit(1)
	}
	cancel()
	fmt.Println("Successfully connected to UniFi controller")

	// Create sync service
	syncer := sync.New(cfg, unifiClient)

	// Start health check server if enabled
	var healthServer *http.HealthServer
	if cfg.Health.Enabled {
		healthServer = http.NewHealthServer(cfg.Health.Port, Version)
		if err := healthServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start health server: %v\n", err)
			os.Exit(1)
		}
		// Connect health recorder to syncer
		syncer.SetHealthRecorder(healthServer)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run initial sync
	fmt.Println("Running initial sync...")
	if err := syncer.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Initial sync failed: %v\n", err)
		if healthServer != nil {
			healthServer.RecordError()
		}
		// Don't exit, continue with periodic sync
	}

	// Start periodic sync
	ticker := time.NewTicker(cfg.Sync.Interval)
	defer ticker.Stop()

	fmt.Printf("Sync loop started (interval: %s)\n", cfg.Sync.Interval)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nShutdown signal received, cleaning up...")
			
			// Shutdown health server
			if healthServer != nil {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := healthServer.Stop(shutdownCtx); err != nil {
					fmt.Fprintf(os.Stderr, "Error stopping health server: %v\n", err)
				}
			}
			
			return
		case <-ticker.C:
			fmt.Printf("\n[%s] Starting scheduled sync...\n", time.Now().Format(time.RFC3339))
			if err := syncer.Run(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "Sync failed: %v\n", err)
				if healthServer != nil {
					healthServer.RecordError()
				}
			}
		}
	}
}
