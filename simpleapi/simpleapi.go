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

	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoChassi is returned when no chassis is found during the search
	ErrChassiNotFound = errors.New("Chassis not found")

	// ErrNoBlade is returned when no blade is found during the search
	ErrBladeNotFound = errors.New("Blade not found")

	// ErrRackNotFound is returned when no rack is found during the search
	ErrRackNotFound = errors.New("Rack not found")

	// ErrChassisNotFound is returned when no chassis is found during the search
	ErrChassisNotFound = errors.New("Chassis not found")
)

type SimpleAPI struct {
	username     string
	password     string
	simpleapiurl string
	servers      *SimpleApiServers
	racks        *SimpleApiRacks
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

type SimpleApiRacks struct {
	Racks []*Rack `json:"racks"`
}

type SimpleApiServer struct {
	Server *Server `json:"server"`
}

type SimpleApiServers struct {
	Servers []*Server `json:"servers"`
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
	log.WithFields(log.Fields{"step": "simpleapi", "url": url}).Debug("Requesting data from SimpleAPI")

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
	if s.servers != nil {
		for _, server := range s.servers.Servers {
			if server.Fqdn == *hostname {
				return server, err
			}
		}
	}

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

// GetServer retrieves all servers
func (s *SimpleAPI) Servers() (servers SimpleApiServers, err error) {
	if s.servers != nil {
		return *s.servers, err
	}

	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/servers", s.simpleapiurl))
	if err != nil {
		return servers, err
	}
	sas := SimpleApiServers{}
	err = json.Unmarshal(payload, &sas)
	if err != nil {
		return servers, err
	}
	return sas, err
}

// Cache all servers to speed up the run
func (s *SimpleAPI) CacheServers() (err error) {
	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/servers", s.simpleapiurl))
	if err != nil {
		return err
	}
	sas := SimpleApiServers{}
	err = json.Unmarshal(payload, &sas)
	if err != nil {
		return err
	}
	s.servers = &sas
	return err
}

// Cache all racks to speed up the run
func (s *SimpleAPI) CacheRacks() (err error) {
	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/racks", s.simpleapiurl))
	if err != nil {
		return err
	}
	sar := SimpleApiRacks{}
	err = json.Unmarshal(payload, &sar)
	if err != nil {
		return err
	}
	s.racks = &sar
	return err
}

func (c *Chassis) GetBladeNameByBay(bladePosition int) (fqdn string, err error) {
	log.WithFields(log.Fields{"step": "simpleapi", "slot": bladePosition}).Debug("Retrieving blade from SimpleAPI using Slot")
	for _, blade := range c.Blades {
		for hostname, properties := range blade {
			if properties.BladePosition == bladePosition {
				return hostname, err
			}
		}
	}
	return fqdn, ErrBladeNotFound
}

func (s *Server) MainRole() *string {
	role := "Unknown"
	for _, r := range s.Roles {
		if r != "staging" {
			return &r
		}
	}
	return &role
}

func (s *SimpleApiChassis) GetChassis(fqdn string) (chassis Chassis, err error) {
	log.WithFields(log.Fields{"step": "simpleapi", "chassis": fqdn}).Debug("Retrieving chassis from SimpleAPI using fqdn")
	for _, c := range s.Chassis {
		if c.Fqdn == fqdn {
			return *c, err
		}
	}
	return chassis, ErrChassisNotFound
}

func (s *SimpleAPI) GetRack(name *string) (rack Rack, err error) {
	if s.racks != nil {
		for _, rack := range s.racks.Racks {
			if rack.Name == *name {
				return *rack, err
			}
		}
	}

	payload, err := s.httpGet(fmt.Sprintf("%s/sdb/api/v1/racks/name/%s", s.simpleapiurl, *name))
	if err != nil {
		return rack, err
	}
	r := &SimpleApiRacks{}

	err = json.Unmarshal(payload, &r)
	if err != nil {
		return rack, err
	}

	if len(r.Racks) == 0 {
		return rack, ErrRackNotFound
	}

	return *r.Racks[0], err
}

func (c *Chassis) GetBlade(fqdn string) (blade Blade, err error) {
	log.WithFields(log.Fields{"step": "simpleapi", "blade": fqdn}).Debug("Retrieving blade from SimpleAPI using fqdn")
	for _, b := range c.Blades {
		if _, ok := b[fqdn]; ok {
			return *b[fqdn], err
		}
	}
	return blade, ErrBladeNotFound
}

func New(username string, password string, simpleapiurl string) (s *SimpleAPI) {
	s = &SimpleAPI{username: username, password: password, simpleapiurl: simpleapiurl}
	err := s.CacheServers()
	if err != nil {
		log.WithFields(log.Fields{"step": "simpleapi", "error": err}).Warning("Problem caching servers")
	}
	err = s.CacheRacks()
	if err != nil {
		log.WithFields(log.Fields{"step": "simpleapi", "error": err}).Warning("Problem caching racks")
	}
	return s
}
