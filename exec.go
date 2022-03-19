package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mdocker/container"
	_ "mdocker/nsenter"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func ExecContainer(containerName string, commandArrays []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		fmt.Printf("Exec container getContainerPidByName %s error %v\n", containerName, err)
	}
	cmdStr := strings.Join(commandArrays, " ")
	fmt.Printf("container pid %s\n", pid)
	fmt.Printf("command %s\n", cmdStr)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	if err = cmd.Run(); err != nil {
		fmt.Printf("Exec container %s error %v", containerName, err)
	}
}

func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
