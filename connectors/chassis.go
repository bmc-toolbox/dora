package connectors

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"gitlab.booking.com/infra/thermalnator/collectors"

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
)

var (
	bladeDevice        = "blade"
	chassisDevice      = "chassis"
	storageBladeDevice = "storageblade"
	ErrPageNotFound    = errors.New("Requested page couldn't be found in the server")
)

type Blade struct {
	name          string  `json:"name"`
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

func (b *Blade) isStorageBlade(bool) {
	return b.isStorageBlade
}

func (c *ChassisConnection) httpGet(url string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "url": url}).Debug("Requesting data from BMC")

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

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

func (c *ChassisConnection) httpGetDell(hostname *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "hostname": *hostname}).Debug("Requesting data from BMC")

	form := url.Values{}
	form.Add("user", c.username)
	form.Add("password", c.password)

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

func (c *ChassisConnection) Dell(ip *string) (payload []byte, err error) {
	return c.httpGetDell(ip)
}

func (c *ChassisConnection) Hp(ip *string) (chassis Chassis, err error) {
	result, err = c.httpGet(fmt.Sprintf("https://%s/xmldata?item=infra2", *ip))
	if err != nil {
		return chassis, err
	}
	iloXML := &Rimp{}
	err = xml.Unmarshal(result, iloXML)
	if err != nil {
		return chassis, err
	}

	var previousSlot collectors.Blade

	if iloXML.Infra2 != nil {
		chassis.Name = iloXML.Infra2.Encl
		chassis.Serial = iloXML.Infra2.EnclSn
		chassis.Model = iloXML.Infra2.Pn
		chassis.Rack = iloXML.Infra2.Rack
		chassis.Power = iloXML.Infra2.Power.PowerConsumed / 1000.00
		chassis.Temp = iloXML.Infra2.Temps.Temp.C

		log.WithFields(log.Fields{"operation": "connection", "ip": ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis"}).Debug("Auditing chassis")

		if iloXML.Infra2.Blades != nil {
			for _, blade := range iloXML.Infra2.Blades.Blade {
				b := Blade{}

				b.BladePosition = blade.Bay.Connection
				b.MgmtIPAddress = blade.MgmtIPAddress
				b.Power = blade.Power.PowerConsumed / 1000.00
				b.Temp = blade.Temps.Temp.C
				b.Serial = blade.Bsn

				if strings.Contains(blade.Spn, "Storage") {
					b.storageBlade = true
				} else {
					b.storageBlade = false
				}
				chassis.Blades = append(chassis.Blades, b)
			}
		}
	}

	return chassis, err
}

func NewChassisConnection(username string, password string) *ChassisConnection {
	return &ChassisConnection{username: username, password: password}
}
