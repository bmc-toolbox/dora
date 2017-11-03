package connectors

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

var (
	macFinder = regexp.MustCompile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
)

// DellCmcReader holds the status and properties of a connection to a CMC device
type DellCmcReader struct {
	ip       *string
	username *string
	password *string
	cmcJSON  *DellCMC
	cmcTemp  *DellCMCTemp
	cmcWWN   *DellCMCWWN
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
		return chassis, ErrUnableToReadData
	}

	payload, err = httpGetDell(ip, "json?method=blades-wwn-info", username, password)
	if err != nil {
		return chassis, err
	}

	dellCMCWWN := &DellCMCWWN{}
	err = json.Unmarshal(payload, dellCMCWWN)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	return &DellCmcReader{ip: ip, username: username, password: password, cmcJSON: dellCMC, cmcWWN: dellCMCWWN}, err
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

// Status returns health string status from the bmc
func (d *DellCmcReader) Status() (status string, err error) {
	if d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError == "No Errors" {
		status = "OK"
	} else {
		status = d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError
	}
	return status, err
}

// PowerSupplyCount returns the total count of the power supply
func (d *DellCmcReader) PowerSupplyCount() (count int64, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.PsuCount, err
}

// FwVersion returns the current firmware version of the bmc
func (d *DellCmcReader) FwVersion() (version string, err error) {
	return d.cmcJSON.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString, err
}

// Nics returns all found Nics in the device
func (d *DellCmcReader) Nics() (nics []*model.Nic, err error) {
	payload, err := httpGetDell(d.ip, "cmc_status?cat=C01&tab=T11&id=P31", d.username, d.password)
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

			if bmcData, ok := d.cmcWWN.SlotMacWwn.SlotMacWwnList[blade.BladePosition]; ok {
				n := &model.Nic{
					Name:       "bmc",
					MacAddress: strings.ToLower(bmcData.IsNotDoubleHeight.PortFMAC),
				}
				blade.Nics = append(blade.Nics, n)
			}

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
				db.Where("ip = ?", blade.BmcAddress).Find(&scans)
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
					idrac, err := NewIDracReader(&blade.BmcAddress, d.username, d.password)
					if err != nil {
						log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					} else {
						err = idrac.Login()
						if err == nil {
							defer idrac.Logout()
							blade.BmcAuth = true

							blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = idrac.CPU()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							blade.Memory, err = idrac.Memory()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							blade.BmcLicenceType, blade.BmcLicenceStatus, err = idrac.License()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}
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
