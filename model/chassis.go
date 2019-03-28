package model

import (
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// NewChassisFromDevice will create a new object coming from the bmc discrete devices
func NewChassisFromDevice(c *devices.Chassis) (chassis *Chassis) {
	chassis = &Chassis{}
	chassis.Vendor = c.Vendor
	chassis.BmcAddress = c.BmcAddress
	chassis.Name = c.Name
	chassis.Serial = c.Serial
	chassis.Model = c.Model
	chassis.PowerKw = c.PowerKw
	chassis.TempC = c.TempC
	chassis.Status = c.Status
	chassis.FwVersion = c.FwVersion
	chassis.PassThru = c.PassThru
	chassis.Blades = make([]*Blade, 0)
	for _, b := range c.Blades {
		blade := NewBladeFromDevice(b)
		if blade.Serial == "" || blade.Serial == "[unknown]" || blade.Serial == "0000000000" || blade.Serial == "_" {
			chassis.FaultySlots = append(chassis.FaultySlots, blade.BladePosition)
			log.WithFields(log.Fields{"operation": "chassis scan", "position": blade.BladePosition, "type": "chassis", "chassis_serial": chassis.Serial}).Error(errors.ErrInvalidSerial)
			continue
		}
		blade.ChassisSerial = c.Serial
		chassis.Blades = append(chassis.Blades, blade)
	}
	chassis.StorageBlades = make([]*StorageBlade, 0)
	for _, s := range c.StorageBlades {
		storageBlade := NewStorageBladeFromDevice(s)
		if storageBlade.Serial == "" || storageBlade.Serial == "[unknown]" || storageBlade.Serial == "0000000000" || storageBlade.Serial == "_" {
			chassis.FaultySlots = append(chassis.FaultySlots, storageBlade.BladePosition)
			log.WithFields(log.Fields{"operation": "chassis scan", "position": storageBlade.BladePosition, "type": "chassis", "chassis_serial": chassis.Serial}).Error(errors.ErrInvalidSerial)
			continue
		}
		chassis.StorageBlades = append(chassis.StorageBlades, storageBlade)

	}
	chassis.Nics = make([]*Nic, 0)
	for _, nic := range c.Nics {
		chassis.Nics = append(chassis.Nics, &Nic{
			MacAddress:    nic.MacAddress,
			Name:          nic.Name,
			ChassisSerial: c.Serial,
		})
	}
	chassis.Psus = make([]*Psu, 0)
	for psuPosition, psu := range c.Psus {
		if psu.Serial == "" || psu.Serial == "[unknown]" || psu.Serial == "0000000000" || psu.Serial == "_" {
			log.WithFields(log.Fields{"operation": "chassis scan", "psu": psuPosition, "type": "chassis", "chassis_serial": chassis.Serial}).Error(errors.ErrInvalidSerial)
			continue
		}
		chassis.Psus = append(chassis.Psus, &Psu{
			Serial:        psu.Serial,
			CapacityKw:    psu.CapacityKw,
			PowerKw:       psu.PowerKw,
			Status:        psu.Status,
			ChassisSerial: c.Serial,
		})
	}

	return chassis
}

// Chassis contains all the chassis the information we will expose across different vendors
type Chassis struct {
	Serial          string          `json:"serial" gorm:"primary_key"`
	Name            string          `json:"name"`
	BmcAddress      string          `json:"bmc_address"`
	BmcSSHReachable bool            `json:"bmc_ssh_reachable"`
	BmcWEBReachable bool            `json:"bmc_web_reachable"`
	BmcAuth         bool            `json:"bmc_auth"`
	Blades          []*Blade        `json:"-" gorm:"ForeignKey:ChassisSerial"`
	FaultySlots     []int           `json:"faulty_slots" gorm:"type:int(2)[]"`
	StorageBlades   []*StorageBlade `json:"-" gorm:"ForeignKey:ChassisSerial"`
	Nics            []*Nic          `json:"-" gorm:"ForeignKey:ChassisSerial"`
	Psus            []*Psu          `json:"-" gorm:"ForeignKey:ChassisSerial"`
	TempC           int             `json:"temp_c"`
	PassThru        string          `json:"pass_thru"`
	Status          string          `json:"status"`
	PowerKw         float64         `json:"power_kw"`
	Model           string          `json:"model"`
	Vendor          string          `json:"vendor"`
	FwVersion       string          `json:"fw_version"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (c Chassis) GetID() string {
	return c.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (c Chassis) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToManyRelationship,
		},
		{
			Type:         "storage_blades",
			Name:         "storage_blades",
			Relationship: jsonapi.ToManyRelationship,
		},
		{
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		},
		{
			Type:         "psus",
			Name:         "psus",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (c Chassis) GetReferencedIDs() []jsonapi.ReferenceID {
	var result []jsonapi.ReferenceID
	for _, blade := range c.Blades {
		result = append(result, jsonapi.ReferenceID{
			ID:           blade.GetID(),
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	for _, storageBlade := range c.StorageBlades {
		result = append(result, jsonapi.ReferenceID{
			ID:           storageBlade.GetID(),
			Type:         "storage_blades",
			Name:         "storage_blades",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	for _, nic := range c.Nics {
		result = append(result, jsonapi.ReferenceID{
			ID:           nic.GetID(),
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	for _, psu := range c.Psus {
		result = append(result, jsonapi.ReferenceID{
			ID:           psu.GetID(),
			Type:         "psus",
			Name:         "psus",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	return result
}

// Diff compare to objects and return list of string with their differences
func (c *Chassis) Diff(chassis *Chassis) (differences []string) {
	if len(c.StorageBlades) != len(chassis.StorageBlades) {
		return []string{"Number of StorageBlades is different"}
	}

	if len(c.Blades) != len(chassis.Blades) {
		return []string{"Number of Blades is different"}
	}

	if len(c.Nics) != len(chassis.Nics) {
		return []string{"Number of Nics is different"}
	}

	sort.Sort(byStorageBladeSerial(c.StorageBlades))
	sort.Sort(byStorageBladeSerial(chassis.StorageBlades))

	sort.Sort(byBladeSerial(c.Blades))
	sort.Sort(byBladeSerial(chassis.Blades))

	for id := range c.Blades {
		sort.Sort(byMacAddress(c.Blades[id].Nics))
		sort.Sort(byMacAddress(chassis.Blades[id].Nics))
	}

	sort.Sort(byMacAddress(c.Nics))
	sort.Sort(byMacAddress(chassis.Nics))

	for _, diff := range pretty.Diff(c, chassis) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "TempC") {
			differences = append(differences, diff)
		}
	}

	return differences
}

// HasBlade checks whether a blade is connected to the chassis
func (c *Chassis) HasBlade(serial string) bool {
	for _, blade := range c.Blades {
		if blade.Serial == serial {
			return true
		}
	}
	return false
}

// HasStorageBlade checks whether a storageblade is connected to the chassis
func (c *Chassis) HasStorageBlade(serial string) bool {
	for _, storageBlade := range c.StorageBlades {
		if storageBlade.Serial == serial {
			return true
		}
	}
	return false
}

// HasNic checks whether a nic is connected to the chassis
func (c *Chassis) HasNic(macAddress string) bool {
	for _, nic := range c.Nics {
		if nic.MacAddress == macAddress {
			return true
		}
	}
	return false
}

// HasPsu checks whether a psu is connected to the chassis
func (c *Chassis) HasPsu(serial string) bool {
	for _, psu := range c.Psus {
		if psu.Serial == serial {
			return true
		}
	}
	return false
}
