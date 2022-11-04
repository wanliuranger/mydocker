package main

import (
	"os"
	"strings"
	"syscall"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
)

func Run(cmdArr []string, tty bool, resourceConfig *cgroups.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
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
		cgroupName := uuid.New().String()
		log.Printf("cgroupName: %v\n", cgroupName)
		cgroupMgr := cgroups.NewCgroupManager(cgroupName)
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
	container.DeleteWorkSpace("/home/ourupf/", "/home/ourupf/mnt")
}

func sendInitCommand(cmdArr []string, writePipe *os.File) {
	command := strings.Join(cmdArr, " ")
	log.Infof("user command is: %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
