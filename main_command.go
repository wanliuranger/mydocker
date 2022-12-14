package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
	_ "mukdenranger.com/mydocker/network"

	_ "github.com/vishvananda/netlink"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `reate a container with namespace and cgroups limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "memory",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "volume",
			Usage: "volume to be mounted",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "name of the container about to run",
		},
		cli.StringFlag{
			Name:  "image",
			Usage: "the image that used to build the container",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "the program will run in detach mode if set",
		},
		cli.BoolFlag{
			Name:  "log",
			Usage: "save log to pointed file",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmds []string
		for _, arg := range context.Args() {
			cmds = append(cmds, arg)
		}
		tty := context.Bool("ti")
		detach := context.Bool("d")
		resourceConfig := &cgroups.ResourceConfig{
			MemoryLimit: context.String("memory"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		useLog := context.Bool("log")
		containerName := context.String("name")
		imageName := context.String("image")
		volumes := context.String("volume")
		envSlices := context.StringSlice("e")
		Run(cmds, containerName, imageName, volumes, tty, detach, useLog, envSlices, resourceConfig)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: `Init container process run user's process in container. Don't call it outside`,
	Action: func(context *cli.Context) error {
		log.Infof("init")
		cmd := context.Args().Get(0)
		log.Infof("command: %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: `commit current container to a tar file`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "The id of the contaner you want to commit",
		},
	},
	Action: func(context *cli.Context) error {
		containerId := context.String("id")
		if containerId == "" {
			return fmt.Errorf("missing container id")
		}
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(containerId, imageName)
		return nil
	},
}

var psCommand = cli.Command{
	Name:  "ps",
	Usage: `show container info`,
	Action: func(context *cli.Context) error {
		showcurrentContainer()
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container with id",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("stop command needs a container id")
		}
		return stopContainer(context.Args().Get(0))
	},
}

var showLogCommand = cli.Command{
	Name:  "log",
	Usage: "show log of certain container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("log command requeres a container id")
		}
		return showLog(context.Args().Get(0))
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "execute command in a certain container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "the id of the container which you want to execute command in",
		},
	},
	Action: func(context *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			return nil
		}
		containerId := context.String("id")
		fmt.Printf("containerId: %v\n", containerId)
		if len(containerId) == 0 {
			fmt.Printf("no containerId %s", containerId)
			return fmt.Errorf("a container Id is required")
		}
		var cmds []string
		for _, cmd := range context.Args() {
			cmds = append(cmds, cmd)
		}

		return execCommandInContainer(containerId, cmds)
	},
}
