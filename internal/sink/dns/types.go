package dns

import (
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

type Sink struct {
	*config.Common
	*config.DNS

	// Runtime state
	localIPs  []net.IP
	localSrvs []*dns.Server

	// Data
	mu    sync.RWMutex
	nodes []types.Node
}
