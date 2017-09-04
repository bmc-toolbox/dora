package connectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/model"
)

// Connection is used to connect and later dicover the hardware information we have for each vendor
type Connection struct {
	username string
	password string
	host     string
	vendor   string
	hwtype   string
}

// VendorAndType returns the vendor and hwtype of the current connection
func (c *Connection) VendorAndType() (vendor string, hwtype string) {
	return c.vendor, c.hwtype
}

func (c *Connection) detect() (err error) {
	log.WithFields(log.Fields{"step": "onnection", "host": c.host}).Info("Detecting vendor")

	client, err := buildClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", c.host))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		payload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		iloXML := &HpRimpBlade{}
		err = xml.Unmarshal(payload, iloXML)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if iloXML.HpBladeBlade != nil {
			c.vendor = HP
			c.hwtype = Blade
			return err
		} else if iloXML.HpMP != nil && iloXML.HpBladeBlade == nil {
			c.vendor = HP
			c.hwtype = Discrete
			return err
		}

		iloXMLC := &HpRimp{}
		err = xml.Unmarshal(payload, iloXMLC)
		if err != nil {
			return err
		}

		if iloXMLC.HpInfra2 != nil {
			c.vendor = HP
			c.hwtype = Chassis
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

func (c *Connection) iLO() (b model.Blade, err error) {
	ilo, err := NewIloReader(&c.host, &c.username, &c.password)
	if err != nil {
		log.WithFields(log.Fields{"operation": "create ilo connection", "ip": b.BmcAddress, "name": b.Name, "serial": b.Serial, "type": "chassis", "error": err}).Warning("Auditing blade")
	}

	b.BmcAddress = c.host

	b.BmcType, err = ilo.BmcType()
	if err != nil {
		return b, err
	}

	b.BmcVersion, err = ilo.BmcVersion()
	if err != nil {
		return b, err
	}

	b.Serial, err = ilo.Serial()
	if err != nil {
		return b, err
	}

	b.Model, err = ilo.Model()
	if err != nil {
		return b, err
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
	return b, err
}

// Collect collects all relevant data of the current hardwand and returns the populated object
func (c *Connection) Collect() (i interface{}, err error) {
	if c.vendor == HP && (c.hwtype == Blade || c.hwtype == Discrete) {
		return c.iLO()
	}

	return i, err
}
