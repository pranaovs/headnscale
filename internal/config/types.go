package config

import (
	"time"

	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

type Config struct {
	Common
	TSNet  TSNet
	Source Source
	Sink   Sink
}

type Common struct {
	NoBaseDomain bool
	BaseDomain   string
	Wildcard     bool
	Refresh      time.Duration
	StateDir     string
}

type Source struct {
	Docker    Docker
	Tailscale Tailscale
}

type (
	Docker struct {
		Enabled  bool
		Host     string
		Node     types.Node
		Context  string
		LabelKey string
	}
	Tailscale struct {
		Enabled     bool
		LoginServer string
		AuthKey     string
		Hostname    string
	}
)

type Sink struct {
	TailscaleServe bool
	Headscale      Headscale
	Hosts          Hosts
	DNS            DNS
}

type (
	Headscale struct {
		Enabled bool

		ExtraRecordsFile string
	}
	Hosts struct {
		Enabled bool
		Path    string
		Port    int
		TSPort  int
	}
	DNS struct {
		Enabled bool
		Port    int
		TSPort  int
	}
)

// TSNet holds Tailscale tsnet server and local client instances
type TSNet struct {
	Srv *tsnet.Server
	Cli *local.Client
	types.Node
}

// GetServer returns the tsnet server if available, or nil
func (t *TSNet) GetServer() *tsnet.Server {
	if t == nil {
		return nil
	}
	return t.Srv
}

// GetClient returns the local client if available, or nil
func (t *TSNet) GetClient() *local.Client {
	if t == nil {
		return nil
	}
	return t.Cli
}
