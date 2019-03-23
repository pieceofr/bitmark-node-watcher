package main

// Check Current containers for bitmark-node container node

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	log "github.com/google/logger"
)

const newImageDownloadIndicator = "Downloaded newer image"
const (
	containerStateRunning = "running"
)

// pullImage Pull Specific Image
func (w *NodeWatcher) pullImage() (updated bool, err error) {
	reader, err := w.DockerClient.ImagePull(w.BackgroundContex, w.Repo, types.ImagePullOptions{})
	if err != nil {
		return false, err
	}
	defer reader.Close()
	response, err := ioutil.ReadAll(reader)
	updated = strings.Contains(string(response), newImageDownloadIndicator)
	return
}
func (w *NodeWatcher) createContainer(config CreateConfig) (container.ContainerCreateCreatedBody, error) {
	container, err := w.DockerClient.ContainerCreate(w.BackgroundContex, config.Config, config.HostConfig, config.NetworkingConfig, w.ContainerName)
	return container, err
}

// getTartgetContainers get the containers which has the same image name
// all needs to stop and remove
func (w *NodeWatcher) getContainersWithImage() ([]types.Container, error) {
	//get all containers
	containers, err := w.DockerClient.ContainerList(w.BackgroundContex, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	targetContainers := []types.Container{}
	for _, container := range containers {
		_, err := w.DockerClient.ContainerInspect(w.BackgroundContex, container.ID[:10])
		if err != nil {
			return targetContainers, err
		}

		if container.Image == w.ImageName {
			targetContainers = append(targetContainers, container)
		}
	}
	return targetContainers, nil
}

func (w *NodeWatcher) getOldContainer() (*types.Container, error) {
	//get all containers
	containers, err := w.DockerClient.ContainerList(w.BackgroundContex, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	oldContainerName := "/" + w.ContainerName + w.Postfix
	for _, container := range containers {
		if container.Names[0] == oldContainerName {
			return &container, nil
		}
	}
	return nil, nil
}

// stop the containers that needs to be stoped
func (w *NodeWatcher) stopContainers(containers []types.Container, stopTimeout time.Duration) error {
	for _, container := range containers {
		//stop container
		if container.State == containerStateRunning {
			err := w.DockerClient.ContainerStop(w.BackgroundContex, container.ID, &stopTimeout)
			if err != nil {
				return err
			}
			log.Info("container:", container.ID, " is stoped")
		}
	}
	return nil
}

func (w *NodeWatcher) startContainer(containerID string) error {
	if err := w.DockerClient.ContainerStart(w.BackgroundContex, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func (w *NodeWatcher) forceRemoveContainer(containerID string) error {
	if err := w.DockerClient.ContainerRemove(w.BackgroundContex, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func (w *NodeWatcher) removeContainer(containerID string) error {
	if err := w.DockerClient.ContainerRemove(w.BackgroundContex, containerID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func (w *NodeWatcher) renameContainer(container *types.Container) error {
	newName := container.Names[0] + w.Postfix
	err := w.DockerClient.ContainerRename(w.BackgroundContex, container.ID, newName)
	if err != nil {
		return err
	}
	log.Info("Container:", container.ID, " is rename to ", newName)
	return nil
}

func (w *NodeWatcher) getNamedContainer(c []types.Container) *types.Container {
	compareName := "/" + w.ContainerName
	for _, container := range c {
		for _, n := range container.Names {
			log.Info("getNamedContainer name:", n)
			if n == compareName {
				return &container
			}
		}
	}
	return nil
}
func (w *NodeWatcher) checkOldContainer(c []types.Container) types.Container {
	compareName := "/" + w.ContainerName + w.Postfix
	for _, container := range c {
		for _, n := range container.Names {
			log.Info("checkOldContainer name:", n)
			if n == compareName {
				return container
			}
		}
	}
	return types.Container{}
}
