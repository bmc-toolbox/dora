package connectors

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

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
	bladeDevice        = "blade"
	chassisDevice      = "chassis"
	storageBladeDevice = "storageblade"
	ErrPageNotFound    = errors.New("Requested page couldn't be found in the server")
)

type Blade struct {
	Name          string  `json:"name"`
	MgmtIPAddress string  `json:"mgmt_ip_address"`
	BladePosition int     `json:"blade_position"`
	Temp          int     `json:"temp_c"`
	Serial        string  `json:"serial"`
	Power         float64 `json:"power_kw"`
	storageBlade  bool    `json:"is_storage_blade"`
	Vandor        string  `json:"vendor"`
}

type Chassis struct {
	Name   string   `json:"name"`
	Rack   string   `json:"rack"`
	Blades []*Blade `json:"blades"`
	Temp   int      `json:"temp_c"`
	Power  float64  `json:"power_kw"`
	Serial string   `json:"serial"`
	Model  string   `json:"model"`
	Vandor string   `json:"vendor"`
}

type ChassisConnection struct {
	username string
	password string
}

func httpGet(url string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "url": url}).Debug("Requesting data from BMC")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payload, err
	}
	req.SetBasicAuth(*username, *password)
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

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

func httpGetDell(hostname *string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "hostname": *hostname}).Debug("Requesting data from BMC")

	form := url.Values{}
	form.Add("user", *username)
	form.Add("password", *password)

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

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

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

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	return bytes.Replace(payload, []byte("\"bladeTemperature\":-1"), []byte("\"bladeTemperature\":\"0\""), -1), err
}

func (b *Blade) IsStorageBlade() bool {
	return b.storageBlade
}

func (c *ChassisConnection) Dell(ip *string) (payload []byte, err error) {
	return httpGetDell(ip, &c.username, &c.password)
}

func (c *ChassisConnection) Hp(ip *string) (chassis Chassis, err error) {
	result, err := httpGet(fmt.Sprintf("https://%s/xmldata?item=infra2", *ip), &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	iloXML := &HpRimp{}
	err = xml.Unmarshal(result, iloXML)
	if err != nil {
		return chassis, err
	}

	if iloXML.HpInfra2 != nil {
		chassis.Name = iloXML.HpInfra2.Encl
		chassis.Serial = iloXML.HpInfra2.EnclSn
		chassis.Model = iloXML.HpInfra2.Pn
		chassis.Rack = iloXML.HpInfra2.Rack
		chassis.Power = iloXML.HpInfra2.HpPower.PowerConsumed / 1000.00

		chassis.Temp = iloXML.HpInfra2.HpTemps.HpTemp.C

		log.WithFields(log.Fields{"operation": "connection", "ip": ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis"}).Debug("Auditing chassis")

		if iloXML.HpInfra2.HpBlades != nil {
			for _, blade := range iloXML.HpInfra2.HpBlades.HpBlade {
				b := Blade{}

				b.BladePosition = blade.HpBay.Connection
				b.MgmtIPAddress = blade.MgmtIPAddr
				b.Power = blade.HpPower.PowerConsumed / 1000.00
				b.Temp = blade.HpTemps.HpTemp.C
				b.Serial = blade.Bsn

				if strings.Contains(blade.Spn, "Storage") {
					b.Name = b.Serial
					b.storageBlade = true
				} else {
					b.Name = blade.Name
					b.storageBlade = false
				}
				chassis.Blades = append(chassis.Blades, &b)
			}
		}
	}

	return chassis, err
}

func NewChassisConnection(username string, password string) *ChassisConnection {
	return &ChassisConnection{username: username, password: password}
}
