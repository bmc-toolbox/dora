package collectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"../simpleapi"
)

func (c *Collector) CollectChassis(input <-chan simpleapi.Chassis) {
	for chassis := range input {
		rack, err := c.simpleAPI.GetRack(chassis.Rack)
		if err != nil {
			fmt.Printf("Received error: %s\n", err)
		}

		for ifname, ifdata := range chassis.Interfaces {
			if ifdata.IPAddress == "" {
				continue
			}

			if strings.HasPrefix(chassis.Model, "BladeSystem") {
				err := c.collectHPChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
				if err == nil {
					break
				}
			} else if strings.HasPrefix(chassis.Model, "P") {
				err := c.collectDellChassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
				if err == nil {
					break
				}
			} else {
				fmt.Printf("I dunno what to do with this device %s, skipping...\n", chassis.Fqdn)
			}
		}
	}
}

func (c *Collector) createAndSendBladeMessage(metric *string, site *string, zone *string, pod *string, row *string, rack *string, chassis *string, role *string, device *string, blade *string, value string, now int32) {
	if *site == "" || *zone == "" || *pod == "" || *row == "" || *rack == "" {
		fmt.Printf("%s position in the datacenter is missing, please verify\n", *blade)
	} else {
		c.pushToTelegraph(fmt.Sprintf("%s,site=%s,zone=%s,pod=%s,row=%s,rack=%s,chassis=%s,role=%s,device_type=%s,device_name=%s value=%s %d\n", *metric, *site, *zone, *pod, *row, *rack, *chassis, *role, *device, *blade, value, now))
	}
}

func (c *Collector) createAndSendChassisMessage(metric *string, site *string, zone *string, pod *string, row *string, rack *string, chassis *string, value string, now int32) {
	if *site == "" || *zone == "" || *pod == "" || *row == "" || *rack == "" {
		fmt.Printf("%s position in the datacenter is missing, please verify\n", *chassis)
	} else {
		c.pushToTelegraph(fmt.Sprintf("%s,site=%s,zone=%s,pod=%s,row=%s,rack=%s,chassis=-,role=chassis,device_type=chassis,device_name=%s value=%s %d\n", *metric, *site, *zone, *pod, *row, *rack, *chassis, value, now))
	}
}

func (c *Collector) collectHPChassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {
	fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via ILOXML %s", chassis.Fqdn, *ip, *iname))
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
				var role string
				var device string
				if strings.Contains(blade.Spn, "Storage") {
					blade.Name = blade.Bsn
					role = "storageblade"
					device = "storageblade"
				} else if blade.Name == "" || blade.Name == "[Unknown]" || blade.Name == "host is unnamed" {
					blade.Name, err = chassis.GetBladeNameByBay(blade.Bay.Connection)
					if err == simpleapi.ErrNoBladeFound {
						fmt.Printf("Blade %d with serial %s hasn't been found in ServerDB %s, please verify...\n", blade.Bay.Connection, blade.Bsn, chassis.Fqdn)
					}
					blade.Name = blade.Bsn
					role = "UnknownBlade"
					device = "blade"
				} else {
					role = "CouldNotFind"
					server, err := c.simpleAPI.GetServer(&blade.Name)
					if err != nil {
						fmt.Println(err)
					} else if server != nil {
						for _, r := range server.Roles {
							if r != "staging" {
								role = r
								break
							}
						}
					}
					device = "blade"
				}

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
	fmt.Println(fmt.Sprintf("Collecting data from %s[%s] via RedFish %s", chassis.Fqdn, *ip, *iname))
	for _, blade := range chassis.Blades {
		for hostname := range blade {
			if strings.HasSuffix(hostname, ".com") && !strings.HasPrefix(hostname, "spare") {
				// Fix tomorrow the spare-

				bmcAddress := bmcAddressBuild.ReplaceAllString(hostname, ".lom.")
				result, err := c.viaRedFish(&bmcAddress, Dell, RFPower)
				if err != nil {
					fmt.Println(err)
					break
				}
				rp := &RedFishPower{}
				err = json.Unmarshal(result, rp)
				if err != nil {
					fmt.Println(err)
					break
				}

				server, err := c.simpleAPI.GetServer(&hostname)
				role := "CouldNotFind"
				if err != nil {
					fmt.Println(err)
				} else if server != nil {
					for _, r := range server.Roles {
						if r != "staging" {
							role = r
							break
						}
					}
				}
				device := "blade"

				for _, item := range rp.PowerControl {
					if item.Name == "System Power Control" {
						// Power Consumption
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
							&hostname,
							fmt.Sprintf("%.2f", item.PowerConsumedWatts/1000.00),
							int32(time.Now().Unix()),
						)
					}
				}

				result, err = c.viaRedFish(&bmcAddress, Dell, RFThermal)
				if err != nil {
					fmt.Println(err)
					break
				}

				rt := &RedFishThermal{}
				err = json.Unmarshal(result, rt)
				if err != nil {
					fmt.Println(err)
					break
				}

				for _, item := range rt.Temperatures {
					if item.Name == "System Board Inlet Temp" {
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
							&hostname,
							fmt.Sprintf("%d", item.ReadingCelsius),
							int32(time.Now().Unix()),
						)
					}
				}
			}
		}
	}

	return err
}
