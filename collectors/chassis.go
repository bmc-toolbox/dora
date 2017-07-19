package collectors

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.booking.com/infra/thermalnator/connectors"
	"gitlab.booking.com/infra/thermalnator/simpleapi"
)

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

			err := c.collectHPChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
			if err == nil {
				break
			} else {
				log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "error": err}).Error("Error collecting chassis data")
			}
		}
	}
}

func (c *Collector) collectChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": *ip, "interface": *iname, "level": "chasssis", "vendor": HP}).Info("Collecting data via Chassi XML Interface")

	conn := connectors.NewChassisConnection(a.username, a.password)
	if strings.HasPrefix(chassis.Model, "BladeSystem") {
		chassisData, err := a.HpChassis(ip)
	} else if strings.HasPrefix(chassis.Model, "P") {
		chassisData, err := a.DellChassis(ip)
	} else {
		return ErrChassiCollectionNotSupported
	}

	if err != nil {
		return err
	}

	now := int32(time.Now().Unix())

	log.WithFields(log.Fields{"metric": powerMetric, "site": rack.Site, "zone": rack.Sitezone, "pod": rack.Sitepod, "row": rack.Siterow, "rack": chassis.Rack, "fqdn": chassis.Fqdn, "type": "chassis"}).Debug("Pushing metric to telegraf")

	// Power Usage
	c.createAndSendMessage(
		&powerMetric,
		&rack.Site,
		&rack.Sitezone,
		&rack.Sitepod,
		&rack.Siterow,
		&chassis.Rack,
		&chassis.Fqdn,
		&chassisDevice,
		&chassisDevice,
		&chassis.Fqdn,
		chassisData.Power,
		now,
	)

	log.WithFields(log.Fields{"metric": thermalMetric, "site": rack.Site, "zone": rack.Sitezone, "pod": rack.Sitepod, "row": rack.Siterow, "rack": chassis.Rack, "fqdn": chassis.Fqdn, "type": "chassis"}).Debug("Pushing metric to telegraf")

	// Thermal
	c.createAndSendMessage(
		&thermalMetric,
		&rack.Site,
		&rack.Sitezone,
		&rack.Sitepod,
		&rack.Siterow,
		&chassis.Rack,
		&chassis.Fqdn,
		&chassisDevice,
		&chassisDevice,
		&chassis.Fqdn,
		chassisData.Temp,
		now,
	)

	for _, blade := range iloXML.Infra2.Blades.Blade {
		role := "CouldNotFind"
		device := "blade"

		if blade.IsStorageBlade {
			blade.Name = blade.Bsn
			role = "storageblade"
			device = "storageblade"
		} else if blade.Name == "" || blade.Name == "[Unknown]" || blade.Name == "host is unnamed" || blade.Name == "localhost.localdomain" {
			blade.Name, err = chassis.GetBladeNameByBay(blade.Bay.Connection)
			if err == simpleapi.ErrBladeNotFound {
				log.WithFields(log.Fields{"slot": blade.Bay.Connection, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Warning("Blade not found in SimpleAPI")
				blade.Name = blade.Bsn
				role = "UnknownBlade"
			}
		}

		if role == "CouldNotFind" {
			server, _err := c.simpleAPI.GetServer(&blade.Name)
			if _err != nil && _err != simpleapi.ErrBladeNotFound {
				log.WithFields(log.Fields{"blade": blade.Name, "chassis": chassis.Fqdn, "error": _err}).Warning("Error retrieving data from SimpleAPI")
			} else if server != nil {
				for _, r := range server.Roles {
					if r != "staging" {
						role = r
						break
					}
				}
			}
		}

		log.WithFields(log.Fields{"metric": powerMetric, "site": rack.Site, "zone": rack.Sitezone, "pod": rack.Sitepod, "row": rack.Siterow, "rack": chassis.Rack, "fqdn": blade.Name, "type": "blade"}).Debug("Pushing metric to telegraf")

		// Power Usage
		c.createAndSendMessage(
			&powerMetric,
			&rack.Site,
			&rack.Sitezone,
			&rack.Sitepod,
			&rack.Siterow,
			&chassis.Rack,
			&chassis.Fqdn,
			&role,
			&device,
			&blade.Name,
			blade.Power,
			now,
		)

		log.WithFields(log.Fields{"metric": thermalMetric, "site": rack.Site, "zone": rack.Sitezone, "pod": rack.Sitepod, "row": rack.Siterow, "rack": chassis.Rack, "fqdn": blade.Name, "type": "blade"}).Debug("Pushing metric to telegraf")

		// Thermal
		c.createAndSendMessage(
			&thermalMetric,
			&rack.Site,
			&rack.Sitezone,
			&rack.Sitepod,
			&rack.Siterow,
			&chassis.Rack,
			&chassis.Fqdn,
			&role,
			&device,
			&blade.Name,
			blade.Temp,
			now,
		)
	}

	return nil
}
