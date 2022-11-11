package cgroups

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"
)

const (
	cgroupBase = "/sys/fs/cgroup"
)

type CgroupManager struct {
	containerName string
	cgroupPath    string
}

func NewCgroupManager(name string) *CgroupManager {
	mgr := &CgroupManager{
		containerName: name,
		cgroupPath:    path.Join(cgroupBase, name),
	}
	return mgr
}

func (mgr *CgroupManager) CreateCgroup() error {
	return os.Mkdir(mgr.cgroupPath, 0755)
}

func (mgr *CgroupManager) RemoveCgroup() error {
	err := os.RemoveAll(mgr.cgroupPath)
	if err != nil {
		logrus.Errorf("remove cgroup error: %v\n", err)
	}
	return err
}

func (mgr *CgroupManager) AddProcess(pid int) error {
	logrus.Printf("cgroupPath: %v", mgr.cgroupPath)
	return ioutil.WriteFile(path.Join(mgr.cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644)
}

func (mgr *CgroupManager) ConfigResource(conf *ResourceConfig) error {
	var err error
	if conf.MemoryLimit != "" {
		err = ioutil.WriteFile(path.Join(mgr.cgroupPath, "memory.max"), []byte(conf.MemoryLimit), 0644)
		if err != nil {
			return err
		}
	}
	if conf.CpuShare != "" {
		err = ioutil.WriteFile(path.Join(mgr.cgroupPath, "cpu.max"), []byte(conf.CpuShare), 0644)
		if err != nil {
			return err
		}
	}
	if conf.CpuSet != "" {
		err = ioutil.WriteFile(path.Join(mgr.cgroupPath, "cpuset.cpus"), []byte(conf.CpuSet), 0644)
	}
	return err
}
