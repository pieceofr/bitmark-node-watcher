package main

import (
	"errors"
	"fmt"
)

var ( // Error variable
	// Process Error
	ErrorGetAPIFail              = errors.New("Get Docker API failed")
	ErrorStartMonitorService     = errors.New("StartMonitor failed")
	ErrorListenContainer         = errors.New("Listen Container failed")
	ErrorHandleExistingContainer = errors.New("Handle existing container failed")
	ErrorImageUpdateRoutine      = errors.New("Image update routine failed")
	// Container Errors
	ErrorContainerCreate        = errors.New("Container create failed")
	ErrorContainerStart         = errors.New("Container start  failed")
	ErrorConfigCreateNew        = errors.New("Create a new Config error")
	ErrorNamedContainerNotFound = errors.New("Named container is not found")

	// Image Errors
	ErrorImagePull = errors.New("Image pull failed")
)

func ErrCombind(cause, detail error) error {
	return fmt.Errorf("%s-%s", cause, detail)
}
