package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const usage = "hello, world"

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage
	app.Commands = []*cli.Command{
		&initCommand,
		&runCommand,
		&commitCommand,
		&listCommand,
		&logCommand,
		&execCommand,
		&stopCommand,
		&removeCommand,
	}
	app.Before = func(context *cli.Context) error {
		// log init
		return nil
	}
	fmt.Println("os.Args: ", os.Args)
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
