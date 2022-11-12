package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"mukdenranger.com/mydocker/container"
	_ "mukdenranger.com/mydocker/nsenter"
)

const (
	ENV_EXEC_PID string = "my_pid"
	ENV_EXEC_CMD string = "my_cmd"
)

func execCommandInContainer(containerId string, cmds []string) error {
	containerConfigPath := path.Join(container.DefaultLocationBase, containerId)
	if _, err := os.Stat(containerConfigPath); err != nil {
		return fmt.Errorf("container %s does not exist, error: %v", containerId, err)
	}

	conf := &container.ContainerInfo{}
	configPath := path.Join(containerConfigPath, container.ConfigName)
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("open container %s config file error: %v", containerId, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.Decode(conf)

	cmdStr := strings.Join(cmds, " ")
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, conf.Pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	containerEnv := getEnvSlices(conf.Pid)
	cmd.Env = append(os.Environ(), containerEnv...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec command error: %v", err)
	}

	return nil
}

func getEnvSlices(pid string) []string {
	envdir := path.Join("/proc", pid, "environ")
	contentBytes, err := ioutil.ReadFile(envdir)
	if err != nil {
		log.Errorf("get envslices error: %v", err)
		return nil
	}
	env := strings.Split(string(contentBytes), "\u0000")
	return env
}
