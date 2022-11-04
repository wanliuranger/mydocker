package cgroups

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}
