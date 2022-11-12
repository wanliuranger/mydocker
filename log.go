package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"mukdenranger.com/mydocker/container"
)

func showLog(containerId string) error {
	containerUrl := path.Join(container.DefaultLocationBase, containerId)
	if _, err := os.Stat(containerUrl); err != nil {
		return fmt.Errorf("no such container")
	}
	logUrl := path.Join(containerUrl, container.LogName)
	file, err := os.Open(logUrl)
	if err != nil {
		fmt.Printf("container %s does not have a log file", containerId)
		return nil
	}
	defer file.Close()
	logContext, _ := ioutil.ReadAll(file)
	fmt.Fprint(os.Stdout, string(logContext))
	return nil
}
