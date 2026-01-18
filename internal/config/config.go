package config

import (
	"log"
	"net"
)

func Load() Config {
	// Initialize the Config struct with embedded Common and nested Sink
	cfg := Config{
		Common: Common{
			BaseDomain:     GetEnv("HEADNSCALE_BASE_DOMAIN", "ts.net"),
			NoBaseDomain:   GetEnvBool("HEADNSCALE_NO_BASE_DOMAIN", false),
			Wildcard:       GetEnvBool("HEADNSCALE_WILDCARD", false),
			Refresh:        GetEnvDuration("HEADNSCALE_REFRESH_SECONDS", 60),
			StateDir:       GetEnv("HEADNSCALE_STATE_DIR", "/var/lib/headnscale"),
			TailscaleServe: GetEnvBool("HEADNSCALE_TS_SERVE", false),
		},
	}

	// ==========================================
	// SOURCES
	// ==========================================

	// Docker Source
	if GetEnvBool("HEADNSCALE_DOCKER_ENABLED", false) {
		// Initialize Docker Config first
		cfg.Source.Docker.Enabled = true
		cfg.Source.Docker.Host = GetEnv("DOCKER_HOST", "unix:///var/run/docker.sock")
		cfg.Source.Docker.Context = GetEnv("DOCKER_CONTEXT", "")
		cfg.Source.Docker.LabelKey = GetEnv("HEADNSCALE_LABEL_KEY", "headnscale.subdomain")

		// Load Node Information (Required for Docker)
		cfg.Source.Docker.Node.Hostname = GetEnvRequired("HEADNSCALE_NODE_HOSTNAME")

		// Validate IPv4
		ip4Str := GetEnvRequired("HEADNSCALE_NODE_IP")
		ip4 := net.ParseIP(ip4Str)
		if ip4 == nil {
			log.Fatalf("Config Error: HEADNSCALE_NODE_IP '%s' is not a valid IP address", ip4Str)
		}
		cfg.Source.Docker.Node.IPv4 = ip4

		// Validate IPv6 (Optional)
		if ip6Str := GetEnv("HEADNSCALE_NODE_IP6", ""); ip6Str != "" {
			ip6 := net.ParseIP(ip6Str)
			if ip6 == nil {
				log.Fatalf("Config Error: HEADNSCALE_NODE_IP6 '%s' is not a valid IP address", ip6Str)
			}
			cfg.Source.Docker.Node.IPv6 = ip6
		}
	}

	// 2. Tailscale Source
	if GetEnvBool("HEADNSCALE_TS_ENABLED", false) {
		cfg.Source.Tailscale = Tailscale{
			Enabled:     true,
			LoginServer: GetEnv("HEADNSCALE_TS_LOGIN_SERVER", ""),
			AuthKey:     GetEnv("TS_AUTHKEY", ""),
			Hostname:    GetEnv("HEADNSCALE_TS_HOSTNAME", "headnscale"),
		}
	}

	// ==========================================
	// SINKS
	// ==========================================

	// 1. Headscale JSON Sink
	if GetEnvBool("HEADSCALE_JSON_ENABLED", false) {
		cfg.Sink.Headscale = Headscale{
			Enabled:          true,
			ExtraRecordsFile: GetEnvRequired("HEADNSCALE_JSON_PATH"),
		}
	}

	// 2. Hosts File Sink
	if GetEnvBool("HEADSCALE_HOSTS_ENABLED", false) {
		localPort := GetEnvPort("HEADNSCALE_HOSTS_PORT", 80)

		cfg.Sink.Hosts = Hosts{
			Enabled: true,
			Path:    GetEnvRequired("HEADNSCALE_HOSTS_PATH"),
			Port:    localPort,
			// Default TS port to the same as local port if not explicitly set
			TSPort: GetEnvPort("HEADNSCALE_HOSTS_TS_PORT", localPort),
		}
	}

	// 3. DNS Sink
	if GetEnvBool("HEADNSCALE_DNS_ENABLED", false) {
		localPort := GetEnvPort("HEADNSCALE_DNS_PORT", 53)

		cfg.Sink.DNS = DNS{
			Enabled: true,
			Port:    localPort,
			// Default TS port to the same as local port if not explicitly set
			TSPort: GetEnvPort("HEADNSCALE_DNS_TS_PORT", localPort),
		}
	}

	return cfg
}
