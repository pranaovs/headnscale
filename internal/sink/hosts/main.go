package hosts

import (
	"context"
	"log"
	"net"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Sink {
	return &Sink{
		filePath:     config.Sink.Hosts.Path,
		noBaseDomain: config.NoBaseDomain,
		baseDomain:   config.BaseDomain,
		httpIP:       net.IPv4(0, 0, 0, 0),
		httpPort:     config.Sink.Hosts.Port,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	s.server = newHTTPServer(s.httpIP, s.httpPort)
	s.server.serve("/hosts", s.filePath)
	s.server.serve("/hosts.txt", s.filePath)

	if err := s.server.start(ctx); err != nil {
		return err
	}

	log.Printf("Hosts sink initialized, HTTP server at %s:%d", s.httpIP, s.httpPort)
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
	if s.server != nil {
		return s.server.stop()
	}
	return nil
}
