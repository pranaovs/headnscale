package tailscale

import (
	"context"
	"log"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/tsnet"
)

func New(config config.Config) *Source {
	return &Source{
		authKey:     config.TailscaleAuthKey,
		hostname:    config.TailscaleHostname,
		loginServer: config.TailscaleLoginServer,
		forceReauth: false,
	}
}

func (s *Source) Initialize(ctx context.Context) error {
	srv := &tsnet.Server{
		Hostname:   s.hostname,
		AuthKey:    s.authKey,
		ControlURL: s.loginServer,
	}

	if _, err := srv.Up(ctx); err != nil {
		log.Printf("Failed to start Tailscale server: %v", err)
		return err
	}

	cli, err := srv.LocalClient()
	if err != nil {
		log.Printf("Failed to create Tailscale local client: %v", err)
		return err
	}

	watcher, err := cli.WatchIPNBus(ctx, 0)
	if err != nil {
		log.Printf("Failed to create IPN bus watcher: %v", err)
		return err
	}

	s.ts = ts{
		srv:     srv,
		cli:     cli,
		watcher: watcher,
	}

	log.Printf("Tailscale source initialized")
	return nil
}

func (s *Source) Fetch(ctx context.Context) ([]types.Node, error) {
	status, err := s.ts.cli.Status(ctx)
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
			} else {
				log.Printf("unknown IP type for peer %s: %v", peer.HostName, ip)
			}
		}
		nodes = append(nodes, types.Node{
			Hostname: peer.HostName,
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

		// Send initial state
		nodes, err := s.Fetch(ctx)
		if err != nil {
			log.Printf("error fetching initial nodes: %v", err)
			errChan <- err
			return
		}

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
				msg, err := s.ts.watcher.Next()
				if err != nil {
					log.Printf("error watching IPN bus: %v", err)
					errChan <- err
					return
				}

				// Check if NetMap has peer updates
				if msg.NetMap != nil && msg.NetMap.Peers != nil {
					log.Printf("Tailscale peer update detected")

					// Fetch updated nodes
					nodes, err := s.Fetch(ctx)
					if err != nil {
						log.Printf("error fetching nodes after event: %v", err)
						continue
					}

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
	if err := s.ts.srv.Close(); err != nil {
		log.Printf("Failed to close Tailscale server: %v", err)
		return err
	}

	log.Printf("Tailscale server closed")
	return nil
}
