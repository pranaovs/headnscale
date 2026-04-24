package dns

import (
	"net"
	"sync"

	"codeberg.org/pranaovs/headnscale/internal/config"
	"codeberg.org/pranaovs/headnscale/internal/types"
	"github.com/miekg/dns"
	"tailscale.com/client/local"
)

type Sink struct {
	config.Common
	config.DNS

	// Runtime state
	localIPs  []net.IP
	localSrvs []*dns.Server
	tsSrvs    []*dns.Server

	// Tailscale
	TSNet    *config.TSNet
	tsClient *local.Client

	// Data
	mu    sync.RWMutex
	nodes []types.Node
}
