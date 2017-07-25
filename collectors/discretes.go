package collectors

// import (
// 	"encoding/json"
// 	"strings"
// 	"time"

// 	log "github.com/sirupsen/logrus"

// 	"fmt"

// 	"gitlab.booking.com/infra/dora/connectors"
// 	"gitlab.booking.com/infra/dora/simpleapi"
// )

// func (c *Collector) CollectDiscrete(input <-chan simpleapi.Server) {
// 	for server := range input {
// 		if server.Rack == "" {
// 			log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete"}).Warning("Position in the datacenter missing")
// 			continue
// 		}

// 		rack, err := c.simpleAPI.GetRack(&server.Rack)
// 		if err != nil {
// 			log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "rack": server.Rack, "error": err}).Warning("Problem retrieving rack from ServerDB")
// 			continue
// 		}

// 		for ifname, ifdata := range server.Interfaces {
// 			if ifname != "ilo" {
// 				continue
// 			}

// 			if ifdata.IPAddress == "" {
// 				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete"}).Error("I coudn't find an ILO address")
// 				continue
// 			}

// 			if strings.HasPrefix(server.Model, "ProLiant") || strings.HasPrefix(server.Model, "CL") {
// 				continue
// 				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": HP, "model": server.Model, "method": "redfish"}).Debug("Collecting Discrete data")
// 				fmt.Println(rack.Name, "HP", server.Model, ifname, ifdata)
// 			} else if strings.HasPrefix(server.Model, "PowerEdge") {
// 				continue
// 				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Dell, "model": server.Model, "method": "redfish"}).Debug("Collecting Discrete data")
// 				err := c.collectDiscreteViaRedFish(&server, &rack, &ifdata.IPAddress, &ifname, Dell)
// 				if err != nil {
// 					log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Supermicro, "model": server.Model, "method": "redfish", "error": err}).Error("Collecting Discrete data")
// 				}
// 			} else if strings.HasPrefix(server.Model, "SSG") || strings.HasPrefix(server.Model, "X") || strings.HasPrefix(server.Model, "Super") {
// 				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Supermicro, "model": server.Model, "method": "redfish"}).Debug("Collecting Discrete data")
// 				err := c.collectDiscreteViaRedFish(&server, &rack, &ifdata.IPAddress, &ifname, Supermicro)
// 				if err != nil {
// 					log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": Supermicro, "model": server.Model, "method": "redfish"}).Info(err)
// 				}
// 				//fmt.Println(rack.Name, "SuperMicro", server.Model, ifname, ifdata)
// 			} else {
// 				log.WithFields(log.Fields{"fqdn": server.Fqdn, "type": "discrete", "vendor": "Unknown", "model": server.Model, "error": "Unknown model"}).Error("Collecting Discrete data")
// 			}
// 		}
// 	}
// }

// func (c *Collector) collectDiscreteViaRedFish(server *simpleapi.Server, rack *simpleapi.Rack, ip *string, iname *string, vendor string) (err error) {
// 	chassis := "-"

// 	result, err := c.viaRedFish(ip, vendor, connectors.RFPower)
// 	if err != nil {
// 		return err
// 	}

// 	rp := &connectors.RedFishPower{}
// 	err = json.Unmarshal(result, rp)
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range rp.PowerControl {
// 		if item.Name == redfishVendorLabels[vendor][connectors.RFPower] {
// 			// Power Consumption
// 			c.createAndSendMessage(
// 				&powerMetric,
// 				&rack.Site,
// 				&rack.Sitezone,
// 				&rack.Sitepod,
// 				&rack.Siterow,
// 				&server.Rack,
// 				&chassis,
// 				server.MainRole(),
// 				&discreteDevice,
// 				&server.Fqdn,
// 				fmt.Sprintf("%.2f", item.PowerConsumedWatts/1000.00),
// 				int32(time.Now().Unix()),
// 			)
// 		}
// 	}

// 	result, err = c.viaRedFish(ip, vendor, connectors.RFThermal)
// 	if err != nil {
// 		return err
// 	}

// 	rt := &connectors.RedFishThermal{}
// 	err = json.Unmarshal(result, rt)
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range rt.Temperatures {
// 		if item.Name == redfishVendorLabels[vendor][connectors.RFThermal] {
// 			// Thermal
// 			c.createAndSendMessage(
// 				&thermalMetric,
// 				&rack.Site,
// 				&rack.Sitezone,
// 				&rack.Sitepod,
// 				&rack.Siterow,
// 				&server.Rack,
// 				&chassis,
// 				server.MainRole(),
// 				&discreteDevice,
// 				&server.Fqdn,
// 				fmt.Sprintf("%d", item.ReadingCelsius),
// 				int32(time.Now().Unix()),
// 			)
// 		}
// 	}

// 	return err
// }
