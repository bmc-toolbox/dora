package connectors

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	BiosVersion string `xml:" BIOS_VERSION,attr"`
	MbMacAddr1  string `xml:" MB_MAC_ADDR1,attr"`
	MbMacAddr2  string `xml:" MB_MAC_ADDR2,attr"`
}

// SupermicroPowerSupply holds the power supply information
type SupermicroPowerSupply struct {
	Location  string `xml:" LOCATION,attr"`
	Status    string `xml:" STATUS,attr"`
	Unplugged string `xml:" UNPLUGGED,attr"`
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

	return err
}

func (s *SupermicroReader) query(requestType string) (ipmi *SupermicroIPMI, err error) {
	bmcURL := fmt.Sprintf("https://%s/cgi/ipmi.cgi", *s.ip)
	log.WithFields(log.Fields{"step": "BMC Connection Supermicro", "ip": *s.ip, "url": bmcURL}).Debug("Retrieving data via BMC")

	req, err := http.NewRequest("POST", bmcURL, bytes.NewBufferString(requestType))
	if err != nil {
		return ipmi, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	u, err := url.Parse(bmcURL)
	if err != nil {
		return ipmi, err
	}
	for _, cookie := range s.client.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return ipmi, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ipmi, err
	}

	ipmi = &SupermicroIPMI{}
	err = xml.Unmarshal(payload, ipmi)
	if err != nil {
		DumpInvalidPayload(*s.ip, payload)
		return ipmi, err
	}

	return ipmi, err
}

// Logout logs out of the bmc
func (s *SupermicroReader) Logout() (err error) {
	bmcURL := fmt.Sprintf("https://%s/cgi/logout.cgi", *s.ip)
	log.WithFields(log.Fields{"step": "BMC Connection Supermicro", "ip": *s.ip, "url": bmcURL}).Debug("Logout from BMC")

	req, err := http.NewRequest("POST", bmcURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	u, err := url.Parse(bmcURL)
	if err != nil {
		return err
	}
	for _, cookie := range s.client.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// Serial returns the device serial
func (s *SupermicroReader) Serial() (serial string, err error) {
	ipmi, err := s.query("FRU_INFO.XML=(0,0)")
	if err != nil {
		return serial, err
	}

	serial = fmt.Sprintf("%s@%s", ipmi.FruInfo.Chassis.SerialNum, ipmi.FruInfo.Board.SerialNum)
	return serial, err
}

// Model returns the device model
func (s *SupermicroReader) Model() (model string, err error) {
	return
}
