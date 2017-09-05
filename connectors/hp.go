package connectors

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// HpBlade contains the unmarshalled data from the hp chassis
type HpBlade struct {
	HpBay       *HpBay   `xml:" BAY,omitempty"`
	Bsn         string   `xml:" BSN,omitempty"`
	MgmtIPAddr  string   `xml:" MGMTIPADDR,omitempty"`
	MgmtType    string   `xml:" MGMTPN,omitempty"`
	MgmtVersion string   `xml:" MGMTFWVERSION,omitempty"`
	Name        string   `xml:" NAME,omitempty"`
	HpPower     *HpPower `xml:" POWER,omitempty"`
	Status      string   `xml:" STATUS,omitempty"`
	Spn         string   `xml:" SPN,omitempty"`
	HpTemps     *HpTemps `xml:" TEMPS,omitempty"`
}

// HpBay contains the position of the blade within the chassis
type HpBay struct {
	Connection int `xml:" CONNECTION,omitempty"`
}

// HpInfra2 is the data retrieved from the chassis xml interface that contains all components
type HpInfra2 struct {
	Addr           string          `xml:" ADDR,omitempty"`
	HpBlades       *HpBlades       `xml:" BLADES,omitempty"`
	HpSwitches     *HpSwitches     `xml:" SWITCHES,omitempty"`
	HpChassisPower *HpChassisPower `xml:" POWER,omitempty"`
	Status         string          `xml:" STATUS,omitempty"`
	HpTemps        *HpTemps        `xml:" TEMPS,omitempty"`
	EnclSn         string          `xml:" ENCL_SN,omitempty"`
	Pn             string          `xml:" PN,omitempty"`
	Encl           string          `xml:" ENCL,omitempty"`
	Rack           string          `xml:" RACK,omitempty"`
}

// HpMP contains the firmware version and the model of the chassis or blade
type HpMP struct {
	Pn   string `xml:" PN,omitempty"`
	Sn   string `xml:" SN,omitempty"`
	Fwri string `xml:" FWRI,omitempty"`
}

// HpSwitches contains all the switches we have within the chassis
type HpSwitches struct {
	HpSwitch []*HpSwitch `xml:" SWITCH,omitempty"`
}

// HpSwitch contains the type of the switch
type HpSwitch struct {
	Spn string `xml:" SPN,omitempty"`
}

// HpBlades contains all the blades we have within the chassis
type HpBlades struct {
	HpBlade []*HpBlade `xml:" BLADE,omitempty"`
}

// HpPower contains the power information of a blade
type HpPower struct {
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty"`
}

// HpChassisPower contains the power information of the chassis
type HpChassisPower struct {
	PowerConsumed float64          `xml:" POWER_CONSUMED,omitempty"`
	HpPowersupply []*HpPowersupply `xml:" POWERSUPPLY,omitempty"`
}

// HpRimp is the entry data structure for the chassis
type HpRimp struct {
	HpInfra2 *HpInfra2 `xml:" INFRA2,omitempty"`
	HpMP     *HpMP     `xml:" MP,omitempty"`
}

// HpPowersupply contains the data of the power supply of the chassis
type HpPowersupply struct {
	Status string `xml:" STATUS,omitempty"`
}

// HpTemp contains the thermal data of a chassis or blade
type HpTemp struct {
	C    int    `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty"`
}

// HpTemps contains the thermal data of a chassis or blade
type HpTemps struct {
	HpTemp *HpTemp `xml:" TEMP,omitempty"`
}

// HpRimpBlade is the entry data structure for the blade when queries directly
type HpRimpBlade struct {
	HpMP         *HpMP         `xml:" MP,omitempty"`
	HpHSI        *HpHSI        `xml:" HSI,omitempty"`
	HpBladeBlade *HpBladeBlade `xml:" BLADESYSTEM,omitempty"`
}

// HpBladeBlade blade information from the hprimp of blades
type HpBladeBlade struct {
	Bay int `xml:" BAY,omitempty"`
}

// HpHSI contains the information about the components of the blade
type HpHSI struct {
	HpNICS *HpNICS `xml:" NICS,omitempty"`
	Sbsn   string  `xml:" SBSN,omitempty" json:"SBSN,omitempty"`
	Spn    string  `xml:" SPN,omitempty" json:"SPN,omitempty"`
}

// HpNICS contains the list of nics that a blade has
type HpNICS struct {
	HpNIC []*HpNIC `xml:" NIC,omitempty"`
}

// HpNIC contains the nic information of a blade
type HpNIC struct {
	Description string `xml:" DESCRIPTION,omitempty"`
	MacAddr     string `xml:" MACADDR,omitempty"`
	Status      string `xml:" STATUS,omitempty"`
}

// HpFirmware is the struct used to render the data from https://$ip/json/fw_info, it contains firmware data of the blade
type HpFirmware struct {
	Firmware []struct {
		FwName    string `json:"fw_name"`
		FwVersion string `json:"fw_version"`
	} `json:"firmware"`
}

// HpProcs is the struct used to render the data from https://$ip/json/proc_info, it contains the processor data
type HpProcs struct {
	Processors []struct {
		ProcName       string `json:"proc_name"`
		ProcNumCores   int    `json:"proc_num_cores"`
		ProcNumThreads int    `json:"proc_num_threads"`
	} `json:"processors"`
}

// HpMem is the struct used to render the data from https://$ip/json/mem_info, it contains the ram data
type HpMem struct {
	MemTotalMemSize int          `json:"mem_total_mem_size"`
	Memory          []*HpMemSlot `json:"memory"`
}

// HpMemSlot is part of the payload returned from https://$ip/json/mem_info
type HpMemSlot struct {
	MemDevLoc string `json:"mem_dev_loc"`
	MemSize   int    `json:"mem_size"`
	MemSpeed  int    `json:"mem_speed"`
}

// HpOverview is the struct used to render the data from https://$ip/json/overview, it contains information about bios version, ilo license and a bit more
type HpOverview struct {
	ServerName    string `json:"server_name"`
	ProductName   string `json:"product_name"`
	SerialNum     string `json:"serial_num"`
	SystemRom     string `json:"system_rom"`
	SystemRomDate string `json:"system_rom_date"`
	BackupRomDate string `json:"backup_rom_date"`
	License       string `json:"license"`
	IloFwVersion  string `json:"ilo_fw_version"`
	IPAddress     string `json:"ip_address"`
	SystemHealth  string `json:"system_health"`
	Power         string `json:"power"`
}

// HpPowerSummary is the struct used to render the data from https://$ip/json/power_summary, it contains the basic information about the power usage of the machine
type HpPowerSummary struct {
	HostpwrState          string `json:"hostpwr_state"`
	PowerSupplyInputPower int    `json:"power_supply_input_power"`
}

// HpHelthTemperature is the struct used to render the data from https://$ip/json/health_temperature, it contains the information about the thermal status of the machine
type HpHelthTemperature struct {
	HostpwrState string           `json:"hostpwr_state"`
	InPost       int              `json:"in_post"`
	Temperature  []*HpTemperature `json:"temperature"`
}

// HpTemperature is part of the data rendered from https://$ip/json/health_temperature, it contains the names of each component and their current temp
type HpTemperature struct {
	Label          string `json:"label"`
	Location       string `json:"location"`
	Status         string `json:"status"`
	Currentreading int    `json:"currentreading"`
	TempUnit       string `json:"temp_unit"`
}

// IloReader holds the status and properties of a connection to an iLO device
type IloReader struct {
	ip          *string
	username    *string
	password    *string
	client      *http.Client
	loginURL    *url.URL
	hpRimpBlade *HpRimpBlade
}

// NewIloReader returns a new IloReader ready to be used
func NewIloReader(ip *string, username *string, password *string) (ilo *IloReader, err error) {
	loginURL, err := url.Parse(fmt.Sprintf("https://%s/json/login_session", *ip))
	if err != nil {
		return nil, err
	}

	client, err := buildClient()
	if err != nil {
		return ilo, err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", *ip))
	if err != nil {
		return ilo, err
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ilo, err
	}
	defer resp.Body.Close()

	hpRimpBlade := &HpRimpBlade{}
	err = xml.Unmarshal(payload, hpRimpBlade)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return ilo, err
	}

	return &IloReader{ip: ip, username: username, password: password, loginURL: loginURL, hpRimpBlade: hpRimpBlade, client: client}, err
}

// Login initiates the connection to an iLO device
func (i *IloReader) Login() (err error) {
	log.WithFields(log.Fields{"step": "Ilo Connection HP", "ip": *i.ip}).Debug("Connecting to iLO")

	data := fmt.Sprintf("{\"method\":\"login\", \"user_login\":\"%s\", \"password\":\"%s\" }", *i.username, *i.password)

	req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(req)
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

	if strings.Contains(string(payload), "Invalid login attempt") {
		return ErrLoginFailed
	}

	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IloReader) get(endpoint string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "Ilo Connection HP", "ip": *i.ip, "endpoint": endpoint}).Debug("Retrieving data from iLO")

	resp, err := i.client.Get(fmt.Sprintf("https://%s/%s", *i.ip, endpoint))
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	return payload, err
}

// Serial returns the device model
func (i *IloReader) Serial() (serial string, err error) {
	return i.hpRimpBlade.HpHSI.Sbsn, err
}

// Model returns the device model
func (i *IloReader) Model() (model string, err error) {
	return i.hpRimpBlade.HpHSI.Spn, err
}

// BmcType returns the device model
func (i *IloReader) BmcType() (bmcType string, err error) {
	switch i.hpRimpBlade.HpMP.Pn {
	case "Integrated Lights-Out 4 (iLO 4)":
		return "iLO4", err
	case "Integrated Lights-Out 3 (iLO 3)":
		return "iLO3", err
	default:
		return i.hpRimpBlade.HpMP.Pn, err
	}
}

// BmcVersion returns the device model
func (i *IloReader) BmcVersion() (bmcVersion string, err error) {
	return i.hpRimpBlade.HpMP.Fwri, err
}

// Name returns the name of this server from the iLO point of view
func (i *IloReader) Name() (name string, err error) {
	payload, err := i.get("json/overview")
	if err != nil {
		return name, err
	}

	hpOverview := &HpOverview{}
	err = json.Unmarshal(payload, hpOverview)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return name, err
	}

	return hpOverview.ServerName, err
}

// SystemHealth returns health string status from the bmc
func (i *IloReader) SystemHealth() (health string, err error) {
	payload, err := i.get("json/overview")
	if err != nil {
		return health, err
	}

	hpOverview := &HpOverview{}
	err = json.Unmarshal(payload, hpOverview)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return health, err
	}

	if hpOverview.SystemHealth == "OP_STATUS_OK" {
		return "OK", err
	}

	return hpOverview.SystemHealth, err
}

// Memory returns the total amount of memory of the server
func (i *IloReader) Memory() (mem int, err error) {
	payload, err := i.get("json/mem_info")
	if err != nil {
		return mem, err
	}

	hpMemData := &HpMem{}
	err = json.Unmarshal(payload, hpMemData)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return mem, err
	}

	if hpMemData.MemTotalMemSize != 0 {
		return hpMemData.MemTotalMemSize / 1024, err
	}

	for _, slot := range hpMemData.Memory {
		mem = mem + slot.MemSize
	}

	return mem / 1024, err
}

// CPU returns the cpu, cores and hyperthreads of the server
func (i *IloReader) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	payload, err := i.get("json/proc_info")
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	hpProcData := &HpProcs{}
	err = json.Unmarshal(payload, hpProcData)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	for _, proc := range hpProcData.Processors {
		return strings.TrimSpace(proc.ProcName), len(hpProcData.Processors), proc.ProcNumCores, proc.ProcNumThreads, err
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// BiosVersion returns the current version of the bios
func (i *IloReader) BiosVersion() (version string, err error) {
	payload, err := i.get("json/overview")
	if err != nil {
		return version, err
	}

	hpOverview := &HpOverview{}
	err = json.Unmarshal(payload, hpOverview)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return version, err
	}

	if hpOverview.SystemRom != "" {
		return hpOverview.SystemRom, err
	}

	return version, ErrBiosNotFound
}

// PowerKw returns the current power usage in Kw
func (i *IloReader) PowerKw() (power float64, err error) {
	payload, err := i.get("json/power_summary")
	if err != nil {
		return power, err
	}

	hpPowerSummary := &HpPowerSummary{}
	err = json.Unmarshal(payload, hpPowerSummary)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return power, err
	}

	return float64(hpPowerSummary.PowerSupplyInputPower) / 1024, err
}

// TempC returns the current verion of the bios
func (i *IloReader) TempC() (temp int, err error) {
	payload, err := i.get("json/health_temperature")
	if err != nil {
		return temp, err
	}

	hpHelthTemperature := &HpHelthTemperature{}
	err = json.Unmarshal(payload, hpHelthTemperature)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return temp, err
	}

	for _, item := range hpHelthTemperature.Temperature {
		if item.Location == "Ambient" {
			return item.Currentreading, err
		}
	}

	return temp, err
}

// Logout logs out and close the iLo connection
func (i *IloReader) Logout() (err error) {
	log.WithFields(log.Fields{"step": "Ilo Connection HP", "ip": *i.ip}).Debug("Logout from iLO")

	data := []byte(`{"method":"logout"}`)

	req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return err
}

// BladeSystemC7000Reader holds the status and properties of a connection to an BladeSystem device
type BladeSystemC7000Reader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	hpRimp   *HpRimp
}

// NewBladeSystemC7000Reader returns a new IloReader ready to be used
func NewBladeSystemC7000Reader(ip *string, username *string, password *string) (chassis *BladeSystemC7000Reader, err error) {
	client, err := buildClient()
	if err != nil {
		return chassis, err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", *ip))
	if err != nil {
		return chassis, err
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return chassis, err
	}
	defer resp.Body.Close()

	hpRimp := &HpRimp{}
	err = xml.Unmarshal(payload, hpRimp)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	return &BladeSystemC7000Reader{ip: ip, username: username, password: password, hpRimp: hpRimp, client: client}, err
}

func (c *BladeSystemC7000Reader) Chassis() (err error) {
	return
}
