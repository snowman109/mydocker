package main

import (
	"fmt"
	"io/ioutil"
	"mdocker/container"
	"os"
	"path"
)

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := path.Join(dirURL, container.ContainerLogFile)
	file, err := os.Open(logFileLocation)
	if err != nil {
		fmt.Println(fmt.Sprintf("Log container open file %s error %v", logFileLocation, err))
		return
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(fmt.Sprintf("Log container read file %s error %v", logFileLocation, err))
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}
