package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`         // 容器的init进程在宿主机上的pid
	Id          string `json:"id"`          // 容器的id
	Name        string `json:"name"`        // 容器名
	Command     string `json:"command"`     // 容器内init进程的运行命令
	CreatedTime string `json:"createdTime"` // 创建时间
	Status      string `json:"status"`      // 容器的状态
	Volume      string `json:"volume"`      // 容器的数据卷
}

const (
	RUNNING             = "running"
	STOP                = "stop"
	Exit                = "exit"
	DefaultInfoLocation = "/var/run/mydocker/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
	RootUrl             = "/home/wyt"
	MntUrl              = "/home/wyt/mnt/%s"
	WriteLayerUrl       = "/home/wyt/writeLayer/%s"
	WorkUrl             = "/home/wyt/work/%s"
)

// NewParentProcess
// 这里是父进程
// 1.这里的/proc/self/exe调用中，/proc/self指的是当前进程运行自己的环境，exec其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
// 2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会调用initCommand取初始化进程的一些环境和资源
// 3.下来的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境
// 4.如果用户指定了-ti参数，就需要把当前进程的输入输出导入到标准输入输出上
func NewParentProcess(tty bool, containerName, volume, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		fmt.Println("New pipe error ", err.Error())
		return nil, nil
	}
	// 加入pipe后，这里不用再通过cmd将剩余参数（/bin/sh）这些传给容器进程，而是通过管道去传输
	//args := []string{"init", command}
	//cmd := exec.Command("/proc/self/exe", args...)
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 生成容器对应目录的container.log文件
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err = os.MkdirAll(dirURL, 0622); err != nil {
			fmt.Println(fmt.Sprintf("NewParentProcess mkdir %s error %v", dirURL, err))
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, fileErr := os.Create(stdLogFilePath)
		if fileErr != nil {
			fmt.Println(fmt.Sprintf("NewParentProcess create file %s error %v", stdLogFilePath, fileErr))
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	cmd.ExtraFiles = []*os.File{readPipe} // 带着文件句柄去创建子进程
	cmd.Env = append(os.Environ(), envSlice...)
	//cmd.Dir = "/home/wyt/busybox"
	NewWorkSpace(containerName, volume, imageName)
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	return cmd, writePipe
}
func NewPipe() (*os.File, *os.File, error) {
	return os.Pipe()
}
func NewWorkSpace(containerName string, volume string, imageName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(containerName, volumeURLs)
			fmt.Println(volumeURLs)
		} else {
			fmt.Println("Volume parameter input is not correct.")
		}
	}
	//CreateMountPoint(containerName, "busybox")
}
