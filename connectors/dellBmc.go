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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
)

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
