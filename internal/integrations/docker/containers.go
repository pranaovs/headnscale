package docker

import (
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/client"
)

func GetRunning(cli client.SDKClient, ctx context.Context) ([]container.Summary, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Fatalf("Error listing containers: %v", err)
		return nil, err
	}

	containersRunning := []container.Summary{}

	for _, container := range containers {
		if container.State == "running" {
			containersRunning = append(containersRunning, container)
		}
	}

	return containersRunning, nil
}

func GetLabelled(containers []container.Summary, labelKey string) ([]container.Summary, error) {
	labeledContainers := []container.Summary{}

	for _, container := range containers {
		if _, ok := container.Labels[labelKey]; ok {
			labeledContainers = append(labeledContainers, container)
		}
	}

	return labeledContainers, nil
}

func GetLabels(containers []container.Summary, labelKey string) ([]string, error) {
	values := []string{}

	for _, container := range containers {
		if labelValue, ok := container.Labels[labelKey]; ok {
			if trimmedHostname := strings.TrimSpace(labelValue); trimmedHostname != "" {
				values = append(values, trimmedHostname)
			}
		}
	}

	return values, nil
}
