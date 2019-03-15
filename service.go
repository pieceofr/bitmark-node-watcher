package main

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

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
			log.Error(ErrCombind(ErrorHandleExistingContainer, err))
			continue
		}
		// get ports and attach volumns because they are key information to create bitmark-node-container
		if createConf != nil { // err == nil and createConf == nil => container does not exist
			createdContainer, err := watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, createConf.Config,
				createConf.HostConfig, createConf.NetworkingConfig, watcher.ContainerName)

			if err != nil {
				log.Error(ErrCombind(ErrorContainerCreate, err))
				continue
			}
			err = watcher.startContainer(createdContainer.ID)
			if err != nil {
				log.Error(ErrCombind(ErrorContainerStart, err))
				continue
			}
			log.Info("Start container successfully")
		} else {
			log.Info("Creating a brand new container")
			newContainerConfig, err := getDefaultConfig(&watcher)
			if err != nil {
				log.Error(ErrCombind(ErrorConfigCreateNew, err))
				continue
			}
			newContainer, err := watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, newContainerConfig.Config,
				newContainerConfig.HostConfig, nil, watcher.ContainerName)
			if err != nil {
				log.Error(ErrCombind(ErrorContainerCreate, err))
				continue
			}
			err = watcher.startContainer(newContainer.ID)
			if err != nil {
				log.Error(ErrCombind(ErrorContainerStart, err))
				continue
			}
			log.Info("Start container successfully")
		}

	}
	return nil
}

// imageUpdateRoutine check image periodically
func imageUpdateRoutine(w *NodeWatcher, updateStatus chan bool) {
	ticker := time.NewTicker(pullImageInterval)
	defer func() {
		ticker.Stop()
		close(updateStatus)
	}()
	// For the first time
	newImage, err := w.pullImage()
	if err != nil {
		log.Info(ErrCombind(ErrorImagePull, err))
	}
	if newImage {
		log.Info("imageUpdateRoutine update a new image")
		updateStatus <- true
	}
	// End of the first time ---- can be delete later
	for { // start  periodically check routine
		select {
		case <-ticker.C:
			newImage, err := w.pullImage()
			if err != nil {
				log.Info(ErrCombind(ErrorImagePull, err))
				continue
			}
			if newImage {
				log.Info("imageUpdateRoutine update a new image")
				updateStatus <- true
				break

			}
			log.Info("no new image found")
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

	for _, container := range nodeContainers {
		log.Println("Container Info", dbgContainerInfo(container))

	}

	if len(nodeContainers) != 0 {
		nameContainer := watcher.getNamedContainer(nodeContainers)
		if nameContainer == nil { //not found is not an error
			log.Warnf(ErrorNamedContainerNotFound.Error())
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

		namedContainers := append([]types.Container{}, *nameContainer)
		err = watcher.stopContainers(namedContainers, containerStopWaitTime)

		if err != nil {
			return nil, err
		}

		oldContainers, err := watcher.getOldContainer()
		if err == nil && oldContainers != nil {
			watcher.forceRemoveContainer(oldContainers.ID)
		}
		err = watcher.renameContainer(nameContainer, "_old")
		if err != nil {
			return nil, err
		}
		log.Println("===Config ===")
		log.Println(jsonConfig.Config)
		log.Println("====HostConfig====")
		log.Println(jsonConfig.HostConfig)
		log.Println("===networkSetting===")
		log.Println(jsonConfig.NetworkSettings)

		return &CreateConfig{Config: &newConfig, HostConfig: jsonConfig.HostConfig, NetworkingConfig: &newNetworkConf}, err
	}
	// no container
	return nil, nil
}

func getDefaultConfig(watcher *NodeWatcher) (*CreateConfig, error) {
	config := CreateConfig{}

	baseDir, err := builVolumSrcBaseDir(watcher)
	if err != nil {
		return nil, err
	}
	baseTargetDir := "/.config/bitmark-node"
	publicIP := os.Getenv("PUBLIC_IP")
	chain := os.Getenv("NETWORK")
	if len(publicIP) == 0 {
		publicIP = "127.0.0.1"
	}
	if len(chain) == 0 {
		chain = "BITMARK"
	}
	additionEnv := append([]string{}, "PUBLIC_IP="+publicIP, "NETWORK="+chain)
	exposePorts := nat.PortMap{
		"2136/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "2136",
			},
		},
		"2130/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "2130",
			},
		},
		"2131/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "2131",
			},
		},
		"9980/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9980",
			},
		},
	}
	hconfig := container.HostConfig{
		NetworkMode:  "default",
		PortBindings: exposePorts,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/db",
				Target: baseTargetDir + "/db",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/data",
				Target: baseTargetDir + "/bitmarkd/bitmark/data",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/data-test",
				Target: baseTargetDir + "/bitmarkd/testing/data",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/log",
				Target: baseTargetDir + "/bitmarkd/bitmark/log",
			},
			{
				Type:   mount.TypeBind,
				Source: baseDir + "/log-test",
				Target: baseTargetDir + "/bitmarkd/testing/log",
			},
		},
	}
	config.HostConfig = &hconfig
	portmap := nat.PortSet{
		"2136/tcp": struct{}{},
		"2130/tcp": struct{}{},
		"2131/tcp": struct{}{},
		"9980/tcp": struct{}{},
	}
	config.Config = &container.Config{
		Image:        watcher.ImageName,
		Env:          additionEnv,
		ExposedPorts: portmap,
	}

	log.Println("HostConfig=======")
	log.Println(hconfig)
	return &config, nil
}

func removeDefaultDB() {

}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func builVolumSrcBaseDir(watcher *NodeWatcher) (string, error) {
	splitDir := strings.Split(watcher.ImageName, "/")
	if len(splitDir) < 2 {
		return "", errors.New("wrong image name")
	}
	// insert data into directory
	splitName := strings.Split(splitDir[1], "-")
	var sb strings.Builder
	if len(splitName) == 3 { //stage directory
		for idx, s := range splitName {
			sb.WriteString(s)
			if s == "node" {
				sb.WriteString("-data")
			}
			if idx < len(splitName)-1 {
				sb.WriteString("-")
			}
		}
	} else {
		return "", errors.New("source of attach volumne name is not expected")
	}

	// Prepare Host Config
	baseDir := userHomeDir() + "/" + sb.String()
	log.Println("baseDir", baseDir)
	return baseDir, nil
}
