package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mdocker/cgroups"
	"mdocker/cgroups/subsystems"
	"mdocker/container"
	"os"
	"strconv"
	"strings"
	"time"
)

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string, name string, imageName string, envSlice []string) {
	parent, writePipe := container.NewParentProcess(tty, name, volume, imageName, envSlice)
	if parent == nil {
		fmt.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Println(err.Error())
	}
	// 记录容器信息
	containerName, err := recordContainerInfo(parent.Process.Pid, cmdArray, name, volume)
	if err != nil {
		fmt.Println(fmt.Sprintf("Record container info error %v", err))
		return
	}
	// use mydocker-cgroup as cgroup name
	// 创建cgroup manager，并通过调用set和apply设置资源限制并使用限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	// 设置资源限制
	cgroupManager.Set(res)
	// 将容器进程加入到各个subsystem挂载对应的cgroup中
	cgroupManager.Apply(parent.Process.Pid)

	// 对容器设置完限制之后，初始化容器
	// 发送用户命令
	sendInitCommand(cmdArray, writePipe)
	// 加入detach后，这里加入了判断。原来parent.wait()主要用于父进程等待子进程结束，这在交互式创建容器的步骤里没问题。但如果创建了detach容器，就不能等待了，创建完成后父进程就结束了。
	if tty {
		parent.Wait() //
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(containerName, volume)
	}
	//os.Exit(0)
}

// 如果使用tty方式的容器，那么容器退出后，就会删除容器的相关信息
func deleteContainerInfo(name string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, name)
	if err := os.RemoveAll(dirURL); err != nil {
		fmt.Println(fmt.Sprintf("Remove dir %s error %v", dirURL, err))
	}
}
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	fmt.Println("command all is ", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

// randStringBytes Id生成器
func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// recordContainerInfo 保存容器信息到宿主机文件系统
func recordContainerInfo(containerPID int, commandArray []string, containerName, volume string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	// 如果用户不指定容器名，以容器Id当名
	if containerName == "" {
		containerName = id
	}
	containerInfo := &container.ContainerInfo{
		Pid:         strconv.Itoa(containerPID),
		Id:          id,
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Volume:      volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		fmt.Println(fmt.Sprintf("Record container info error %v", err))
		return "", err
	}
	jsonStr := string(jsonBytes)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err = os.Mkdir(dirUrl, 0622); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir error %s error %v", dirUrl, err))
		//return "", err
	} //这里在创建log的时候已经创建文件夹了
	fileName := dirUrl + container.ConfigName
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(fmt.Sprintf("Create file %s error %v", file, err))
		return "", err
	}
	defer file.Close()
	if _, err = file.WriteString(jsonStr); err != nil {
		fmt.Println(fmt.Sprintf("File write string error %v", err))
		return "", err
	}
	return containerName, nil
}
