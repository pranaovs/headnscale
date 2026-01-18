package dns

import (
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

type Sink struct {
	baseDomain   string
	noBaseDomain bool
	wildcard     bool

	// Local Configuration
	dnsPort  int
	localIPs []net.IP // Slice of IPs to listen on locally

	// Server instances (Slices to handle multiple listeners)
	localSrvs []*dns.Server
	tsSrvs    []*dns.Server

	// Tailscale
	tsServer *tsnet.Server
	tsClient *local.Client

	// Data
	mu    sync.RWMutex
	nodes []types.Node
}
