package network

import "net"

type Network struct {
	Name    string     `json:"name"`
	IpRange *net.IPNet `json:"ipRange"`
	Driver  string     `json:"driver"`
}
