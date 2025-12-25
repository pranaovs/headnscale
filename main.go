package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/sink/dns"
	"github.com/pranaovs/headnscale/internal/sink/headscale"
	"github.com/pranaovs/headnscale/internal/sink/hosts"
	"github.com/pranaovs/headnscale/internal/source/docker"
	"github.com/pranaovs/headnscale/internal/source/tailscale"
	"github.com/pranaovs/headnscale/internal/types"
)

func main() {
	cfg := config.Load()
	logStartup(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize sources based on config
	var sources []types.Source
	if cfg.DockerEnabled {
		sources = append(sources, docker.New(cfg))
	}
	if cfg.TailscaleEnabled {
		sources = append(sources, tailscale.New(cfg))
	}

	// Initialize sinks based on config
	var sinks []types.Sink
	if cfg.HostsFile != "" {
		sinks = append(sinks, hosts.New(cfg))
	}
	if cfg.ExtraRecordsFile != "" {
		sinks = append(sinks, headscale.New(cfg))
	}
	if cfg.DnsPort != 0 {
		sinks = append(sinks, dns.New(cfg))
	}
	// Setup and start all modules
	if err := initializeModules(ctx, sources, sinks); err != nil {
		log.Fatalf("Failed to initialize modules: %v", err)
	}
	defer closeModules(sources, sinks)

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
		var allNodes []types.Node
		for _, nodes := range sourceState {
			allNodes = append(allNodes, nodes...)
		}

		log.Printf("Writing merged update: %d total nodes", len(allNodes))

		for _, sink := range sinks {
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
					writeToSinks()
					mu.Unlock()
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

func closeModules(sources []types.Source, sinks []types.Sink) {
	for _, source := range sources {
		if err := source.Close(); err != nil {
			log.Printf("Error closing source: %v", err)
		}
	}

	for _, sink := range sinks {
		if err := sink.Close(); err != nil {
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

func logStartup(config config.Config) {
	log.Printf("Using configuration:")
	log.Printf(" - Base Domain: %s", config.BaseDomain)
	log.Printf(" - No Base Domain: %t", config.NoBaseDomain)
	log.Printf(" - Refresh Interval: %s", config.Refresh)
	log.Printf(" - HTTP Port: %d", config.Port)

	// Sources
	log.Printf(" - Docker Source: %t", config.DockerEnabled)
	if config.DockerEnabled {
		log.Printf("   - Label Key: %s", config.LabelKey)
		log.Printf("   - Node Hostname: %s", config.Node.Hostname)
		log.Printf("   - Node IPv4: %s", config.Node.IP.IPv4.String())
		if config.Node.IP.IPv6 != nil {
			log.Printf("   - Node IPv6: %s", config.Node.IP.IPv6.String())
		}
	}

	log.Printf(" - Tailscale Source: %t", config.TailscaleEnabled)
	if config.TailscaleEnabled {
		log.Printf("   - Tailscale Hostname: %s", config.TailscaleHostname)
		if config.TailscaleLoginServer != "" {
			log.Printf("   - Tailscale Login Server: %s", config.TailscaleLoginServer)
		}
	}

	// Sinks
	if config.HostsFile != "" {
		log.Printf(" - Hosts File: %s", config.HostsFile)
	}
	if config.ExtraRecordsFile != "" {
		log.Printf(" - Extra Records File: %s", config.ExtraRecordsFile)
	}
	if config.DnsPort != 0 {
		log.Printf(" - DNS Port: %d", config.DnsPort)
	}
}
