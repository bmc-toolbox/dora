package collectors

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.booking.com/infra/dora/connectors"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/simpleapi"
	"gitlab.booking.com/infra/dora/storage"
)

type Collector struct {
	username    string
	password    string
	telegrafURL string
	simpleAPI   *simpleapi.SimpleAPI
}

type RawCollectedData struct {
	PowerUsage  string
	Temperature string
	Vendor      string
}

func (c *Collector) CollectChassis(input <-chan simpleapi.Chassis) {
	for chassis := range input {
		rack, err := c.simpleAPI.GetRack(&chassis.Rack)
		if err != nil {
			log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis"}).Info("Received errors %s trying to to get data of rack %s\n", err, chassis.Rack)
			continue
		}

		if rack.Site == "" || rack.Sitezone == "" || rack.Sitepod == "" || rack.Siterow == "" || chassis.Rack == "" {
			log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis"}).Error("Position in the datacenter missing")
			continue
		}

		for ifname, ifdata := range chassis.Interfaces {
			if ifdata.IPAddress == "" {
				continue
			}

			err := c.collectChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
			if err == nil {
				break
			} else {
				log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "error": err}).Error("Error collecting chassis data")
			}
		}
	}
}

func (c *Collector) collectChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": *ip, "interface": *iname, "level": "chasssis"}).Info("Collecting data")

	conn := connectors.NewChassisConnection(c.username, c.password)
	var chassisData model.Chassis
	if strings.HasPrefix(chassis.Model, "BladeSystem") {
		chassisData, err = conn.Hp(ip)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(chassis.Model, "P") {
		chassisData, err = conn.Dell(ip)
		if err != nil {
			return err
		}
	} else {
		log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": *ip, "interface": *iname, "level": "chasssis", "Error": "Vendor unknown"}).Error("Collecting data")
	}

	db, err := storage.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	chassisStorage := storage.NewChassisStorage(db)
	_, err = chassisStorage.UpdateOrCreate(&chassisData)
	if err != nil {
		return err
	}

	return nil
}

func New(username string, password string, telegrafURL string, simpleApi *simpleapi.SimpleAPI) *Collector {
	return &Collector{username: username, password: password, telegrafURL: telegrafURL, simpleAPI: simpleApi}
}
