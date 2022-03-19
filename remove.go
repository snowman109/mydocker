package main

import (
	"fmt"
	"mdocker/container"
	"os"
)

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		fmt.Printf("Get container %s info error %v\n", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP {
		fmt.Printf("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err = os.RemoveAll(dirURL); err != nil {
		fmt.Printf("Remove file %s error %v\n", dirURL, err)
		return
	}
	container.DeleteWorkSpace(containerName, containerInfo.Volume)
}
