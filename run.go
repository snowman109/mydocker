package main

import (
	"fmt"
	"log"
	"mdocker/cgroups"
	"mdocker/cgroups/subsystems"
	"mdocker/container"
	"os"
	"strings"
)

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig) {

	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		fmt.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Println(err.Error())
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
	parent.Wait()
	os.Exit(-1)

}
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	fmt.Println("command all is ", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
