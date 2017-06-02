package simpleapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	// ErrNoChassi is returned when no chassis is found during the search
	ErrNoChassi = errors.New("no chassi found")

	// ErrNoBlade is returned when no blade is found during the search
	ErrNoBlade = errors.New("no blade found")
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
	SwitchID   int    `json:"switch_id"`
}

type Blade struct {
	IPAddress     string `json:"ip_address"`
	BladePosition int    `json:"blade_position"`
	ResourceURI   string `json:"resource_uri"`
	State         string `json:"state"`
}

type Chassi struct {
	Environment string                  `json:"environment"`
	Rack        string                  `json:"rack"`
	Blades      []map[string]Blade      `json:"servers"`
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
	Racks []Rack `json:"racks"`
}

type SimpleApiChassis struct {
	Chassis []Chassi `json:"chassis"`
}

// Chassis retrieves information from all chassis on SimpleAPI
func (s *SimpleAPI) Chassis() (chassis SimpleApiChassis, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sdb/api/v1/chassis", s.simpleapiurl), nil)
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error simpleapi:", err)
		return chassis, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading the response:", err)
		return chassis, err
	}

	err = json.Unmarshal(payload, &chassis)
	if err != nil {
		fmt.Println("error unmarshalling:", err)
		return chassis, err
	}
	return chassis, err
}

func (s *SimpleApiChassis) GetChassi(fqdn string) (chassi Chassi, err error) {
	for _, c := range s.Chassis {
		if c.Fqdn == fqdn {
			return c, err
		}
	}
	return chassi, ErrNoChassi
}

func (s *SimpleAPI) GetRack(name string) (rack Rack, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sdb/api/v1/racks/name/%s", s.simpleapiurl, name), nil)
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error simpleapi:", err)
		return rack, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading the response:", err)
		return rack, err
	}
	r := &SImpleApiRacks{}

	err = json.Unmarshal(payload, &r)
	if err != nil {
		fmt.Println("error unmarshalling:", err)
		return rack, err
	}
	return r.Racks[0], err
}

func (c *Chassi) GetBlade(fqdn string) (blade Blade, err error) {
	for _, b := range c.Blades {
		if _, ok := b[fqdn]; ok {
			return b[fqdn], err
		}
	}
	return blade, ErrNoBlade
}

func New(username string, password string, simpleapiurl string) *SimpleAPI {
	return &SimpleAPI{username: username, password: password, simpleapiurl: simpleapiurl}
}
