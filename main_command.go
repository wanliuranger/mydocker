package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
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
		resourceConfig := &cgroups.ResourceConfig{
			MemoryLimit: context.String("memory"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		// volumes:=context.String("volume")
		Run(cmds, tty, resourceConfig)
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