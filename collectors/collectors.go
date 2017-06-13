package collectors

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"../simpleapi"
)

const (
	HP         = "HP"
	Dell       = "Dell"
	Supermicro = "Supermicro"
	Unknown    = "Unknown"
	RFPower    = "power"
	RFThermal  = "thermal"
)

var (
	powerMetric                     = "power_kw"
	thermalMetric                   = "temp_c"
	ErrChassiCollectionNotSupported = errors.New("It's not possible to collect metric via chassi on this model")
	redfishVendorEndPoints          = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
		},
		HP: map[string]string{
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
		Supermicro: map[string]string{
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
	}
	redfishVendorLabels = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Board Inlet Temp",
		},
		HP: map[string]string{
			//			RFPower:   "PowerMetrics",
			RFThermal: "30-System Board",
		},
		Supermicro: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Temp",
		},
	}
	bmcAddressBuild = regexp.MustCompile(".(prod|corp|dqs).")
)

type Collector struct {
	username    string
	password    string
	telegrafURL string
	simpleAPI   *simpleapi.SimpleAPI
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

func (c *Collector) viaILOXML(ip *string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/xmldata?item=infra2", *ip))
}

func (c *Collector) viaRedFish(ip *string, collectType string, vendor string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/%s", *ip, redfishVendorEndPoints[collectType][vendor]))
}

func (c *Collector) pushToTelegraph(metric string) (err error) {
	//fmt.Println(metric)
	//return err
	req, err := http.NewRequest("POST", c.telegrafURL, strings.NewReader(metric))
	if err != nil {
		return err
	}
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
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return err
}

func New(username string, password string, telegrafURL string, simpleApi *simpleapi.SimpleAPI) *Collector {
	return &Collector{username: username, password: password, telegrafURL: telegrafURL, simpleAPI: simpleApi}
}
