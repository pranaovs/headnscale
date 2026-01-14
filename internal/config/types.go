package config

import (
	"time"

	"github.com/pranaovs/headnscale/internal/types"
)

type Config struct {
	// Common
	NoBaseDomain bool
	BaseDomain   string
	Wildcard     bool
	Node         types.Node
	Refresh      time.Duration
	Port         int
	StateDir     string

	// Docker source
	DockerEnabled bool
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
