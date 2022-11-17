package network

import (
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"
	"sync"
)

//这里的实现不对，数太大了，不应该写成文件的形式保存，应该常驻在内存里。但还是先这么写，先实现功能

const (
	defaultIPAllocatorDir  string = "/var/run/mydocker/network/ipam/"
	defaultIPAllocatorName string = "subnet.json"
)

type IPAM struct {
	mpLock     sync.Mutex
	AllocateMp map[string]string `json:"subnet"`
}

func (ipam *IPAM) load() {
	f, err := os.Open(path.Join(defaultIPAllocatorDir, defaultIPAllocatorName))
	if err != nil {
		return
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	decoder.Decode(ipam)
}

func (ipam *IPAM) dump() {
	if _, err := os.Stat(defaultIPAllocatorDir); err != nil {
		if err := os.MkdirAll(defaultIPAllocatorDir, 0644); err != nil {
			return
		}
	}
	f, err := os.Create(path.Join(defaultIPAllocatorDir, defaultIPAllocatorName))
	if err != nil {
		return
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.Encode(ipam)
}

func (ipam *IPAM) addSubnet(cidr *net.IPNet) {
	ones, bits := cidr.Mask.Size()
	ipam.AllocateMp[cidr.String()] = strings.Repeat("0", (1<<(bits-ones))-2)
}

func (ipam *IPAM) Allocate(cidr *net.IPNet) (*net.IP, error) {
	ipam.mpLock.Lock()
	defer ipam.mpLock.Unlock()
	ipam.load()
	cidrStr := cidr.String()
	if _, ok := ipam.AllocateMp[cidrStr]; !ok {
		ipam.addSubnet(cidr)
	}
	cnt := 0
	for ipam.AllocateMp[cidrStr][cnt] == '1' && cnt < len(ipam.AllocateMp[cidrStr]) {
		cnt++
	}
	if cnt == len(ipam.AllocateMp[cidrStr]) {
		return nil, nil
	}
	cnt = cnt + 1
	allocatedBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		num := byte(cnt % 128)
		cnt = cnt >> 8
		allocatedBytes = append(allocatedBytes, num)
	}
	ans := net.IP(make([]byte, 4))
	for i := 0; i < 4; i++ {
		ans[i] = cidr.IP[i] | allocatedBytes[i]
	}

	mpByte := []byte(ipam.AllocateMp[cidrStr])
	mpByte[cnt] = '1'
	ipam.AllocateMp[cidrStr] = string(mpByte)
	ipam.dump()
	return &ans, nil
}
