package main

// Check Current containers for bitmark-node container node

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
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
	//io.Copy(os.Stdout, reader) // printout result
	defer reader.Close()
	response, err := ioutil.ReadAll(reader)
	updated = strings.Contains(string(response), newImageDownloadIndicator)
	return
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
	for _, container := range containers {
		if container.Names[0] == "/"+w.ContainerName+"_old" {
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
			log.Infof("container:", container.ID, " is stoped")
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
func (w *NodeWatcher) waitForStopContainer(container types.Container, waitTime time.Duration) error {
	timeout := time.After(waitTime)
	for {
		select {
		case <-timeout:
			return nil
		default:
			if ci, err := w.DockerClient.ContainerInspect(w.BackgroundContex, container.ID); err != nil {
				return err
			} else if !ci.State.Running {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *NodeWatcher) renameContainer(container *types.Container, postfix string) error {
	newName := container.Names[0] + postfix
	err := w.DockerClient.ContainerRename(w.BackgroundContex, container.ID, newName)
	if err != nil {
		return err
	}
	log.Infof("Container:", container.ID, " is rename to ", newName)
	return nil
}

func (w *NodeWatcher) getNamedContainer(c []types.Container) *types.Container {
	compareName := "/" + w.ContainerName
	for _, container := range c {
		for _, n := range container.Names {
			log.Infof("getNamedContainer name:", n)
			if n == compareName {
				return &container
			}
		}
	}
	return nil
}
func (w *NodeWatcher) checkOldContainer(c []types.Container) types.Container {
	compareName := "/" + w.ContainerName + "_old"
	for _, container := range c {
		for _, n := range container.Names {
			log.Infof("checkOldContainer name:", n)
			if n == compareName {
				return container
			}
		}
	}
	return types.Container{}
}
