package dns

import (
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

type Sink struct {
	config.Common
	config.DNS

	// Runtime state
	localIPs  []net.IP
	localSrvs []*dns.Server
	tsSrvs    []*dns.Server

	// Tailscale
	tsServer *tsnet.Server
	tsClient *local.Client

	// Data
	mu    sync.RWMutex
	nodes []types.Node
}
