package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Endpoint struct {
	Id          string           `json:"id"`
	Device      netlink.Veth     `json:"veth"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portmapping"`
	Network     *Network
}
