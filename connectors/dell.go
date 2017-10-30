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
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

var (
	macFinder = regexp.MustCompile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
)

// DellCMC is the entry of the json exposed by dell
// We don't need to use an maps[string] with DellChassis, because we don't have clusters
type DellCMC struct {
	DellChassis *DellChassis `json:"0"`
}

// DellCMCTemp is the entry of the json exposed by dell when reading the temp metrics
type DellCMCTemp struct {
	DellChassisTemp *DellChassisTemp `json:"1"`
}

// DellChassisTemp is where the chassis thermal data is kept
type DellChassisTemp struct {
	TempHealth                 int    `json:"TempHealth"`
	TempUpperCriticalThreshold int    `json:"TempUpperCriticalThreshold"`
	TempSensorID               int    `json:"TempSensorID"`
	TempCurrentValue           int    `json:"TempCurrentValue"`
	TempLowerCriticalThreshold int    `json:"TempLowerCriticalThreshold"`
	TempPresence               int    `json:"TempPresence"`
	TempSensorName             string `json:"TempSensorName"`
}

// DellChassis groups all the interresting stuff we will ready from the chassis
type DellChassis struct {
	DellChassisGroupMemberHealthBlob *DellChassisGroupMemberHealthBlob `json:"ChassisGroupMemberHealthBlob"`
}

// DellChassisGroupMemberHealthBlob has a collection of metrics from the chassis, psu and blades
type DellChassisGroupMemberHealthBlob struct {
	DellBlades        map[string]*DellBlade `json:"blades_status"`
	DellPsuStatus     *DellPsuStatus        `json:"psu_status"`
	DellChassisStatus *DellChassisStatus    `json:"chassis_status"`
	DellCMCStatus     *DellCMCStatus        `json:"cmc_status"`
	// TODO: active_alerts
}

// DellChassisStatus expose the basic information that identify the chassis
type DellChassisStatus struct {
	ROCmcFwVersionString string `json:"RO_cmc_fw_version_string"`
	ROChassisServiceTag  string `json:"RO_chassis_service_tag"`
	ROChassisProductname string `json:"RO_chassis_productname"`
	CHASSISName          string `json:"CHASSIS_name"`
}

// DellCMCStatus brings the information about the cmc status itself we will use it to know if the chassis has errors
type DellCMCStatus struct {
	CMCActiveError string `json:"cmcActiveError"`
}

// DellNic is the nic we have on a servers
type DellNic struct {
	BladeNicName string `json:"bladeNicName"`
	BladeNicVer  string `json:"bladeNicVer"`
}

// DellBlade contains all the blade information
type DellBlade struct {
	BladeTemperature    string              `json:"bladeTemperature"`
	BladePresent        int                 `json:"bladePresent"`
	IdracURL            string              `json:"idracURL"`
	BladeLogDescription string              `json:"bladeLogDescription"`
	StorageNumDrives    int                 `json:"storageNumDrives"`
	BladeCPUInfo        string              `json:"bladeCpuInfo"`
	Nics                map[string]*DellNic `json:"nic"`
	BladeMasterSlot     int                 `json:"bladeMasterSlot"`
	BladeUSCVer         string              `json:"bladeUSCVer"`
	BladeSvcTag         string              `json:"bladeSvcTag"`
	BladeBIOSver        string              `json:"bladeBIOSver"`
	ActualPwrConsump    int                 `json:"actualPwrConsump"`
	IsStorageBlade      int                 `json:"isStorageBlade"`
	BladeModel          string              `json:"bladeModel"`
	BladeName           string              `json:"bladeName"`
	BladeSerialNum      string              `json:"bladeSerialNum"`
}

// DellPsuStatus contains the information and power usage of the pdus
type DellPsuStatus struct {
	AcPower  string `json:"acPower"`
	PsuCount int    `json:"psuCount"`
}

// DellBladeMemoryEndpoint is the struct used to collect data from "https://$ip/sysmgmt/2012/server/memory" when passing the header X_SYSMGMT_OPTIMIZE:true
type DellBladeMemoryEndpoint struct {
	Memory *DellBladeMemory `json:"Memory"`
}

// DellBladeMemory is part of the payload returned by "https://$ip/sysmgmt/2012/server/memory"
type DellBladeMemory struct {
	Capacity       int `json:"capacity"`
	ErrCorrection  int `json:"err_correction"`
	MaxCapacity    int `json:"max_capacity"`
	SlotsAvailable int `json:"slots_available"`
	SlotsUsed      int `json:"slots_used"`
}

// DellBladeProcessorEndpoint is the struct used to collect data from "https://$ip/sysmgmt/2012/server/processor" when passing the header X_SYSMGMT_OPTIMIZE:true
type DellBladeProcessorEndpoint struct {
	Proccessors map[string]*DellBladeProcessor `json:"Processor"`
}

// DellBladeProcessor contains the processor data information
type DellBladeProcessor struct {
	Brand             string                     `json:"brand"`
	CoreCount         int                        `json:"core_count"`
	CurrentSpeed      int                        `json:"current_speed"`
	DeviceDescription string                     `json:"device_description"`
	HyperThreading    []*DellBladeHyperThreading `json:"hyperThreading"`
}

// DellBladeHyperThreading contains the hyperthread information
type DellBladeHyperThreading struct {
	Capable int `json:"capable"`
	Enabled int `json:"enabled"`
}

// IDracAuth is the struct used to verify the iDrac authentication
type IDracAuth struct {
	Status     string `xml:"status"`
	AuthResult int    `xml:"authResult"`
	ForwardURL string `xml:"forwardUrl"`
	ErrorMsg   string `xml:"errorMsg"`
}

// IDracLicense is the struct used to collect data from "https://$ip/sysmgmt/2012/server/license" and it contains the license information for the bmc
type IDracLicense struct {
	License struct {
		VConsole int `json:"VCONSOLE"`
	} `json:"License"`
}

// IDracRoot is the structure used to render the data when querying -> https://$ip/data?get
type IDracRoot struct {
	BiosVer          string                 `xml:"biosVer"`
	FwVersion        string                 `xml:"fwVersion"`
	SysDesc          string                 `xml:"sysDesc"`
	Powermonitordata *IDracPowermonitordata `xml:"powermonitordata,omitempty"`
}

// IDracPowermonitordata contains the power consumption data for the iDrac
type IDracPowermonitordata struct {
	PresentReading *IDracPresentReading `xml:"presentReading,omitempty"`
}

// IDracPresentReading contains the present reading data
type IDracPresentReading struct {
	Reading *IDracReading `xml:" reading,omitempty"`
}

// IDracReading is used to express the power data
type IDracReading struct {
	ProbeName string `xml:" probeName,omitempty"`
	Reading   string `xml:" reading"`
	//Text             string            `xml:",chardata" json:",omitempty"`
}

// DellSVMInventory is the struct used to collect data from "https://$ip/sysmgmt/2012/server/inventory/software"
type DellSVMInventory struct {
	Device []*DellIDracDevice `xml:"Device"`
}

// DellIDracDevice contains the list of devices and their information
type DellIDracDevice struct {
	Display     string                `xml:" display,attr"`
	Application *DellIDracApplication `xml:" Application"`
}

// DellIDracApplication contains the name of the device and it's version
type DellIDracApplication struct {
	Display string `xml:" display,attr"`
	Version string `xml:" version,attr"`
}

// DellSystemServerOS contains the hostname, os name and os version
type DellSystemServerOS struct {
	SystemServerOS struct {
		HostName  string `json:"HostName"`
		OSName    string `json:"OSName"`
		OSVersion string `json:"OSVersion"`
	} `json:"system.ServerOS"`
}

// IDracInventory contains the whole hardware inventory exposed thru https://$ip/sysmgmt/2012/server/inventory/hardware
type IDracInventory struct {
	Version   string            `xml:" version,attr"`
	Component []*IDracComponent `xml:" Component,omitempty"`
}

// IDracComponent holds the information from each component detected by the iDrac
type IDracComponent struct {
	Classname  string           `xml:" Classname,attr"`
	Key        string           `xml:" Key,attr"`
	Properties []*IDracProperty `xml:" PROPERTY,omitempty"`
}

// IDracProperty is the property of each component exposed to iDrac
type IDracProperty struct {
	Name         string `xml:" NAME,attr"`
	Type         string `xml:" TYPE,attr"`
	DisplayValue string `xml:" DisplayValue,omitempty"`
	Value        string `xml:" VALUE,omitempty"`
}

// IDracTemp contains the data structure to render the thermal data from iDrac http://$ip/sysmgmt/2012/server/temperature
type IDracTemp struct {
	Statistics   string `json:"Statistics"`
	Temperatures struct {
		IDRACEmbedded1SystemBoardInletTemp struct {
			MaxFailure         int    `json:"max_failure"`
			MaxWarning         int    `json:"max_warning"`
			MaxWarningSettable int    `json:"max_warning_settable"`
			MinFailure         int    `json:"min_failure"`
			MinWarning         int    `json:"min_warning"`
			MinWarningSettable int    `json:"min_warning_settable"`
			Name               string `json:"name"`
			Reading            int    `json:"reading"`
			SensorStatus       int    `json:"sensor_status"`
		} `json:"iDRAC.Embedded.1#SystemBoardInletTemp"`
	} `json:"Temperatures"`
	IsFreshAirCompliant int `json:"is_fresh_air_compliant"`
}

// DellCMCWWN is the structure used to render the data when querying /json?method=blades-wwn-info
type DellCMCWWN struct {
	SlotMacWwn struct {
		SlotMacWwnList map[string]DellCMCWWNBlade `json:"slot_mac_wwn_list"`
	} `json:"slot_mac_wwn"`
}

// DellCMCWWNBlade contains the blade structure used by DellCMCWWN
type DellCMCWWNBlade struct {
	BladeSlotName     string `json:"bladeSlotName"`
	IsFullHeight      int    `json:"is_full_height"`
	IsNotDoubleHeight struct {
		IsInstalled string `json:"isInstalled"`
		PortFMAC    string `json:"portFMAC"`
	} `json:"is_not_double_height"`
}

// IDracReader holds the status and properties of a connection to an iDrac device
type IDracReader struct {
	ip             *string
	username       *string
	password       *string
	client         *http.Client
	st1            string
	st2            string
	iDracInventory *IDracInventory
}

// NewIDracReader returns a new IloReader ready to be used
func NewIDracReader(ip *string, username *string, password *string) (iDrac *IDracReader, err error) {
	client, err := buildClient()
	if err != nil {
		return iDrac, err
	}

	return &IDracReader{ip: ip, username: username, password: password, client: client}, err
}

// Login initiates the connection to an iLO device
func (i *IDracReader) Login() (err error) {
	log.WithFields(log.Fields{"step": "iDrac Connection Dell", "ip": *i.ip}).Debug("Connecting to iDrac")

	data := fmt.Sprintf("user=%s&password=%s", *i.username, *i.password)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/data/login", *i.ip), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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

	iDracAuth := &IDracAuth{}
	err = xml.Unmarshal(payload, iDracAuth)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return err
	}

	if iDracAuth.AuthResult == 1 {
		return ErrLoginFailed
	}

	stTemp := strings.Split(iDracAuth.ForwardURL, ",")
	i.st1 = strings.TrimLeft(stTemp[0], "index.html?ST1=")
	i.st2 = strings.TrimLeft(stTemp[1], "ST2=")

	err = i.loadHwData()
	if err != nil {
		return err
	}

	return err
}

// loadHwData load the full hardware information from the iDrac
func (i *IDracReader) loadHwData() (err error) {
	payload, err := i.get("sysmgmt/2012/server/inventory/hardware", nil)
	if err != nil {
		return err
	}

	iDracInventory := &IDracInventory{}
	err = xml.Unmarshal(payload, iDracInventory)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return err
	}

	if iDracInventory == nil || iDracInventory.Component == nil {
		return ErrUnableToReadData
	}

	i.iDracInventory = iDracInventory

	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IDracReader) get(endpoint string, extraHeaders *map[string]string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "iDrac Connection Dell", "ip": *i.ip, "endpoint": endpoint}).Debug("Retrieving data from iDrac")

	bmcURL := fmt.Sprintf("https://%s", *i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return payload, err
	}
	req.Header.Add("ST2", i.st2)
	if extraHeaders != nil {
		for key, value := range *extraHeaders {
			req.Header.Add(key, value)
		}
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return payload, err
	}

	for _, cookie := range i.client.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" {
			req.AddCookie(cookie)
		}
	}

	resp, err := i.client.Do(req)
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

// Nics returns all found Nics in the device
func (i *IDracReader) Nics() (nics []*model.Nic, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_NICView" {
			for _, property := range component.Properties {
				if property.Name == "ProductName" && property.Type == "string" {
					data := strings.Split(property.Value, " - ")
					if len(data) == 2 {
						if nics == nil {
							nics = make([]*model.Nic, 0)
						}

						n := &model.Nic{
							Name:       data[0],
							MacAddress: strings.ToLower(data[1]),
						}
						nics = append(nics, n)
					} else {
						log.WithFields(log.Fields{"operation": "connection", "ip": *i.ip, "type": "blade", "error": "Invalid network card, please review"}).Error("Auditing blade")
					}
				}
			}
		} else if component.Classname == "DCIM_iDRACCardView" {
			for _, property := range component.Properties {
				if property.Name == "PermanentMACAddress" && property.Type == "string" {
					if nics == nil {
						nics = make([]*model.Nic, 0)
					}

					n := &model.Nic{
						Name:       "bmc",
						MacAddress: strings.ToLower(property.Value),
					}
					nics = append(nics, n)
				}
			}
		}
	}
	return nics, err
}

// Serial returns the device serial
func (i *IDracReader) Serial() (serial string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "NodeID" && property.Type == "string" {
					return strings.ToLower(property.Value), err
				}
			}
		}
	}
	return serial, err
}

// Status returns health string status from the bmc
func (i *IDracReader) Status() (serial string, err error) {
	return "NotSupported", err
}

// PowerKw returns the current power usage in Kw
func (i *IDracReader) PowerKw() (power float64, err error) {
	payload, err := i.get("data?get=powermonitordata", nil)
	if err != nil {
		return power, err
	}

	iDracRoot := &IDracRoot{}
	err = xml.Unmarshal(payload, iDracRoot)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return power, err
	}

	if iDracRoot.Powermonitordata != nil && iDracRoot.Powermonitordata.PresentReading != nil && iDracRoot.Powermonitordata.PresentReading.Reading != nil {
		value, err := strconv.Atoi(iDracRoot.Powermonitordata.PresentReading.Reading.Reading)
		if err != nil {
			return power, err
		}
		return float64(value) / 1000.00, err
	}

	return power, err
}

// BiosVersion returns the current version of the bios
func (i *IDracReader) BiosVersion() (version string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "BIOSVersionString" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}

	return version, err
}

// Name returns the name of this server from the bmc point of view
func (i *IDracReader) Name() (name string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "HostName" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}

	return name, err
}

// BmcVersion returns the version of the bmc we are running
func (i *IDracReader) BmcVersion() (bmcVersion string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_iDRACCardView" {
			for _, property := range component.Properties {
				if property.Name == "FirmwareVersion" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}
	return bmcVersion, err
}

// Model returns the device model
func (i *IDracReader) Model() (model string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "Model" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}
	return model, err
}

// BmcType returns the type of bmc we are talking to
func (i *IDracReader) BmcType() (bmcType string, err error) {
	return "iDrac", err
}

// License returns the bmc license information
func (i *IDracReader) License() (name string, licType string, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	payload, err := i.get("sysmgmt/2012/server/license", extraHeaders)
	if err != nil {
		return name, licType, err
	}

	iDracLicense := &IDracLicense{}
	err = json.Unmarshal(payload, iDracLicense)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return name, licType, err
	}

	if iDracLicense.License.VConsole == 1 {
		return "Enterprise", "Licensed", err
	}
	return "-", "Unlicensed", err
}

// Memory return the total amount of memory of the server
func (i *IDracReader) Memory() (mem int, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	payload, err := i.get("sysmgmt/2012/server/memory", extraHeaders)
	if err != nil {
		return mem, err
	}

	dellBladeMemory := &DellBladeMemoryEndpoint{}
	err = json.Unmarshal(payload, dellBladeMemory)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return mem, err
	}

	return dellBladeMemory.Memory.Capacity / 1024, err
}

// TempC returns the current temperature of the machine
func (i *IDracReader) TempC() (temp int, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	payload, err := i.get("sysmgmt/2012/server/temperature", extraHeaders)
	if err != nil {
		return temp, err
	}

	iDracTemp := &IDracTemp{}
	err = json.Unmarshal(payload, iDracTemp)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return temp, err
	}

	return iDracTemp.Temperatures.IDRACEmbedded1SystemBoardInletTemp.Reading, err
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDracReader) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	payload, err := i.get("sysmgmt/2012/server/processor", extraHeaders)
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	dellBladeProc := &DellBladeProcessorEndpoint{}
	err = json.Unmarshal(payload, dellBladeProc)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	for _, proc := range dellBladeProc.Proccessors {
		hasHT := 0
		for _, ht := range proc.HyperThreading {
			if ht.Capable == 1 {
				hasHT = 2
			}
		}
		return strings.TrimSpace(proc.Brand), len(dellBladeProc.Proccessors), proc.CoreCount, proc.CoreCount * hasHT, err
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// Logout logs out and close the iLo connection
func (i *IDracReader) Logout() (err error) {
	log.WithFields(log.Fields{"step": "iDrac Connection Dell", "ip": *i.ip}).Debug("Logout from iDrac")

	resp, err := i.client.Get(fmt.Sprintf("https://%s/data/logout", *i.ip))
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return err
}

// DellCmcReader holds the status and properties of a connection to a CMC device
type DellCmcReader struct {
	ip       *string
	username *string
	password *string
	cmcJSON  *DellCMC
	cmcTemp  *DellCMCTemp
}

// NewDellCmcReader returns a connection to DellCmcReader
func NewDellCmcReader(ip *string, username *string, password *string) (chassis *DellCmcReader, err error) {
	payload, err := httpGetDell(ip, "json?method=groupinfo", username, password)
	if err != nil {
		return chassis, err
	}

	dellCMC := &DellCMC{}
	err = json.Unmarshal(payload, dellCMC)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	if dellCMC.DellChassis == nil {
		return chassis, ErrUnableToReadData
	}

	//
	// payload, err = httpGetDell(ip, "json?method=blades-wwn-info", username, password)
	// if err != nil {
	// 	return chassis, err
	// }

	// dellCMCWWN := &DellCMCWWN{}
	// err = json.Unmarshal(payload, dellCMCWWN)
	// if err != nil {
	// 	DumpInvalidPayload(*ip, payload)
	// 	return chassis, err
	// }
	// fmt.Printf("%v", dellCMCWWN)

	return &DellCmcReader{ip: ip, username: username, password: password, cmcJSON: dellCMC}, err
}

// Name returns the hostname of the machine
func (d *DellCmcReader) Name() (name string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.CHASSISName, err
}

// Model returns the device model
func (d *DellCmcReader) Model() (model string, err error) {
	return strings.TrimSpace(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisProductname), err
}

// Serial returns the device serial
func (d *DellCmcReader) Serial() (serial string, err error) {
	return strings.ToLower(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisServiceTag), err
}

// PowerKw returns the current power usage in Kw
func (d *DellCmcReader) PowerKw() (power float64, err error) {
	p, err := strconv.Atoi(strings.TrimRight(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.AcPower, " W"))
	if err != nil {
		return power, err
	}
	return float64(p) / 1000.00, err
}

// TempC returns the current temperature of the machine
func (d *DellCmcReader) TempC() (temp int, err error) {
	payload, err := httpGetDell(d.ip, "json?method=temp-sensors", d.username, d.password)
	if err != nil {
		return temp, err
	}

	dellCMCTemp := &DellCMCTemp{}
	err = json.Unmarshal(payload, dellCMCTemp)
	if err != nil {
		DumpInvalidPayload(*d.ip, payload)
		return temp, err
	}

	if dellCMCTemp.DellChassisTemp != nil {
		return dellCMCTemp.DellChassisTemp.TempCurrentValue, err
	}

	return temp, err
}

// Status returns health string status from the bmc
func (d *DellCmcReader) Status() (status string, err error) {
	if d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError == "No Errors" {
		status = "OK"
	} else {
		status = d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError
	}
	return status, err
}

// PowerSupplyCount returns the total count of the power supply
func (d *DellCmcReader) PowerSupplyCount() (count int, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.PsuCount, err
}

// FwVersion returns the current firmware version of the bmc
func (d *DellCmcReader) FwVersion() (version string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString, err
}

// Nics returns all found Nics in the device
func (d *DellCmcReader) Nics() (nics []*model.Nic, err error) {
	payload, err := httpGetDell(d.ip, "cmc_status?cat=C01&tab=T11&id=P31", d.username, d.password)
	if err != nil {
		return nics, err
	}

	mac := macFinder.FindString(string(payload))
	if mac != "" {
		nics = make([]*model.Nic, 0)
		n := &model.Nic{
			Name:       "OA1",
			MacAddress: strings.ToLower(mac),
		}
		nics = append(nics, n)
	}

	return nics, err
}

// PassThru returns the type of switch we have for this chassis
func (d *DellCmcReader) PassThru() (passthru string, err error) {
	passthru = "1G"
	for _, dellBlade := range d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 0 {
			for _, nic := range dellBlade.Nics {
				if strings.Contains(nic.BladeNicName, "10G") {
					passthru = "10G"
				} else {
					passthru = "1G"
				}
				return passthru, err
			}
		}
	}
	return passthru, err
}

// StorageBlades returns all StorageBlades found in this chassis
func (d *DellCmcReader) StorageBlades() (storageBlades []*model.StorageBlade, err error) {
	// db := storage.InitDB()
	for _, dellBlade := range d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 1 {
			storageBlade := model.StorageBlade{}

			storageBlade.BladePosition = dellBlade.BladeMasterSlot
			storageBlade.Serial = strings.ToLower(dellBlade.BladeSvcTag)
			chassisSerial, _ := d.Serial()
			if storageBlade.Serial == "" || storageBlade.Serial == "[unknown]" || storageBlade.Serial == "0000000000" {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": storageBlade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": ErrInvalidSerial}).Error("Auditing blade")
				continue
			}

			storageBlade.Model = dellBlade.BladeModel
			storageBlade.PowerKw = float64(dellBlade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(dellBlade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": storageBlade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": err}).Warning("Auditing blade")
				continue
			}
			storageBlade.TempC = temp
			if dellBlade.BladeLogDescription == "No Errors" {
				storageBlade.Status = "OK"
			} else {
				storageBlade.Status = dellBlade.BladeLogDescription
			}
			storageBlade.Vendor = Dell
			storageBlade.FwVersion = dellBlade.BladeBIOSver

			// Todo: We will fix the association as soon as we get a storage blade :)
			// blade := model.Blade{}
			// db.Where("chassis_serial = ? and blade_position = ?", chassisSerial, hpBlade.AssociatedBlade).First(&blade)
			// if blade.Serial != "" {
			// 	storageBlade.BladeSerial = blade.Serial
			// }
			storageBlades = append(storageBlades, &storageBlade)
		}
	}
	return storageBlades, err
}

func (d *DellCmcReader) Blades() (blades []*model.Blade, err error) {
	db := storage.InitDB()
	for _, dellBlade := range d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 0 {
			blade := model.Blade{}

			blade.BladePosition = dellBlade.BladeMasterSlot
			blade.Serial = strings.ToLower(dellBlade.BladeSvcTag)
			chassisSerial, _ := d.Serial()

			if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": "Review this blade. The chassis identifies it as connected, but we have no data"}).Error("Auditing blade")
				continue
			}

			blade.Model = dellBlade.BladeModel
			blade.PowerKw = float64(dellBlade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(dellBlade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": err}).Warning("Auditing blade")
				continue
			} else {
				blade.TempC = temp
			}
			if dellBlade.BladeLogDescription == "No Errors" {
				blade.Status = "OK"
			} else {
				blade.Status = dellBlade.BladeLogDescription
			}
			blade.Vendor = Dell
			blade.BiosVersion = dellBlade.BladeBIOSver

			blade.BmcType = "iDRAC"
			blade.Name = dellBlade.BladeName
			idracURL := strings.TrimLeft(dellBlade.IdracURL, "https://")
			idracURL = strings.TrimLeft(idracURL, "http://")
			idracURL = strings.Split(idracURL, ":")[0]
			blade.BmcAddress = idracURL
			blade.BmcVersion = dellBlade.BladeUSCVer

			for _, nic := range dellBlade.Nics {
				if nic.BladeNicName == "" {
					log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": "Network card information missing, please verify"}).Error("Auditing blade")
					continue
				}
				n := &model.Nic{
					Name:       strings.ToLower(nic.BladeNicName[:len(nic.BladeNicName)-17]),
					MacAddress: strings.ToLower(nic.BladeNicName[len(nic.BladeNicName)-17:]),
				}
				blade.Nics = append(blade.Nics, n)
			}

			if blade.BmcAddress == "0.0.0.0" || blade.BmcAddress == "" || blade.BmcAddress == "[]" {
				blade.BmcAddress = "unassigned"
				blade.BmcWEBReachable = false
				blade.BmcSSHReachable = false
				blade.BmcIpmiReachable = false
				blade.BmcAuth = false
			} else {
				scans := []model.ScannedPort{}
				db.Where("ip = ?", blade.BmcAddress).Find(&scans)
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
					idrac, err := NewIDracReader(&blade.BmcAddress, d.username, d.password)
					if err != nil {
						log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					} else {
						err = idrac.Login()
						if err == nil {
							defer idrac.Logout()
							blade.BmcAuth = true

							blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = idrac.CPU()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							blade.Nics, err = idrac.Nics()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading nics", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							blade.Memory, err = idrac.Memory()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							blade.BmcLicenceType, blade.BmcLicenceStatus, err = idrac.License()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}
						}
					}
				} else {
					log.WithFields(log.Fields{"operation": "create ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
				}
			}
			blades = append(blades, &blade)
		}
	}
	return blades, err
}
