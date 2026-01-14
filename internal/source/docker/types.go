package docker

import (
	sdkclient "github.com/docker/go-sdk/client"
	"github.com/pranaovs/headnscale/internal/types"
)

type Source struct {
	cli           sdkclient.SDKClient
	dockerHost    string
	dockerContext string
	labelKey      string
	node          types.Node
}
