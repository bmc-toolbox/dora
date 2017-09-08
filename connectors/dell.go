package connectors

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
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

// IDracReader holds the status and properties of a connection to an iDrac device
type IDracReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	st1      string
	st2      string
}

// NewIDracReader returns a new IloReader ready to be used
func NewIDracReader(ip *string, username *string, password *string) (iDrac *IDracReader) {
	return &IDracReader{ip: ip, username: username, password: password}
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

	client, err := buildClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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

	i.client = client

	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IDracReader) get(endpoint string, extraHeaders *map[string]string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "iDrac Connection Dell", "ip": *i.ip, "endpoint": endpoint}).Debug("Retrieving data from iDrac")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", *i.ip, endpoint), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("ST2", i.st2)
	for key, value := range *extraHeaders {
		req.Header.Add(key, value)
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
		return chassis, ErrUnabletoReadData
	}

	return &DellCmcReader{ip: ip, username: username, password: password, cmcJSON: dellCMC}, err
}

func (d *DellCmcReader) Name() (name string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.CHASSISName, err
}

func (d *DellCmcReader) Model() (model string, err error) {
	return strings.TrimSpace(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisProductname), err
}

func (d *DellCmcReader) Serial() (serial string, err error) {
	return strings.ToLower(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisServiceTag), err
}

func (d *DellCmcReader) PowerKw() (power float64, err error) {
	p, err := strconv.Atoi(strings.TrimRight(d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.AcPower, " W"))
	if err != nil {
		return power, err
	}
	return float64(p) / 1000.00, err
}

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

func (d *DellCmcReader) Status() (status string, err error) {
	if d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError == "No Errors" {
		status = "OK"
	} else {
		status = d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError
	}
	return status, err
}

func (d *DellCmcReader) PowerSupplyCount() (count int, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.PsuCount, err
}

func (d *DellCmcReader) FwVersion() (version string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString, err
}

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
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": storageBlade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": "Review this blade. The chassis identifies it as connected, but we have no data"}).Error("Auditing blade")
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
					redFish, err := NewRedFishReader(&blade.BmcAddress, d.username, d.password)
					if err != nil {
						blade.BmcAuth = true

						blade.Memory, err = redFish.Memory()
						if err != nil {
							log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
						}

						blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = redFish.CPU()
						if err != nil {
							log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
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
