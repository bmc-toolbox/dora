package simpleapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	// ErrNoChassi is returned when no chassis is found during the search
	ErrNoChassiFound = errors.New("no chassi found")

	// ErrNoBlade is returned when no blade is found during the search
	ErrNoBladeFound = errors.New("no blade found")
)

type SimpleAPI struct {
	username     string
	password     string
	simpleapiurl string
}

type NetInterface struct {
	IPAddress  string `json:"ip_address"`
	MacAddress string `json:"mac_address"`
	Vlan       int    `json:"vlan"`
	SwitchFqdn string `json:"switch_fqdn"`
	SwitchID   int    `json:"switch_id"`
}

type Blade struct {
	IPAddress     string `json:"ip_address"`
	BladePosition int    `json:"blade_position"`
	ResourceURI   string `json:"resource_uri"`
	State         string `json:"state"`
}

type Chassis struct {
	Environment string                  `json:"environment"`
	Rack        string                  `json:"rack"`
	Blades      []map[string]*Blade     `json:"servers"`
	Fqdn        string                  `json:"fqdn"`
	State       string                  `json:"state"`
	Location    string                  `json:"location"`
	Interfaces  map[string]NetInterface `json:"interfaces"`
	Position    int                     `json:"position"`
	Model       string                  `json:"model"`
}

type Rack struct {
	Sitezone    string `json:"sitezone"`
	Sitepod     string `json:"sitepod"`
	Environment string `json:"environment"`
	Siterow     string `json:"siterow"`
	Name        string `json:"name"`
	Site        string `json:"site"`
}

type SImpleApiRacks struct {
	Racks []*Rack `json:"racks"`
}

type SimpleApiServer struct {
	Server *Server `json:"server"`
}

type SimpleApiChassis struct {
	Chassis []*Chassis `json:"chassis"`
}

type Server struct {
	Chassis      string                  `json:"chassis"`
	Environment  string                  `json:"environment"`
	Fqdn         string                  `json:"fqdn"`
	Ilo          string                  `json:"ilo"`
	Interfaces   map[string]NetInterface `json:"interfaces"`
	IPAddress    string                  `json:"ip_address"`
	IsVMHost     bool                    `json:"is_vm_host"`
	KvmUser      string                  `json:"kvm_user"`
	LastEdit     string                  `json:"last_edit"`
	Location     string                  `json:"location"`
	Model        string                  `json:"model"`
	Pci          bool                    `json:"pci"`
	Rack         string                  `json:"rack"`
	RackID       int                     `json:"rack_id"`
	RackPosition int                     `json:"rack_position"`
	Roles        []string                `json:"roles"`
	SerialNumber string                  `json:"serial_number"`
	State        string                  `json:"state"`
}

func (s *SimpleAPI) httpGet(url string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payload, err
	}
	req.SetBasicAuth(s.username, s.password)
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

// Chassis retrieves information from all chassis on SimpleAPI
func (s *SimpleAPI) Chassis() (chassis SimpleApiChassis, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sdb/api/v1/chassis", s.simpleapiurl), nil)
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return chassis, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return chassis, err
	}

	err = json.Unmarshal(payload, &chassis)
	if err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetServer retrieves information for a give server
func (s *SimpleAPI) GetServer(hostname *string) (server *Server, err error) {
	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/servers/%s", s.simpleapiurl, *hostname))
	if err != nil {
		return server, err
	}
	sas := SimpleApiServer{}
	err = json.Unmarshal(payload, &sas)
	if err != nil {
		return server, err
	}
	return sas.Server, err
}

func (c *Chassis) GetBladeNameByBay(bladePosition int) (fqdn string, err error) {
	for _, blade := range c.Blades {
		for hostname, properties := range blade {
			if properties.BladePosition == bladePosition {
				return hostname, err
			}
		}
	}
	return fqdn, ErrNoBladeFound
}

func (s *SimpleApiChassis) GetChassis(fqdn string) (chassis Chassis, err error) {
	for _, c := range s.Chassis {
		if c.Fqdn == fqdn {
			return *c, err
		}
	}
	return chassis, ErrNoChassiFound
}

func (s *SimpleAPI) GetRack(name *string) (rack Rack, err error) {
	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/racks/name/%s", s.simpleapiurl, *name))
	if err != nil {
		return rack, err
	}
	r := &SImpleApiRacks{}

	err = json.Unmarshal(payload, &r)
	if err != nil {
		return rack, err
	}
	return *r.Racks[0], err
}

func (c *Chassis) GetBlade(fqdn string) (blade Blade, err error) {
	for _, b := range c.Blades {
		if _, ok := b[fqdn]; ok {
			return *b[fqdn], err
		}
	}
	return blade, ErrNoBladeFound
}

func New(username string, password string, simpleapiurl string) *SimpleAPI {
	return &SimpleAPI{username: username, password: password, simpleapiurl: simpleapiurl}
}
