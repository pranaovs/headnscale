package types

import (
	"time"
)

type Config struct {
	LabelKey         string
	ExtraRecordsFile string
	NoBaseDomain     bool
	Refresh          time.Duration
	Node             Node
}

type Node struct {
	BaseDomain string
	Hostname   string
	IP         NodeIP
}
