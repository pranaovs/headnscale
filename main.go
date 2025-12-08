package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	sdkclient "github.com/docker/go-sdk/client"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/dns"
	docker "github.com/pranaovs/headnscale/internal/integrations/docker"
	"github.com/pranaovs/headnscale/internal/types"
)

func main() {
	// Load configuration
	cfg := loadConfig()

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

	log.Printf("Successfully wrote %d DNS records", len(sorted))
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadConfig() types.Config {
	cfg := types.Config{
		LabelKey:         config.GetEnv("HEADNSCALE_LABEL_KEY", "headnscale.subdomain"),
		ExtraRecordsFile: config.GetEnv("HEADNSCALE_JSON_PATH", "/var/lib/headscale/extra-records.json"),
		NoBaseDomain:     config.GetEnv("HEADNSCALE_NO_BASE_DOMAIN", "false") == "true",
		Refresh:          getDuration("HEADNSCALE_REFRESH_SECONDS", 60),
		BaseDomain:       config.GetEnv("HEADNSCALE_BASE_DOMAIN", "ts.net"),
		Node: types.Node{
			Hostname: config.GetEnv("HEADNSCALE_NODE_HOSTNAME", ""),
		},
	}

	ip4 := config.GetEnv("HEADNSCALE_NODE_IP", "")
	if ip4 == "" {
		log.Fatal("HEADNSCALE_NODE_IP is required")
	}

	ip6 := config.GetEnv("HEADNSCALE_NODE_IP6", "")

	cfg.Node.IP.IPv4 = net.ParseIP(ip4)
	if cfg.Node.IP.IPv4 == nil {
		log.Fatalf("Invalid IPv4 address: %s", ip4)
	}

	if ip6 != "" {
		cfg.Node.IP.IPv6 = net.ParseIP(ip6)
		if cfg.Node.IP.IPv6 == nil {
			log.Fatalf("Invalid IPv6 address: %s", ip6)
		}
	}

	if cfg.ExtraRecordsFile == "" {
		log.Fatal("HEADNSCALE_JSON_PATH is required")
	}
	if cfg.Node.Hostname == "" {
		log.Fatal("HEADNSCALE_NODE_HOSTNAME is required")
	}

	return cfg
}

func getDuration(env string, defSeconds int) time.Duration {
	val := config.GetEnv(env, "")
	if val == "" {
		return time.Duration(defSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("invalid value for %s: %v", env, err)
	}
	return time.Duration(seconds) * time.Second
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
