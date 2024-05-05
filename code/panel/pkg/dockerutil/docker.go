package dockerutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	cli *client.Client
}

var dockerClient *DockerClient

func SetupClient() error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		return err
	}

	dockerClient = &DockerClient{cli: cli}

	return nil
}

func GetClient() *DockerClient {
	return dockerClient
}

type DockerErr string

func (d DockerClient) StopAndRemoveContainer(ctx context.Context, containerName string) error {
	timeout := 100
	err := d.cli.ContainerStop(ctx, containerName, container.StopOptions{
		Signal:  "SIGKILL",
		Timeout: &timeout,
	})
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		return err
	}

	err = d.cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		return err
	}
	return nil
}

func (d DockerClient) GetContainerStatusByName(ctx context.Context, containerName string) (string, error) {
	// List containers to find the one with the specified name
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return "", err
	}

	// Search for the container by name
	var containerID string
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				containerID = container.ID
				break
			}
		}
	}

	if containerID == "" {
		return "", fmt.Errorf("Container with name %s not found", containerName)
	}

	// Inspect the container to get its status
	containerInfo, err := d.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}

	return containerInfo.State.Status, nil
}

func (d DockerClient) GetIDFromName(ctx context.Context, containerName string) string {
	// List containers to find the one with the specified name
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return ""
	}

	// Search for the container by name
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				return container.ID
			}
		}
	}

	return ""
}
