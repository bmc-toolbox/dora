package connectors

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
)

// SupermicroIPMI is the base structure that holds the information on queries to https://$ip/cgi/ipmi.cgi
type SupermicroIPMI struct {
	Bios         *SupermicroBios          `xml:" BIOS,omitempty"`
	CPU          []*SupermicroCPU         `xml:" CPU,omitempty"`
	ConfigInfo   *SupermicroConfigInfo    `xml:" CONFIG_INFO,omitempty"`
	Dimm         []*SupermicroDimm        `xml:" DIMM,omitempty"`
	FruInfo      *SupermicroFruInfo       `xml:" FRU_INFO,omitempty"`
	GenericInfo  *SupermicroGenericInfo   `xml:" GENERIC_INFO,omitempty"`
	PlatformInfo *SupermicroPlatformInfo  `xml:" PLATFORM_INFO,omitempty"`
	PowerSupply  []*SupermicroPowerSupply `xml:" PowerSupply,omitempty"`
	NodeInfo     *SupermicroNodeInfo      `xml:" NodeInfo,omitempty"`
	BiosLicense  *SupermicroBiosLicense   `xml:" BIOS_LINCESNE,omitempty" json:"BIOS_LINCESNE,omitempty"`
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
	Version string `xml:" VER,attr"`
}

// SupermicroConfigInfo holds the bmc configuration
type SupermicroConfigInfo struct {
	Hostname *SupermicroHostname `xml:" HOSTNAME,omitempty"`
}

// SupermicroHostname is the bmc hostname
type SupermicroHostname struct {
	Name string `xml:" NAME,attr"`
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
	MbMacAddr3  string `xml:" MB_MAC_ADDR3,attr"`
	MbMacAddr4  string `xml:" MB_MAC_ADDR4,attr"`
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

// SupermicroNodeInfo contains a lists of boards in the chassis
type SupermicroNodeInfo struct {
	Nodes []*SupermicroNode `xml:" Node,omitempty"`
}

// SupermicroNode contains the power and thermal information of each board in the chassis
type SupermicroNode struct {
	IP          string `xml:" IP,attr"`
	Power       string `xml:" Power,attr"`
	PowerStatus string `xml:" PowerStatus,attr"`
	SystemTemp  string `xml:" SystemTemp,attr"`
}

// SupermicroBiosLicense contains the license of bmc
type SupermicroBiosLicense struct {
	Check string `xml:" CHECK,attr"  json:",omitempty"`
}

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
	defer resp.Body.Close()

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

	serial = strings.TrimSpace(fmt.Sprintf("%s_%s", strings.TrimSpace(ipmi.FruInfo.Chassis.SerialNum), strings.TrimSpace(ipmi.FruInfo.Board.SerialNum)))
	return strings.ToLower(serial), err
}

// Model returns the device model
func (s *SupermicroReader) Model() (model string, err error) {
	ipmi, err := s.query("FRU_INFO.XML=(0,0)")
	if err != nil {
		return model, err
	}

	if ipmi.FruInfo != nil && ipmi.FruInfo.Board != nil {
		return ipmi.FruInfo.Board.ProdName, err
	}

	return model, err
}

// BmcType returns the type of bmc we are talking to
func (s *SupermicroReader) BmcType() (bmcType string, err error) {
	return "Supermicro", err
}

// BmcVersion returns the version of the bmc we are running
func (s *SupermicroReader) BmcVersion() (bmcVersion string, err error) {
	ipmi, err := s.query("GENERIC_INFO.XML=(0,0)")
	if err != nil {
		return bmcVersion, err
	}

	if ipmi.GenericInfo != nil && ipmi.GenericInfo.Generic != nil {
		return ipmi.GenericInfo.Generic.IpmiFwVersion, err
	}

	return bmcVersion, err
}

// Name returns the hostname of the machine
func (s *SupermicroReader) Name() (name string, err error) {
	ipmi, err := s.query("CONFIG_INFO.XML=(0,0)")
	if err != nil {
		return name, err
	}

	if ipmi.ConfigInfo != nil && ipmi.ConfigInfo.Hostname != nil {
		return ipmi.ConfigInfo.Hostname.Name, err
	}

	return name, err
}

// Status returns health string status from the bmc
func (s *SupermicroReader) Status() (health string, err error) {
	return "Not Supported", err
}

// Memory returns the total amount of memory of the server
func (s *SupermicroReader) Memory() (mem int, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")

	for _, dimm := range ipmi.Dimm {
		dimm := strings.TrimSuffix(dimm.Size, " MB")
		size, err := strconv.Atoi(dimm)
		if err != nil {
			return mem, err
		}
		mem += size
	}

	return mem / 1024, err
}

// CPU returns the cpu, cores and hyperthreads of the server
func (s *SupermicroReader) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")
	for _, entry := range ipmi.CPU {
		cpu = entry.Version
		cpuCount = len(ipmi.CPU)

		coreCount, err = strconv.Atoi(entry.Core)
		if err != nil {
			return cpu, cpuCount, coreCount, hyperthreadCount, err
		}

		hyperthreadCount = coreCount
		break
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// BiosVersion returns the current version of the bios
func (s *SupermicroReader) BiosVersion() (version string, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")
	if err != nil {
		return version, err
	}

	if ipmi.Bios != nil {
		return ipmi.Bios.Version, err
	}

	return version, err
}

// PowerKw returns the current power usage in Kw
func (s *SupermicroReader) PowerKw() (power float64, err error) {
	ipmi, err := s.query("Get_NodeInfoReadings.XML=(0,0)")
	if err != nil {
		return power, err
	}

	if ipmi.NodeInfo != nil {
		for _, node := range ipmi.NodeInfo.Nodes {
			if node.IP == strings.Split(*s.ip, ":")[0] {
				value, err := strconv.Atoi(node.Power)
				if err != nil {
					return power, err
				}

				return float64(value) / 1000.00, err
			}
		}
	}

	return power, err
}

// TempC returns the current temperature of the machine
func (s *SupermicroReader) TempC() (temp int, err error) {
	ipmi, err := s.query("Get_NodeInfoReadings.XML=(0,0)")
	if err != nil {
		return temp, err
	}

	if ipmi.NodeInfo != nil {
		for _, node := range ipmi.NodeInfo.Nodes {
			if node.IP == strings.Split(*s.ip, ":")[0] {
				temp, err := strconv.Atoi(node.SystemTemp)
				if err != nil {
					return temp, err
				}

				return temp, err
			}
		}
	}

	return temp, err
}

// Nics returns all found Nics in the device
func (s *SupermicroReader) Nics() (nics []*model.Nic, err error) {
	nics = make([]*model.Nic, 0)
	ipmi, err := s.query("GENERIC_INFO.XML=(0,0)")
	if err != nil {
		return nics, err
	}

	bmcNic := &model.Nic{
		Name:       "bmc",
		MacAddress: ipmi.GenericInfo.Generic.BmcMac,
	}

	nics = append(nics, bmcNic)

	ipmi, err = s.query("Get_PlatformInfo.XML=(0,0)")
	if err != nil {
		return nics, err
	}

	if ipmi.PlatformInfo != nil {
		if ipmi.PlatformInfo.MbMacAddr1 != "" {
			bmcNic := &model.Nic{
				Name:       "eth0",
				MacAddress: ipmi.PlatformInfo.MbMacAddr1,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr2 != "" {
			bmcNic := &model.Nic{
				Name:       "eth1",
				MacAddress: ipmi.PlatformInfo.MbMacAddr2,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr3 != "" {
			bmcNic := &model.Nic{
				Name:       "eth2",
				MacAddress: ipmi.PlatformInfo.MbMacAddr3,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr4 != "" {
			bmcNic := &model.Nic{
				Name:       "eth3",
				MacAddress: ipmi.PlatformInfo.MbMacAddr4,
			}
			nics = append(nics, bmcNic)
		}
	}

	return nics, err
}

// License returns the iLO's license information
func (s *SupermicroReader) License() (name string, licType string, err error) {
	ipmi, err := s.query("BIOS_LINCENSE_ACTIVATE.XML=(0,0)")
	if err != nil {
		return name, licType, err
	}

	if ipmi.BiosLicense != nil {
		switch ipmi.BiosLicense.Check {
		case "0":
			return "oob", "Activated", err
		case "1":
			return "oob", "Not Activated", err
		}
	}

	return name, licType, err
}
