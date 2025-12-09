package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
	"time"

	sdkclient "github.com/docker/go-sdk/client"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/dns"
	"github.com/pranaovs/headnscale/internal/httpserver"
	docker "github.com/pranaovs/headnscale/internal/integrations/docker"
	"github.com/pranaovs/headnscale/internal/types"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Build Docker client
	ctx := context.Background()
	cli, err := sdkclient.New(ctx, docker.GetClientOption()...)
	if err != nil {
		log.Fatalf("failed to initialize Docker client: %v", err)
	}
	defer func() {
		err = cli.Close()
		log.Printf("error closing Docker client: %v", err)
	}()

	logStartup(cfg)

	if cfg.HostsFile != "" {
		httpserver.ServeFile("/hosts", cfg.HostsFile)
		httpserver.ServeFile("/hosts.txt", cfg.HostsFile)
		log.Printf("Serving hosts file at /hosts and /hosts.txt")
	}

	go httpserver.Start(net.ParseIP("0.0.0.0"), 8080)
	log.Printf("HTTP server started on port %d", cfg.Port)

	// Perform one scan immediately
	process(ctx, cli, cfg)

	// Schedule recurring scans
	ticker := time.NewTicker(cfg.Refresh)
	defer ticker.Stop()

	for range ticker.C {
		process(ctx, cli, cfg)
	}
}

func process(ctx context.Context, cli sdkclient.SDKClient, cfg types.Config) {
	containers, err := docker.GetRunning(cli, ctx)
	if err != nil {
		log.Printf("error listing containers: %v", err)
		return
	}

	labeled, err := docker.GetLabelled(containers, cfg.LabelKey)
	if err != nil {
		log.Printf("error filtering labeled containers: %v", err)
		return
	}

	subdomains, err := docker.GetLabels(labeled, cfg.LabelKey)
	if err != nil {
		log.Printf("error retrieving labels: %v", err)
		return
	}

	trimmedSubdomains := []string{}

	for _, subdomain := range subdomains {
		// Split the label value by | to support multiple hostnames
		for hostname := range strings.SplitSeq(subdomain, "|") {
			if trimmedHostname := strings.TrimSpace(hostname); trimmedHostname != "" {
				trimmedSubdomains = append(trimmedSubdomains, trimmedHostname)
			}
		}
	}

	log.Printf("Found %d labeled containers, %d subdomains", len(labeled), len(trimmedSubdomains))

	// Create DNS JSON records
	records := dns.CreateJSON(trimmedSubdomains, cfg.Node.Hostname+"."+cfg.BaseDomain, cfg.Node)
	if cfg.NoBaseDomain {
		records = append(records, dns.CreateJSON(trimmedSubdomains, cfg.Node.Hostname, cfg.Node)...)
	}
	sorted := dns.SortJSON(records)

	// Write file
	if err := writeJSON(cfg.ExtraRecordsFile, sorted); err != nil {
		log.Printf("error writing JSON: %v", err)
		return
	}

	if cfg.HostsFile != "" {
		records := dns.CreateHosts(trimmedSubdomains, cfg.Node.Hostname+"."+cfg.BaseDomain, cfg.Node)
		if cfg.NoBaseDomain {
			records = append(records, dns.CreateHosts(trimmedSubdomains, cfg.Node.Hostname, cfg.Node)...)
		}
		sorted := dns.SortHosts(records)

		// Write the hosts file
		if err := writeHosts(cfg.HostsFile, sorted); err != nil {
			log.Printf("error writing hosts file: %v", err)
		}
	}

	log.Printf("Successfully wrote %d DNS records", len(sorted))
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writeHosts(path string, hosts []string) error {
	data := strings.Join(hosts, "")
	return os.WriteFile(path, []byte(data), 0o644)
}

func logStartup(cfg types.Config) {
	log.Printf("Using configuration:")
	log.Printf(" - Label Key: %s", cfg.LabelKey)
	log.Printf(" - Extra Records File: %s", cfg.ExtraRecordsFile)
	log.Printf(" - Base Domain: %s", cfg.BaseDomain)
	log.Printf(" - Hostname: %s", cfg.Node.Hostname)
	log.Printf(" - No Base Domain: %t", cfg.NoBaseDomain)
	log.Printf(" - Refresh Interval: %s", cfg.Refresh)
	log.Printf(" - Node IPv4: %s", cfg.Node.IP.IPv4.String())
	if cfg.Node.IP.IPv6 != nil {
		log.Printf(" - Node IPv6: %s", cfg.Node.IP.IPv6.String())
	}
}
