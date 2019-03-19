package main

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func createWatcher() (NodeWatcher, error) {
	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.39"))
	if err != nil {
		log.Println("Get Docker API Failed")
		return NodeWatcher{}, err
	}
	return NodeWatcher{DockerAPI: client, BackgroundContex: ctx, Repo: "hello-world"}, nil
}

func TestPullImage(t *testing.T) {
	log.Println("start TestTargetContainer ...")
	watcher, err := createWatcher()
	assert.NoError(t, err, "TestPullImage fail to create NodeWatcher")
	err = watcher.pullImage()
	assert.NoError(t, err, "TestPullImage fail to pull NodeWatcher")
}

func TestTargetContainer(t *testing.T) {
	log.Println("start TestTargetContainer ...")
	watcher, err := createWatcher()
	assert.NoError(t, err, "TestTargetContainer fail to create NodeWatcher")
	containers, err := watcher.getTartgetContainers()
	assert.NoError(t, err, "TestTargetContainer fail error")
	assert.True(t, (len(containers) > 0), "no contianers exist")
}
