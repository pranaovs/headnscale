package tailscale

import (
	"bytes"

	"github.com/pranaovs/headnscale/internal/types"
)

// NodesEqual checks if two node slices are equal
func NodesEqual(a, b []types.Node) bool {
	if len(a) != len(b) {
		return false
	}

	// Handle empty slices
	if len(a) == 0 {
		return true
	}

	// Create map for O(n) lookup instead of O(n²)
	index := make(map[string]types.Node, len(a))
	for _, n := range a {
		index[n.Hostname] = n
	}

	for _, n := range b {
		other, ok := index[n.Hostname]
		if !ok {
			return false
		}

		if !equalNodeIP(other.IP, n.IP) {
			return false
		}
	}

	return true
}

// equalNodeIP efficiently compares two NodeIP structs
func equalNodeIP(a, b types.NodeIP) bool {
	// Compare IPv4
	if !equalIP(a.IPv4, b.IPv4) {
		return false
	}

	// Compare IPv6
	if !equalIP(a.IPv6, b.IPv6) {
		return false
	}

	return true
}

// equalIP compares two net.IP values using bytes.Equal for efficiency
func equalIP(a, b []byte) bool {
	// Both nil
	if a == nil && b == nil {
		return true
	}

	// One nil, one not
	if a == nil || b == nil {
		return false
	}

	// Compare bytes directly
	return bytes.Equal(a, b)
}
