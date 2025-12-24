package tailscale

import "github.com/pranaovs/headnscale/internal/types"

func NodesEqual(a, b []types.Node) bool {
	if len(a) != len(b) {
		return false
	}

	index := make(map[string]types.Node, len(a))
	for _, n := range a {
		index[n.Hostname] = n
	}

	for _, n := range b {
		other, ok := index[n.Hostname]
		if !ok {
			return false
		}

		// Hostname is identity, but still checked for completeness
		if other.Hostname != n.Hostname {
			return false
		}

		if !equalNodeIP(other.IP, n.IP) {
			return false
		}
	}

	return true
}

func equalNodeIP(a, b types.NodeIP) bool {
	switch {
	case a.IPv4 == nil && b.IPv4 != nil:
		return false
	case a.IPv4 != nil && b.IPv4 == nil:
		return false
	case a.IPv4 != nil && !a.IPv4.Equal(b.IPv4):
		return false
	}

	switch {
	case a.IPv6 == nil && b.IPv6 != nil:
		return false
	case a.IPv6 != nil && b.IPv6 == nil:
		return false
	case a.IPv6 != nil && !a.IPv6.Equal(b.IPv6):
		return false
	}

	return true
}
