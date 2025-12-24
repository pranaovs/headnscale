package tailscale

import (
	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/tailcfg"
)

// peersToNodes converts NetMap peers to Node slice
func peersToNodes(peers []tailcfg.NodeView) []types.Node {
	nodes := make([]types.Node, 0, len(peers))

	for _, peer := range peers {
		ips := types.NodeIP{}

		// Get all addresses from the peer
		addrs := peer.Addresses()
		for i := 0; i < addrs.Len(); i++ {
			addr := addrs.At(i).Addr()
			if addr.Is4() {
				ips.IPv4 = addr.AsSlice()
			} else if addr.Is6() {
				ips.IPv6 = addr.AsSlice()
			}
		}

		nodes = append(nodes, types.Node{
			Hostname: peer.DisplayName(false),
			IP:       ips,
		})
	}

	return nodes
}
