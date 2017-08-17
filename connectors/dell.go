package connectors

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/publicsuffix"
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

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
		Jar:       jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return ErrPageNotFound
	}

	iDracAuth := &IDracAuth{}
	err = xml.Unmarshal(payload, iDracAuth)
	if err != nil {
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

	result, err := i.get("sysmgmt/2012/server/memory", extraHeaders)
	if err != nil {
		return mem, err
	}

	dellBladeMemory := &DellBladeMemoryEndpoint{}
	err = json.Unmarshal(result, dellBladeMemory)
	if err != nil {
		return mem, err
	}

	return dellBladeMemory.Memory.Capacity / 1024, err
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDracReader) CPU() (cpu string, coreCount int, hyperthreadCount int, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	result, err := i.get("sysmgmt/2012/server/processor", extraHeaders)
	if err != nil {
		return cpu, coreCount, hyperthreadCount, err
	}

	dellBladeProc := &DellBladeProcessorEndpoint{}
	err = json.Unmarshal(result, dellBladeProc)
	if err != nil {
		return cpu, coreCount, hyperthreadCount, err
	}

	for _, proc := range dellBladeProc.Proccessors {
		hasHT := 0
		for _, ht := range proc.HyperThreading {
			if ht.Capable == 1 {
				hasHT = 2
			}
		}
		return fmt.Sprintf("%d x %s", len(dellBladeProc.Proccessors), strings.TrimSpace(proc.Brand)), proc.CoreCount, proc.CoreCount * hasHT, err
	}

	return cpu, coreCount, hyperthreadCount, err
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
