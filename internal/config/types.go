package config

import (
	"time"

	"github.com/pranaovs/headnscale/internal/types"
)

type Config struct {
	LabelKey             string
	ExtraRecordsFile     string
	HostsFile            string
	NoBaseDomain         bool
	BaseDomain           string
	Node                 types.Node
	Refresh              time.Duration
	Port                 int
	TailscaleLoginServer string
	TailscaleAuthKey     string
	TailscaleHostname    string
}
