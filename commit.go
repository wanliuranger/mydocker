package main

import (
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func commitContainer(imageName string) {
	containerWorkDir := "/home/ourupf/mnt"
	cur, _ := os.Getwd()
	imageTar := path.Join(cur, imageName+".tar")
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", containerWorkDir, ".").CombinedOutput(); err != nil {
		log.Errorf("tar error: %v", err)
	}
}
