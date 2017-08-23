package connectors

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrRedFishNotSupported = errors.New("RedFish not supported")
	redfishVendorEndPoints = map[string]map[string]string{
		Dell: map[string]string{
			RFEntry:   "redfish/v1/Systems/System.Embedded.1/",
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
		},
		HP: map[string]string{
			RFEntry:   "rest/v1/Systems/1",
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
		Supermicro: map[string]string{
			RFEntry:   "redfish/v1/Systems/1",
			RFPower:   "redfish/v1/Chassis/1/Power",
			RFThermal: "redfish/v1/Chassis/1/Thermal",
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
	OdataContext     string                        `json:"@odata.context"`
	OdataID          string                        `json:"@odata.id"`
	OdataType        string                        `json:"@odata.type"`
	BiosVersion      string                        `json:"BiosVersion"`
	Description      string                        `json:"Description"`
	HostName         string                        `json:"HostName"`
	ID               string                        `json:"Id"`
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
	payload, err := r.get("/redfish/v1/")
	if err != nil {
		return err
	}

	if strings.Contains(string(payload), "iLO") {
		r.vendor = HP
		return err
	}

	payload, err = r.get("/redfish/v1/odata/")
	if err != nil {
		return err
	}

	if strings.Contains(string(payload), "iDrac") {
		r.vendor = HP
		return err
	}

	if strings.Contains(string(payload), "/redfish/v1/JsonSchemas") {
		r.vendor = Supermicro
		return err
	}

	return ErrVendorUnknown
}

// get, so theoretically we should be able to use a session for the whole RedFish connection, but it doesn't seems to be properly supported by any vendors
func (r *RedFishReader) get(endpoint string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", *r.ip, endpoint), nil)
	if err != nil {
		return payload, err
	}
	req.SetBasicAuth(*r.username, *r.password)
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

// Memory returns the current memory installed in a given server
func (r *RedFishReader) Memory() (mem int, err error) {

}
