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

func (s *Source) Fetch(ctx context.Context) ([]types.Node, error) { return nil, nil }

func (s *Source) Watch(ctx context.Context) (<-chan []types.Node, <-chan error) { return nil, nil }

func (s *Source) Close() error {
	if err := s.ts.srv.Close(); err != nil {
		log.Printf("Failed to close Tailscale server: %v", err)
		return err
	}

	log.Printf("Tailscale server closed")
	return nil
}
