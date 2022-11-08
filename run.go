package main

import (
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
)

func Run(cmdArr []string, containerName string, imageName string, volume string, tty bool, resourceConfig *cgroups.ResourceConfig) {
	if containerName == "" {
		containerName = randName(10)
	}
	if imageName == "" {
		imageName = "busybox"
	}
	parent, writePipe := container.NewParentProcess(tty, containerName, imageName, volume)
	if parent == nil {
		log.Errorf("New parent proces error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	defer func() {
		syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NODEV), "")
	}()
	if resourceConfig != nil {
		log.Printf("cgroupName: %v\n", containerName)
		cgroupMgr := cgroups.NewCgroupManager(containerName)
		err := cgroupMgr.CreateCgroup()
		if err != nil {
			log.Printf("create cgroup error: %v", err)
		}
		defer cgroupMgr.RemoveCgroup()
		err = cgroupMgr.ConfigResource(resourceConfig)
		if err != nil {
			log.Printf("config resource error: %v", err)
		}
		err = cgroupMgr.AddProcess(parent.Process.Pid)
		if err != nil {
			log.Printf("add process error: %v", err)
		}
	}

	sendInitCommand(cmdArr, writePipe)
	parent.Wait()
	//os.Chdir("/home/ourupf")
	container.DeleteWorkSpace(containerName, volume)
}

func sendInitCommand(cmdArr []string, writePipe *os.File) {
	command := strings.Join(cmdArr, " ")
	log.Infof("user command is: %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func randName(nameLen int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234566780")
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, nameLen)
	for i := 0; i < nameLen; i++ {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
