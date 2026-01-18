package types

import "net"

type IP struct {
	IPv4 net.IP
	IPv6 net.IP
}
