package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
)

func Run(cmdArr []string, containerName string, imageName string, volume string, tty bool, detach bool, useLog bool, resourceConfig *cgroups.ResourceConfig) {
	containerId := randName(10)
	if containerName == "" {
		containerName = containerId
	}
	if imageName == "" {
		imageName = "busybox"
	}
	parent, writePipe := container.NewParentProcess(tty, containerId, containerName, imageName, volume, useLog)
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
		cgroupMgr := cgroups.NewCgroupManager(containerId)
		err := cgroupMgr.CreateCgroup()
		if err != nil {
			log.Printf("create cgroup error: %v", err)
		}
		if !detach {
			defer cgroupMgr.RemoveCgroup()
		}
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
	recordContainer(parent.Process.Pid, containerId, containerName, cmdArr, volume, time.Now(), container.RUNNING)

	if !detach {
		parent.Wait()
		//os.Chdir("/home/ourupf")
		container.DeleteWorkSpace(containerName, volume)
		os.RemoveAll(path.Join(container.DefaultLocationBase, containerId))
	}
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

func recordContainer(pid int, id string, name string, cmdArr []string, volume string, createTime time.Time, status string) {
	pidStr := fmt.Sprintf("%v", pid)
	cmdAll := strings.Join(cmdArr, " ")
	createTimeStr := fmt.Sprintf("%v", createTime)
	containerInfo := &container.ContainerInfo{
		Pid:        pidStr,
		Id:         id,
		Name:       name,
		Command:    cmdAll,
		CreateTime: createTimeStr,
		Volume:     volume,
		Status:     status,
	}

	configUrl := path.Join(container.DefaultLocationBase, id)
	if err := os.MkdirAll(configUrl, 0644); err != nil {
		log.Errorf("create config file base error: %v", err)
		return
	}
	filePath := path.Join(configUrl, container.ConfigName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Errorf("create config file failed, error: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(containerInfo)
	if err != nil {
		log.Errorf("write config file failed, error: %v", err)
	}

}
