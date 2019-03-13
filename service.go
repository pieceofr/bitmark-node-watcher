package main

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"

	//	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
)

const (
	containerStopWaitTime = 10 * time.Second
	pullImageInterval     = 20 * time.Second
)

// StartMonitor  Monitor process
func StartMonitor(watcher NodeWatcher) error {
	for {
		updated := make(chan bool)
		go imageUpdateRoutine(&watcher, updated)
		<-updated
		createConf, err := handleExistingContainer(watcher)

		if err != nil {
			log.Println("handleExistingContainer:", err)
			continue
		}

		// get ports and attach volumns because they are key information to create bitmark-node-container
		if createConf != nil { // err == nil and createConf == nil => container does not exist
			createdContainer, err := watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, createConf.Config,
				createConf.HostConfig, createConf.NetworkingConfig, watcher.ContainerName)
			if err != nil {
				log.Println("conatiner create failed")
				continue
			}
			err = watcher.startContainers(createdContainer.ID)
			if err != nil {
				log.Println("start container fail:", err)
				continue
			}
			log.Println("start container successfully:", err)
		} else { // Should not happend .. but ...
			log.Println("brand new container, implement later", err)
		}

	}
	return nil
}

// imageUpdateRoutine check image periodically
func imageUpdateRoutine(w *NodeWatcher, updateStatus chan bool) {
	// pull Image and check if image has been updated
	ticker := time.NewTicker(pullImageInterval)
	defer func() {
		ticker.Stop()
		close(updateStatus) // use the new channel, old channel should
	}()
	for { // start  periodically check routine
		log.Println("start a new Image Check routine")
		select {
		case <-ticker.C:
			newImage, err := w.pullImage()
			if err != nil {
				log.Println("imageUpdateRoutine Check Image Error", err)
				continue
			}
			if newImage {
				log.Println("imageUpdateRoutine update the image")
				updateStatus <- true
				break

			}
			log.Println("no new image found")
		}
	}
}

// handleExistingContainer handle existing container and return old container config for recreating a new container
func handleExistingContainer(watcher NodeWatcher) (*CreateConfig, error) {
	nodeContainers, err := watcher.getContainersWithImage()
	if err != nil { //not found is not an error
		log.Println("getContainersWithImage err", err)
		return nil, nil
	}

	for _, item := range nodeContainers {
		log.Println("Container Info", dbgContainerInfo(item))
	}
	// stop containers and rename the containers with postfix
	if len(nodeContainers) != 0 {
		nameContainer := watcher.getContainer(nodeContainers)
		if nameContainer == nil { //not found is not an error
			return nil, nil
		}
		jsonConfig, err := watcher.DockerClient.ContainerInspect(watcher.BackgroundContex, nameContainer.ID)

		if err != nil { //inspect fail is an error because we can not do anything about existing error
			return nil, err
		}
		newConfig := container.Config{
			Image:        watcher.ImageName,
			ExposedPorts: jsonConfig.Config.ExposedPorts,
			Env:          jsonConfig.Config.Env,
			Volumes:      jsonConfig.Config.Volumes,
			Cmd:          jsonConfig.Config.Cmd,
		}

		newNetworkConf := network.NetworkingConfig{
			EndpointsConfig: jsonConfig.NetworkSettings.Networks,
		}

		err = watcher.stopContainers(nodeContainers, containerStopWaitTime)
		if err != nil {
			return nil, err
		}

		oldContainers, err := watcher.getOldContainers()
		if err == nil && len(oldContainers) != 0 {
			for _, container := range oldContainers {
				log.Println("old container id", container.ID)
				watcher.forceRemoveContainers(container.ID)
			}

		}
		err = watcher.renameContainers(nodeContainers, "_old")
		if err != nil {
			return nil, err
		}
		return &CreateConfig{Config: &newConfig, HostConfig: jsonConfig.HostConfig, NetworkingConfig: &newNetworkConf}, err
	}
	// no container
	return nil, nil
}

func getDefaultConfig(watcher *NodeWatcher) (*CreateConfig, error) {
	config := CreateConfig{}
	splites := strings.Split(watcher.ImageName, "/")
	if len(splites) < 2 {
		return nil, errors.New("wrong image name")
	}
	baseDir := UserHomeDir() + splites[2]
	log.Println("baseDir", baseDir)
	/*
		exposedPorts, portBindings, _ := nat.ParsePortSpecs([]string{
			"0.0.0.0:8080:2368",
		})
	*/
	hconfig := container.HostConfig{
		//PortBindings: portBindings,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/db",
				Target: "/.config/bitmark-node/db",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/data",
				Target: "/.config/bitmark-node/bitmarkd/bitmark/data",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/data-test",
				Target: "/.config/bitmark-node/bitmarkd/bitmark/data-test",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/log",
				Target: "/.config/bitmark-node/bitmarkd/bitmark/log",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/logtest",
				Target: "/.config/bitmark-node/bitmarkd/testing/log",
			},
		},
	}
	config.HostConfig = &hconfig

	return &config, nil
}

func removeDefaultDB() {

}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
