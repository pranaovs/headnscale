package hosts

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(cfg config.Config) *Sink {
	return &Sink{
		Common: cfg.Common,
		Hosts:  cfg.Sink.Hosts,
		tsNet:  &cfg.TSNet,
		ips:    []net.IP{net.IPv4zero, net.IPv6zero},
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
	// 1. Start Local Servers
	for _, ip := range s.ips {
		// Determine network (tcp4 vs tcp6)
		network := "tcp4"
		addr := fmt.Sprintf("%s:%d", ip.String(), s.Port)

		if ip.To4() == nil {
			network = "tcp6"
			addr = fmt.Sprintf("[%s]:%d", ip.String(), s.Port)
		}

		// Create standard OS listener
		ln, err := net.Listen(network, addr)
		if err != nil {
			log.Printf("Failed to listen on local %s: %v", addr, err)
			return err
		}

		// Hand off to HTTP starter
		s.servers = append(s.servers, startHTTP(ln, s.Path))
	}

	// 2. Start Tailscale Servers
	if s.tsNet != nil && s.tsNet.Srv != nil {
		// Collect IPs populated in main.go
		var tsIPs []net.IP
		if s.tsNet.IPv4 != nil {
			tsIPs = append(tsIPs, s.tsNet.IPv4)
		}
		if s.tsNet.IPv6 != nil {
			tsIPs = append(tsIPs, s.tsNet.IPv6)
		}

		for _, ip := range tsIPs {
			// Use s.TSPort for Tailscale interface
			addr := net.JoinHostPort(ip.String(), strconv.Itoa(s.TSPort))

			// Create Tailscale listener
			ln, err := s.tsNet.Srv.Listen("tcp", addr)
			if err != nil {
				log.Printf("Failed to listen on Tailscale %s: %v", addr, err)
				continue
			}

			// Hand off to HTTP starter
			s.servers = append(s.servers, startHTTP(ln, s.Path))
		}
	}

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

func (s *Sink) Close(ctx context.Context) error {
	// Use the provided context for graceful shutdown
	for _, srv := range s.servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error stopping HTTP server: %v", err)
		}
	}
	return nil
}
