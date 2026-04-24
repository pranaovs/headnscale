package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"codeberg.org/pranaovs/headnscale/internal/config"
	"codeberg.org/pranaovs/headnscale/internal/sink/dns"
	"codeberg.org/pranaovs/headnscale/internal/sink/headscale"
	"codeberg.org/pranaovs/headnscale/internal/sink/hosts"
	"codeberg.org/pranaovs/headnscale/internal/source/docker"
	"codeberg.org/pranaovs/headnscale/internal/source/tailscale"
	"codeberg.org/pranaovs/headnscale/internal/types"
	"tailscale.com/tsnet"
)

func main() {
	cfg := config.Load()
	logStartup(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create Tailscale server if enabled
	if cfg.Source.Tailscale.Enabled || cfg.TailscaleServe {
		cfg.TSNet.Srv = &tsnet.Server{
			Hostname:   cfg.Source.Tailscale.Hostname,
			AuthKey:    cfg.Source.Tailscale.AuthKey,
			ControlURL: cfg.Source.Tailscale.LoginServer,
			Dir:        cfg.StateDir + "/" + cfg.Source.Tailscale.Hostname,
		}

		var err error

		// Start Tailscale server
		if _, err = cfg.TSNet.Srv.Up(ctx); err != nil {
			log.Fatalf("Failed to start Tailscale server: %v", err)
		}

		// Cleanup Tailscale server on exit
		defer func() {
			if err := cfg.TSNet.Srv.Close(); err != nil {
				log.Printf("Failed to close Tailscale server: %v", err)
			}
		}()

		// Create and save the local client
		cfg.TSNet.Cli, err = cfg.TSNet.Srv.LocalClient()
		if err != nil {
			log.Fatalf("Failed to create Tailscale local client: %v", err)
		}

		status, err := cfg.TSNet.Cli.Status(ctx)
		if err != nil {
			log.Fatalf("Failed to get Tailscale status: %v", err)
		}
		cfg.TSNet.Hostname = tailscale.HostNameFromDNSName(status.Self.DNSName)

		for _, ip := range status.TailscaleIPs {
			if ip.Is4() {
				// ip.AsSlice() returns []byte, which is compatible with net.IP
				cfg.TSNet.IPv4 = ip.AsSlice()
			} else if ip.Is6() {
				cfg.TSNet.IPv6 = ip.AsSlice()
			} else {
				log.Printf("Unknown IP type: %s", ip.String())
			}
		}
		log.Printf("Tailscale server started (source: %t, serve: %t)", cfg.Source.Tailscale.Enabled, cfg.TailscaleServe)
	}

	// Initialize sources based on config
	var sources []types.Source
	if cfg.Source.Docker.Enabled {
		sources = append(sources, docker.New(cfg))
	}
	if cfg.Source.Tailscale.Enabled {
		sources = append(sources, tailscale.New(cfg))
	}

	// Initialize sinks based on config
	var sinks []types.Sink
	if cfg.Sink.Hosts.Enabled {
		sinks = append(sinks, hosts.New(cfg))
	}
	if cfg.Sink.Headscale.Enabled {
		sinks = append(sinks, headscale.New(cfg))
	}
	if cfg.Sink.DNS.Enabled {
		sinks = append(sinks, dns.New(cfg))
	}
	// Setup and start all modules
	if err := initializeModules(ctx, sources, sinks); err != nil {
		log.Fatalf("Failed to initialize modules: %v", err)
	}
	defer func() {
		// Create a new context with timeout for cleanup since main context will be cancelled
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		closeModules(cleanupCtx, sources, sinks)
	}()

	// Start watching for changes
	go watchSources(ctx, sources, sinks)

	// Wait for interrupt signal
	waitForShutdown()
}

func initializeModules(ctx context.Context, sources []types.Source, sinks []types.Sink) error {
	for _, source := range sources {
		if err := source.Initialize(ctx); err != nil {
			return err
		}
	}

	for _, sink := range sinks {
		if err := sink.Initialize(ctx); err != nil {
			return err
		}
	}

	return nil
}

func watchSources(ctx context.Context, sources []types.Source, sinks []types.Sink) {
	var mu sync.Mutex
	sourceState := make(map[types.Source][]types.Node)

	// Initialize state for each source
	for _, src := range sources {
		sourceState[src] = nil
	}

	// Helper to merge all source states and write to sinks
	writeToSinks := func() {
		// Check if context is cancelled before processing
		if ctx.Err() != nil {
			log.Printf("Context cancelled, skipping sink writes")
			return
		}

		// Copy merged state while holding the lock
		mu.Lock()
		var allNodes []types.Node
		for _, nodes := range sourceState {
			allNodes = append(allNodes, nodes...)
		}
		mu.Unlock()

		log.Printf("Writing merged update: %d total nodes", len(allNodes))

		for _, sink := range sinks {
			// Check context before each sink write
			if ctx.Err() != nil {
				log.Printf("Context cancelled, stopping sink writes")
				return
			}
			if err := sink.Process(ctx, allNodes); err != nil {
				log.Printf("Error writing to sink: %v", err)
			}
		}
	}

	for _, source := range sources {
		go func(src types.Source) {
			nodesChan, errChan := src.Watch(ctx)

			for {
				select {
				case nodes, ok := <-nodesChan:
					if !ok {
						log.Println("Source watch channel closed")
						return
					}

					log.Printf("Received update from source: %d nodes", len(nodes))

					mu.Lock()
					sourceState[src] = nodes
					mu.Unlock()

					writeToSinks()
				case err, ok := <-errChan:
					if !ok {
						return
					}
					if err != nil {
						log.Printf("Error from source watch: %v", err)
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(source)
	}
}

func processOnce(ctx context.Context, sources []types.Source, sinks []types.Sink) error {
	var allNodes []types.Node

	for _, source := range sources {
		nodes, err := source.Fetch(ctx)
		if err != nil {
			log.Printf("Error fetching from source: %v", err)
			continue
		}
		allNodes = append(allNodes, nodes...)
	}

	log.Printf("Fetched %d nodes from %d source(s)", len(allNodes), len(sources))

	for _, sink := range sinks {
		if err := sink.Process(ctx, allNodes); err != nil {
			log.Printf("Error writing to sink: %v", err)
		}
	}

	return nil
}

func closeModules(ctx context.Context, sources []types.Source, sinks []types.Sink) {
	for _, source := range sources {
		if err := source.Close(ctx); err != nil {
			log.Printf("Error closing source: %v", err)
		}
	}

	for _, sink := range sinks {
		if err := sink.Close(ctx); err != nil {
			log.Printf("Error closing sink: %v", err)
		}
	}
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received, cleaning up...")
}

func logStartup(cfg config.Config) {
	log.Printf("Using configuration:")
	log.Printf(" - Base Domain: %s", cfg.BaseDomain)
	log.Printf(" - No Base Domain: %t", cfg.NoBaseDomain)
	log.Printf(" - Refresh Interval: %s", cfg.Refresh)
	log.Printf(" - State Directory: %s", cfg.StateDir)

	// Docker Source
	dockerEnabled := cfg.Source.Docker.Enabled
	log.Printf(" - Docker Source: %t", dockerEnabled)
	if dockerEnabled {
		log.Printf("   - Label Key: %s", cfg.Source.Docker.LabelKey)
		log.Printf("   - Docker Host: %s", cfg.Source.Docker.Host)
		// Node info is primarily used by Docker source
		log.Printf("   - Node Hostname: %s", cfg.Source.Docker.Node.Hostname)
		log.Printf("   - Node IPv4: %s", cfg.Source.Docker.Node.IPv4.String())
		if cfg.Source.Docker.Node.IPv6 != nil {
			log.Printf("   - Node IPv6: %s", cfg.Source.Docker.Node.IPv6.String())
		}
	}

	// Tailscale Source & Serve
	tsEnabled := cfg.Source.Tailscale.Enabled
	log.Printf(" - Tailscale Source: %t", tsEnabled)
	log.Printf(" - Tailscale Serve: %t", cfg.TailscaleServe)

	if tsEnabled {
		log.Printf("   - Tailscale Hostname: %s", cfg.Source.Tailscale.Hostname)
		if cfg.Source.Tailscale.LoginServer != "" {
			log.Printf("   - Tailscale Login Server: %s", cfg.Source.Tailscale.LoginServer)
		}
	}

	// Sinks
	if cfg.Sink.Hosts.Enabled {
		log.Printf(" - Hosts File Sink: Enabled")
		log.Printf("   - Path: %s", cfg.Sink.Hosts.Path)
		log.Printf("   - HTTP Port: %d", cfg.Sink.Hosts.Port)
	}

	if cfg.Sink.Headscale.Enabled {
		log.Printf(" - Headscale Sink: Enabled")
		log.Printf("   - Path: %s", cfg.Sink.Headscale.ExtraRecordsFile)
	}

	if cfg.Sink.DNS.Enabled {
		log.Printf(" - DNS Sink: Enabled")
		log.Printf("   - UDP Port: %d", cfg.Sink.DNS.Port)
	}
}
