package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/storage"
)

var (
	macFinder = regexp.MustCompile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
	findBmcIP = regexp.MustCompile("bladeIpAddress\">((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})")
)

// DellCmcReader holds the status and properties of a connection to a CMC device
type DellCmcReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	cmcJSON  *DellCMC
	cmcTemp  *DellCMCTemp
	cmcWWN   *DellCMCWWN
}

// NewDellCmcReader returns a connection to DellCmcReader
func NewDellCmcReader(ip *string, username *string, password *string) (chassis *DellCmcReader, err error) {
	return &DellCmcReader{ip: ip, username: username, password: password}, err
}

// Login initiates the connection to a chassis device
func (d *DellCmcReader) Login() (err error) {
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": Dell, "ip": *d.ip}).Debug("connecting to chassis")

	form := url.Values{}
	form.Add("user", *d.username)
	form.Add("password", *d.password)

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", *d.ip))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	d.client, err = buildClient()
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	auth, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(auth), "Try Again") {
		return ErrLoginFailed
	}

	if resp.StatusCode == 404 {
		return ErrPageNotFound
	}

	err = d.loadHwData()
	if err != nil {
		return err
	}

	return err
}

func (d *DellCmcReader) loadHwData() (err error) {
	payload, err := d.get("json?method=groupinfo")
	if err != nil {
		return err
	}

	d.cmcJSON = &DellCMC{}
	err = json.Unmarshal(payload, d.cmcJSON)
	if err != nil {
		DumpInvalidPayload(*d.ip, payload)
		return err
	}

	if d.cmcJSON.DellChassis == nil {
		return ErrUnableToReadData
	}

	payload, err = d.get("json?method=blades-wwn-info")
	if err != nil {
		return err
	}

	d.cmcWWN = &DellCMCWWN{}
	err = json.Unmarshal(payload, d.cmcWWN)
	if err != nil {
		DumpInvalidPayload(*d.ip, payload)
		return err
	}

	return err
}

// Logout logs out and close the chassis connection
func (d *DellCmcReader) Logout() (err error) {
	_, err = d.get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", *d.ip))
	return err
}

func (d *DellCmcReader) get(endpoint string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": Dell, "ip": *d.ip, "endpoint": endpoint}).Debug("retrieving data from chassis")

	resp, err := d.client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", *d.ip, endpoint))
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

	// Dell has a really shitty consistency of the data type returned, here we fix what's possible
	payload = bytes.Replace(payload, []byte("\"bladeTemperature\":-1"), []byte("\"bladeTemperature\":\"0\""), -1)
	payload = bytes.Replace(payload, []byte("\"nic\": [],"), []byte("\"nic\": {},"), -1)
	payload = bytes.Replace(payload, []byte("N\\/A"), []byte("0"), -1)

	return payload, err
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
	payload, err := d.get("json?method=temp-sensors")
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

// FwVersion returns the current firmware version of the bmc
func (d *DellCmcReader) FwVersion() (version string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString, err
}

// Nics returns all found Nics in the device
func (d *DellCmcReader) Nics() (nics []*model.Nic, err error) {
	payload, err := d.get("cmc_status?cat=C01&tab=T11&id=P31")
	if err != nil {
		return nics, err
	}

	serial, _ := d.Serial()

	mac := macFinder.FindString(string(payload))
	if mac != "" {
		nics = make([]*model.Nic, 0)
		n := &model.Nic{
			Name:          "OA1",
			MacAddress:    strings.ToLower(mac),
			ChassisSerial: serial,
		}
		nics = append(nics, n)
	}

	return nics, err
}

// IsActive returns health string status from the bmc
func (d *DellCmcReader) IsActive() bool {
	return true
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

// Psus returns a list of psus installed on the device
func (d *DellCmcReader) Psus() (psus []*model.Psu, err error) {
	serial, _ := d.Serial()
	for _, psu := range d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.Psus {
		if psu.PsuPresent == 0 {
			continue
		}

		i, err := strconv.ParseFloat(strings.TrimSuffix(psu.PsuAcCurrent, " A"), 64)
		if err != nil {
			return psus, err
		}

		e, err := strconv.ParseFloat(psu.PsuAcVolts, 64)
		if err != nil {
			return psus, err
		}

		var status string
		if psu.PsuActiveError == "No Errors" {
			status = "OK"
		} else {
			status = psu.PsuActiveError
		}

		p := &model.Psu{
			Serial:        fmt.Sprintf("%s_%s", serial, psu.PsuPosition),
			CapacityKw:    float64(psu.PsuCapacity) / 1000.00,
			PowerKw:       (i * e) / 1000.00,
			Status:        status,
			ChassisSerial: serial,
		}

		psus = append(psus, p)
	}

	return psus, err
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

// Blades returns all StorageBlades found in this chassis
func (d *DellCmcReader) Blades() (blades []*model.Blade, err error) {
	db := storage.InitDB()
	for _, dellBlade := range d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 0 {
			blade := model.Blade{}

			blade.BladePosition = dellBlade.BladeMasterSlot
			blade.Serial = strings.ToLower(dellBlade.BladeSvcTag)
			chassisSerial, _ := d.Serial()

			if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial}).Error("Review this blade. The chassis identifies it as connected, but we have no data")
				continue
			}

			blade.Model = dellBlade.BladeModel
			blade.PowerKw = float64(dellBlade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(dellBlade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial}).Warning(err)
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

			if bmcData, ok := d.cmcWWN.SlotMacWwn.SlotMacWwnList[blade.BladePosition]; ok {
				n := &model.Nic{
					Name:       "bmc",
					MacAddress: strings.ToLower(bmcData.IsNotDoubleHeight.PortFMAC),
				}
				blade.Nics = append(blade.Nics, n)
			}

			if strings.Count(blade.BmcAddress, ".") != 4 {
				payload, err := d.get(fmt.Sprintf("blade_status?id=%d&cat=C10&tab=T41&id=P78", blade.BladePosition))
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial}).Warning(err)
				} else {
					ip := findBmcIP.FindStringSubmatch(string(payload))
					if len(ip) > 0 {
						blade.BmcAddress = ip[1]
					}
				}
			}

			for _, nic := range dellBlade.Nics {
				if nic.BladeNicName == "" {
					log.WithFields(log.Fields{"operation": "connection", "ip": *d.ip, "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial}).Error("Network card information missing, please verify")
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
					} else if scan.Port == 623 && scan.Protocol == "ipmi" && scan.State == "open" {
						blade.BmcIpmiReachable = true
					}
				}

				if blade.BmcWEBReachable {
					bmcUser := viper.GetString("bmc_user")
					bmcPass := viper.GetString("bmc_pass")
					idrac, err := NewIDrac8Reader(&blade.BmcAddress, &bmcUser, &bmcPass)
					if err != nil {
						log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
					} else {
						err = idrac.Login()
						if err == nil {
							defer idrac.Logout()
							blade.BmcAuth = true

							blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = idrac.CPU()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis"}).Warning(err)
							}

							blade.Memory, err = idrac.Memory()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
							}

							blade.BmcLicenceType, blade.BmcLicenceStatus, err = idrac.License()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
							}
						}
					}
				} else {
					log.WithFields(log.Fields{"operation": "create idrac connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Error("Not reachable")
				}
			}
			blades = append(blades, &blade)
		}
	}
	return blades, err
}
