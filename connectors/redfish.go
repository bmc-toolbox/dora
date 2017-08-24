package connectors

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// HP is the constant that defines the vendor HP
	HP = "HP"
	// Dell is the constant that defines the vendor Dell
	Dell = "Dell"
	// Supermicro is the constant that defines the vendor Supermicro
	Supermicro = "Supermicro"
	// Common is the constant of thinks we could use across multiple vendors
	Common = "Common"
	// Unknown is the constant that defines Unknowns vendors
	Unknown = "Unknown"
	// RFPower is the constant for power definition on RedFish
	RFPower = "power"
	// RFThermal is the constant for thermal definition on RedFish
	RFThermal = "thermal"
	// RFEntry is used to identify the vendor of the redfish we are using
	RFEntry = "entry"
	// RFCPU is the constant for CPU definition on RedFish
	RFCPU = "cpu"
)

var (
	redfishVendorEndPoints = map[string]map[string]string{
		Dell: map[string]string{
			RFEntry:   "redfish/v1/Systems/System.Embedded.1/",
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
			RFCPU:     "redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1",
		},
		HP: map[string]string{
			RFEntry:   "rest/v1/Systems/1",
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
			RFCPU:     "rest/v1/Systems/1/Processors/1",
		},
		Supermicro: map[string]string{
			RFEntry:   "redfish/v1/Systems/1",
			RFPower:   "redfish/v1/Chassis/1/Power",
			RFThermal: "redfish/v1/Chassis/1/Thermal",
			// Review	RFCPU:     "redfish/v1/Systems/1/Processors/1",
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

	payload, err = r.get("redfish/v1/Systems/1")
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
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   20 * time.Second,
			KeepAlive: 20 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 20 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: tr,
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
