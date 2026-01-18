package docker

import (
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	sdkclient "github.com/docker/go-sdk/client"
	"github.com/pranaovs/headnscale/internal/config"
	"github.com/pranaovs/headnscale/internal/types"
)

func New(config config.Config) *Source {
	return &Source{
		Common: config.Common,
		Docker: config.Source.Docker,
	}
}

func (s *Source) Initialize(ctx context.Context) error {
	cli, err := sdkclient.New(ctx, GetClientOption(s)...)
	if err != nil {
		return err
	}
	s.cli = cli
	log.Printf("Docker source initialized")
	return nil
}

func (s *Source) Fetch(ctx context.Context) ([]types.Node, error) {
	containers, err := GetRunning(ctx, s.cli)
	if err != nil {
		log.Printf("error listing containers: %v", err)
		return nil, err
	}

	labeled, err := GetLabelled(containers, s.LabelKey)
	if err != nil {
		log.Printf("error filtering labeled containers: %v", err)
		return nil, err
	}

	subdomains, err := GetLabels(labeled, s.LabelKey)
	if err != nil {
		log.Printf("error retrieving labels: %v", err)
		return nil, err
	}

	nodes := []types.Node{}
	trimmedSubdomains := []string{}

	for _, subdomain := range subdomains {
		for hostname := range strings.SplitSeq(subdomain, "|") {
			if trimmedHostname := strings.TrimSpace(hostname); trimmedHostname != "" {
				trimmedSubdomains = append(trimmedSubdomains, trimmedHostname)
			}
		}
	}

	for _, subdomain := range trimmedSubdomains {
		nodes = append(nodes, types.Node{Hostname: subdomain + "." + s.Node.Hostname, IP: s.Node.IP})
	}

	return nodes, nil
}

func (s *Source) Watch(ctx context.Context) (<-chan []types.Node, <-chan error) {
	nodesChan := make(chan []types.Node)
	errChan := make(chan error, 1)

	go func() {
		defer close(nodesChan)
		defer close(errChan)

		// Set up event filters for container events
		filterArgs := filters.NewArgs()
		filterArgs.Add("type", "container")
		filterArgs.Add("event", "start")
		filterArgs.Add("event", "stop")
		filterArgs.Add("event", "die")
		filterArgs.Add("event", "kill")
		filterArgs.Add("event", "destroy")

		eventOptions := events.ListOptions{
			Filters: filterArgs,
		}

		eventsChan, errorsChan := s.cli.Events(ctx, eventOptions)

		// Send initial state
		nodes, err := s.Fetch(ctx)
		if err != nil {
			log.Printf("error fetching initial nodes: %v", err)
			errChan <- err
			return
		}

		select {
		case nodesChan <- nodes:
		case <-ctx.Done():
			return
		}

		// Watch for events
		for {
			select {
			case event := <-eventsChan:
				// Check if container has our label
				if event.Actor.Attributes != nil {
					if _, hasLabel := event.Actor.Attributes[s.LabelKey]; hasLabel {
						log.Printf("Container event detected: %s - %s", event.Action, event.Actor.Attributes["name"])

						// Fetch updated nodes
						nodes, err := s.Fetch(ctx)
						if err != nil {
							log.Printf("error fetching nodes after event: %v", err)
							continue
						}

						select {
						case nodesChan <- nodes:
						case <-ctx.Done():
							return
						}
					}
				}
			case err := <-errorsChan:
				if err != nil {
					log.Printf("error from Docker events stream: %v", err)
					errChan <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nodesChan, errChan
}

func (s *Source) Close(ctx context.Context) error {
	if s.cli != nil {
		if err := s.cli.Close(); err != nil {
			log.Printf("error closing Docker client: %v", err)
			return err
		}
	}
	return nil
}
