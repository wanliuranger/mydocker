package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultNetworkPath string = "/var/run/mydocker/network/network"
)

var (
	drivers  = map[string]NetworkDriver{}
	networks = map[string]Network{}
)

type Network struct {
	Name    string     `json:"name"`
	IpRange *net.IPNet `json:"ipRange"`
	Driver  string     `json:"driver"`
}

func (nw *Network) load(basePath string) error {
	if _, err := os.Stat(basePath); err != nil {
		if err := os.MkdirAll(basePath, 0644); err != nil {
			return err
		}
	}
	nwPath := path.Join(basePath, nw.Name)
	f, err := os.Open(nwPath)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(f)
	decoder.Decode(nw)
	return nil
}

func (nw *Network) dump(basePath string) error {
	if _, err := os.Stat(basePath); err != nil {
		if err := os.MkdirAll(basePath, 0644); err != nil {
			return err
		}
	}

	nwPath := path.Join(basePath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer nwFile.Close()
	encoder := json.NewEncoder(nwFile)
	encoder.Encode(nw)
	return nil
}

func Init() error {
	var bridgeDriver NetworkDriver = &BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = bridgeDriver
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if err := os.MkdirAll(defaultNetworkPath, 0644); err != nil {
			return err
		}
	}
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		newNetwork := &Network{
			Name: nwName,
		}
		if err := newNetwork.load(nwPath); err != nil {
			return err
		}
		networks[newNetwork.Name] = *newNetwork
		return nil
	})
	return nil
}

func CreateNetwork(driver string, subnet string, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = *ip
	dri, ok := drivers[driver]
	if !ok {
		return fmt.Errorf("no such driver")
	}
	nw, err := dri.Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}
