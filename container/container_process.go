package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess
// 这里是父进程
// 1.这里的/proc/self/exe调用中，/proc/self指的是当前进程运行自己的环境，exec其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
// 2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会调用initCommand取初始化进程的一些环境和资源
// 3.下来的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境
// 4.如果用户指定了-ti参数，就需要把当前进程的输入输出导入到标准输入输出上
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	}
	cmd.ExtraFiles = []*os.File{readPipe} // 带着文件句柄去创建子进程
	//cmd.Dir = "/home/wyt/busybox"
	mntURL := "/home/wyt/mnt/"
	rootURL := "/home/wyt/"
	NewWorkSpace(rootURL, mntURL)
	cmd.Dir = mntURL
	return cmd, writePipe
}
func NewPipe() (*os.File, *os.File, error) {
	return os.Pipe()
}
func NewWorkSpace(rootURL string, mntURL string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
}
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		fmt.Println(fmt.Sprintf("Fail to judge whether dir %s exists. %v", busyboxURL, err))
		return
	}
	if !exist {
		if err = os.Mkdir(busyboxURL, 0777); err != nil {
			fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", busyboxURL, err))
			return
		}
		if _, err = exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			fmt.Println(fmt.Sprintf("unTar dir %s error %v", busyboxTarURL, err))
		}

	}
}
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	exist, err := PathExists(writeURL)
	if err != nil {
		fmt.Println(fmt.Sprintf("Fail to judge whether dir %s exists. %v", writeURL, err))
		return
	}
	if !exist {
		if err := os.Mkdir(writeURL, 0777); err != nil {
			fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", writeURL, err))
		}
	}
}
func CreateMountPoint(rootURL string, mntURL string) {
	// 创建mnt文件夹作为挂载点
	if err := os.Mkdir(mntURL, 0777); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", mntURL, err))
	}
	workDir := rootURL + "work"
	if err := os.Mkdir(workDir, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", workDir, err))
	}
	// 把writeLayer目录和busybox目录mount到mnt目录下
	dirs := "-olowerdir=" + rootURL + "busybox,upperdir=" + rootURL + "writeLayer,workdir=" + rootURL + "work"
	cmd := exec.Command("mount", "-t", "overlay", "overlay", dirs, mntURL)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Println("mount mnt false: ", err)
	}
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
