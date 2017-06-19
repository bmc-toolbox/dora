package collectors

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"../simpleapi"

	log "github.com/sirupsen/logrus"
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
	bladeDevice                     = "blade"
	chassisDevice                   = "chassis"
	discreteDevice                  = "discrete"
	storageBladeDevice              = "storageblade"
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
	log.WithFields(log.Fields{"step": "collectoers", "url": url}).Debug("Requesting data from BMC")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
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
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

func (c *Collector) httpGetDell(hostname *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "collectoers", "hostname": *hostname}).Debug("Requesting data from BMC")

	form := url.Values{}
	form.Add("user", "Administrator")
	form.Add("password", "D4rkne55")

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", *hostname))
	if err != nil {
		return payload, err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
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
		return payload, err
	}

	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
		Jar:       jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/json?method=temp-sensors", *hostname))
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", *hostname))
	if err != nil {
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return bytes.Replace(payload, []byte("\"bladeTemperature\":-1"), []byte("\"bladeTemperature\":\"0\""), -1), err
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

func (c *Collector) createAndSendMessage(metric *string, site *string, zone *string, pod *string, row *string, rack *string, chassis *string, role *string, device *string, fqdn *string, value string, now int32) {
	err := c.pushToTelegraph(fmt.Sprintf("%s,site=%s,zone=%s,pod=%s,row=%s,rack=%s,chassis=%s,role=%s,device_type=%s,device_name=%s value=%s %d\n", *metric, *site, *zone, *pod, *row, *rack, *chassis, *role, *device, *fqdn, value, now))
	if err != nil {
		log.WithFields(log.Fields{"fqdn": *fqdn, "type": "blade", "metric": *metric}).Info("Unable to push data to telegraf")
	}
}

func (c *Collector) pushToTelegraph(metric string) (err error) {
	log.WithFields(log.Fields{"step": "collectoers", "metric": metric}).Debug("Pushing data to telegraf")

	return
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
