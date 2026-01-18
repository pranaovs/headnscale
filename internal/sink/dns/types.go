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
	dnsIP        net.IP
	dnsPort      int
	srv          *dns.Server

	// Tailscale specific
	tsServer *tsnet.Server
	tsClient *local.Client
	tsSrvs   []*dns.Server

	// Mutex for async struct writes
	mu    sync.RWMutex
	nodes []types.Node
}
