package collectors

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"../simpleapi"

	"encoding/json"

	"golang.org/x/crypto/ssh"
)

const (
	HP        = "HP"
	Dell      = "Dell"
	Unknown   = "Unknown"
	RFPower   = "power"
	RFThermal = "thermal"
)

var (
	// ErrIsNotActive is returned when a chassi is in standby mode
	ErrIsNotActive                  = errors.New("This is a standby chassi")
	ErrChassiCollectionNotSupported = errors.New("It's not possible to collect metric via chassi on this model")
	redfish                         = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
		},
		HP: map[string]string{
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
	}
	bmcAddressBuild = regexp.MustCompile(".(prod|corp|dqs).")
)

type Collector struct {
	username    string
	password    string
	telegrafURL string
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

func (c *Collector) CollectViaChassi(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	if strings.HasPrefix(chassis.Model, "BladeSystem") {
		return
		fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via web %s", chassis.Fqdn, *ip, *iname))
		result, err := c.viaILOXML(ip)
		if err != nil {
			return err
		}
		iloXML := &Rimp{}
		err = xml.Unmarshal(result, iloXML)
		if err != nil {
			return err
		}
		for _, blade := range iloXML.Infra2.Blades.Blade {
			if blade.Name != "" {
				now := int32(time.Now().Unix())
				fmt.Printf("power_kw,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassis=%s,subdevice=%s value=%.2f %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, blade.Bay.Connection, chassis.Fqdn, blade.Name, blade.Power.PowerConsumed/1000.00, now)
				fmt.Printf("temp_c,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassis=%s,subdevice=%s value=%s %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, blade.Bay.Connection, chassis.Fqdn, blade.Name, blade.Temps.Temp.C, now)
			}
		}
	} else if strings.HasPrefix(chassis.Model, "P") {
		fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via RedFish %s", chassis.Fqdn, *ip, *iname))
		for _, blade := range chassis.Blades {
			for hostname, properties := range blade {
				if strings.HasSuffix(hostname, ".com") {
					// Fix tomorrow the spare-
					bmcAddress := bmcAddressBuild.ReplaceAllString(hostname, ".lom.")
					result, err := c.viaRedFish(&bmcAddress, Dell, RFPower)
					if err != nil {
						fmt.Println(err)
						break
					}
					rp := &DellRedFishPower{}
					err = json.Unmarshal(result, rp)
					if err != nil {
						fmt.Println(err)
						break
					}

					for _, item := range rp.PowerControl {
						if strings.Compare(item.Name, "System Power Control") == 0 {
							fmt.Printf("power_kw,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%d,device=chassis,chassis=%s,subdevice=%s value=%.2f %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, properties.BladePosition, chassis.Fqdn, hostname, item.PowerConsumedWatts/1000.00, int32(time.Now().Unix()))
						}
					}

					result, err = c.viaRedFish(&bmcAddress, Dell, RFThermal)
					if err != nil {
						fmt.Println(err)
						break
					}

					rt := &DellRedFishThermal{}
					err = json.Unmarshal(result, rt)
					if err != nil {
						fmt.Println(err)
						break
					}

					for _, item := range rt.Temperatures {
						if strings.Compare(item.Name, "System Board Inlet Temp") == 0 {
							fmt.Printf("temp_c,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%d,device=chassis,chassis=%s,subdevice=%s value=%d %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, properties.BladePosition, chassis.Fqdn, hostname, item.ReadingCelsius, int32(time.Now().Unix()))
						}
					}
				}
			}
		}
	} // else {
	// 	fmt.Println(fmt.Sprintf("Trying to collect data from %s[%s] via console %s", chassis.Fqdn, *ip, *iname))
	// 	// result, err := collector.ViaConsole(ip)
	// 	// if err == nil {
	// 	// 	parseHPPower(result.PowerUsage)
	// 	// 	continue
	// 	// } else if err == collectors.ErrIsNotActive {
	// 	// 	continue
	// 	// } else {
	// 	// 	fmt.Println(err)
	// 	// }
	// }
	return err
}

func (c *Collector) viaILOXML(ip *string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/xmldata?item=infra2", *ip), nil)
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

func (c *Collector) viaRedFish(ip *string, collectType string, vendor string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", *ip, redfish[collectType][vendor]), nil)
	if err != nil {
		fmt.Println("error building request:", err)
		return payload, err
	}

	req.SetBasicAuth(c.username, c.password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error readfish:", err)
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

func (c *Collector) pushToTelegraph(metric string) (err error) {
	_, err = http.NewRequest("POST", c.telegrafURL, strings.NewReader(metric))
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func New(username string, password string) *Collector {
	return &Collector{username: username, password: password}
}
