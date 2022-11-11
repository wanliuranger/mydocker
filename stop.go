package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"syscall"

	"mukdenranger.com/mydocker/cgroups"
	"mukdenranger.com/mydocker/container"
)

func stopContainer(containerId string) error {
	containerPath := path.Join(container.DefaultLocationBase, containerId, container.ConfigName)
	_, err := os.Stat(containerPath)
	if err != nil {
		return fmt.Errorf("no such container")
	}
	file, err := os.Open(containerPath)
	if err != nil {
		return fmt.Errorf("can't open container %s config, you need to clean manually", containerId)
	}
	decoder := json.NewDecoder(file)
	conf := &container.ContainerInfo{}
	decoder.Decode(conf)
	//kill掉工作进程
	pid, _ := strconv.Atoi(conf.Pid)
	syscall.Kill(pid, syscall.SIGKILL)
	//卸掉mount点
	container.DeleteWorkSpace(conf.Name, conf.Volume)
	os.RemoveAll(path.Join(container.DefaultLocationBase, conf.Id))
	//清理cgroup
	cgroupMgr := cgroups.NewCgroupManager(conf.Id)
	cgroupMgr.RemoveCgroup()
	return nil
}
