package docker

import (
	sdkclient "github.com/docker/go-sdk/client"
	"github.com/pranaovs/headnscale/internal/types"
)

type Source struct {
	cli      sdkclient.SDKClient
	labelKey string
	node     types.Node
}
