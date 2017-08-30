package scanner

// Based on https://raw.githubusercontent.com/lair-framework/go-nmap/master/nmap.go

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

var (
	ErrInvalidProtocol = errors.New("Invalid protocol")
)

// NmapRun is contains all the data for a single nmap scan.
type NmapRun struct {
	Hosts []Host `xml:"host"`
}

// Host contains all information about a single host.
type Host struct {
	Addresses []Address `xml:"address"`
	Ports     []Port    `xml:"ports>port"`
	Status    Status    `xml:"status"`
}

// Status is the host's status. Up, down, etc.
type Status struct {
	State string `xml:"state,attr"`
}

// Address contains a IPv4 or IPv6 address for a Host.
type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

// Port contains all the information about a scanned port.
type Port struct {
	Protocol string `xml:"protocol,attr"`
	PortId   int    `xml:"portid,attr"`
	State    State  `xml:"state"`
}

// State contains information about a given ports
// status. State will be open, closed, etc.
type State struct {
	State string `xml:"state,attr"`
}

// Parse takes a byte array of nmap xml data and unmarshals it into an
// NmapRun struct. All elements are returned as strings, it is up to the caller
// to check and cast them to the proper type.
func Parse(content []byte) (*NmapRun, error) {
	r := &NmapRun{}
	err := xml.Unmarshal(content, r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func scan(subnet string, ports string, protocol string) (err error) {
	xmlDir := viper.GetString("nmap_xml_dir")
	err = os.MkdirAll(xmlDir, 0755)
	if err != nil {
		return err
	}

	scanType := ""
	fileName := fmt.Sprintf("%s/%s-%s.xml", xmlDir, strings.Replace(subnet, "/", "_", -1), protocol)
	switch protocol {
	case "udp":
		scanType = "-sU"
	case "tcp":
		scanType = "-sT"
	default:
		return ErrInvalidProtocol
	}

	cmd := exec.Command("nmap", "-oX", fileName, scanType, subnet, "-p", ports, "--open")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return err
}
