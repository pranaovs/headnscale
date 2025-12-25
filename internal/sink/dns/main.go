package dns

import (
	"context"
	"net"

	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Sink {
	return &Sink{
		noBaseDomain: config.NoBaseDomain,
		baseDomain:   config.BaseDomain,
		dnsIP:        net.IPv4(0, 0, 0, 0),
		dnsPort:      config.DnsPort,
	}
}

func (s *Sink) Initialize(ctx context.Context) error {
}

func (s *Sink) Process(ctx context.Context, nodes []types.Node) error {
}

func (s *Sink) Close() error {
}
