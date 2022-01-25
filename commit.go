package main

import (
	"fmt"
	"os/exec"
)

func commitContainer(imageName string) {
	mntURL := "/home/wyt/mnt"
	imageTar := "/home/wyt/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		fmt.Println(fmt.Sprintf("Tar folder %s error %v", mntURL, err))
	}
}
