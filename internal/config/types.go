package config

import (
	"time"

	"github.com/pranaovs/headnscale/internal/types"
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

type Config struct {
	// Common
	NoBaseDomain   bool
	BaseDomain     string
	Wildcard       bool
	Node           types.Node
	Refresh        time.Duration
	Port           int
	StateDir       string
	TailscaleServe bool
	TSNet          TSNet

	// Docker source
	DockerEnabled bool
	DockerHost    string
	DockerContext string
	LabelKey      string

	// Tailscale source
	TailscaleEnabled     bool
	TailscaleLoginServer string
	TailscaleAuthKey     string
	TailscaleHostname    string

	// Sinks
	ExtraRecordsFile string
	HostsFile        string
	DNSPort          int
}

// TSNet holds Tailscale tsnet server and local client instances
type TSNet struct {
	Srv *tsnet.Server
	Cli *local.Client
}
