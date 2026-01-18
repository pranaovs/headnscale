package docker

import (
	"github.com/docker/go-sdk/client"
)

func GetClientOption(s *Source) []client.ClientOption {
	options := []client.ClientOption{}

	options = append(options, client.WithDockerHost(s.Host))
	options = append(options, client.WithDockerContext(s.Context))

	return options
}
