package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (bd *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (bd *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  bd.Name(),
	}
	err := bd.initBridge(n)
	if err != nil {
		log.Errorf("error init bridge: %v", err)
	}

	return n, err
}

func (bd *BridgeNetworkDriver) Delete(network *Network) error {
	err := bd.deleteBridge(network)
	return err
}

func (bd *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.Id[:5]
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.Id[:5],
	}
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error adding endpoint device: %v", err)
	}
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error bringing up endpoint device: %v", err)
	}
	return nil
}

//to be implemented
func (bd *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	return nil
}

func (bd *BridgeNetworkDriver) initBridge(n *Network) error {
	bridgeName := n.Name
	if err := bd.createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error add bridge: %s, error: %v", bridgeName, err)
	}

	if err := bd.setInterfaceIP(bridgeName, n.IpRange.String()); err != nil {
		return fmt.Errorf("error assigning address %s on bridge: %s", n.IpRange.String(), bridgeName)
	}

	if err := bd.setInterfaceUp(bridgeName); err != nil {
		return fmt.Errorf("error bringing interface: %s up", bridgeName)
	}

	if err := bd.setIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
	}

	return nil
}

func (bd *BridgeNetworkDriver) deleteBridge(n *Network) error {
	bridgeName := n.Name
	link, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("getting link with name %s failed, error: %v", bridgeName, err)
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to remove bridge interface %s, error: %v", bridgeName, err)
	}
	return nil
}

func (bd *BridgeNetworkDriver) createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{
		LinkAttrs: la,
	}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s, err: %v", bridgeName, err)
	}
	return nil
}

func (bd *BridgeNetworkDriver) setInterfaceIP(bridgeName string, rawIP string) error {
	ifname, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	ipnet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{
		IPNet: ipnet,
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}
	return netlink.AddrAdd(ifname, addr)
}

func (bd *BridgeNetworkDriver) setInterfaceUp(bridgeName string) error {
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ], error: %v", bridgeName, err)
	}
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error bringing link %s up, error: %v", bridgeName, err)
	}
	return nil
}

func (bd *BridgeNetworkDriver) setIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables output: %v", output)
	}
	return err
}
