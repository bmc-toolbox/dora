package connectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// Connection is used to connect and later dicover the hardware information we have for each vendor
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

// Collect collects all relevant data of the current hardwand and returns the populated object
func (c *Connection) Collect() (i interface{}, err error) {
	if c.vendor == HP && (c.hwtype == Blade || c.hwtype == Discrete) {
		ilo, err := NewIloReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return ilo.Blade()
	} else if c.vendor == HP && c.hwtype == Chassis {
		c7000, err := NewHpChassisReader(&c.host, &c.username, &c.password)
		if err != nil {
			return i, err
		}
		return c7000.Chassis()
	}

	return i, err
}

func collect(input <-chan *Connection, db *gorm.DB) {
	for hosts := range input {

	}
}

// DataCollection collects the data of all given ips
func DataCollection(ips []string) {
	concurrency := viper.GetInt("concurrency")

	cc := make(chan string, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()
	bmcUser := viper.GetString("bmc_user")
	bmcPass := viper.GetString("bmc_pass")

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan *Connection, db *gorm.DB, wg *sync.WaitGroup) {
			collect(input, db)
			wg.Done()
		}(cc, db, &wg)
	}

	if ips[0] == "all" {
		hosts := []model.ScannedPort{}
		db.Where("port = 443 and protocol = 'tcp' and state = 'open'").Find(&scans)
		for _, host := range hosts {
			c := NewConnection(bmcUser, bmcPass, host.ScannedHostIP)
			if c.HwType != Blade {
				cc <- c
			}
		}
	} else {
		for _, ip := range ips {
			host := model.ScannedPort{}
			db.Where("scanned_host_ip = ? and port = 443 and protocol = 'tcp' and state = 'open'", ip).First(&scans)
			for _, host := range hosts {
				c := NewConnection(bmcUser, bmcPass, host.ScannedHostIP)
				cc <- c
			}
		}
	}

	close(cc)
	wg.Wait()
}
