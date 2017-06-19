package collectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"../simpleapi"
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

			if strings.HasPrefix(chassis.Model, "BladeSystem") {
				err := c.collectHPChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
				if err == nil {
					break
				} else {
					log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "error": err}).Error("Error collecting chassis data")
				}
			} else if strings.HasPrefix(chassis.Model, "P") {
				err := c.collectDellChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
				if err == nil {
					break
				} else {
					log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "error": err}).Error("Error collecting chassis data")
				}
			} else {
				log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis"}).Warning("I dunno what to do with this device, skipping..")
			}
		}
	}
}

func (c *Collector) collectHPChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": *ip, "interface": *iname, "level": "chasssis", "vendor": HP}).Info("Collecting data via Chassi XML Interface")

	result, err := c.viaILOXML(ip)
	if err != nil {
		return err
	}
	iloXML := &Rimp{}
	err = xml.Unmarshal(result, iloXML)
	if err != nil {
		return err
	}

	if iloXML.Infra2 != nil {
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
			fmt.Sprintf("%.2f", iloXML.Infra2.Power.PowerConsumed/1000.00),
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
			iloXML.Infra2.Temps.Temp.C,
			now,
		)

		if iloXML.Infra2.Blades != nil {
			for _, blade := range iloXML.Infra2.Blades.Blade {
				role := "CouldNotFind"
				device := "blade"

				if strings.Contains(blade.Spn, "Storage") {
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
					fmt.Sprintf("%.2f", blade.Power.PowerConsumed/1000.00),
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
					blade.Temps.Temp.C,
					now,
				)
			}
		}
	}

	return nil
}

func (c *Collector) collectDellChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": *ip, "interface": *iname, "level": "chasssis", "vendor": Dell}).Info("Collecting data via Chassis JSON Interface")

	result, err := c.dellCMC(ip)
	if err != nil {
		return err
	}
	cmcJSON := &CMC{}
	err = json.Unmarshal(result, cmcJSON)
	if err != nil {
		return err
	}

	for _, blade := range cmcJSON.Blades {
		if blade.Present == 1 {
			now := int32(time.Now().Unix())

			role := "CouldNotFind"
			device := "blade"

			if blade.IsStorageBlade == 1 {
				blade.Name = blade.Bsn
				role = "storageblade"
				device = "storageblade"
			} else if blade.Name == "" || blade.Name == "localhost.localdomain" {
				pos, err := strconv.Atoi(blade.Slot)
				if err == nil {
					blade.Name, err = chassis.GetBladeNameByBay(pos)
					if err == simpleapi.ErrBladeNotFound {
						log.WithFields(log.Fields{"slot": blade.Slot, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Warning("Blade not found in SimpleAPI")
						blade.Name = blade.Bsn
						role = "UnknownBlade"
					}
				} else {
					log.WithFields(log.Fields{"slot": blade.Slot, "serial": blade.Bsn, "chassis": chassis.Fqdn, "error": err}).Error("Could not convert the Blade slot")
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
				fmt.Sprintf("%.2f", blade.CurrentConsumption/1000.00),
				now,
			)

			if blade.Temperature == "N/A" {
				blade.Temperature = "0"
			}

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
				blade.Temperature,
				now,
			)
		}
	}

	return nil
}
