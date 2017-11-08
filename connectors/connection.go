package connectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

const (
	// Blade is the constant defining the blade hw type
	Blade = "blade"
	// Discrete is the constant defining the Discrete hw type
	Discrete = "discrete"
	// Chassis is the constant defining the chassis hw type
	Chassis = "chassis"
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
)

// Connection is used to connect and later discover the hardware information we have for each vendor
type Connection struct {
	username string
	password string
	host     string
	vendor   string
	hwtype   string
}

// Vendor returns the vendor of the current connection
func (c *Connection) Vendor() (vendor string) {
	return c.vendor
}

// HwType returns hwtype of the current connection
func (c *Connection) HwType() (hwtype string) {
	return c.hwtype
}

func (c *Connection) detect() (err error) {
	log.WithFields(log.Fields{"step": "connection", "host": c.host}).Debug("detecting vendor")

	client, err := buildClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", c.host))
	if err != nil {
		return err
	}
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		iloXMLC := &HpRimp{}
		err = xml.Unmarshal(payload, iloXMLC)
		if err != nil {
			return err
		}

		if iloXMLC.HpInfra2 != nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": HP}).Debug("it's a chassis")
			c.vendor = HP
			c.hwtype = Chassis
			return err
		}

		iloXML := &HpRimpBlade{}
		err = xml.Unmarshal(payload, iloXML)
		if err != nil {
			return err
		}

		if iloXML.HpBladeBlade != nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": HP}).Debug("it's a blade")
			c.vendor = HP
			c.hwtype = Blade
			return err
		} else if iloXML.HpMP != nil && iloXML.HpBladeBlade == nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": HP}).Debug("it's a discrete")
			c.vendor = HP
			c.hwtype = Discrete
			return err
		}

		return err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/session?aimGetProp=hostname,gui_str_title_bar,OEMHostName,fwVersion,sysDesc", c.host))
	if err != nil {
		return err
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		dellJSON := &DellHwDetection{}
		err = json.Unmarshal(payload, dellJSON)
		if err != nil {
			return err
		}

		c.vendor = Dell

		if strings.HasPrefix(dellJSON.AimGetProp.SysDesc, "PowerEdge M") {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": Dell}).Debug("it's a blade")
			c.hwtype = Blade
		} else {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": Dell}).Debug("it's a discrete")
			c.hwtype = Discrete
		}

		return err
	}
	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", c.host))
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": Dell}).Debug("it's a chassis")
		c.vendor = Dell
		c.hwtype = Chassis
		return err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi/login.cgi", c.host))
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.WithFields(log.Fields{"step": "connection", "host": c.host, "vendor": Supermicro}).Debug("it's a discrete")
		c.vendor = Supermicro
		c.hwtype = Discrete
		return err
	}

	return ErrVendorUnknown
}

// NewConnection creates a new connection and detects the vendor and model of the given hardware
func NewConnection(username string, password string, host string) (c *Connection, err error) {
	c = &Connection{username: username, password: password, host: host}
	err = c.detect()
	return c, err
}

func (c *Connection) blade(bmc Bmc) (blade *model.Blade, err error) {
	err = bmc.Login()
	if err != nil {
		return blade, err
	}

	defer bmc.Logout()
	blade = &model.Blade{}
	db := storage.InitDB()

	blade.BmcAuth = true
	blade.BmcWEBReachable = true
	blade.BmcAddress = c.host
	blade.Vendor = c.Vendor()

	blade.Serial, err = bmc.Serial()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
		return nil, err
	}

	if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" || blade.Serial == "_" {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": ErrInvalidSerial}).Warning("auditing hardware")
		return nil, ErrInvalidSerial
	}

	blade.BmcType, err = bmc.BmcType()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc type", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.BmcVersion, err = bmc.BmcVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc version", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Model, err = bmc.Model()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading model", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Nics, err = bmc.Nics()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading nics", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.BiosVersion, err = bmc.BiosVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bios version", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = bmc.CPU()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading cpu", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Memory, err = bmc.Memory()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading memory", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Status, err = bmc.Status()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading status", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.Name, err = bmc.Name()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading name", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.TempC, err = bmc.TempC()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading thermal data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.PowerKw, err = bmc.PowerKw()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading power usage data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	blade.BmcLicenceType, blade.BmcLicenceStatus, err = bmc.License()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	scans := []model.ScannedPort{}
	db.Where("ip = ?", blade.BmcAddress).Find(&scans)
	for _, scan := range scans {
		if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
			blade.BmcSSHReachable = true
		} else if scan.Port == 623 && scan.Protocol == "udp" && (scan.State == "open|filtered" || scan.State == "open") {
			blade.BmcIpmiReachable = true
		}
	}

	return blade, nil
}

func (c *Connection) discrete(bmc Bmc) (discrete *model.Discrete, err error) {
	err = bmc.Login()
	if err != nil {
		return discrete, err
	}

	defer bmc.Logout()
	discrete = &model.Discrete{}
	db := storage.InitDB()

	discrete.BmcAuth = true
	discrete.BmcWEBReachable = true
	discrete.BmcAddress = c.host
	discrete.Vendor = c.Vendor()

	discrete.Serial, err = bmc.Serial()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
		return nil, err
	}

	if discrete.Serial == "" || discrete.Serial == "[unknown]" || discrete.Serial == "0000000000" || discrete.Serial == "_" {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": ErrInvalidSerial}).Warning("auditing hardware")
		return nil, ErrInvalidSerial
	}

	discrete.BmcType, err = bmc.BmcType()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc type", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.BmcVersion, err = bmc.BmcVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc version", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Model, err = bmc.Model()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading model", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Nics, err = bmc.Nics()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading nics", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.BiosVersion, err = bmc.BiosVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bios version", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Processor, discrete.ProcessorCount, discrete.ProcessorCoreCount, discrete.ProcessorThreadCount, err = bmc.CPU()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading cpu", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Memory, err = bmc.Memory()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading memory", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Status, err = bmc.Status()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading status", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.Name, err = bmc.Name()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading name", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.TempC, err = bmc.TempC()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading thermal data", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.PowerKw, err = bmc.PowerKw()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading power usage data", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	discrete.BmcLicenceType, discrete.BmcLicenceStatus, err = bmc.License()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading license data", "ip": discrete.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	scans := []model.ScannedPort{}
	db.Where("ip = ?", discrete.BmcAddress).Find(&scans)
	for _, scan := range scans {
		if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
			discrete.BmcSSHReachable = true
		} else if scan.Port == 623 && scan.Protocol == "udp" && (scan.State == "open|filtered" || scan.State == "open") {
			discrete.BmcIpmiReachable = true
		}
	}

	return discrete, nil
}

func (c *Connection) chassis(ch BmcChassis) (chassis *model.Chassis, err error) {
	err = ch.Login()
	if err != nil {
		return chassis, err
	}
	defer ch.Logout()

	chassis = &model.Chassis{}

	chassis.Vendor = c.Vendor()
	chassis.BmcAddress = c.host
	chassis.Name, err = ch.Name()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading name", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Serial, err = ch.Serial()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Model, err = ch.Model()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading model", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.PowerKw, err = ch.PowerKw()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading power usage", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.TempC, err = ch.TempC()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading thermal data", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Status, err = ch.Status()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading status", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.FwVersion, err = ch.FwVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading firmware version", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.PassThru, err = ch.PassThru()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading passthru", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Blades, err = ch.Blades()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading blades", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.StorageBlades, err = ch.StorageBlades()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading storage blades", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Nics, err = ch.Nics()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading nics", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	chassis.Psus, err = ch.Psus()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading psus", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("auditing hardware")
	}

	db := storage.InitDB()
	scans := []model.ScannedPort{}
	db.Where("ip = ?", chassis.BmcAddress).Find(&scans)
	for _, scan := range scans {
		if scan.Port == 443 && scan.Protocol == "tcp" && scan.State == "open" {
			chassis.BmcWEBReachable = true
		} else if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
			chassis.BmcSSHReachable = true
		}
	}

	return chassis, nil
}

// Collect collects all relevant data of the current hardwand and returns the populated object
func (c *Connection) Collect() (i interface{}, err error) {
	switch c.Vendor() {
	case HP:
		if c.HwType() == Blade {
			ilo, err := NewIloReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.blade(ilo)
		} else if c.HwType() == Discrete {
			ilo, err := NewIloReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.blade(ilo)
		} else if c.HwType() == Chassis {
			c7000, err := NewHpChassisReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.chassis(c7000)
		}
	case Dell:
		if c.HwType() == Blade {
			idrac, err := NewIDracReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.blade(idrac)
		} else if c.HwType() == Discrete {
			idrac, err := NewIDracReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.discrete(idrac)
		} else if c.HwType() == Chassis {
			m1000e, err := NewDellCmcReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.chassis(m1000e)
		}
	case Supermicro:
		if c.HwType() == Discrete {
			smBmc, err := NewSupermicroReader(&c.host, &c.username, &c.password)
			if err != nil {
				return i, err
			}
			return c.discrete(smBmc)
		}
	}

	return i, ErrVendorUnknown
}
