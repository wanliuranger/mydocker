package network

import (
	"net"
	"sync"
	"testing"
)

var ipAllocatorTest = &IPAM{
	mpLock:     sync.Mutex{},
	AllocateMp: make(map[string]string),
}

func TestIPAllocator(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("192.168.11.1/24")

	ip1, _ := ipAllocatorTest.Allocate(cidr)
	t.Logf("%s\n", ip1.String())
	if err := ipAllocatorTest.Release(cidr, ip1); err != nil {
		t.Errorf("release error")
	}
}
