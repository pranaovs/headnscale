package tailscale

import (
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

type Source struct {
	authKey     string
	hostname    string
	loginServer string
	forceReauth bool
	dir         string
	ts          ts
}

type ts struct {
	srv     *tsnet.Server
	cli     *local.Client
	watcher *local.IPNBusWatcher
}
