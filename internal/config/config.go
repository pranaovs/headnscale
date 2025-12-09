package config

import (
	"log"
	"net"
	"strconv"

	"github.com/pranaovs/headnscale/internal/types"
	"github.com/pranaovs/headnscale/internal/utils"
)

func Load() types.Config {
	cfg := types.Config{
		LabelKey:         GetEnv("HEADNSCALE_LABEL_KEY", "headnscale.subdomain"),
		ExtraRecordsFile: GetEnv("HEADNSCALE_JSON_PATH", ""),
		HostsFile:        GetEnv("HEADNSCALE_HOSTS_PATH", ""),
		NoBaseDomain:     GetEnv("HEADNSCALE_NO_BASE_DOMAIN", "false") == "true",
		BaseDomain:       GetEnv("HEADNSCALE_BASE_DOMAIN", "ts.net"),
		Node: types.Node{
			Hostname: GetEnv("HEADNSCALE_NODE_HOSTNAME", ""),
		},
	}

	refreshDuration, err := utils.GetDuration(GetEnv("HEADNSCALE_REFRESH_SECONDS", "60"))
	if err != nil {
		log.Fatal("Invalid HEADNSCALE_REFRESH_SECONDS value")
	}
	cfg.Refresh = refreshDuration

	port, err := strconv.Atoi(GetEnv("HEADNSCALE_PORT", "8080"))
	if err != nil || port <= 0 || port > 65535 {
		log.Fatal("Invalid HEADNSCALE_PORT value")
	}
	cfg.Port = port

	ip4 := GetEnv("HEADNSCALE_NODE_IP", "")
	if ip4 == "" {
		log.Fatal("HEADNSCALE_NODE_IP is required")
	}

	ip6 := GetEnv("HEADNSCALE_NODE_IP6", "")

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
