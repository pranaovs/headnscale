package docker

import (
	sdkclient "github.com/docker/go-sdk/client"
	"codeberg.org/pranaovs/headnscale/internal/config"
)

type Source struct {
	config.Common
	config.Docker
	cli sdkclient.SDKClient
}
