package types

import "net"

type NodeIP struct {
	IPv4 net.IP
	IPv6 net.IP
}
