package dns

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Sink {
	return &Sink{
		noBaseDomain: config.NoBaseDomain,
		baseDomain:   config.BaseDomain,
		dnsIP:        net.IPv4(0, 0, 0, 0),
		dnsPort:      config.DNSPort,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	s.srv = &dns.Server{
		Addr:    net.JoinHostPort(s.dnsIP.String(), strconv.Itoa(s.dnsPort)),
		Net:     "udp",
		Handler: s,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			log.Printf("DNS Server exited: %v\n", err)
		}
	}()

	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes = nodes

	return nil
}

func (s *Sink) Close() error {
	if err := s.srv.Shutdown(); err != nil {
		return err
	}
	return nil
}
