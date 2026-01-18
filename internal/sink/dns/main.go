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
		tsServer: config.TSNet.GetServer(),
		tsClient: config.TSNet.GetClient(),
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

	// Tailscale DNS Servers
	if s.tsServer != nil && s.tsClient != nil {
		status, err := s.tsClient.Status(ctx)
		if err != nil {
			log.Printf("Failed to get Tailscale status: %v", err)
			return err
		}

		for _, ip := range status.TailscaleIPs {
			// Construct specific address (e.g., "100.100.100.100:53")
			addr := net.JoinHostPort(ip.String(), strconv.Itoa(s.dnsPort))

			// tsnet requires the specific IP address to be passed to ListenPacket
			pc, err := s.tsServer.ListenPacket("udp", addr)
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

func (s *Sink) Close() error {
	// Close all local servers
	for _, srv := range s.localSrvs {
		if err := srv.Shutdown(); err != nil {
			log.Printf("Error shutting down local DNS: %v", err)
		}
	}

	// Close all Tailscale servers
	for _, tsSrv := range s.tsSrvs {
		if err := tsSrv.Shutdown(); err != nil {
			log.Printf("Error shutting down Tailscale DNS: %v", err)
		}
	}
	return nil
}
