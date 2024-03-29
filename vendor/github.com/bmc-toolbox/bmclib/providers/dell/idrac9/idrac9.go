package idrac9

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
)

const (
	// BMCType defines the bmc model that is supported by this package
	BMCType = "idrac9"
)

// IDrac9 holds the status and properties of a connection to an iDrac device
type IDrac9 struct {
	ip                   string
	username             string
	password             string
	xsrfToken            string
	httpClient           *http.Client
	sshClient            *sshclient.SSHClient
	iDracInventory       *dell.IDracInventory
	ctx                  context.Context
	log                  logr.Logger
	httpClientSetupFuncs []func(*http.Client)
}

// IDrac9Option is a type that can configure an *IDrac9
type IDrac9Option func(*IDrac9)

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) IDrac9Option {
	return func(i *IDrac9) {
		i.httpClientSetupFuncs = append(i.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// WithHTTPClient sets an HTTP client on an *IDrac9
func WithHTTPClient(c *http.Client) IDrac9Option {
	return func(i *IDrac9) {
		i.httpClient = c
	}
}

// New returns a new IDrac9 ready to be used
func New(ctx context.Context, host string, httpHost string, username string, password string, log logr.Logger) (*IDrac9, error) {
	return NewWithOptions(ctx, host, httpHost, username, password, log)
}

// NewWithOptions returns a new IDrac9 with options ready to be used
func NewWithOptions(ctx context.Context, host, httpHost string, username string, password string, log logr.Logger, opts ...IDrac9Option) (*IDrac9, error) {
	sshClient, err := sshclient.New(host, username, password)
	if err != nil {
		return nil, err
	}

	idrac := &IDrac9{ip: httpHost, username: username, password: password, sshClient: sshClient, ctx: ctx, log: log}

	for _, opt := range opts {
		opt(idrac)
	}
	if idrac.httpClient != nil {
		for _, setupFunc := range idrac.httpClientSetupFuncs {
			setupFunc(idrac.httpClient)
		}
	}
	return idrac, nil
}

// CheckCredentials verify whether the credentials are valid or not
func (i *IDrac9) CheckCredentials() (err error) {
	err = i.httpLogin()
	if err != nil {
		return err
	}
	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IDrac9) get(endpoint string, extraHeaders *map[string]string) (statusCode int, payload []byte, err error) {
	i.log.V(1).Info("retrieving data from bmc", "step", "bmc connection", "vendor", dell.VendorID, "ip", i.ip, "endpoint", endpoint)

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Add("XSRF-TOKEN", i.xsrfToken)

	if extraHeaders != nil {
		for key, value := range *extraHeaders {
			req.Header.Add(key, value)
		}
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("%s/%s", bmcURL, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if resp.StatusCode == 404 {
		return 404, payload, errors.ErrPageNotFound
	}

	return resp.StatusCode, payload, err
}

// PUTs data
func (i *IDrac9) put(endpoint string, body []byte) (statusCode int, payload []byte, err error) {
	bmcURL := fmt.Sprintf("https://%s", i.ip)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", bmcURL, endpoint), bytes.NewReader(body))
	if err != nil {
		return statusCode, payload, err
	}

	req.Header.Add("XSRF-TOKEN", i.xsrfToken)

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("%s/%s", bmcURL, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return statusCode, payload, err
	}
	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, payload, err
	}

	if resp.StatusCode == 500 {
		return resp.StatusCode, payload, errors.Err500
	}

	return resp.StatusCode, payload, err
}

// calls delete on the given endpoint
func (i *IDrac9) delete(endpoint string) (statusCode int, payload []byte, err error) {
	bmcURL := fmt.Sprintf("https://%s", i.ip)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return 0, []byte{}, err
	}

	req.Header.Add("XSRF-TOKEN", i.xsrfToken)

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("%s/%s", bmcURL, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return statusCode, payload, err
	}

	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, payload, err
	}

	return resp.StatusCode, payload, err
}

// posts the payload to the given endpoint
func (i *IDrac9) post(endpoint string, data []byte, formDataContentType string) (statusCode int, body []byte, err error) {
	u, err := url.Parse(fmt.Sprintf("https://%s/%s", i.ip, endpoint))
	if err != nil {
		return 0, []byte{}, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, []byte{}, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" || cookie.Name == "tokenvalue" {
			req.AddCookie(cookie)
		}
	}
	req.Header.Add("XSRF-TOKEN", i.xsrfToken)

	if formDataContentType != "" {
		// Set multipart form content type
		req.Header.Set("Content-Type", formDataContentType)
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("https://%s/%s", i.ip, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, []byte{}, err
	}
	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, []byte{}, err
	}

	return resp.StatusCode, body, err
}

// Nics returns all found Nics in the device
func (i *IDrac9) Nics() (nics []*devices.Nic, err error) {
	err = i.loadHwData()
	if err != nil {
		return nics, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_NICView" {
			var speed string
			var up bool
			var macAddress string
			var name string
			for _, property := range component.Properties {
				if property.Name == "LinkSpeed" && property.Type == "uint8" && property.DisplayValue != "Unknown" {
					speed = property.DisplayValue
					up = true
				} else if property.Name == "PermanentMACAddress" && property.Type == "string" {
					macAddress = strings.ToLower(property.Value)
				} else if property.Name == "ProductName" && property.Type == "string" {
					name = strings.Split(property.Value, " - ")[0]
				}
			}

			if macAddress != "" {
				if nics == nil {
					nics = make([]*devices.Nic, 0)
				}
				n := &devices.Nic{
					Name:       name,
					Speed:      speed,
					Up:         up,
					MacAddress: macAddress,
				}
				nics = append(nics, n)
			}
		} else if component.Classname == "DCIM_iDRACCardView" {
			for _, property := range component.Properties {
				if property.Name == "PermanentMACAddress" && property.Type == "string" {
					if nics == nil {
						nics = make([]*devices.Nic, 0)
					}

					n := &devices.Nic{
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

// Returns the device serial or an empty string in case it doesn't find it.
func (i *IDrac9) Serial() (serial string, err error) {
	err = i.loadHwData()
	if err != nil {
		return "", err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "NodeID" && property.Type == "string" {
					return strings.ToLower(property.Value), err
				}
			}
		}
	}

	return "", fmt.Errorf("IDrac9 Serial(): Serial not found!")
}

// ChassisSerial returns the serial number of the chassis where the blade is attached
func (i *IDrac9) ChassisSerial() (serial string, err error) {
	err = i.loadHwData()
	if err != nil {
		return serial, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "ChassisServiceTag" && property.Type == "string" {
					return strings.ToLower(property.Value), err
				}
			}
		}
	}
	return serial, err
}

// Status returns health string status from the bmc
func (i *IDrac9) Status() (status string, err error) {
	err = i.httpLogin()
	if err != nil {
		return status, err
	}

	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2016/server/extended_health"
	statusCode, payload, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", err
	}

	iDracHealthStatus := &dell.IDracHealthStatus{}
	err = json.Unmarshal(payload, iDracHealthStatus)
	if err != nil {
		return status, err
	}

	for _, entry := range iDracHealthStatus.HealthStatus {
		if entry != 0 && entry != 2 {
			return "Degraded", err
		}
	}

	return "OK", err
}

// PowerKw returns the current power usage in Kw
func (i *IDrac9) PowerKw() (power float64, err error) {
	err = i.httpLogin()
	if err != nil {
		return 0, err
	}

	endpoint := "sysmgmt/2015/server/sensor/power"
	statusCode, payload, err := i.get(endpoint, nil)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return 0, err
	}

	iDracPowerData := &dell.IDrac9PowerData{}
	err = json.Unmarshal(payload, iDracPowerData)
	if err != nil {
		return power, err
	}

	if len(iDracPowerData.Root.Powermonitordata.PresentReading.Reading) == 0 {
		return power, err
	}

	return iDracPowerData.Root.Powermonitordata.PresentReading.Reading[0].Reading / 1000.00, err
}

// PowerState returns the current power state of the machine
func (i *IDrac9) PowerState() (state string, err error) {
	err = i.loadHwData()
	if err != nil {
		return state, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "PowerState" && property.Type == "uint16" {
					return strings.ToLower(property.DisplayValue), err
				}
			}
		}
	}
	return state, err
}

// BiosVersion returns the current version of the bios
func (i *IDrac9) BiosVersion() (version string, err error) {
	err = i.loadHwData()
	if err != nil {
		return version, err
	}

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
func (i *IDrac9) Name() (name string, err error) {
	err = i.loadHwData()
	if err != nil {
		return name, err
	}

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

// Version returns the version of the bmc we are running
func (i *IDrac9) Version() (bmcVersion string, err error) {
	err = i.loadHwData()
	if err != nil {
		return bmcVersion, err
	}

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

// Slot returns the current slot within the chassis
func (i *IDrac9) Slot() (slot int, err error) {
	err = i.loadHwData()
	if err != nil {
		return -1, err
	}

	model, err := i.Model()
	if err != nil {
		return -1, err
	}

	if model == "PowerEdge C6420" {
		return i.slotC6420()
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "BaseBoardChassisSlot" && property.Type == "string" {
					if property.Value == "NA" {
						return -1, err
					}
					v := strings.Split(property.Value, " ")
					if len(v) < 2 {
						return -1, fmt.Errorf("Looks like the BaseBoardChassisSlot is ill-formatted!")
					}
					slot, err = strconv.Atoi(v[1])
					if err != nil {
						return -1, err
					}

					return slot, err
				}
			}
		}
	}

	return -1, err
}

// slotC6420 returns the current slot for the C6420 blade within the chassis
func (i *IDrac9) slotC6420() (slot int, err error) {
	endpoint := "sysmgmt/2012/server/configgroup/System.ServerTopology"
	statusCode, payload, err := i.get(endpoint, nil)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return -1, err
	}

	iDracSystemTopology := &dell.SystemTopology{}
	err = json.Unmarshal(payload, iDracSystemTopology)
	if err != nil {
		return -1, err
	}

	if iDracSystemTopology.SystemServerTopology.BladeSlotNumInChassis == "" {
		return -1, err
	}

	slot, err = strconv.Atoi(iDracSystemTopology.SystemServerTopology.BladeSlotNumInChassis)
	if err != nil {
		return -1, err
	}

	return slot, err
}

// Returns the device model or an empty string in case it doesn't find it.
func (i *IDrac9) Model() (model string, err error) {
	err = i.loadHwData()
	if err != nil {
		return "", err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "Model" && property.Type == "string" {
					return property.Value, nil
				}
			}
		}
	}

	return "", fmt.Errorf("IDrac9 Model(): Model not found!")
}

// HardwareType returns the type of bmc we are talking to
func (i *IDrac9) HardwareType() (bmcType string) {
	return BMCType
}

// License returns the bmc license information
func (i *IDrac9) License() (name string, licType string, err error) {
	err = i.httpLogin()
	if err != nil {
		return name, licType, err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2012/server/license"
	statusCode, payload, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", "", err
	}

	iDracLicense := &dell.IDracLicense{}
	err = json.Unmarshal(payload, iDracLicense)
	if err != nil {
		return "", "", err
	}

	if iDracLicense.License.VConsole == 1 {
		return "Enterprise", "Licensed", err
	}
	return "-", "Unlicensed", err
}

// Memory return the total amount of memory of the server
func (i *IDrac9) Memory() (mem int, err error) {
	err = i.loadHwData()
	if err != nil {
		return mem, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "SysMemTotalSize" && property.Type == "uint32" {
					size, err := strconv.Atoi(property.Value)
					if err != nil {
						return mem, err
					}
					return size / 1024, err
				}
			}
		}
	}
	return mem, err
}

// TempC returns the current temperature of the machine
func (i *IDrac9) TempC() (temp int, err error) {
	err = i.httpLogin()
	if err != nil {
		return 0, err
	}

	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2012/server/temperature"
	statusCode, payload, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return 0, err
	}

	iDracTemp := &dell.IDracTemp{}
	err = json.Unmarshal(payload, iDracTemp)
	if err != nil {
		return 0, err
	}

	return iDracTemp.Temperatures.IDRACEmbedded1SystemBoardInletTemp.Reading, nil
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDrac9) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	err = i.loadHwData()
	if err != nil {
		return "", 0, 0, 0, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_CPUView" {
			cpuCount++
			if component.Key == "CPU.Socket.1" {
				var e error
				for _, property := range component.Properties {
					if property.Name == "Model" && property.Type == "string" {
						cpu = httpclient.StandardizeProcessorName(property.DisplayValue)
					} else if property.Name == "NumberOfProcessorCores" && property.Type == "uint32" {
						if coreCount, e = strconv.Atoi(property.Value); e != nil {
							err = multierror.Append(err, fmt.Errorf("invalid core count %s", e))
						}
					} else if property.Name == "NumberOfEnabledThreads" && property.Type == "uint32" {
						if hyperthreadCount, e = strconv.Atoi(property.Value); e != nil {
							err = multierror.Append(err, fmt.Errorf("invalid thread count %s", e))
						}
					}
				}
			}
		}
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// IsBlade returns if the current hardware is a blade or not
func (i *IDrac9) IsBlade() (isBlade bool, err error) {
	err = i.httpLogin()
	if err != nil {
		return isBlade, err
	}

	serial, err := i.Serial()
	if err != nil {
		return isBlade, err
	}

	chassisSerial, err := i.ChassisSerial()
	if err != nil {
		return isBlade, err
	}

	if serial != chassisSerial {
		return true, err
	}

	return isBlade, err
}

// Psus returns a list of psus installed on the device
func (i *IDrac9) Psus() (psus []*devices.Psu, err error) {
	err = i.httpLogin()
	if err != nil {
		return psus, err
	}

	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2013/server/sensor/powersupplyunit"
	statusCode, payload, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return psus, err
	}

	iDracPowersupplyunit := &dell.IDracPowersupplyunit{}
	err = json.Unmarshal(payload, iDracPowersupplyunit)
	if err != nil {
		return psus, err
	}

	serial, _ := i.Serial()

	for _, psu := range iDracPowersupplyunit.Powersupplyunits {
		if psus == nil {
			psus = make([]*devices.Psu, 0)
		}

		var status string
		if psu.Health == 2 {
			status = "OK"
		} else {
			status = "BROKEN"
		}

		p := &devices.Psu{
			Serial:     strings.ToLower(fmt.Sprintf("%s_%s", serial, strings.Split(psu.Name, " ")[0])),
			Status:     status,
			PartNumber: strings.ToLower(psu.PartNumber),
			CapacityKw: float64(psu.OutputWattage) / 1000.00,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// Vendor returns bmc's vendor
func (i *IDrac9) Vendor() (vendor string) {
	return dell.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *IDrac9) ServerSnapshot() (server interface{}, err error) { // nolint: gocyclo
	err = i.httpLogin()
	if err != nil {
		return server, err
	}

	if isBlade, _ := i.IsBlade(); isBlade {
		blade := &devices.Blade{}
		blade.Vendor = i.Vendor()
		blade.BmcAddress = i.ip
		blade.BmcType = i.HardwareType()

		blade.Serial, err = i.Serial()
		if err != nil {
			return nil, err
		}
		blade.BmcVersion, err = i.Version()
		if err != nil {
			return nil, err
		}
		blade.Model, err = i.Model()
		if err != nil {
			return nil, err
		}
		blade.Nics, err = i.Nics()
		if err != nil {
			return nil, err
		}
		blade.Disks, err = i.Disks()
		if err != nil {
			return nil, err
		}
		blade.BiosVersion, err = i.BiosVersion()
		if err != nil {
			return nil, err
		}
		blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = i.CPU()
		if err != nil {
			return nil, err
		}
		blade.Memory, err = i.Memory()
		if err != nil {
			return nil, err
		}
		blade.Status, err = i.Status()
		if err != nil {
			return nil, err
		}
		blade.Name, err = i.Name()
		if err != nil {
			return nil, err
		}
		blade.TempC, err = i.TempC()
		if err != nil {
			return nil, err
		}
		blade.PowerKw, err = i.PowerKw()
		if err != nil {
			return nil, err
		}
		blade.PowerState, err = i.PowerState()
		if err != nil {
			return nil, err
		}
		blade.BmcLicenceType, blade.BmcLicenceStatus, err = i.License()
		if err != nil {
			return nil, err
		}
		blade.BladePosition, err = i.Slot()
		if err != nil {
			return nil, err
		}
		blade.ChassisSerial, err = i.ChassisSerial()
		if err != nil {
			return nil, err
		}
		server = blade
	} else {
		discrete := &devices.Discrete{}
		discrete.Vendor = i.Vendor()
		discrete.BmcAddress = i.ip
		discrete.BmcType = i.HardwareType()

		discrete.Serial, err = i.Serial()
		if err != nil {
			return nil, err
		}
		discrete.BmcVersion, err = i.Version()
		if err != nil {
			return nil, err
		}
		discrete.Model, err = i.Model()
		if err != nil {
			return nil, err
		}
		discrete.Nics, err = i.Nics()
		if err != nil {
			return nil, err
		}
		discrete.Disks, err = i.Disks()
		if err != nil {
			return nil, err
		}
		discrete.BiosVersion, err = i.BiosVersion()
		if err != nil {
			return nil, err
		}
		discrete.Processor, discrete.ProcessorCount, discrete.ProcessorCoreCount, discrete.ProcessorThreadCount, err = i.CPU()
		if err != nil {
			return nil, err
		}
		discrete.Memory, err = i.Memory()
		if err != nil {
			return nil, err
		}
		discrete.Status, err = i.Status()
		if err != nil {
			return nil, err
		}
		discrete.Name, err = i.Name()
		if err != nil {
			return nil, err
		}
		discrete.TempC, err = i.TempC()
		if err != nil {
			return nil, err
		}
		discrete.PowerKw, err = i.PowerKw()
		if err != nil {
			return nil, err
		}
		discrete.PowerState, err = i.PowerState()
		if err != nil {
			return nil, err
		}
		discrete.BmcLicenceType, discrete.BmcLicenceStatus, err = i.License()
		if err != nil {
			return nil, err
		}
		discrete.Psus, err = i.Psus()
		if err != nil {
			return nil, err
		}
		server = discrete
	}

	return server, err
}

// Disks returns a list of disks installed on the device
func (i *IDrac9) Disks() (disks []*devices.Disk, err error) {
	err = i.loadHwData()
	if err != nil {
		return disks, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_PhysicalDiskView" {
			if disks == nil {
				disks = make([]*devices.Disk, 0)
			}
			disk := &devices.Disk{}

			for _, property := range component.Properties {
				if property.Name == "Model" {
					disk.Model = strings.ToLower(property.Value)
				} else if property.Name == "SerialNumber" {
					disk.Serial = strings.ToLower(property.Value)
				} else if property.Name == "MediaType" {
					if property.DisplayValue == "Solid State Drive" {
						disk.Type = "SSD"
					} else if property.DisplayValue == "Hard Disk Drive" {
						disk.Type = "HDD"
					} else {
						disk.Type = property.DisplayValue
					}
				} else if property.Name == "PrimaryStatus" {
					disk.Status = property.DisplayValue
				} else if property.Name == "DeviceDescription" {
					disk.Location = property.DisplayValue
				} else if property.Name == "SizeInBytes" {
					size, err := strconv.Atoi(property.Value)
					if err != nil {
						return disks, err
					}
					disk.Size = fmt.Sprintf("%d GB", size/1024/1024/1024)
				} else if property.Name == "Revision" {
					disk.FwVersion = strings.ToLower(property.Value)
				}
			}

			if disk.Serial != "" {
				disks = append(disks, disk)
			}
		}
	}
	return disks, err
}

// UpdateCredentials updates login credentials
func (i *IDrac9) UpdateCredentials(username string, password string) {
	i.username = username
	i.password = password
}

// BiosVersion returns the BIOS version from the BMC, implements the Firmware interface
func (i *IDrac9) GetBIOSVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// BMCVersion returns the BMC version, implements the Firmware interface
func (i *IDrac9) GetBMCVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// Updates the BMC firmware, implements the Firmware interface
func (i *IDrac9) FirmwareUpdateBMC(ctx context.Context, filePath string) error {
	return errors.ErrNotImplemented
}
