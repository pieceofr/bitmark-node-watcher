package main

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var mockData *MockData

type MockData struct {
	Watcher NodeWatcher
	BaseDir string
	Chain   string
	SubDir  map[string]string
	Env     map[string]string
}

func TestMain(m *testing.M) {
	mockData = &MockData{}
	mockData.init()
	if err := cleanContainers(); err != nil {
		log.Info("Setup Error:", err)
		panic("cleanContainer Error")
	}
	os.Exit(m.Run())
}

func TestPullImage(t *testing.T) {
	watcher := mockData.getWatcher()
	_, err := watcher.pullImage()
	assert.NoError(t, err, ErrorImagePull.Error())
}

func TestStartContainer(t *testing.T) {
	watcher := mockData.getWatcher()
	newContainerConfig, err := getDefaultConfig(watcher)
	assert.NoError(t, err, ErrorConfigCreateNew.Error())
	newContainer, err := watcher.DockerClient.ContainerCreate(watcher.BackgroundContex, newContainerConfig.Config,
		newContainerConfig.HostConfig, nil, watcher.ContainerName)
	assert.NoError(t, err, ErrorContainerCreate.Error())

	err = watcher.startContainer(newContainer.ID)
	assert.NoError(t, err, ErrorContainerStart.Error())

}

func TestStopContainer(t *testing.T) {
	watcher := mockData.getWatcher()
	containers, _ := watcher.getContainersWithImage()
	err := watcher.stopContainers(containers, 10*time.Second)
	assert.NoError(t, err, ErrorContainerStop.Error())
}

func (mock *MockData) init() error {
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	mock.Watcher = NodeWatcher{DockerClient: client, BackgroundContex: ctx,
		Repo:          "docker.io/bitmark/bitmark-node-test",
		ImageName:     "bitmark/bitmark-node-test",
		ContainerName: "bitmarkNodeTest"}
	// Create Directory For Test
	mock.createDir()
	mock.BaseDir = userHomeDir() + "/bitmark-node-data-test"
	mock.Chain = "bitmark"
	// Create Environment Variable
	mock.Env = make(map[string]string)
	mock.Env["PUBLIC_IP"] = "127.0.0.1"
	mock.Env["NETWORK"] = mock.Chain
	mock.Env["DOCKER_HOST"] = "unix:///var/run/docker.sock"
	mock.Env["NODE_IMAGE"] = "bitmark/bitmark-node-test"
	mock.Env["NODE_NAME"] = "bitmarkNodeTest"
	mock.Env["USER_NODE_BASE_DIR"] = mock.BaseDir
	for k, v := range mock.Env {
		log.Info("key:", k, " val:", v)
	}
	mock.SubDir = make(map[string]string)
	mock.SubDir["dirNodeDB"] = "/db"
	mock.SubDir["dirMainDB"] = "/data"
	mock.SubDir["dirTestDB"] = "/data-test"
	mock.SubDir["dirMainLog"] = "/log"
	mock.SubDir["dirTestLog"] = "/log-test"
	// Create sub directory  names
	return nil
}
func (mock *MockData) getWatcher() *NodeWatcher {
	return &mock.Watcher
}

func (mock *MockData) getSubDir(sub string) string {
	log.Info("getSubDir:", mock.BaseDir+mock.SubDir[sub])
	return mock.BaseDir + mock.SubDir[sub]
}

func (mock *MockData) createDir() error {
	// Remove file if exist
	if _, err := os.Stat(mock.BaseDir); nil == err {
		os.RemoveAll(mock.BaseDir)
	}
	//Create file Dirs
	if err := os.MkdirAll(mock.getSubDir("dirNodeDB"), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(mock.getSubDir("dirNodeDB")); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(mock.getSubDir("dirMainDB"), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(mock.getSubDir("dirMainDB")); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(mock.getSubDir("dirTestDB"), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(mock.getSubDir("dirTestDB")); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(mock.getSubDir("dirMainLog"), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(mock.getSubDir("dirMainLog")); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(mock.getSubDir("dirTestLog"), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(mock.getSubDir("dirTestLog")); os.IsNotExist(err) {
		return err
	}
	return nil
}

// Utilies
func cleanContainers() error {
	// Clean Container
	watcher := mockData.getWatcher()
	containers, err := watcher.getContainersWithImage()
	if err != nil {
		return err
	}
	for _, container := range containers {
		log.Infof("Container ID:%s is going to be removed", container.ID)
		watcher.forceRemoveContainer(container.ID)
	}
	containers, err = watcher.getContainersWithImage()
	if len(containers) > 0 {
		return errors.New("clean container error")
	}
	return nil
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
