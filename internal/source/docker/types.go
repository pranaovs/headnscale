package docker

import (
	sdkclient "github.com/docker/go-sdk/client"
	"github.com/pranaovs/headnscale/internal/config"
)

type Source struct {
	*config.Common
	*config.Docker
	cli sdkclient.SDKClient
}
