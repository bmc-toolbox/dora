package collectors

import (
	"encoding/json"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"fmt"

	"../simpleapi"
)

func (c *Collector) CollectDiscrete(input <-chan simpleapi.Server) {
	for server := range input {
		if server.Rack == "" {
			log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete"}).Warning("Position in the datacenter missing")
			continue
		}

		rack, err := c.simpleAPI.GetRack(&server.Rack)
		if err != nil {
			log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "rack": server.Rack, "error": err}).Warning("Problem retrieving rack from ServerDB")
			continue
		}

		for ifname, ifdata := range server.Interfaces {
			if ifname != "ilo" {
				continue
			}

			if ifdata.IPAddress == "" {
				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete"}).Error("I coudn't find an ILO address")
				continue
			}

			if strings.HasPrefix(server.Model, "ProLiant") || strings.HasPrefix(server.Model, "CL") {
				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": HP, "method": "redfish"}).Debug("Collecting Discrete data")
				fmt.Println(rack.Name, "HP", server.Model, ifname, ifdata)
			} else if strings.HasPrefix(server.Model, "PowerEdge") {
				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Dell, "method": "redfish"}).Debug("Collecting Discrete data")
				c.collectDellDiscrete(&server, &rack, &ifdata.IPAddress, &ifname)
			} else if strings.HasPrefix(server.Model, "SSG") || strings.HasPrefix(server.Model, "X") || strings.HasPrefix(server.Model, "Super") {
				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Supermicro, "method": "redfish"}).Debug("Collecting Discrete data")
				fmt.Println(rack.Name, "SuperMicro", server.Model, ifname, ifdata)
			} else {
				fmt.Println(rack.Name, "No Idea", server.Model, ifname, ifdata)
			}

			// if strings.HasPrefix(chassis.Model, "BladeSystem") {
			// 	err := c.collectHPChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
			// 	if err == nil {
			// 		break
			// 	} else {
			// 		log.WithFields(log.Fields{"chassis": chassis.Fqdn, "error": err}).Error("Error collecting chassis data")
			// 	}
			// } else if strings.HasPrefix(chassis.Model, "P") {
			// 	err := c.collectDellChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
			// 	if err == nil {
			// 		break
			// 	} else {
			// 		log.WithFields(log.Fields{"chassis": chassis.Fqdn, "error": err}).Error("Error collecting chassis data")
			// 	}
			// } else {
			// 	log.WithFields(log.Fields{"chassis": chassis.Fqdn}).Warning("I dunno what to do with this device, skipping..")
			// }
		}
	}
}

func (c *Collector) collectDellDiscrete(server *simpleapi.Server, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	chassis := "-"

	result, err := c.viaRedFish(ip, Dell, RFPower)
	if err != nil {
		return err
	}

	rp := &RedFishPower{}
	err = json.Unmarshal(result, rp)
	if err != nil {
		return err
	}

	for _, item := range rp.PowerControl {
		if item.Name == "System Power Control" {
			// Power Consumption
			c.createAndSendMessage(
				&powerMetric,
				&rack.Site,
				&rack.Sitezone,
				&rack.Sitepod,
				&rack.Siterow,
				&server.Rack,
				&chassis,
				server.MainRole(),
				&discreteDevice,
				&server.Fqdn,
				fmt.Sprintf("%.2f", item.PowerConsumedWatts/1000.00),
				int32(time.Now().Unix()),
			)
		}
	}

	rt := &RedFishThermal{}
	err = json.Unmarshal(result, rt)
	if err != nil {
		return err
	}

	for _, item := range rt.Temperatures {
		if item.Name == "System Board Inlet Temp" {
			// Thermal
			c.createAndSendMessage(
				&thermalMetric,
				&rack.Site,
				&rack.Sitezone,
				&rack.Sitepod,
				&rack.Siterow,
				&server.Rack,
				&chassis,
				server.MainRole(),
				&discreteDevice,
				&server.Fqdn,
				fmt.Sprintf("%d", item.ReadingCelsius),
				int32(time.Now().Unix()),
			)
		}
	}

	return err
}
