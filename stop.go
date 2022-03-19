package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mdocker/container"
	"strconv"
	"syscall"
)

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Read file %s error %v\n", configFilePath, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		fmt.Printf("GetContainerInfoByName unmarshal error %v\n", err)
		return nil, err
	}
	return &containerInfo, nil
}

func stopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		fmt.Printf("Get container pid by name %s error %v\n", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		fmt.Printf("Convert pid from string to int error %v\n", err)
		return
	}
	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		fmt.Printf("Stop container %s error %v\n", containerName, err)
		return
	}

	// 根据容器名获取对应的信息对象
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		fmt.Printf("Get container %s info error %v\n", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		fmt.Printf("Json marshal %s error %v\n", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	if err = ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		fmt.Printf("Write file %s error %v\n", containerName, err)
	}
}
