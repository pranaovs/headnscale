package config

import (
	"log"
	"net"

	"github.com/pranaovs/headnscale/internal/types"
	"github.com/pranaovs/headnscale/internal/utils"
)

func Load() types.Config {
	refreshDuration, err := utils.GetDuration(GetEnv("HEADNSCALE_REFRESH_SECONDS", "60"))
	if err != nil {
		log.Fatal("Invalid HEADNSCALE_REFRESH_SECONDS value")
	}

	cfg := types.Config{
		LabelKey:         GetEnv("HEADNSCALE_LABEL_KEY", "headnscale.subdomain"),
		ExtraRecordsFile: GetEnv("HEADNSCALE_JSON_PATH", "/var/lib/headscale/extra-records.json"),
		NoBaseDomain:     GetEnv("HEADNSCALE_NO_BASE_DOMAIN", "false") == "true",
		Refresh:          refreshDuration,
		BaseDomain:       GetEnv("HEADNSCALE_BASE_DOMAIN", "ts.net"),
		Node: types.Node{
			Hostname: GetEnv("HEADNSCALE_NODE_HOSTNAME", ""),
		},
	}

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
