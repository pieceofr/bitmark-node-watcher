package main

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
)

func dbgContainerInfo(c types.Container) string {
	portsString := ""
	for idx, port := range c.Ports {
		portStr := "{" + "IP: " + port.IP + "," + "PrivatePort: " + strconv.Itoa(int(port.PrivatePort)) + "," + "PublicPort: " + strconv.Itoa(int(port.PublicPort)) + "," + "Type: " + port.Type + "}"
		if idx != 0 {
			portsString = portsString + ", " + portStr
		} else {
			portsString = portsString + portStr
		}
	}
	return fmt.Sprintf("\n{\nID:%s\n  Names:%s\n  Image:%s\n  ImageID:%s\n  Command:%s\n  Created:%d\n  Ports:%s\n  SizeRw:%d\n  SizeRootFs:%d\n  Labels:%s\n  State:%s\n  Status:%s\n  HostConfig:%s\n  NetworkSettings:%s\n  Mounts:%s\n}\n",
		c.ID, c.Names, c.Image, c.ImageID, c.Command, c.Created, portsString, c.SizeRw, c.SizeRootFs, c.Labels, c.State, c.Status, c.HostConfig, *c.NetworkSettings, c.Mounts)
}
