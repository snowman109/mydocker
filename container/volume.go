package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func MountVolume(containerName string, volumeURLs []string) {
	// 创建宿主机文件目录
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir host dir %s error. %v", parentUrl, err))
		//return
	}
	// 在容器文件系统里创建挂载点
	containerUrl := volumeURLs[1]
	containerVolumeUrl := fmt.Sprintf(MntUrl, containerName) + containerUrl
	if err := os.Mkdir(containerVolumeUrl, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir container dir %s error. %v", containerVolumeUrl, err))
		//return
	}
	// 把宿主机文件目录挂载到容器挂载点
	workDir := parentUrl + "work"
	if err := os.Mkdir(workDir, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir work dir %s error. %v", workDir, err))
		//return
	}
	readonlyDir := "/tmp/read"
	if err := os.Mkdir(readonlyDir, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir readonly dir %s error. %v", readonlyDir, err))
		//return
	}
	// 把writeLayer目录和busybox目录mount到mnt目录下
	dirs := "-olowerdir=" + readonlyDir + ",upperdir=" + parentUrl + ",workdir=" + workDir
	cmd := exec.Command("mount", "-t", "overlay", "overlay", dirs, containerVolumeUrl)
	fmt.Println(cmd.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprintf("Mount volume failed. %v", err))
	}
}

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}
func CreateReadOnlyLayer(imageName string) {
	busyboxURL := RootUrl + "/" + imageName + "/"
	busyboxTarURL := RootUrl + "/" + imageName + ".tar"
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
func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
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
func CreateMountPoint(containerName, imageName string) {
	// 创建mnt文件夹作为挂载点
	mntURL := fmt.Sprintf(MntUrl, containerName)
	if err := os.Mkdir(mntURL, 0777); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", mntURL, err))
	}
	workDir := fmt.Sprintf(WorkUrl, containerName)
	if err := os.Mkdir(workDir, 0755); err != nil {
		fmt.Println(fmt.Sprintf("Mkdir dir %s error. %v", workDir, err))
	}
	// 把writeLayer目录和busybox目录mount到mnt目录下
	dirs := "-olowerdir=" + RootUrl + "/" + imageName + ",upperdir=" + fmt.Sprintf(WriteLayerUrl, containerName) +
		",workdir=" + workDir
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
func DeleteWorkSpace(containerName, volume string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(containerName, volumeURLs)
		}
	}
	DeleteMountPoint(containerName)
	DeleteWriteLayer(containerName)
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	workURL := fmt.Sprintf(WorkUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		fmt.Println(fmt.Sprintf("Remove writelayer %s error %v", writeURL, err))
	}
	if err := os.RemoveAll(workURL); err != nil {
		fmt.Println(fmt.Sprintf("Remove word dir %s error %v", workURL, err))
	}
}

func DeleteMountPoint(containerName string) {
	// 卸载整个容器文件系统的挂载点
	mntURL := fmt.Sprintf(MntUrl, containerName)
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		fmt.Println(fmt.Sprintf("Unmount %s error %v", mntURL, err))
		return
	}
	// 删除容器文件系统挂载点
	if err := os.RemoveAll(mntURL); err != nil {
		fmt.Println(fmt.Sprintf("Remove mountpoint dir %s error %v", mntURL, err))
	}

}
func DeleteMountPointWithVolume(containerName string, volumeURLs []string) {
	// 卸载容器里volume挂载点的文件系统
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntUrl + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprintf("Umount volume failed. %v", err))
	}
}
