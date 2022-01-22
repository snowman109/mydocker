package container

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

/**
这里的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，
这是本容器执行的第一个进程。
使用mount先去挂载proc文件系统，以便后面通过ps等系统命令查看当前进程资源情况
*/
func RunContainerInitProcess() error {
	cmdArray := ReadUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run contaner get user command error, cmdArray is nil")
	}
	// MS_NOEXEC 在本文件系统中不允许运行其他程序
	// MS_NOSUID 在本文件系统中运行程序的时候，不允许set-user-ID或get-group-ID
	// MS_NODEV 这个参数是自从Linux2.4以来，所有mount的系统都会默认设定的参数
	// 放到setUpMount中了
	setUpMount()
	//defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	//syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Println("Exec loop path error ", err.Error())
		return err
	}
	log.Println("Find path", path)
	// 不再使用传参的方式执行子进程，而是管道的方式
	//argv := []string{command}

	// 这一步是因为容器的第一个进程是init初始化进程，但ps的时候想看到pid为1的是用户进程，因此用syscall.Exec
	// 这个系统调用能将执行command，覆盖当前进程(init)的镜像、数据和堆栈等信息，包括PID
	if err = syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
func ReadUserCommand() []string {
	// uintptr(3)就是指index为3的文件描述符，也就是传递进来的管道的一端
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		fmt.Println("init read pipe error ", err.Error())
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
func setUpMount() {
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		log.Println("Get current location error ", err)
		return
	}
	log.Println("Current location is ", pwd)
	pivotRoot(pwd)
	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}
func pivotRoot(root string) error {
	// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount了一次
	// bind mount 是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error %v", err)
	}
	// 创建rootfs/.pivot_root存储old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的rootfs，老的old_root现在挂载在rootfs/.pivot_root上
	// 挂载点目前依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录当根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
