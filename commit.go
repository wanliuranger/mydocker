package main

import (
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func commitContainer(containerId string, imageName string) {
	containerWorkBase := "/home/ourupf/mnt"
	containerWorkDir := path.Join(containerWorkBase, containerId)
	cur, _ := os.Getwd()
	imageTar := path.Join(cur, imageName+".tar")
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", containerWorkDir, ".").CombinedOutput(); err != nil {
		log.Errorf("tar error: %v", err)
	}
}
