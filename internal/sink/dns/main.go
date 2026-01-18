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
		Common:   config.Common,
		DNS:      config.Sink.DNS,
		localIPs: []net.IP{net.IPv4zero, net.IPv6zero},
		TSNet:    &config.TSNet,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	for _, ip := range s.localIPs {
		// Explicitly use udp4/udp6 to ensure clean binding on dual-stack OSs
		network := "udp4"
		if ip.To4() == nil {
			network = "udp6"
		}

		addr := net.JoinHostPort(ip.String(), strconv.Itoa(s.Port))
		srv := &dns.Server{
			Addr:    addr,
			Net:     network,
			Handler: s,
		}
		s.localSrvs = append(s.localSrvs, srv)

		go func(server *dns.Server, address string) {
			log.Printf("Starting Local DNS server on %s (%s)", address, server.Net)
			if err := server.ListenAndServe(); err != nil {
				log.Printf("Local DNS server error (%s): %v", address, err)
			}
		}(srv, addr)
	}

	if s.TailscaleServe {
		var ips []net.IP
		if s.TSNet.IPv4 != nil {
			ips = append(ips, s.TSNet.IPv4)
		}
		if s.TSNet.IPv6 != nil {
			ips = append(ips, s.TSNet.IPv6)
		}

		for _, ip := range ips {
			// Use s.TSPort (from *config.DNS) for the Tailnet listener
			addr := net.JoinHostPort(ip.String(), strconv.Itoa(s.TSPort))

			// Use the existing tsnet server to create the listener
			pc, err := s.TSNet.Srv.ListenPacket("udp", addr)
			if err != nil {
				log.Printf("Failed to listen on Tailscale UDP (%s): %v", addr, err)
				continue
			}

			tsSrv := &dns.Server{
				PacketConn: pc,
				Handler:    s,
			}
			s.tsSrvs = append(s.tsSrvs, tsSrv)

			go func(server *dns.Server, address string) {
				log.Printf("Starting Tailscale DNS server on %s", address)
				if err := server.ActivateAndServe(); err != nil {
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

func (s *Sink) Close(ctx context.Context) error {
	// Note: DNS server Shutdown() doesn't accept context parameter
	// Close all local servers
	for _, srv := range s.localSrvs {
		if err := srv.ShutdownContext(ctx); err != nil {
			log.Printf("Error shutting down local DNS: %v", err)
		}
	}

	// Close all Tailscale servers
	for _, tsSrv := range s.tsSrvs {
		if err := tsSrv.ShutdownContext(ctx); err != nil {
			log.Printf("Error shutting down Tailscale DNS: %v", err)
		}
	}
	return nil
}
