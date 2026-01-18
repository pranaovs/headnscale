package tailscale

import (
	"context"
	"log"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
	"github.com/pranaovs/headnscale/internal/utils"
)

func New(config config.Config) *Source {
	return &Source{
		cli: config.TSNet.Cli,
	}
}

func (s *Source) Initialize(ctx context.Context) error {
	var err error
	s.watcher, err = s.cli.WatchIPNBus(ctx, 0)
	if err != nil {
		log.Printf("Failed to create IPN bus watcher: %v", err)
		return err
	}

	log.Printf("Tailscale source initialized")
	return nil
}

func (s *Source) Fetch(ctx context.Context) ([]types.Node, error) {
	status, err := s.cli.Status(ctx)
	if err != nil {
		log.Printf("Failed to get Tailscale status: %v", err)
		return nil, err
	}

	nodes := []types.Node{}

	for _, peer := range status.Peer {
		ips := types.NodeIP{}
		for _, ip := range peer.TailscaleIPs {
			if ip.Is4() {
				ips.IPv4 = ip.AsSlice()
			} else if ip.Is6() {
				ips.IPv6 = ip.AsSlice()
			}
		}
		nodes = append(nodes, types.Node{
			Hostname: hostNameFromDNSName(peer.DNSName),
			IP:       ips,
		})
	}

	return nodes, nil
}

func (s *Source) Watch(ctx context.Context) (<-chan []types.Node, <-chan error) {
	nodesChan := make(chan []types.Node)
	errChan := make(chan error, 1)

	go func() {
		defer close(nodesChan)
		defer close(errChan)

		var previousNodes []types.Node

		// Send initial state
		nodes, err := s.Fetch(ctx)
		if err != nil {
			log.Printf("error fetching initial nodes: %v", err)
			errChan <- err
			return
		}

		previousNodes = nodes

		select {
		case nodesChan <- nodes:
		case <-ctx.Done():
			return
		}

		// Watch for IPN bus events
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := s.watcher.Next()
				if err != nil {
					log.Printf("error watching IPN bus: %v", err)
					errChan <- err
					return
				}

				// Check if NetMap has peer updates
				if msg.NetMap != nil && msg.NetMap.Peers != nil {
					// Fetch updated nodes
					nodes, err := s.Fetch(ctx)
					if err != nil {
						log.Printf("error fetching nodes after event: %v", err)
						continue
					}

					// If nodes are unchanged, skip
					if utils.NodesEqual(previousNodes, nodes) {
						continue
					}

					log.Printf("Tailscale peer update detected")

					previousNodes = nodes

					select {
					case nodesChan <- nodes:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return nodesChan, errChan
}

func (s *Source) Close() error {
	return nil
}
