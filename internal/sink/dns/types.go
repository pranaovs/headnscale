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
	dnsIP        net.IP
	dnsPort      int
	srv          *dns.Server

	// Mutex for async struct writes
	mu    sync.RWMutex
	nodes []types.Node
}
