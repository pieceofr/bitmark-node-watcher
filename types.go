package main

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
)

// NodeWatcher main data structure of service
type NodeWatcher struct {
	DockerClient     *dockerClient.Client
	BackgroundContex context.Context
	Repo             string
	ImageName        string
	ContainerName    string
	Postfix          string
}

// CreateConfig collect configs to create a container
type CreateConfig struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}
