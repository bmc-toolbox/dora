package connectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/storage"
)

// HpChassisReader holds the status and properties of a connection to a BladeSystem device
type HpChassisReader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
	hpRimp   *HpRimp
}

// NewHpChassisReader returns a connection to HpChassisReader
func NewHpChassisReader(ip *string, username *string, password *string) (chassis *HpChassisReader, err error) {
	client, err := buildClient()
	if err != nil {
		return chassis, err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", *ip))
	if err != nil {
		return chassis, err
	}
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return chassis, err
	}
	defer resp.Body.Close()

	hpRimp := &HpRimp{}
	err = xml.Unmarshal(payload, hpRimp)
	if err != nil {
		DumpInvalidPayload(*ip, payload)
		return chassis, err
	}

	if hpRimp.HpInfra2 == nil {
		return chassis, ErrUnableToReadData
	}

	return &HpChassisReader{ip: ip, username: username, password: password, hpRimp: hpRimp, client: client}, err
}

// Login initiates the connection to a chassis device
func (h *HpChassisReader) Login() (err error) {
	return err
}

// Logout logs out and close the chassis connection
func (h *HpChassisReader) Logout() (err error) {
	return err
}

// Name returns the hostname of the machine
func (h *HpChassisReader) Name() (name string, err error) {
	return h.hpRimp.HpInfra2.Encl, err
}

// Model returns the device model
func (h *HpChassisReader) Model() (model string, err error) {
	return h.hpRimp.HpMP.Pn, err
}

// Serial returns the device serial
func (h *HpChassisReader) Serial() (serial string, err error) {
	return strings.ToLower(strings.TrimSpace(h.hpRimp.HpInfra2.EnclSn)), err
}

// PowerKw returns the current power usage in Kw
func (h *HpChassisReader) PowerKw() (power float64, err error) {
	return h.hpRimp.HpInfra2.HpChassisPower.PowerConsumed / 1000.00, err
}

// TempC returns the current temperature of the machine
func (h *HpChassisReader) TempC() (temp int, err error) {
	return h.hpRimp.HpInfra2.HpTemp.C, err
}

// Psus returns a list of psus installed on the device
func (h *HpChassisReader) Psus() (psus []*model.Psu, err error) {
	serial, _ := h.Serial()

	for _, psu := range h.hpRimp.HpInfra2.HpChassisPower.HpPowersupply {
		if psus == nil {
			psus = make([]*model.Psu, 0)
		}

		p := &model.Psu{
			Serial:        strings.ToLower(psu.Sn),
			Status:        psu.Status,
			PowerKw:       psu.ActualOutput / 1000.00,
			CapacityKw:    psu.Capacity / 1000.00,
			ChassisSerial: serial,
		}
		psus = append(psus, p)
	}

	return psus, err
}

// Nics returns all found Nics in the device
func (h *HpChassisReader) Nics() (nics []*model.Nic, err error) {
	serial, _ := h.Serial()

	for _, manager := range h.hpRimp.HpInfra2.HpManagers {
		if nics == nil {
			nics = make([]*model.Nic, 0)
		}

		n := &model.Nic{
			Name:          manager.Name,
			MacAddress:    strings.ToLower(manager.MacAddr),
			ChassisSerial: serial,
		}
		nics = append(nics, n)
	}

	return nics, err
}

// Status returns health string status from the bmc
func (h *HpChassisReader) Status() (status string, err error) {
	return h.hpRimp.HpInfra2.Status, err
}

// IsActive returns health string status from the bmc
func (h *HpChassisReader) IsActive() bool {
	for _, manager := range h.hpRimp.HpInfra2.HpManagers {
		if manager.MgmtIPAddr == strings.Split(*h.ip, ":")[0] && manager.Role == "ACTIVE" {
			return true
		}
	}
	return false
}

// FwVersion returns the current firmware version of the bmc
func (h *HpChassisReader) FwVersion() (version string, err error) {
	return h.hpRimp.HpMP.Fwri, err
}

// PassThru returns the type of switch we have for this chassis
func (h *HpChassisReader) PassThru() (passthru string, err error) {
	passthru = "1G"
	for _, hpswitch := range h.hpRimp.HpInfra2.HpSwitches {
		if strings.Contains(hpswitch.Spn, "10G") {
			passthru = "10G"
		}
		break
	}
	return passthru, err
}

// StorageBlades returns all StorageBlades found in this chassis
func (h *HpChassisReader) StorageBlades() (storageBlades []*model.StorageBlade, err error) {
	if h.hpRimp.HpInfra2.HpBlades != nil {
		chassisSerial, _ := h.Serial()
		db := storage.InitDB()
		for _, hpBlade := range h.hpRimp.HpInfra2.HpBlades {
			if hpBlade.Type == "STORAGE" {
				storageBlade := model.StorageBlade{}
				storageBlade.Serial = strings.ToLower(strings.TrimSpace(hpBlade.Bsn))

				if storageBlade.Serial == "" || storageBlade.Serial == "[unknown]" || storageBlade.Serial == "0000000000" {
					log.WithFields(log.Fields{"operation": "connection", "ip": *h.ip, "position": storageBlade.BladePosition, "type": "chassis", "chassis_serial": chassisSerial, "error": ErrInvalidSerial}).Error("Auditing blade")
					continue
				}
				storageBlade.BladePosition = hpBlade.HpBay.Connection
				storageBlade.Status = hpBlade.Status
				storageBlade.PowerKw = hpBlade.HpPower.PowerConsumed / 1000.00
				storageBlade.TempC = hpBlade.HpTemp.C
				storageBlade.Vendor = HP
				storageBlade.FwVersion = hpBlade.BladeRomVer
				storageBlade.Model = hpBlade.Spn
				storageBlade.ChassisSerial = chassisSerial

				blade := model.Blade{}
				db.Where("chassis_serial = ? and blade_position = ?", chassisSerial, hpBlade.AssociatedBlade).First(&blade)
				if blade.Serial != "" {
					storageBlade.BladeSerial = blade.Serial
				}
				storageBlades = append(storageBlades, &storageBlade)
			}
		}
	}
	return storageBlades, err
}

// Blades returns all StorageBlades found in this chassis
func (h *HpChassisReader) Blades() (blades []*model.Blade, err error) {
	name, _ := h.Name()
	if h.hpRimp.HpInfra2.HpBlades != nil {
		chassisSerial, _ := h.Serial()
		db := storage.InitDB()
		for _, hpBlade := range h.hpRimp.HpInfra2.HpBlades {
			if hpBlade.Type == "SERVER" {
				blade := model.Blade{}
				blade.BladePosition = hpBlade.HpBay.Connection
				blade.Status = hpBlade.Status
				blade.Serial = strings.ToLower(strings.TrimSpace(hpBlade.Bsn))
				blade.ChassisSerial = chassisSerial

				if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" {
					nb := model.Blade{}
					db.Where("bmc_address = ? and blade_position = ?", hpBlade.MgmtIPAddr, hpBlade.HpBay.Connection).First(&nb)
					log.WithFields(log.Fields{"operation": "connection", "ip": *h.ip, "name": name, "position": blade.BladePosition, "type": "chassis"}).Error("Review this blade. The chassis identifies it as connected, but we have no data")

					if nb.Serial == "" {
						continue
					}

					blade.Status = "Require Reseat"
					blade.Serial = nb.Serial
				}

				blade.PowerKw = hpBlade.HpPower.PowerConsumed / 1000.00
				blade.TempC = hpBlade.HpTemp.C
				blade.Vendor = HP
				blade.Model = hpBlade.Spn
				blade.Name = hpBlade.Name
				blade.BmcAddress = hpBlade.MgmtIPAddr
				blade.BmcVersion = hpBlade.MgmtVersion
				blade.BmcType = hpBlade.MgmtType
				blade.BiosVersion = hpBlade.BladeRomVer

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
						ilo, err := NewIloReader(&blade.BmcAddress, &bmcUser, &bmcPass)
						if err == nil {
							blade.Nics, _ = ilo.Nics()
							err = ilo.Login()
							if err != nil {
								log.WithFields(log.Fields{"operation": "opening ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
							} else {
								defer ilo.Logout()
								blade.BmcAuth = true

								blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = ilo.CPU()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis"}).Warning(err)
								}

								blade.Memory, err = ilo.Memory()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
								}

								blade.BmcLicenceType, blade.BmcLicenceStatus, err = ilo.License()
								if err != nil {
									log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
								}
							}
						}
					} else {
						log.WithFields(log.Fields{"operation": "create ilo connection", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Error("Not reachable")
					}
				}
				blades = append(blades, &blade)
			}
		}
	}
	return blades, err
}
