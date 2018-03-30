package scanner

import (
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// Port statuses
const (
	closed = iota - 1
	open
)

// ErrUnsupportedProtocol returned when requested to scan an unsupported protocol
var ErrUnsupportedProtocol = errors.New("Unsupported protocol")

// Result is the result of the probe; the appropriate "IsXXXX()" function
// should be used to evaluate it.
type Result int

// String returns the probe result as a string.
func (r Result) String() string {
	switch r {
	case closed:
		return "closed"
	case open:
		return "open"
	default:
		return "unsupported"
	}
}

// probeTCP determines whether the indicated TCP port on the target host is
// open.
func probeTCP(node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := net.DialTimeout("tcp4", address, 1*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"dial": "tcp", "address": address}).Debug(err)
		return closed
	}
	defer conn.Close()
	return open
}

// probeTCP determines whether the indicated IPMI port on the target host is
// open.
func probeIPMI(node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := net.DialTimeout("udp4", address, 1*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"dial": "udp", "address": address}).Debug(err)
		return closed
	}
	defer conn.Close()

	// This payload is the rmcp ping from rmcp rfc
	payload := []byte("\x06\x00\xff\x06\x00\x00\x11\xbe\x80\x18\x00\x00")
	_, err = conn.Write(payload)
	if err != nil {
		log.WithFields(log.Fields{"write": "udp", "address": address}).Debug(err)
		return closed
	}

	err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		log.WithFields(log.Fields{"set read timeout": "udp", "address": address}).Debug(err)
		return closed
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		log.WithFields(log.Fields{"read": "udp", "address": address}).Debug(err)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return closed
		}
		return open
	}

	if n > 0 {
		return open
	}

	return closed
}

// Probe determines whether the specified port on the on the specified host is
// potentially accepting input via the specified network protocol.
func Probe(protocol, host string, port int) (r Result, err error) {
	switch protocol {
	case "tcp":
		return probeTCP(host, port), err
	case "ipmi":
		return probeIPMI(host, port), err
	default:
		return r, ErrUnsupportedProtocol
	}
}
