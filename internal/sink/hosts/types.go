package hosts

import (
	"net"
	"net/http"

	"codeberg.org/pranaovs/headnscale/internal/config"
)

type Sink struct {
	config.Common
	config.Hosts

	// Tailscale State
	tsNet *config.TSNet

	// Runtime state
	ips     []net.IP
	servers []*http.Server
}
