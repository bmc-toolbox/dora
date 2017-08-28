package connectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"

	"gitlab.booking.com/infra/dora/model"

	"strings"

	log "github.com/sirupsen/logrus"
)

// ChassisConnection is the basic
type ChassisConnection struct {
	username string
	password string
}

func (c *ChassisConnection) Dell(ip *string) (chassis model.Chassis, err error) {
	payload, err := httpGetDell(ip, "json?method=groupinfo", &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	chassis.BmcAuth = true
	dellCMC := &DellCMC{}
	err = json.Unmarshal(payload, dellCMC)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	chassis.Name = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.CHASSISName
	chassis.Serial = strings.ToLower(dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisServiceTag)
	chassis.Model = strings.TrimSpace(dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisProductname)
	chassis.FwVersion = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString
	chassis.PowerSupplyCount = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.PsuCount
	chassis.BmcAddress = *ip
	if dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError == "No Errors" {
		chassis.Status = "OK"
	} else {
		chassis.Status = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError
	}

	power, err := strconv.Atoi(strings.TrimRight(dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.AcPower, " W"))
	if err != nil {
		log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis", "error": err}).Error("Auditing chassis")
		return
	}
	chassis.PowerKw = float64(power) / 1000
	chassis.Vendor = Dell

	log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis"}).Debug("Auditing chassis")

	for _, blade := range dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if blade.BladePresent == 1 {
			b := model.Blade{}

			b.BladePosition = blade.BladeMasterSlot
			b.Serial = strings.ToLower(blade.BladeSvcTag)
			if b.Serial == "" {
				log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "position": b.BladePosition, "type": "chassis", "error": "Review this blade. The chassis identifies it as connected, but we have no data"}).Error("Auditing blade")
				continue
			}

			b.Model = blade.BladeModel
			b.PowerKw = float64(blade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(blade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis", "error": err, "blade": blade.BladeSvcTag}).Error("Auditing blade")
				continue
			}
			b.TempC = temp
			if blade.BladeLogDescription == "No Errors" {
				b.Status = "OK"
			} else {
				chassis.Status = blade.BladeLogDescription
			}
			b.Vendor = Dell
			b.BmcType = "iDRAC"
			b.BiosVersion = blade.BladeBIOSver

			if chassis.PassThru == "" {
				for _, nic := range blade.Nics {
					if strings.Contains(nic.BladeNicName, "10G") {
						chassis.PassThru = "10G"
					} else {
						chassis.PassThru = "1G"
					}
					break
				}
			}

			if blade.IsStorageBlade == 1 {
				b.IsStorageBlade = true
				b.Name = b.Serial
			} else {
				b.IsStorageBlade = false
				b.Name = blade.BladeName
				idracURL := strings.TrimLeft(blade.IdracURL, "https://")
				idracURL = strings.TrimLeft(idracURL, "http://")
				idracURL = strings.Split(idracURL, ":")[0]
				b.BmcAddress = idracURL
				b.BmcVersion = blade.BladeUSCVer

				for _, nic := range blade.Nics {
					if nic.BladeNicName == "" {
						log.WithFields(log.Fields{"operation": "connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": "Network card information missing, please verify"}).Error("Auditing blade")
						continue
					}

					n := &model.Nic{
						MacAddress: strings.ToLower(nic.BladeNicName[len(nic.BladeNicName)-17:]),
					}
					b.Nics = append(b.Nics, n)
				}

			}

			b.TestConnections()
			if b.BmcWEBReachable {
				redFish, err := NewRedFishReader(&b.BmcAddress, &c.username, &c.password)
				//
				if err != nil {
					log.WithFields(log.Fields{"operation": "opening RedFish connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Info("Auditing blade")
					if err != ErrLoginFailed {
						iDrac := NewIDracReader(&b.BmcAddress, &c.username, &c.password)
						err := iDrac.Login()
						if err != nil {
							log.WithFields(log.Fields{"operation": "opening iDrac connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
						} else {
							defer iDrac.Logout()
							b.BmcAuth = true

							b.Memory, err = iDrac.Memory()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading memory data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}

							b.Processor, b.ProcessorCount, b.ProcessorCoreCount, b.ProcessorThreadCount, err = iDrac.CPU()
							if err != nil {
								log.WithFields(log.Fields{"operation": "reading cpu data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
							}
						}
					}
				} else {
					b.BmcAuth = true

					b.Memory, err = redFish.Memory()
					if err != nil {
						log.WithFields(log.Fields{"operation": "reading memory data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					}

					b.Processor, b.ProcessorCount, b.ProcessorCoreCount, b.ProcessorThreadCount, err = redFish.CPU()
					if err != nil {
						log.WithFields(log.Fields{"operation": "reading cpu data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					}
				}
			}

			chassis.Blades = append(chassis.Blades, &b)
		}
	}

	payload, err = httpGetDell(ip, "json?method=temp-sensors", &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	dellCMCTemp := &DellCMCTemp{}
	err = json.Unmarshal(payload, dellCMCTemp)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	if dellCMCTemp.DellChassisTemp != nil {
		chassis.TempC = dellCMCTemp.DellChassisTemp.TempCurrentValue
	}
	chassis.TestConnections()

	return chassis, err
}

func (c *ChassisConnection) Hp(ip *string) (chassis model.Chassis, err error) {
	payload, err := httpGet(fmt.Sprintf("https://%s/xmldata?item=all", *ip), &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	iloXML := &HpRimp{}
	err = xml.Unmarshal(payload, iloXML)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	if iloXML.HpInfra2 != nil {
		chassis.Name = iloXML.HpInfra2.Encl
		chassis.Serial = strings.ToLower(iloXML.HpInfra2.EnclSn)

		chassis.Model = iloXML.HpInfra2.Pn
		chassis.PowerKw = iloXML.HpInfra2.HpChassisPower.PowerConsumed / 1000.00
		chassis.TempC = iloXML.HpInfra2.HpTemps.HpTemp.C
		chassis.Status = iloXML.HpInfra2.Status
		chassis.Vendor = HP
		chassis.FwVersion = iloXML.HpMP.Fwri
		chassis.PowerSupplyCount = len(iloXML.HpInfra2.HpChassisPower.HpPowersupply)
		chassis.BmcAddress = *ip

		for _, hpswitch := range iloXML.HpInfra2.HpSwitches.HpSwitch {
			if strings.Contains(hpswitch.Spn, "10G") {
				chassis.PassThru = "10G"
			} else {
				chassis.PassThru = "1G"
			}
			break
		}

		log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis"}).Debug("Auditing chassis")

		if iloXML.HpInfra2.HpBlades != nil {
			for _, blade := range iloXML.HpInfra2.HpBlades.HpBlade {
				b := model.Blade{}

				b.BladePosition = blade.HpBay.Connection
				b.PowerKw = blade.HpPower.PowerConsumed / 1000.00
				b.TempC = blade.HpTemps.HpTemp.C
				b.Serial = strings.ToLower(strings.TrimSpace(blade.Bsn))
				if b.Serial == "" || b.Serial == "[unknown]" {
					log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "position": b.BladePosition, "type": "chassis", "error": "Review this blade. The chassis identifies it as connected, but we have no data"}).Error("Auditing blade")
					continue
				}
				b.Status = blade.Status
				b.Vendor = HP
				b.BmcType = blade.MgmtType

				if strings.Contains(blade.Spn, "Storage") {
					b.Name = b.Serial
					b.IsStorageBlade = true
				} else {
					b.Name = blade.Name
					b.IsStorageBlade = false
					b.BmcAddress = blade.MgmtIPAddr
					b.BmcVersion = blade.MgmtVersion
					b.Model = blade.Spn

					payload, err := httpGet(fmt.Sprintf("https://%s/xmldata?item=all", b.BmcAddress), &c.username, &c.password)
					if err != nil {
						log.WithFields(log.Fields{"operation": "connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Error("Auditing blade")
					} else {
						bladeIloXML := &HpRimpBlade{}
						err = xml.Unmarshal(payload, bladeIloXML)
						if err != nil {
							DumpInvalidPayload(*ip, payload)
							log.WithFields(log.Fields{"operation": "connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Error("Auditing blade")
						} else if bladeIloXML.HpHSI != nil && bladeIloXML.HpHSI.HpNICS != nil && bladeIloXML.HpHSI.HpNICS.HpNIC != nil {
							for _, nic := range bladeIloXML.HpHSI.HpNICS.HpNIC {
								if strings.Contains("iLo", nic.Description) {
									continue
								}
								n := &model.Nic{
									MacAddress: strings.ToLower(nic.MacAddr),
								}
								b.Nics = append(b.Nics, n)
							}
						}
					}
				}

				b.TestConnections()
				if b.BmcWEBReachable {
					ilo, err := NewIloReader(&b.BmcAddress, &c.username, &c.password)
					if err != nil {
						log.WithFields(log.Fields{"operation": "create ilo connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					}

					err = ilo.Login()
					if err != nil {
						log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
					} else {
						defer ilo.Logout()
						b.BmcAuth = true

						b.BiosVersion, err = ilo.BiosVersion()
						if err != nil {
							log.WithFields(log.Fields{"operation": "reading bios version", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
						}

						b.Processor, b.ProcessorCount, b.ProcessorCoreCount, b.ProcessorThreadCount, err = ilo.CPU()
						if err != nil {
							log.WithFields(log.Fields{"operation": "reading cpu data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
						}

						b.Memory, err = ilo.Memory()
						if err != nil {
							log.WithFields(log.Fields{"operation": "reading memory data", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
						}
					}
				}

				chassis.Blades = append(chassis.Blades, &b)
			}
		}
	}
	chassis.TestConnections()

	return chassis, err
}

func NewChassisConnection(username string, password string) *ChassisConnection {
	return &ChassisConnection{username: username, password: password}
}
