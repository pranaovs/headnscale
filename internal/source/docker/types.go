package docker

import (
	"codeberg.org/pranaovs/headnscale/internal/config"
	sdkclient "github.com/docker/go-sdk/client"
)

type Source struct {
	config.Common
	config.Docker
	cli sdkclient.SDKClient
}
