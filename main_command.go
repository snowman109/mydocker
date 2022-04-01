package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"mdocker/cgroups/subsystems"
	"mdocker/container"
	"os"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit mydocker run -ti [command]",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "ti", Usage: "enable tty"},
		&cli.StringFlag{Name: "m", Usage: "memory limit"},
		&cli.StringFlag{Name: "cpushare", Usage: "cpushare limit"},
		&cli.StringFlag{Name: "cpuset", Usage: "cpuset limit"},
		&cli.StringFlag{Name: "v", Usage: "volume"},
		&cli.BoolFlag{Name: "d", Usage: "detach container"},
		&cli.StringFlag{Name: "name", Usage: "container name"},
		&cli.StringSliceFlag{Name: "e", Usage: "set environment"},
	},
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf("Missing container command")
		}
		cmdArray := make([]string, 0, context.Args().Len())
		for _, arg := range context.Args().Slice() {
			cmdArray = append(cmdArray, arg)
		}
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		//fmt.Println("run cmdArray ", cmdArray)
		//fmt.Println("run context.Args().Get(0)", context.Args().Get(0))
		//cmd := context.Args().Get(0)
		tty := context.Bool("ti") // 前台运行
		detach := context.Bool("d")
		// tty 和 detach不能共存
		if tty && detach {
			return fmt.Errorf("ti and d paramter can not both provided")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		containerName := context.String("name")
		// 把volume参数传给Run函数
		volume := context.String("v")
		envSlice := context.StringSlice("e")
		Run(tty, cmdArray, resConf, volume, containerName, imageName,envSlice)
		return nil
	}, // 这里是run命令执行的真正函数，1.判断参数是否包含command；2.获取用户指定的command；3.调用Run function去准备启动容器
}
var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		//log.Println("init come on")
		//fmt.Println("init context.Args().Get(0)", context.Args().Get(0))
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 2 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		commitContainer(containerName, imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "log",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf("Please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			fmt.Printf("pid callback pid %d\n", os.Getpid())
			return nil
		}
		if context.Args().Len() < 2 {
			return fmt.Errorf("Missing container name or command")
		}
		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		ExecContainer(containerName, commandArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		removeContainer(containerName)
		return nil
	},
}
