package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var version = "master"
var commit = "unknown"
var date = "unknown"
var dockerImage = ""

const (
	dockerAPIVersion string = "1.24"
	logPath          string = "/var/log/bitmark-node-watcher.log"
)

func main() {
	logfile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	log.SetOutput(logfile)
	log.SetLevel(log.InfoLevel)

	defer logfile.Close()

	// assign it to the standard logger
	app := cli.NewApp()
	app.Name = "bitmark-node-watcher"
	app.Version = version + " - " + commit + " - " + date
	app.Usage = "Automatically update running bitmark-node container"
	app.Before = before
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "daemon socket to connect to",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},

		cli.StringFlag{
			Name:  "image, i",
			Usage: "image name to pull",
			Value: "bitmark/bitmark-node-test",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "container name to create",
			Value: "bitmarkNodeTest",
		},
	}

	app.Action = func(c *cli.Context) error {
		ctx := context.Background()
		client, err := client.NewEnvClient()
		if err != nil {
			log.Error(ErrorGetAPIFail)
			return err
		}
		// Create a Docker API Client and current Context
		dockerImage := c.GlobalString("image")
		dockerRepo := "docker.io/" + dockerImage
		containerName := c.GlobalString("name")
		watcher := NodeWatcher{DockerClient: client, BackgroundContex: ctx,
			Repo: dockerRepo, ImageName: dockerImage, ContainerName: containerName}

		err = StartMonitor(watcher)
		if err != nil {
			log.Errorf(ErrorStartMonitorService.Error(), " image name:", watcher.ImageName)
			return err
		}
		log.Infof("Start Monitor host:", c.GlobalString("host"), "image:", c.GlobalString("image"))
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {
	// configure environment vars for client
	err := envConfig(c)
	if err != nil {
		log.Println("envConfig Error", err)
		return err
	}
	return nil
}

// envConfig translates the command-line options into environment variables
// that will initialize the api client
func envConfig(c *cli.Context) error {
	var err error
	err = setEnvOptStr("DOCKER_HOST", c.GlobalString("host"))
	err = setEnvOptStr("DOCKER_API_VERSION", dockerAPIVersion)
	return err
}

func setEnvOptStr(env string, opt string) error {
	if opt != "" && opt != os.Getenv(env) {
		err := os.Setenv(env, opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func setEnvOptBool(env string, opt bool) error {
	if opt == true {
		return setEnvOptStr(env, "1")
	}
	return nil
}
