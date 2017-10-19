package connectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/kr/pretty"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	log.WithFields(log.Fields{"step": "connection", "host": c.host}).Info("Detecting vendor")

	client, err := buildClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", c.host))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		log.WithFields(log.Fields{"step": "connection", "host": c.host, "data": "It seems to be HP"}).Debug("Detecting vendor")

		payload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		iloXMLC := &HpRimp{}
		err = xml.Unmarshal(payload, iloXMLC)
		if err != nil {
			return err
		}

		if iloXMLC.HpInfra2 != nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "data": "It's a chassis"}).Debug("Detecting vendor")
			c.vendor = HP
			c.hwtype = Chassis
			return err
		}

		iloXML := &HpRimpBlade{}
		err = xml.Unmarshal(payload, iloXML)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if iloXML.HpBladeBlade != nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "data": "It's a blade"}).Debug("Detecting vendor")
			c.vendor = HP
			c.hwtype = Blade
			return err
		} else if iloXML.HpMP != nil && iloXML.HpBladeBlade == nil {
			log.WithFields(log.Fields{"step": "connection", "host": c.host, "data": "It's a discrete"}).Debug("Detecting vendor")
			c.vendor = HP
			c.hwtype = Discrete
			return err
		}

		return err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/data/login", c.host))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		c.vendor = Dell
		c.hwtype = Blade
		return err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", c.host))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		c.vendor = Dell
		c.hwtype = Chassis
		return err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi/login.cgi", c.host))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
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
		log.WithFields(log.Fields{"operation": "reading serial", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
		return nil, err
	}

	if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" || blade.Serial == "_" {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": ErrInvalidSerial}).Warning("Auditing hardware")
		return nil, ErrInvalidSerial
	}

	blade.BmcType, err = bmc.BmcType()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc type", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.BmcVersion, err = bmc.BmcVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bmc version", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Model, err = bmc.Model()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading model", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Nics, err = bmc.Nics()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading nics", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.BiosVersion, err = bmc.BiosVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading bios version", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = bmc.CPU()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading cpu", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Memory, err = bmc.Memory()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading memory", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Status, err = bmc.Status()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading status", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.Name, err = bmc.Name()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading name", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.TempC, err = bmc.TempC()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading thermal data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.PowerKw, err = bmc.PowerKw()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading power usage data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	blade.BmcLicenceType, blade.BmcLicenceStatus, err = bmc.License()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
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

func (c *Connection) chassis(ch BmcChassis) (chassis *model.Chassis, err error) {
	chassis = &model.Chassis{}

	chassis.Vendor = c.Vendor()
	chassis.BmcAddress = c.host
	chassis.Name, err = ch.Name()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading name", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.Serial, err = ch.Serial()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading serial", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.Model, err = ch.Model()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading model", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.PowerKw, err = ch.PowerKw()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading power usage", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.TempC, err = ch.TempC()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading thermal data", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.Status, err = ch.Status()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading status", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.FwVersion, err = ch.FwVersion()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading firmware version", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.PowerSupplyCount, err = ch.PowerSupplyCount()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading psu count", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.PassThru, err = ch.PassThru()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading passthru", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.Blades, err = ch.Blades()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading blades", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
	}

	chassis.StorageBlades, err = ch.StorageBlades()
	if err != nil {
		log.WithFields(log.Fields{"operation": "reading blades", "ip": chassis.BmcAddress, "vendor": c.Vendor(), "type": c.HwType(), "error": err}).Warning("Auditing hardware")
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
	if c.vendor == HP && (c.hwtype == Blade || c.hwtype == Discrete) {
		ilo, err := NewIloReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c.blade(ilo)
	} else if c.vendor == HP && c.hwtype == Chassis {
		c7000, err := NewHpChassisReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c.chassis(c7000)
	} else if c.vendor == Dell && c.hwtype == Chassis {
		m1000e, err := NewDellCmcReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c.chassis(m1000e)
	} else if c.vendor == Supermicro && c.hwtype == Discrete {
		smBmc, err := NewSupermicroReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c.blade(smBmc)
	}
	/* else if c.vendor == Dell && (c.hwtype == Blade || c.hwtype == Discrete) {
		redfish, err := NewRedFishReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c.blade(redfish), err
	} */

	return i, ErrVendorUnknown
}

func notifyServerChanges(blade *model.Blade, existingData *model.Blade) {
	hasDiff := true
	if len(blade.Nics) == len(existingData.Nics) {
		count := 0
		for _, nic := range blade.Nics {
			for _, enic := range existingData.Nics {
				if nic.MacAddress == enic.MacAddress {
					count++
				}
			}
		}

		if count == len(blade.Nics) {
			hasDiff = false
		}
	}

	if !hasDiff {
		sort.Slice(blade.Nics, func(i, j int) bool {
			switch strings.Compare(blade.Nics[i].MacAddress, blade.Nics[j].MacAddress) {
			case -1:
				return true
			case 1:
				return false
			}
			return blade.Nics[i].MacAddress > blade.Nics[j].MacAddress
		})

		sort.Slice(existingData.Nics, func(i, j int) bool {
			switch strings.Compare(existingData.Nics[i].MacAddress, existingData.Nics[j].MacAddress) {
			case -1:
				return true
			case 1:
				return false
			}
			return existingData.Nics[i].MacAddress > existingData.Nics[j].MacAddress
		})

		for _, diff := range pretty.Diff(blade, existingData) {
			if strings.Contains(diff, "UpdatedAt") || strings.Contains(diff, "].BladeSerial") || strings.Contains(diff, "PowerKw") || strings.Contains(diff, "TempC") || strings.Contains(diff, "ChassisSerial: \"\"") {
				continue
			}
			hasDiff = true
		}
	}

	if hasDiff {
		callback := fmt.Sprintf("%s/blades/%s", viper.GetString("url"), blade.Serial)
		err := assetNotify(callback)
		if err != nil {
			log.WithFields(log.Fields{"operation": "serverdb callback", "url": callback, "error": err}).Error("Sending ServerDB callback")
		}
	}
}

func collect(input <-chan string, db *gorm.DB) {
	bmcUser := viper.GetString("bmc_user")
	bmcPass := viper.GetString("bmc_pass")

	for host := range input {
		c, err := NewConnection(bmcUser, bmcPass, host)
		if err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "error": err}).Error("Connecting to host")
			continue
		}

		if c.HwType() == Blade {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "data": "We don't want to scan blades directly since the chassis does it for us"}).Debug("Connecting to host")
			continue
		}

		data, err := c.Collect()
		if err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "error": err}).Error("Collecting data")
			continue
		}

		switch data.(type) {
		case *model.Chassis:
			chassis := data.(*model.Chassis)
			if chassis == nil {
				continue
			}

			bladeStorage := storage.NewBladeStorage(db)
			existingBlades := make(map[*model.Blade]*model.Blade)
			for _, blade := range chassis.Blades {
				existingBlade, err := bladeStorage.GetOne(blade.Serial)
				if err == nil || err == gorm.ErrRecordNotFound {
					existingBlades[blade] = &existingBlade
				}
			}

			chassisStorage := storage.NewChassisStorage(db)
			_, err = chassisStorage.UpdateOrCreate(chassis)
			if err != nil {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("Collecting data")
				continue
			}

			for blade, existingBlade := range existingBlades {
				notifyServerChanges(blade, existingBlade)
			}

			count, serials, err := chassisStorage.RemoveOldBladesRefs(chassis)
			if err != nil {
				log.WithFields(log.Fields{"operation": "cleanup", "ip": host, "type": c.HwType(), "error": err}).Error("Collecting data")
				continue
			} else if count > 0 {
				for _, serial := range serials {
					log.WithFields(log.Fields{"operation": "cleanup", "ip": host, "type": c.HwType(), "chassis": chassis.Serial, "serial": serial}).Info("blade has been remove from chassis")
				}
				callback := fmt.Sprintf("%s/chassis/%s", viper.GetString("url"), chassis.Serial)
				err := assetNotify(callback)
				if err != nil {
					log.WithFields(log.Fields{"operation": "serverdb callback", "url": callback, "error": err}).Error("Sending ServerDB callback")
				}
			}

			count, serials, err = chassisStorage.RemoveOldStorageBladesRefs(chassis)
			if err != nil {
				log.WithFields(log.Fields{"operation": "cleanup", "ip": host, "type": c.HwType(), "error": err}).Error("Collecting data")
				continue
			} else if count > 0 {
				for _, serial := range serials {
					log.WithFields(log.Fields{"operation": "cleanup", "ip": host, "type": c.HwType(), "chassis": chassis.Serial, "serial": serial}).Info("storage blade has been remove from chassis")
				}
				callback := fmt.Sprintf("%s/chassis/%s", viper.GetString("url"), chassis.Serial)
				err := assetNotify(callback)
				if err != nil {
					log.WithFields(log.Fields{"operation": "serverdb callback", "url": callback, "error": err}).Error("Sending ServerDB callback")
				}
			}
		case *model.Blade:
			blade := data.(*model.Blade)
			if blade == nil {
				continue
			}

			bladeStorage := storage.NewBladeStorage(db)
			existingData, err := bladeStorage.GetOne(blade.Serial)
			if err == nil || err == gorm.ErrRecordNotFound {
				notifyServerChanges(blade, &existingData)
			}

			_, err = bladeStorage.UpdateOrCreate(blade)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "error": err}).Error("Collecting data")
				continue
			}
		}
	}
}

// DataCollection collects the data of all given ips
func DataCollection(ips []string) {
	concurrency := viper.GetInt("collector.concurrency")

	cc := make(chan string, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan string, db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			collect(input, db)
		}(cc, db, &wg)
	}

	if ips[0] == "all" {
		hosts := []model.ScannedPort{}
		if err := db.Where("port = 443 and protocol = 'tcp' and state = 'open'").Find(&hosts).Error; err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": "all", "error": err}).Error(fmt.Sprintf("Retrieving scanned hosts"))
		} else {
			for _, host := range hosts {
				cc <- host.IP
			}
		}
	} else {
		for _, ip := range ips {
			host := model.ScannedPort{}
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				lookup, err := net.LookupHost(ip)
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": ip, "error": err}).Error(fmt.Sprintf("Retrieving scanned hosts"))
					continue
				}
				ip = lookup[0]
			}

			if err := db.Where("ip = ? and port = 443 and protocol = 'tcp' and state = 'open'", ip).Find(&host).Error; err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": ip, "error": err}).Error(fmt.Sprintf("Retrieving scanned hosts"))
				continue
			}
			cc <- host.IP
		}
	}

	close(cc)
	wg.Wait()
}
