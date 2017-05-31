package collectors

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	HP      = "HP"
	Dell    = "Dell"
	Unknown = "Unknown"
)

var (
	// ErrIsNotActive is returned when a chassi is in standby mode
	ErrIsNotActive = errors.New("This is a standby chassi")
)

type Collector struct {
	username string
	password string
}

type RawCollectedData struct {
	PowerUsage  string
	Temperature string
	Vendor      string
}

func (c *Collector) runCommand(client *ssh.Client, command string) (result string, err error) {
	session, err := client.NewSession()
	if err != nil {
		return result, err
	}
	defer session.Close()

	var r bytes.Buffer
	session.Stdout = &r
	if err := session.Run(command); err != nil {
		return result, err
	}
	return r.String(), err
}

func (c *Collector) ViaILOXML(ip string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/xmldata?item=infra2", ip), nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error ilo:", err)
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading the response:", err)
		return payload, err
	}
	return payload, err
}

func (c *Collector) ViaConsole(ip string) (result RawCollectedData, err error) {
	// var hostKey ssh.PublicKey
	// An SSH client is represented with a ClientConn.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ip), config)
	if err != nil {
		return result, err
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	r, err := c.runCommand(client, "help")
	if err != nil {
		return result, err
	}

	if strings.Count(r, "getpbinfo") != 0 {
		result.Vendor = Dell
	} else if strings.Count(r, "SAVE SEND SET SHOW SLEEP TEST UNASSIGN") != 0 {
		result.Vendor = HP
	} else {
		result.Vendor = Unknown
	}

	if result.Vendor == HP {
		r, err = c.runCommand(client, "show enclosure power_summary")
		if err != nil {
			return result, err
		}
		if strings.Count(r, "standby mode.") != 0 {
			return result, ErrIsNotActive
		}
		result.PowerUsage = r

		r, err = c.runCommand(client, "show enclosure temp")
		if err != nil {
			return result, err
		}
		result.Temperature = r
	} else if result.Vendor == Dell {
		r, err = c.runCommand(client, "show enclosure power_summary")
		if err != nil {
			return result, err
		}
		result.PowerUsage = r
	}

	return result, err
}

func New(username string, password string) *Collector {
	return &Collector{username: username, password: password}
}
