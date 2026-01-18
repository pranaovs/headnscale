package hosts

import (
	"net"
	"net/http"

	"github.com/pranaovs/headnscale/internal/config"
)

type Sink struct {
	config.Common
	config.Hosts

	// Runtime state
	ips     []net.IP
	servers []*http.Server
}
