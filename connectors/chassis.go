package connectors

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	"gitlab.booking.com/infra/dora/model"

	"strings"

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
)

var (
	bladeDevice        = "blade"
	chassisDevice      = "chassis"
	storageBladeDevice = "storageblade"
	// ErrPageNotFound is used to inform the http request that we couldn't find the expected page and/or endpoint
	ErrPageNotFound = errors.New("Requested page couldn't be found in the server")
)

// ChassisConnection is the basic
type ChassisConnection struct {
	username string
	password string
}

func (c *ChassisConnection) Dell(ip *string) (chassis model.Chassis, err error) {
	result, err := httpGetDell(ip, &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	dellCMC := &DellCMC{}
	err = json.Unmarshal(result, dellCMC)
	if err != nil {
		return chassis, err
	}

	chassis.Name = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.CHASSISName
	chassis.Serial = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisServiceTag
	chassis.Model = strings.TrimSpace(dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROChassisProductname)
	chassis.FwVersion = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellChassisStatus.ROCmcFwVersionString
	if dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError == "No Errors" {
		chassis.Status = "OK"
	} else {
		chassis.Status = dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellCMCStatus.CMCActiveError
	}

	power, err := strconv.Atoi(strings.TrimRight(dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellPsuStatus.AcPower, " W"))
	if err != nil {
		log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.ID, "type": "chassis", "error": err}).Error("Auditing chassis")
		return
	}
	chassis.Power = float64(power) / 1000
	chassis.Vendor = Dell
	// chassis.Temp = ?

	log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.ID, "type": "chassis"}).Debug("Auditing chassis")

	for _, blade := range dellCMC.DellChassis.DellChassisGroupMemberHealthBlob.DellBlades {
		if blade.BladePresent == 1 {
			b := model.Blade{}

			b.BladePosition = blade.BladeMasterSlot
			b.Power = float64(blade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(blade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis", "error": err, "blade": blade.BladeSvcTag}).Error("Auditing blade")
				continue
			}
			b.Temp = temp
			b.Serial = blade.BladeSvcTag
			b.Status = blade.BladeLogDescription
			b.Vendor = Dell

			if blade.IsStorageBlade == 1 {
				b.IsStorageBlade = true
				b.Name = blade.BladeSvcTag
			} else {
				b.IsStorageBlade = false
				b.Name = blade.BladeName
				b.BmcAddress = blade.IdracURL
				b.BmcVersion = blade.BladeUSCVer
			}
			chassis.Blades = append(chassis.Blades, &b)
		}
	}

	return chassis, err
}

func (c *ChassisConnection) Hp(ip *string) (chassis model.Chassis, err error) {
	result, err := httpGet(fmt.Sprintf("https://%s/xmldata?item=all", *ip), &c.username, &c.password)
	if err != nil {
		return chassis, err
	}
	iloXML := &HpRimp{}
	err = xml.Unmarshal(result, iloXML)
	if err != nil {
		return chassis, err
	}

	if iloXML.HpInfra2 != nil {
		chassis.Name = iloXML.HpInfra2.Encl
		chassis.Serial = iloXML.HpInfra2.EnclSn
		chassis.Model = iloXML.HpInfra2.Pn
		chassis.Rack = iloXML.HpInfra2.Rack
		chassis.Power = iloXML.HpInfra2.HpPower.PowerConsumed / 1000.00
		chassis.Temp = iloXML.HpInfra2.HpTemps.HpTemp.C
		chassis.Status = iloXML.HpInfra2.Status
		chassis.Vendor = HP
		chassis.FwVersion = iloXML.HpMP.Fwri

		log.WithFields(log.Fields{"operation": "connection", "ip": *ip, "name": chassis.Name, "serial": chassis.Serial, "type": "chassis"}).Debug("Auditing chassis")

		if iloXML.HpInfra2.HpBlades != nil {
			for _, blade := range iloXML.HpInfra2.HpBlades.HpBlade {
				b := model.Blade{}

				b.BladePosition = blade.HpBay.Connection
				b.Power = blade.HpPower.PowerConsumed / 1000.00
				b.Temp = blade.HpTemps.HpTemp.C
				b.Serial = strings.TrimSpace(blade.Bsn)
				b.Status = blade.Status
				b.Vendor = HP

				if strings.Contains(blade.Spn, "Storage") {
					b.Name = b.Serial
					b.IsStorageBlade = true
				} else {
					b.Name = blade.Name
					b.IsStorageBlade = false
					b.BmcAddress = blade.MgmtIPAddr
					b.BmcVersion = blade.MgmtVersion
				}
				chassis.Blades = append(chassis.Blades, &b)
			}
		}
	}

	return chassis, err
}

func NewChassisConnection(username string, password string) *ChassisConnection {
	return &ChassisConnection{username: username, password: password}
}
