package audit

import (
	"gitlab.booking.com/infra/thermalnator/connection"
	"gitlab.booking.com/infra/thermalnator/simpleapi"

	log "github.com/sirupsen/logrus"
)

func (a *Audit) AuditChassis(input <-chan simpleapi.Chassis) {
	for chassis := range input {
		rack, err := a.simpleAPI.GetRack(&chassis.Rack)
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

			log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "address": ifdata.IPAddress, "interface": ifname, "level": "chasssis", "vendor": HP}).Info("Collecting chassis data")
			err := a.chassis(&chassis, &rack, &ifdata.IPAddress, &ifname)
			if err == nil {
				break
			} else {
				log.WithFields(log.Fields{"fqdn": chassis.Fqdn, "type": "chassis", "error": err}).Error("Error collecting chassis data")
			}
		}
	}
}

func (a *Audit) chassis(chassis *simpleapi.Chassis, rack *simpleapi.Rack, ip *string, iname *string) (err error) {

	conn := connection.NewChassisConnection(a.username, a.password)
	chassisData, err := c.HpChassis(ip)
	if err != nil {
		return err
	}

	var previousSlot connection.Blade

	for blade := range chassisData.Blades {
		if (blade.IsStorageBlade && previousSlot.IsStorageBlade) || (blade.IsStorageBlade && previousSlot == nil) || (blade.BladePosition%2 == false) {
			log.WithFields(log.Fields{"slot": blade.Bay.Connection, "serial": blade.Bsn, "chassis": chassis.Fqdn}).Warn("Orphan storage blade found")
		}

	}

	return nil
}
