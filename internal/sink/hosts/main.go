package hosts

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Sink {
	return &Sink{
		Common: config.Common,
		Hosts:  config.Sink.Hosts,
		ips:    []net.IP{net.IPv4zero, net.IPv6zero},
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	// Helper function handles all the setup complexity
	for _, ip := range s.ips {
		srv, err := startServer(ip, s.Port, s.Path)
		if err != nil {
			log.Printf("Failed to start HTTP server on %s: %v", ip, err)
			return err
		}
		s.servers = append(s.servers, srv)
	}

	log.Printf("Hosts sink initialized, serving on port %d (Dual Stack)", s.Port)
	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	records := create(nodes, s.BaseDomain)
	if s.NoBaseDomain {
		records = append(records, create(nodes, "")...)
	}
	sorted := sort(records)

	if err := write(s.Path, sorted); err != nil {
		log.Printf("error writing hosts: %v", err)
		return err
	}

	log.Printf("Wrote %d host records to %s", len(sorted), s.Path)
	return nil
}

func (s *Sink) Close() error {
	for _, srv := range s.servers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error stopping HTTP server: %v", err)
		}
		cancel()
	}
	return nil
}
