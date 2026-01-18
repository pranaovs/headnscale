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
		filePath:     config.Sink.Hosts.Path,
		noBaseDomain: config.NoBaseDomain,
		baseDomain:   config.BaseDomain,
		// Listen on both IPv4 wildcard and IPv6 wildcard
		ips:  []net.IP{net.IPv4zero, net.IPv6zero},
		port: config.Sink.Hosts.Port,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	// Helper function handles all the setup complexity
	for _, ip := range s.ips {
		srv, err := startServer(ip, s.port, s.filePath)
		if err != nil {
			log.Printf("Failed to start HTTP server on %s: %v", ip, err)
			return err
		}
		s.servers = append(s.servers, srv)
	}

	log.Printf("Hosts sink initialized, serving on port %d (Dual Stack)", s.port)
	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	records := create(nodes, s.baseDomain)
	if s.noBaseDomain {
		records = append(records, create(nodes, "")...)
	}
	sorted := sort(records)

	if err := write(s.filePath, sorted); err != nil {
		log.Printf("error writing hosts: %v", err)
		return err
	}

	log.Printf("Wrote %d host records to %s", len(sorted), s.filePath)
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
