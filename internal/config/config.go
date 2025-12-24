package config

import (
	"log"
	"net"
	"strconv"

	"github.com/pranaovs/headnscale/internal/types"
	"github.com/pranaovs/headnscale/internal/utils"
)

func Load() Config {
	cfg := Config{
		// Docker source config
		DockerEnabled: GetEnv("HEADNSCALE_DOCKER_ENABLED", "") != "",
		LabelKey:      GetEnv("HEADNSCALE_LABEL_KEY", "headnscale.subdomain"),

		// Tailscale source config
		TailscaleEnabled:     GetEnv("HEADNSCALE_TS_AUTHKEY", "") != "",
		TailscaleLoginServer: GetEnv("HEADNSCALE_TS_LOGIN_SERVER", ""),
		TailscaleAuthKey:     GetEnv("HEADNSCALE_TS_AUTHKEY", ""),
		TailscaleHostname:    GetEnv("HEADNSCALE_TS_HOSTNAME", "headnscale"),

		// Sink config
		ExtraRecordsFile: GetEnv("HEADNSCALE_JSON_PATH", ""),
		HostsFile:        GetEnv("HEADNSCALE_HOSTS_PATH", ""),

		// Common config
		NoBaseDomain: GetEnv("HEADNSCALE_NO_BASE_DOMAIN", "false") == "true",
		BaseDomain:   GetEnv("HEADNSCALE_BASE_DOMAIN", "ts.net"),
		Node: types.Node{
			Hostname: GetEnv("HEADNSCALE_NODE_HOSTNAME", ""),
		},
	}

	// Parse refresh duration
	refreshDuration, err := utils.GetDuration(GetEnv("HEADNSCALE_REFRESH_SECONDS", "60"))
	if err != nil {
		log.Fatal("Invalid HEADNSCALE_REFRESH_SECONDS value")
	}
	cfg.Refresh = refreshDuration

	// Parse HTTP port
	port, err := strconv.Atoi(GetEnv("HEADNSCALE_PORT", "8080"))
	if err != nil || port <= 0 || port > 65535 {
		log.Fatal("Invalid HEADNSCALE_PORT value")
	}
	cfg.Port = port

	// Parse node IPs (required only if Docker source is enabled)
	if cfg.DockerEnabled {
		ip4 := GetEnv("HEADNSCALE_NODE_IP", "")
		if ip4 == "" {
			log.Fatal("HEADNSCALE_NODE_IP is required when Docker source is enabled")
		}
		cfg.Node.IP.IPv4 = net.ParseIP(ip4)
		if cfg.Node.IP.IPv4 == nil {
			log.Fatalf("Invalid IPv4 address: %s", ip4)
		}

		ip6 := GetEnv("HEADNSCALE_NODE_IP6", "")
		if ip6 != "" {
			cfg.Node.IP.IPv6 = net.ParseIP(ip6)
			if cfg.Node.IP.IPv6 == nil {
				log.Fatalf("Invalid IPv6 address: %s", ip6)
			}
		}

		if cfg.Node.Hostname == "" {
			log.Fatal("HEADNSCALE_NODE_HOSTNAME is required when Docker source is enabled")
		}
	}

	// Validate at least one source is enabled
	if !cfg.DockerEnabled && !cfg.TailscaleEnabled {
		log.Fatal("At least one source must be enabled (set HEADNSCALE_DOCKER_ENABLED or HEADNSCALE_TS_AUTHKEY)")
	}

	// Validate at least one sink is enabled
	if cfg.ExtraRecordsFile == "" && cfg.HostsFile == "" {
		log.Fatal("At least one sink must be enabled (set HEADNSCALE_JSON_PATH or HEADNSCALE_HOSTS_PATH)")
	}

	return cfg
}
