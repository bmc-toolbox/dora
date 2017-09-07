package connectors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (

	// RFPower is the constant for power definition on RedFish
	RFPower = "power"
	// RFThermal is the constant for thermal definition on RedFish
	RFThermal = "thermal"
	// RFEntry is used to identify the vendor of the redfish we are using
	RFEntry = "entry"
	// RFCPU is the constant for CPU definition on RedFish
	RFCPU = "cpu"
	// RFBMC is the constant for BMC definition on RedFish
	RFBMC = "bmc"
	// RFBMCNetwork is the constant for BMC Network definition on RedFish
	RFBMCNetwork = "bmcNetwork"
)

var (
	redfishVendorEndPoints = map[string]map[string]string{
		Dell: map[string]string{
			RFEntry:      "redfish/v1/Systems/System.Embedded.1/",
			RFPower:      "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal:    "redfish/v1/Chassis/System.Embedded.1/Thermal",
			RFCPU:        "redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1",
			RFBMC:        "redfish/v1/Managers/iDRAC.Embedded.1/",
			RFBMCNetwork: "redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces/iDRAC.Embedded.1%23NIC.1",
		},
		HP: map[string]string{
			RFEntry:      "redfish/v1/Systems/1/",
			RFPower:      "redfish/v1/Chassis/1/Power/",
			RFThermal:    "redfish/v1/Chassis/1/Thermal/",
			RFCPU:        "redfish/v1/Systems/1/Processors/1/",
			RFBMC:        "redfish/v1/Managers/1/",
			RFBMCNetwork: "redfish/v1/Managers/1/EthernetInterfaces/1/",
		},
		Supermicro: map[string]string{
			RFEntry:      "redfish/v1/Systems/1/",
			RFPower:      "redfish/v1/Chassis/1/Power/",
			RFThermal:    "redfish/v1/Chassis/1/Thermal/",
			RFCPU:        "redfish/v1/Systems/1/Processors/1/",
			RFBMC:        "redfish/v1/Managers/1/",
			RFBMCNetwork: "redfish/v1/Managers/1/EthernetInterfaces/1/",
		},
	}
	redfishVendorLabels = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Board Inlet Temp",
		},
		HP: map[string]string{
			//			RFPower:   "PowerMetrics",
			RFThermal: "30-System Board",
		},
		Supermicro: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Temp",
		},
	}
	bmcAddressBuild = regexp.MustCompile(".(prod|corp|dqs).")
)

type RedFishEntry struct {
	BiosVersion      string                        `json:"BiosVersion"`
	Description      string                        `json:"Description"`
	HostName         string                        `json:"HostName"`
	Manufacturer     string                        `json:"Manufacturer"`
	MemorySummary    *RedFishEntryMemorySummary    `json:"MemorySummary"`
	Model            string                        `json:"Model"`
	PowerState       string                        `json:"PowerState"`
	ProcessorSummary *RedFishEntryProcessorSummary `json:"ProcessorSummary"`
	SerialNumber     string                        `json:"SerialNumber"`
	Status           *RedFishEntryStatus           `json:"Status"`
	SystemType       string                        `json:"SystemType"`
}

type RedFishEntryMemorySummary struct {
	Status               *RedFishEntryStatus `json:"Status"`
	TotalSystemMemoryGiB float64             `json:"TotalSystemMemoryGiB"`
}

type RedFishEntryProcessorSummary struct {
	Count  int                 `json:"Count"`
	Model  string              `json:"Model"`
	Status *RedFishEntryStatus `json:"Status"`
}

type RedFishEntryStatus struct {
	HealthRollUp string `json:"HealthRollUp"`
}

type RedFishHealth struct {
	Health string `json:"Health"`
}

type RedFishCPU struct {
	Model        string         `json:"Model"`
	Name         string         `json:"Name"`
	Status       *RedFishHealth `json:"Status"`
	TotalCores   int            `json:"TotalCores"`
	TotalThreads int            `json:"TotalThreads"`
}

// RedFishReader holds the status and properties of a connection to an iDrac device
type RedFishReader struct {
	ip       *string
	username *string
	password *string
	vendor   string
}

// RedFishManager holds the information related to the bmc itself
type RedFishManager struct {
	Description        string `json:"Description"`
	EthernetInterfaces struct {
		OdataID string `json:"@odata.id"`
	} `json:"EthernetInterfaces"`
	FirmwareVersion string         `json:"FirmwareVersion"`
	Status          *RedFishStatus `json:"Status"`
}

// RedFishEthernetInterfaces holds the information related to network interfaces of the bmc

// RedFishStatus contains the default RedFish status structure
type RedFishStatus struct {
	State string `json:"State"`
}

// NewRedFishReader returns a new RedFishReader ready to be used
func NewRedFishReader(ip *string, username *string, password *string) (r *RedFishReader, err error) {
	r = &RedFishReader{ip: ip, username: username, password: password}
	err = r.detectVendor()
	return r, err
}

func (r *RedFishReader) detectVendor() (err error) {
	payload, err := r.get("redfish/v1/")
	if err == ErrPageNotFound {
		return ErrRedFishNotSupported
	} else if err != nil {
		return err
	}

	if strings.Contains(string(payload), "iLO") {
		r.vendor = HP
		return err
	}

	if strings.Contains(string(payload), "iDRAC") {
		r.vendor = Dell
		return err
	}

	payload, err = r.get(redfishVendorEndPoints[Supermicro][RFEntry])
	if err != nil {
		return err
	}

	if strings.Contains(string(payload), "Supermicro") {
		r.vendor = Supermicro
		return err
	}

	return ErrVendorUnknown
}

// get, so theoretically we should be able to use a session for the whole RedFish connection, but it doesn't seems to be properly supported by any vendors
func (r *RedFishReader) get(endpoint string) (payload []byte, err error) {
	url := fmt.Sprintf("https://%s/%s", *r.ip, endpoint)
	if r.vendor == "" {
		log.WithFields(log.Fields{"step": fmt.Sprintf("RedFish Connection"), "ip": *r.ip, "url": url}).Debug("Retrieving data via RedFish")
	} else {
		log.WithFields(log.Fields{"step": fmt.Sprintf("RedFish Connection %s", r.vendor), "ip": *r.ip, "url": url}).Debug("Retrieving data via RedFish")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payload, err
	}
	req.SetBasicAuth(*r.username, *r.password)

	client, err := buildClient()
	if err != nil {
		return payload, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return payload, err
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
		return payload, err
	}

	return payload, err
}

// Memory returns the current memory installed in a given server
func (r *RedFishReader) Memory() (mem int, err error) {
	payload, err := r.get(redfishVendorEndPoints[r.vendor][RFEntry])
	if err != nil {
		return mem, err
	}

	redFishEntry := &RedFishEntry{}
	err = json.Unmarshal(payload, redFishEntry)
	if err != nil {
		DumpInvalidPayload(*r.ip, payload)
		return mem, err
	}

	return int(redFishEntry.MemorySummary.TotalSystemMemoryGiB), err
}

// CPU return the cpu, cores and hyperthreads the server
func (r *RedFishReader) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	payload, err := r.get(redfishVendorEndPoints[r.vendor][RFEntry])
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	redFishEntry := &RedFishEntry{}
	err = json.Unmarshal(payload, redFishEntry)
	if err != nil {
		DumpInvalidPayload(*r.ip, payload)
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	payload, err = r.get(redfishVendorEndPoints[r.vendor][RFCPU])
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	redFishCPU := &RedFishCPU{}
	err = json.Unmarshal(payload, redFishCPU)
	if err != nil {
		DumpInvalidPayload(*r.ip, payload)
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	return redFishEntry.ProcessorSummary.Model, redFishEntry.ProcessorSummary.Count, redFishCPU.TotalCores, redFishCPU.TotalThreads, err
}

// BiosVersion returns the current version of the bios
func (r *RedFishReader) BiosVersion() (version string, err error) {
	payload, err := r.get(redfishVendorEndPoints[r.vendor][RFEntry])
	if err != nil {
		return version, err
	}

	redFishEntry := &RedFishEntry{}
	err = json.Unmarshal(payload, redFishEntry)
	if err != nil {
		DumpInvalidPayload(*r.ip, payload)
		return version, err
	}

	return redFishEntry.BiosVersion, err
}

// BmcType returns the device model
func (r *RedFishReader) BmcType() (bmcType string, err error) {
	if r.vendor == Dell {
		return "iDRAC", err
	} else if r.vendor == HP {
		// Since we know that only ilo4 and ilo5 have redfish, if we don't find ilo5 in the fw string it's ilo4
		bmcversion, err := r.BmcVersion()
		if err != nil {
			return bmcType, err
		} else if strings.Contains(bmcversion, "iLO 5") {
			return "iLO5", err
		} else {
			return "iLO4", err
		}
	} else if r.vendor == Supermicro {
		return "Supermicro", err
	}

	return bmcType, err
}

// BmcVersion returns the device model
func (r *RedFishReader) BmcVersion() (bmcVersion string, err error) {
	payload, err := r.get(redfishVendorEndPoints[r.vendor][RFBMC])
	if err != nil {
		return bmcVersion, err
	}

	redFishManager := &RedFishManager{}
	err = json.Unmarshal(payload, redFishManager)
	if err != nil {
		DumpInvalidPayload(*r.ip, payload)
		return bmcVersion, err
	}

	return redFishManager.FirmwareVersion, err
}
