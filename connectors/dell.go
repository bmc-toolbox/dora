package connectors

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
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

var (
	// ErrLoginFailed is returned when we fail to login to a bmc
	ErrLoginFailed = errors.New("Failed to login")
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

type DellBladeMemory struct {
	Capacity       int `json:"capacity"`
	ErrCorrection  int `json:"err_correction"`
	MaxCapacity    int `json:"max_capacity"`
	SlotsAvailable int `json:"slots_available"`
	SlotsUsed      int `json:"slots_used"`
}

type IDracAuth struct {
	Status     string `xml:"status"`
	AuthResult int    `xml:"authResult"`
	ForwardUrl string `xml:"forwardUrl"`
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

	if resp.StatusCode == 404 {
		return ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	iDracAuth := &IDracAuth{}
	err = xml.Unmarshal(payload, iDracAuth)
	if err != nil {
		return err
	}

	if iDracAuth.AuthResult == 1 {
		return ErrLoginFailed
	}

	stTemp := strings.Split(iDracAuth.ForwardUrl, ",")
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
	//req.Header.Add("Referer", fmt.Sprintf("https://%s/index.htm", *i.ip))
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
