package hosts

import (
	"net"
	"net/http"
)

type Sink struct {
	filePath     string
	noBaseDomain bool
	baseDomain   string

	// Networking
	ips     []net.IP
	port    int
	servers []*http.Server // Changed from custom struct to standard type
}
