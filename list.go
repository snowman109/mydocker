package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mdocker/container"
	"os"
	"path"
	"text/tabwriter"
)

func ListContainers() {
	// 找到存储容器信息的路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	// 读取所有文件
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		fmt.Println(fmt.Sprintf("Read dir %s error %v", dirURL, err))
		return
	}
	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, containerErr := getContainerInfo(file)
		if containerErr != nil {
			fmt.Println(fmt.Sprintf("Get container info %s error %v", file.Name(), tmpContainer))
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 使用tabwriter.NewWriter在控制台打印出容器信息
	// tabwriter是引用的text/tabwriter类库，用于在控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	// 控制台输出的新系列
	fmt.Fprint(w, "ID\tName\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	if err = w.Flush(); err != nil {
		fmt.Println(fmt.Sprintf("Flush error %v", err))
	}
}

func getContainerInfo(file fs.FileInfo) (*container.ContainerInfo, error) {
	// 获取文件名
	containerName := file.Name()
	// 根据文件名生成绝对路径
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = path.Join(configFileDir, container.ConfigName)
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		fmt.Println(fmt.Sprintf("Read file %s error %v", configFileDir, err))
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err = json.Unmarshal(content, &containerInfo); err != nil {
		fmt.Println(fmt.Sprintf("Json unmarshal error %v", err))
		return nil, err
	}
	return &containerInfo, err
}
