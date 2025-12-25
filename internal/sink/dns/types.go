package dns

import "net"

type Sink struct {
	baseDomain   string
	noBaseDomain bool
	dnsIP        net.IP
	dnsPort      int
	// server       *httpServer
}
