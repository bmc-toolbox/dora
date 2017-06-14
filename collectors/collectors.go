package collectors

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

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

func (c *Collector) httpGetDell(hostname *string) (payload []byte, err error) {
	form := url.Values{}
	form.Add("user", "Administrator")
	form.Add("password", "D4rkne55")

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", *hostname))
	if err != nil {
		log.Println("error building the url:", err)
		return payload, err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		log.Println("error building the request:", err)
		return payload, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Println("error building the cookie:", err)
		return payload, err
	}

	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
		Jar:       jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error making the login request:", err)
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/json?method=temp-sensors", *hostname))
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

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", *hostname))
	if err != nil {
		fmt.Println("error making the logout request:", err)
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return payload, err
}

func (c *Collector) dellCMC(ip *string) (payload []byte, err error) {
	return c.httpGetDell(ip)
}

func (c *Collector) viaILOXML(ip *string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/xmldata?item=infra2", *ip))
}

func (c *Collector) viaRedFish(ip *string, collectType string, vendor string) (payload []byte, err error) {
	return c.httpGet(fmt.Sprintf("https://%s/%s", *ip, redfishVendorEndPoints[collectType][vendor]))
}

func (c *Collector) pushToTelegraph(metric string) (err error) {
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
