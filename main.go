package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pranaovs/headnscale/internal/config"
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

	// Initialize sources
	sources := []types.Source{
		docker.New(cfg),
		tailscale.New(cfg),
	}

	// Initialize sinks
	sinks := []types.Sink{
		hosts.New(cfg),
		headscale.New(cfg),
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

					log.Printf("Received update: %d nodes", len(nodes))

					for _, sink := range sinks {
						if err := sink.Process(ctx, nodes); err != nil {
							log.Printf("Error writing to sink: %v", err)
						}
					}
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
	log.Printf(" - Label Key: %s", config.LabelKey)
	log.Printf(" - Extra Records File: %s", config.ExtraRecordsFile)
	log.Printf(" - Hosts File: %s", config.HostsFile)
	log.Printf(" - Base Domain: %s", config.BaseDomain)
	log.Printf(" - Hostname: %s", config.Node.Hostname)
	log.Printf(" - No Base Domain: %t", config.NoBaseDomain)
	log.Printf(" - Refresh Interval: %s", config.Refresh)
	log.Printf(" - HTTP Port: %d", config.Port)
	log.Printf(" - Node IPv4: %s", config.Node.IP.IPv4.String())
	if config.Node.IP.IPv6 != nil {
		log.Printf(" - Node IPv6: %s", config.Node.IP.IPv6.String())
	}
}
