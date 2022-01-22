package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"mdocker/cgroups/subsystems"
	"mdocker/container"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit mydocker run -ti [command]",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "ti", Usage: "enable tty"},
		&cli.StringFlag{Name: "m", Usage: "memory limit"},
		&cli.StringFlag{Name: "cpushare", Usage: "cpushare limit"},
		&cli.StringFlag{Name: "cpuset", Usage: "cpuset limit"},
	},
	Action: func(context *cli.Context) error {
		if context.Args().Len() < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args().Slice() {
			cmdArray = append(cmdArray, arg)
		}
		fmt.Println("run cmdArray ", cmdArray)
		fmt.Println("run context.Args().Get(0)", context.Args().Get(0))
		//cmd := context.Args().Get(0)
		tty := context.Bool("ti")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		Run(tty, cmdArray, resConf)
		return nil
	}, // 这里是run命令执行的真正函数，1.判断参数是否包含command；2.获取用户指定的command；3.调用Run function去准备启动容器
}
var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Println("init come on")
		fmt.Println("init context.Args().Get(0)", context.Args().Get(0))
		err := container.RunContainerInitProcess()
		return err
	},
}