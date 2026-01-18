package dns

import (
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/types"
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

	// Data
	mu    sync.RWMutex
	nodes []types.Node
}
