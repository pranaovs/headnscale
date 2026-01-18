package headscale

import "github.com/pranaovs/headnscale/internal/config"

type Sink struct {
	*config.Common
	*config.Headscale
}
