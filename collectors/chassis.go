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
			log.WithFields(log.Fields{"chassis": chassis.Fqdn}).Info("Received errors %s trying to to get data of rack %s\n", err, chassis.Rack)
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
					log.WithFields(log.Fields{"chassis": chassis.Fqdn}).Info("Error collecting chassis data")
				}
			} else if strings.HasPrefix(chassis.Model, "P") {
				err := c.collectDellChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
				if err == nil {
					break
				} else {
					log.WithFields(log.Fields{"chassis": chassis.Fqdn}).Info("Error collecting chassis data")
				}
			} else {
				log.WithFields(log.Fields{"chassis": chassis.Fqdn}).Info("I dunno what to do with this device, skipping..")
			}
		}
	}
}

func (c *Collector) createAndSendBladeMessage(metric *string, site *string, zone *string, pod *string, row *string, rack *string, chassis *string, role *string, device *string, blade *string, value string, now int32) {
	if *site == "" || *zone == "" || *pod == "" || *row == "" || *rack == "" {
		log.WithFields(log.Fields{"blade": *blade, "metric": *metric}).Info("Position in the datacenter missing")
	} else {
		err := c.pushToTelegraph(fmt.Sprintf("%s,site=%s,zone=%s,pod=%s,row=%s,rack=%s,chassis=%s,role=%s,device_type=%s,device_name=%s value=%s %d\n", *metric, *site, *zone, *pod, *row, *rack, *chassis, *role, *device, *blade, value, now))
		if err != nil {
			log.WithFields(log.Fields{"blade": *blade, "metric": *metric}).Info("Unable to push data to telegraf")
		}
	}
}

func (c *Collector) createAndSendChassisMessage(metric *string, site *string, zone *string, pod *string, row *string, rack *string, chassis *string, value string, now int32) {
	if *site == "" || *zone == "" || *pod == "" || *row == "" || *rack == "" {
		log.WithFields(log.Fields{"chassis": *chassis, "metric": *metric}).Info("Position in the datacenter missing")
	} else {
		err := c.pushToTelegraph(fmt.Sprintf("%s,site=%s,zone=%s,pod=%s,row=%s,rack=%s,chassis=-,role=chassis,device_type=chassis,device_name=%s value=%s %d\n", *metric, *site, *zone, *pod, *row, *rack, *chassis, value, now))
		if err != nil {
			log.WithFields(log.Fields{"chassis": *chassis, "metric": *metric}).Info("Unable to push data to telegraf")
		}
	}
}

func (c *Collector) collectHPChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"chassis": chassis.Fqdn, "address": *ip, "interface": *iname, "level": "chasssis"}).Info("Collecting data via Chassi XML Interface")

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
		c.createAndSendChassisMessage(
			&powerMetric,
			&rack.Site,
			&rack.Sitezone,
			&rack.Sitepod,
			&rack.Siterow,
			&chassis.Rack,
			&chassis.Fqdn,
			fmt.Sprintf("%.2f", iloXML.Infra2.Power.PowerConsumed/1000.00),
			now,
		)

		log.WithFields(log.Fields{"metric": thermalMetric, "site": rack.Site, "zone": rack.Sitezone, "pod": rack.Sitepod, "row": rack.Siterow, "rack": chassis.Rack, "fqdn": chassis.Fqdn, "type": "chassis"}).Debug("Pushing metric to telegraf")

		// Thermal
		c.createAndSendChassisMessage(
			&thermalMetric,
			&rack.Site,
			&rack.Sitezone,
			&rack.Sitepod,
			&rack.Siterow,
			&chassis.Rack,
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
					if err == simpleapi.ErrNoBladeFound {
						log.WithFields(log.Fields{"slot": blade.Bay.Connection, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Info("Blade not found in SimpleAPI")
						blade.Name = blade.Bsn
						role = "UnknownBlade"
					}
				}

				if role == "CouldNotFind" {
					server, err := c.simpleAPI.GetServer(&blade.Name)
					if err != nil {
						log.WithFields(log.Fields{"blade": blade.Name, "chassis": chassis.Fqdn}).Info("Error retrieving data from SimpleAPI")
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
				c.createAndSendBladeMessage(
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
				c.createAndSendBladeMessage(
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

	return err
}

func (c *Collector) collectDellChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	log.WithFields(log.Fields{"chassis": chassis.Fqdn, "address": *ip, "interface": *iname, "level": "chasssis"}).Info("Collecting data via Chassi XML Interface")

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
					if err == simpleapi.ErrNoBladeFound {
						log.WithFields(log.Fields{"slot": blade.Slot, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Info("Blade not found in SimpleAPI")
						blade.Name = blade.Bsn
						role = "UnknownBlade"
					}
				} else {
					log.WithFields(log.Fields{"slot": blade.Slot, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Info("Could not convert the Blade slot")
					blade.Name = blade.Bsn
					role = "UnknownBlade"
				}
			}

			if role == "CouldNotFind" {
				server, err := c.simpleAPI.GetServer(&blade.Name)
				if err != nil {
					log.WithFields(log.Fields{"blade": blade.Name, "chassis": chassis.Fqdn}).Info("Error retrieving data from SimpleAPI")
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
			c.createAndSendBladeMessage(
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
			c.createAndSendBladeMessage(
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

	return err
}
