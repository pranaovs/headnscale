package hosts

import "net"

type Sink struct {
	filePath     string
	noBaseDomain bool
	baseDomain   string
	httpIP       net.IP
	httpPort     int
	server       *httpServer
}
