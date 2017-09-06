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
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
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

// HpIloLicense is the struct used to render the data from https://$ip/json/license, it contains the license information of the ilo
type HpIloLicense struct {
	Name string `json:"name"`
	Type string `json:"type"`
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
	return strings.ToLower(strings.TrimSpace(i.hpRimpBlade.HpHSI.Sbsn)), err
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

// Status returns health string status from the bmc
func (i *IloReader) Status() (health string, err error) {
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

// Nics returns all found Nics in the device
func (i *IloReader) Nics() (nics []*model.Nic, err error) {
	nics = make([]*model.Nic, 0)

	if i.hpRimpBlade.HpHSI != nil &&
		i.hpRimpBlade.HpHSI.HpNICS != nil &&
		i.hpRimpBlade.HpHSI.HpNICS.HpNIC != nil {
		for _, nic := range i.hpRimpBlade.HpHSI.HpNICS.HpNIC {
			n := &model.Nic{
				Name:       nic.Description,
				MacAddress: strings.ToLower(nic.MacAddr),
			}
			nics = append(nics, n)
		}
	}
	return nics, err
}

// License returns the iLO's license information
func (i *IloReader) License() (name string, licType string, err error) {
	payload, err := i.get("json/license")
	if err != nil {
		return name, licType, err
	}

	hpIloLicense := &HpIloLicense{}
	err = json.Unmarshal(payload, hpIloLicense)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return name, licType, err
	}

	return hpIloLicense.Name, hpIloLicense.Type, err
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

// HpChassisReader holds the status and properties of a connection to a BladeSystem device
type HpChassisReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	hpRimp   *HpRimp
}

// NewHpChassisReader returns a connection to HpChassisReader
func NewHpChassisReader(ip *string, username *string, password *string) (chassis *HpChassisReader, err error) {
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

	if hpRimp.HpInfra2 == nil {
		return chassis, ErrUnabletoReadData
	}

	return &HpChassisReader{ip: ip, username: username, password: password, hpRimp: hpRimp, client: client}, err
}

func (h *HpChassisReader) Name() (name string, err error) {
	return h.hpRimp.HpInfra2.Encl, err
}

func (h *HpChassisReader) Model() (model string, err error) {
	return h.hpRimp.HpMP.Pn, err
}

func (h *HpChassisReader) Serial() (serial string, err error) {
	return strings.ToLower(strings.TrimSpace(h.hpRimp.HpInfra2.EnclSn)), err
}

func (h *HpChassisReader) PowerKw() (power float64, err error) {
	return h.hpRimp.HpInfra2.HpChassisPower.PowerConsumed / 1000.00, err
}

func (h *HpChassisReader) TempC() (temp int, err error) {
	return h.hpRimp.HpInfra2.HpTemps.HpTemp.C, err
}

func (h *HpChassisReader) Status() (status string, err error) {
	return h.hpRimp.HpInfra2.Status, err
}

func (h *HpChassisReader) FwVersion() (version string, err error) {
	return h.hpRimp.HpMP.Fwri, err
}

func (h *HpChassisReader) PassThru() (passthru string, err error) {
	passthru = "1G"
	for _, hpswitch := range h.hpRimp.HpInfra2.HpSwitches.HpSwitch {
		if strings.Contains(hpswitch.Spn, "10G") {
			passthru = "10G"
		}
		break
	}
	return passthru, err
}

func (h *HpChassisReader) PowerSupplyCount() (count int, err error) {
	return len(h.hpRimp.HpInfra2.HpChassisPower.HpPowersupply), err
}

func (h *HpChassisReader) Blades() (blades []*model.Blade, err error) {
	name, _ := h.Name()
	if h.hpRimp.HpInfra2.HpBlades != nil {
		for _, hpBlade := range h.hpRimp.HpInfra2.HpBlades.HpBlade {
			db := storage.InitDB()

			blade := model.Blade{}
			blade.BladePosition = hpBlade.HpBay.Connection
			blade.Status = hpBlade.Status
			blade.Serial = strings.ToLower(strings.TrimSpace(hpBlade.Bsn))

			if blade.Serial == "" || blade.Serial == "[unknown]" {
				nb := model.Blade{}
				db.Where("bmc_address = ? and blade_position = ?", hpBlade.MgmtIPAddr, hpBlade.HpBay.Connection).First(&nb)
				log.WithFields(log.Fields{"operation": "connection", "ip": *h.ip, "name": name, "position": blade.BladePosition, "type": "chassis", "error": "Review this blade. The chassis identifies it as connected, but we have no data"}).Error("Auditing blade")

				if nb.Serial == "" {
					continue
				}

				blade.Status = "Require Reseat"
				blade.Serial = nb.Serial
			}
			blade.PowerKw = hpBlade.HpPower.PowerConsumed / 1000.00
			blade.TempC = hpBlade.HpTemps.HpTemp.C
			blade.Vendor = HP
			blade.BmcType = hpBlade.MgmtType

			if strings.Contains(hpBlade.Spn, "Storage") {
				blade.Name = blade.Serial
				blade.IsStorageBlade = true
				blade.BmcAddress = "-"
			} else {
				blade.Name = hpBlade.Name
				blade.IsStorageBlade = false
				blade.BmcAddress = hpBlade.MgmtIPAddr
				blade.BmcVersion = hpBlade.MgmtVersion
				blade.Model = hpBlade.Spn

				if blade.BmcAddress == "0.0.0.0" || blade.BmcAddress == "" || blade.BmcAddress == "[]" {
					blade.BmcAddress = "unassigned"
					blade.BmcWEBReachable = false
					blade.BmcSSHReachable = false
					blade.BmcIpmiReachable = false
					blade.BmcAuth = false
				} else {
					scans := []model.ScannedPort{}
					db.Where("scanned_host_ip = ?", blade.BmcAddress).Find(&scans)
					for _, scan := range scans {
						if scan.Port == 443 && scan.Protocol == "tcp" && scan.State == "open" {
							blade.BmcWEBReachable = true
						} else if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
							blade.BmcSSHReachable = true
						} else if scan.Port == 623 && scan.Protocol == "udp" && scan.State == "open" {
							blade.BmcIpmiReachable = true
						}
					}

					if blade.BmcWEBReachable {
						ilo, err := NewIloReader(&blade.BmcAddress, h.username, h.password)
						if err == nil {
							blade.Nics, _ = ilo.Nics()
							err = ilo.Login()
							if err != nil {
								log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							} else {
								defer ilo.Logout()
								blade.BmcAuth = true

								blade.BiosVersion, err = ilo.BiosVersion()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading bios version", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
								}

								blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = ilo.CPU()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
								}

								blade.Memory, err = ilo.Memory()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
								}

								blade.BmcLicenceType, blade.BmcLicenceStatus, err = ilo.License()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
								}
							}
						}
					} else {
						log.WithFields(log.Fields{"operation": "create ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					}
				}
			}
			blades = append(blades, &blade)
		}
	}
	return blades, err
}
