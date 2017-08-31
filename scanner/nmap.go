package scanner

// Based on https://raw.githubusercontent.com/lair-framework/go-nmap/master/nmap.go

import (
	"encoding/xml"
	"errors"
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
	PortID   int    `xml:"portid,attr"`
	State    State  `xml:"state"`
}

// State contains information about a given ports
// status. State will be open, closed, etc.
type State struct {
	State string `xml:"state,attr"`
}

func nmapParse(content []byte) (*NmapRun, error) {
	r := &NmapRun{}
	err := xml.Unmarshal(content, r)
	if err != nil {
		return r, err
	}
	return r, nil
}
