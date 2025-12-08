package types

import (
	"time"
)

type Config struct {
	LabelKey         string
	ExtraRecordsFile string
	NoBaseDomain     bool
	Refresh          time.Duration
	BaseDomain       string
	Node             Node
}

type Node struct {
	Hostname string
	IP       NodeIP
}
