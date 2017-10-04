package connectors

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SupermicroIPMI is the base structure that holds the information on queries to https://$ip/cgi/ipmi.cgi
type SupermicroIPMI struct {
	Bios         *SupermicroBios          `xml:" BIOS,omitempty"`
	CPU          []*SupermicroCPU         `xml:" CPU,omitempty"`
	Dimm         []*SupermicroDimm        `xml:" DIMM,omitempty"`
	FruInfo      *SupermicroFruInfo       `xml:" FRU_INFO,omitempty"`
	GenericInfo  *SupermicroGenericInfo   `xml:" GENERIC_INFO,omitempty"`
	PlatformInfo *SupermicroPlatformInfo  `xml:" PLATFORM_INFO,omitempty"`
	PowerSupply  []*SupermicroPowerSupply `xml:" PowerSupply,omitempty"`
}

// SupermicroBios holds the bios information
type SupermicroBios struct {
	Date    string `xml:" REL_DATE,attr"`
	Vendor  string `xml:" VENDOR,attr"`
	Version string `xml:" VER,attr"`
}

// SupermicroCPU holds the cpu information
type SupermicroCPU struct {
	Core    string `xml:" CORE,attr"`
	Type    string `xml:" TYPE,attr"`
	Version string `xml:" VER,attr"`
}

// SupermicroDimm holds the ram information
type SupermicroDimm struct {
	Size string `xml:" SIZE,attr"`
}

// SupermicroFruInfo holds the fru ipmi information (serial numbers and so on)
type SupermicroFruInfo struct {
	Board   *SupermicroBoard   `xml:" BOARD,omitempty"`
	Chassis *SupermicroChassis `xml:" CHASSIS,omitempty"`
}

// SupermicroChassis holds the chassis information
type SupermicroChassis struct {
	PartNum   string `xml:" PART_NUM,attr"`
	SerialNum string `xml:" SERIAL_NUM,attr"`
}

// SupermicroBoard holds the mother board information
type SupermicroBoard struct {
	MfcName   string `xml:" MFC_NAME,attr"`
	PartNum   string `xml:" PART_NUM,attr"`
	ProdName  string `xml:" PROD_NAME,attr"`
	SerialNum string `xml:" SERIAL_NUM,attr"`
}

// SupermicroGenericInfo holds the bmc information
type SupermicroGenericInfo struct {
	Generic *SupermicroGeneric `xml:" GENERIC,omitempty"`
}

// SupermicroGeneric holds the bmc information
type SupermicroGeneric struct {
	BiosVersion   string `xml:" BIOS_VERSION,attr"`
	BmcIP         string `xml:" BMC_IP,attr"`
	BmcMac        string `xml:" BMC_MAC,attr"`
	IpmiFwVersion string `xml:" IPMIFW_VERSION,attr"`
}

// SupermicroPlatformInfo holds the hardware related information
type SupermicroPlatformInfo struct {
	BiosVersion string `xml:" BIOS_VERSION,attr"  json:",omitempty"`
	MbMacAddr1  string `xml:" MB_MAC_ADDR1,attr"  json:",omitempty"`
	MbMacAddr2  string `xml:" MB_MAC_ADDR2,attr"  json:",omitempty"`
}

// SupermicroPowerSupply holds the power supply information
type SupermicroPowerSupply struct {
	Location  string `xml:" LOCATION,attr"  json:",omitempty"`
	Status    string `xml:" STATUS,attr"  json:",omitempty"`
	Unplugged string `xml:" UNPLUGGED,attr"  json:",omitempty"`
}

// SupermicroReader holds the status and properties of a connection to a supermicro bmc
type SupermicroReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
}

// "FRU_INFO.XML=(0,0)" https://ha150datanode-28.example.com/cgi/ipmi.cgi
// "Get_PlatformCap.XML=(0,0)" https://ha150datanode-28.example.com/cgi/ipmi.cgi
// "GENERIC_INFO.XML=(0,0)" https://ha150datanode-28.example.com/cgi/ipmi.cgi
// "Get_PlatformInfo.XML=(0,0)" https://ha150datanode-28.example.com/cgi/ipmi.cgi
// "SMBIOS_INFO.XML=(0,0)" https://ha150datanode-28.example.com/cgi/ipmi.cgi

// NewSupermicroReader returns a new IloReader ready to be used
func NewSupermicroReader(ip *string, username *string, password *string) (sm *SupermicroReader, err error) {
	client, err := buildClient()
	if err != nil {
		return sm, err
	}
	return &SupermicroReader{ip: ip, username: username, password: password, client: client}, err
}

// Login initiates the connection to an iLO device
func (s *SupermicroReader) Login() (err error) {
	log.WithFields(log.Fields{"step": "BMC Connection Supermicro", "ip": *s.ip}).Debug("Connecting to BMC")

	data := fmt.Sprintf("name=%s&pwd=%s", *s.username, *s.password)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/cgi/login.cgi", *s.ip), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !strings.Contains(string(payload), "../cgi/url_redirect.cgi?url_name=mainmenu") {
		return ErrLoginFailed
	}
}

func (s *SupermicroReader) get(requestType string) (ipmi *SupermicroIPMI, err error) {
	url := fmt.Sprintf("https://%s/ipmi.cgi", *s.ip)
	log.WithFields(log.Fields{"step": "BMC Connection Supermicro", "ip": *s.ip, "url": url}).Debug("Retrieving data via BMC")

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestType))
	if err != nil {
		return ipmi, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return ipmi, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401:
		return payload, ErrLoginFailed
	case 404:
		return payload, ErrPageNotFound
	case 500:
		return payload, ErrRedFishEndPoint500
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return ipmi, err
	}

	err = xml.Unmarshal(payload, ipmi)
	if err != nil {
		DumpInvalidPayload(*s.ip, payload)
		return ipmi, err
	}

	return ipmi, err
}

// Serial returns the device serial
func (s *SupermicroReader) Serial() (serial string, err error) {
	ipmi, err := s.get("FRU_INFO.XML=(0,0)")
	if err != nil {
		return serial, err
	}
	serial := fmt.Sprintf("%s@%s", ipmi.FruInfo.Chassis.SerialNum, ipmi.FruInfo.Board.SerialNum)
	return serial, err
}

// Model returns the device model
func (s *SupermicroReader) Model() (model string, err error) {
	return
}
