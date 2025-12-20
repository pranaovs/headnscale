package types

import (
	"time"
)

type Config struct {
	LabelKey         string
	ExtraRecordsFile string
	HostsFile        string
	NoBaseDomain     bool
	Refresh          time.Duration
	BaseDomain       string
	Port             int
	Node             Node
}
