package utils

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

	// Build a map of hostname -> list of IPs to handle duplicate hostnames
	indexA := make(map[string][]types.IP, len(a))
	for _, n := range a {
		indexA[n.Hostname] = append(indexA[n.Hostname], n.IP)
	}

	indexB := make(map[string][]types.IP, len(b))
	for _, n := range b {
		indexB[n.Hostname] = append(indexB[n.Hostname], n.IP)
	}

	// Check all hostnames match
	if len(indexA) != len(indexB) {
		return false
	}

	for hostname, ipsA := range indexA {
		ipsB, ok := indexB[hostname]
		if !ok || len(ipsA) != len(ipsB) {
			return false
		}

		// For each IP in A, find a match in B (order-independent)
		for _, ipA := range ipsA {
			found := false
			for _, ipB := range ipsB {
				if equalNodeIP(ipA, ipB) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// equalNodeIP efficiently compares two NodeIP structs
func equalNodeIP(a, b types.IP) bool {
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
