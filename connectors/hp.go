package connectors

import (
	"bytes"
	"encoding/json"
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
	MemTotalMemSize int `json:"mem_total_mem_size"`
}

// IloReader holds the status and properties of a connection to an iLO device
type IloReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	loginURL *url.URL
}

// NewIloReader returns a new IloReader ready to be used
func NewIloReader(ip *string, username *string, password *string) (ilo *IloReader, err error) {
	u, err := url.Parse(fmt.Sprintf("https://%s/json/login_session", *ip))
	if err != nil {
		return nil, err
	}
	return &IloReader{ip: ip, username: username, password: password, loginURL: u}, err
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

	if strings.Contains(string(payload), "Invalid login attempt") {
		return ErrLoginFailed
	}

	i.client = client

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

// Memory return the total amount of memory of the server
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

	return hpMemData.MemTotalMemSize / 1024, err
}

// CPU return the cpu, cores and hyperthreads the server
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

// BiosVersion return the current verion of the bios
func (i *IloReader) BiosVersion() (version string, err error) {
	payload, err := i.get("json/fw_info")
	if err != nil {
		return version, err
	}

	hpFwData := &HpFirmware{}
	err = json.Unmarshal(payload, hpFwData)
	if err != nil {
		DumpInvalidPayload(*i.ip, payload)
		return version, err
	}

	for _, entry := range hpFwData.Firmware {
		if entry.FwName == "System ROM" {
			return entry.FwVersion, err
		}
	}

	return version, ErrBiosNotFound
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
