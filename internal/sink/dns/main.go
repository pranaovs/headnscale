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
		wildcard:     config.Wildcard,
		dnsIP:        net.IPv4(0, 0, 0, 0),
		dnsPort:      config.Sink.DNS.Port,
		tsServer:     config.TSNet.GetServer(),
		tsClient:     config.TSNet.GetClient(),
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	// 1. Start Local DNS Server
	s.srv = &dns.Server{
		Addr:    net.JoinHostPort(s.dnsIP.String(), strconv.Itoa(s.dnsPort)),
		Net:     "udp",
		Handler: s,
	}

	go func() {
		log.Printf("Starting Local DNS server on %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil {
			log.Printf("Local DNS server error: %v", err)
		}
	}()

	// 2. Start Tailscale DNS Servers (if enabled)
	if s.tsServer != nil {
		status, err := s.tsClient.Status(ctx)
		if err != nil {
			log.Printf("Failed to get Tailscale status: %v", err)
			return err
		}

		for _, ip := range status.TailscaleIPs {
			// Construct address explicitly (e.g., "100.x.y.z:53" or "[fd7a::1]:53")
			addr := net.JoinHostPort(ip.String(), strconv.Itoa(s.dnsPort))

			// tsnet requires the IP to be explicitly specified
			pc, err := s.tsServer.ListenPacket("udp", "0.0.0.0:53")
			if err != nil {
				log.Printf("Failed to listen on Tailscale UDP (%s): %v", addr, err)
				continue
			}

			tsSrv := &dns.Server{
				PacketConn: pc,
				Handler:    s,
			}
			s.tsSrvs = append(s.tsSrvs, tsSrv)

			go func(srv *dns.Server, address string) {
				log.Printf("Starting Tailscale DNS server on %s", address)
				if err := srv.ActivateAndServe(); err != nil {
					log.Printf("Tailscale DNS server error (%s): %v", address, err)
				}
			}(tsSrv, addr)
		}
	}

	return nil
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes = nodes

	return nil
}

func (s *Sink) Close() error {
	if s.srv != nil {
		if err := s.srv.Shutdown(); err != nil {
			log.Printf("Error shutting down local DNS: %v", err)
		}
	}

	for _, tsSrv := range s.tsSrvs {
		if err := tsSrv.Shutdown(); err != nil {
			log.Printf("Error shutting down Tailscale DNS: %v", err)
		}
	}
	return nil
}
