package tailscale

import (
	"github.com/pranaovs/headnscale/internal/config"
	"tailscale.com/client/local"
)

type Source struct {
	config.Common
	config.Tailscale
	cli     *local.Client
	watcher *local.IPNBusWatcher
}
