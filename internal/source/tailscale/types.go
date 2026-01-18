package tailscale

import (
	"tailscale.com/client/local"
)

type Source struct {
	cli     *local.Client
	watcher *local.IPNBusWatcher
}
