package config

import (
	"time"

	"github.com/pranaovs/headnscale/internal/types"
)

type Config struct {
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

	// Common
	NoBaseDomain bool
	BaseDomain   string
	Node         types.Node
	Refresh      time.Duration
	Port         int
}
