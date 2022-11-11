package container

type ContainerInfo struct {
	Pid        string `json:"pid"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Command    string `json:"command"`
	CreateTime string `json:"createTime"`
	Volume     string `json:"volume"`
	Status     string `json:"status"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	EXIT                string = "exited"
	DefaultLocationBase string = "/var/run/mydocker/"
	ConfigName          string = "config.json"
	LogName             string = "container.log"
)
