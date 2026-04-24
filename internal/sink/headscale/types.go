package headscale

import "codeberg.org/pranaovs/headnscale/internal/config"

type Sink struct {
	config.Common
	config.Headscale
}
