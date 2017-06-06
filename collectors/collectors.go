package collectors

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"../simpleapi"

	"encoding/json"
)

const (
	HP        = "HP"
	Dell      = "Dell"
	Unknown   = "Unknown"
	RFPower   = "power"
	RFThermal = "thermal"
)

var (
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

func (c *Collector) httpGet(url string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error building request:", err)
		return payload, err
	}
	req.SetBasicAuth(c.username, c.password)
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error making the request:", err)
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

func (c *Collector) CollectViaChassi(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	if strings.HasPrefix(chassis.Model, "BladeSystem") {
		fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via ILOXML %s", chassis.Fqdn, *ip, *iname))
		result, err := c.viaILOXML(ip)
		if err != nil {
			return err
		}
		iloXML := &Rimp{}
		err = xml.Unmarshal(result, iloXML)
		if err != nil {
			return err
		}
		if iloXML.Infra2 != nil && iloXML.Infra2.Blades != nil {
			for _, blade := range iloXML.Infra2.Blades.Blade {
				if blade.Name != "" {
					now := int32(time.Now().Unix())
					fmt.Printf("power_kw,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassis=%s,subdevice=%s value=%.2f %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, blade.Bay.Connection, chassis.Fqdn, blade.Name, blade.Power.PowerConsumed/1000.00, now)
					fmt.Printf("temp_c,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassis=%s,subdevice=%s value=%s %d\n", rack.Site, rack.Sitezone, rack.Sitepod, rack.Siterow, chassis.Rack, blade.Bay.Connection, chassis.Fqdn, blade.Name, blade.Temps.Temp.C, now)
				}
			}
		}
	} else if strings.HasPrefix(chassis.Model, "P") {
		fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via RedFish %s", chassis.Fqdn, *ip, *iname))
		for _, blade := range chassis.Blades {
			for hostname, properties := range blade {
				if strings.HasSuffix(hostname, ".com") && !strings.HasPrefix(hostname, "spare") {
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
	} else {
		fmt.Printf("I dunno what to do with this device %s, skipping...\n", chassis.Fqdn)
	}
	return err
}

func (c *Collector) viaILOXML(ip *string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/xmldata?item=infra2", *ip))
}

func (c *Collector) viaRedFish(ip *string, collectType string, vendor string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/%s", *ip, redfish[collectType][vendor]))
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
