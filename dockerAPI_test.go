package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	baseDir     string = "~/bitmark-node-unit-test"
	nodeRepo    string = "docker.io/bitmark/bitmark-node"
	dockerHost  string = "unix:///var/run/docker.sock"
	dockerImage string = "bitmark/bitmark-node"
	nodeName    string = "bitmarkNode"
	dirNodeDB   string = "/db"
	dirMainDB   string = "/data"
	dirTestDB   string = "/data-test"
	publicIP    string = "127.0.0.1"
	chain       string = "bitmark"
)

func TestMain(m *testing.M) {
	if err := cleanContainers(); err != nil {
		log.Println("Setup Error:", err)
		return
	}
	if err := setEnv(); err != nil {
		log.Println("Setup Error:", err)
		return
	}
	os.Exit(m.Run())
}

func createWatcher() (NodeWatcher, error) {
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		return NodeWatcher{}, err
	}
	return NodeWatcher{DockerClient: client, BackgroundContex: ctx, Repo: nodeRepo}, nil
}
func TestCreateWatcher(t *testing.T) {
	_, err := createWatcher()
	assert.NoError(t, err, ErrorCreateWatcher.Error())
}

func TestPullImage(t *testing.T) {
	watcher, _ := createWatcher()
	updated, err := watcher.pullImage()
	assert.NoError(t, err, ErrorImagePull.Error())
	assert.Equal(t, updated, true, "Image not updated")
}

func TestCreateContainer(t *testing.T) {
	watcher, _ := createWatcher()
	newContainerConfig, err := getDefaultConfig(&watcher)
	assert.NoError(t, err, ErrorConfigCreateNew.Error())
	_, err = watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, newContainerConfig.Config,
		newContainerConfig.HostConfig, nil, watcher.ContainerName)
	assert.NoError(t, err, ErrorContainerCreate.Error())
}

func TestStartContainer(t *testing.T) {
	watcher, _ := createWatcher()
	newContainerConfig, _ := getDefaultConfig(&watcher)
	newContainer, _ := watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, newContainerConfig.Config,
		newContainerConfig.HostConfig, nil, watcher.ContainerName)
	err := watcher.startContainer(newContainer.ID)
	assert.NoError(t, err, ErrorContainerStart.Error())
}

func cleanContainers() error {
	// Clean Container
	w, err := createWatcher()
	if err != nil {
		return err
	}
	containers, err := w.getContainersWithImage()
	for _, container := range containers {
		w.DockerClient.ContainerRemove(w.BackgroundContex, container.ID, types.ContainerRemoveOptions{})
	}
	containers, err = w.getContainersWithImage()
	if len(containers) > 0 {
		return errors.New("clean container error")
	}
	return nil
}

func setEnv() error {
	// Set Env variable
	os.Setenv("PUBLIC_IP", publicIP)
	os.Setenv("NETWORK", chain)
	os.Setenv("DOCKER_HOST", dockerHost)
	os.Setenv("NODE_IMAGE", dockerImage)
	os.Setenv("NODE_NAME", nodeName)
	os.Setenv("USER_HOME_BASE_DIR", baseDir)
	// Remove file if exist
	if _, err := os.Stat(baseDir); nil == err {
		os.RemoveAll(baseDir)
	}
	//Create file Dirs
	if err := os.Mkdir(baseDir, 0700); err != nil {
		return err
	}
	if err := os.Mkdir(baseDir+dirNodeDB, 0700); err != nil {
		return err
	}
	if err := os.Mkdir(baseDir+dirMainDB, 0700); err != nil {
		return err
	}
	if err := os.Mkdir(baseDir+dirTestDB, 0700); err != nil {
		return err
	}
	return nil
}
