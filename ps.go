package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"mukdenranger.com/mydocker/container"
)

func showcurrentContainer() {
	containerIds, err := os.ReadDir(container.DefaultLocationBase)
	if err != nil {
		log.Errorf("read container list failed, err: %v", err)
		return
	}

	var containerInfoList []container.ContainerInfo

	for _, c := range containerIds {
		if c.IsDir() {
			configFilePath := path.Join(container.DefaultLocationBase, c.Name(), container.ConfigName)
			file, err := os.Open(configFilePath)
			if err != nil {
				log.Errorf("open configuration file of container %s failed, err: %v", c.Name(), err)
			}
			defer file.Close()
			decoder := json.NewDecoder(file)
			var containerInfo = &container.ContainerInfo{}
			decoder.Decode(containerInfo)
			containerInfoList = append(containerInfoList, *containerInfo)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tName\tPID\tStatus\tCOMMAND\tCREATED\n")
	for _, it := range containerInfoList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s",
			it.Id,
			it.Name,
			it.Pid,
			it.Status,
			it.Command,
			it.CreateTime)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error: %v", err)
	}

}
