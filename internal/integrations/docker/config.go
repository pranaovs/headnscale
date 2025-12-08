package docker

import (
	"github.com/docker/go-sdk/client"
	"github.com/pranaovs/headnscale/internal/config"
)

func GetClientOption() []client.ClientOption {
	options := []client.ClientOption{}

	options = append(options, client.WithDockerHost("unix:///var/run/docker.sock"))
	options = append(options, client.WithDockerContext(config.GetEnv("DOCKER_CONTEXT", "")))

	return options
}
