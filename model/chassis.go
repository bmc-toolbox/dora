package model

import (
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// Chassis contains all the chassis the information we will expose across different vendors
type Chassis struct {
	Serial           string          `json:"serial" gorm:"primary_key"`
	Name             string          `json:"name"`
	BmcAddress       string          `json:"bmc_address"`
	BmcSSHReachable  bool            `json:"bmc_ssh_reachable"`
	BmcWEBReachable  bool            `json:"bmc_web_reachable"`
	BmcAuth          bool            `json:"bmc_auth"`
	Blades           []*Blade        `json:"-" gorm:"ForeignKey:ChassisSerial"`
	StorageBlades    []*StorageBlade `json:"-" gorm:"ForeignKey:ChassisSerial"`
	Nics             []*Nic          `json:"-" gorm:"ForeignKey:ChassisSerial"`
	TempC            int             `json:"temp_c"`
	PowerSupplyCount int             `json:"power_supply_count"`
	PassThru         string          `json:"pass_thru"`
	Status           string          `json:"status"`
	PowerKw          float64         `json:"power_kw"`
	Model            string          `json:"model"`
	Vendor           string          `json:"vendor"`
	FwVersion        string          `json:"fw_version"`
	UpdatedAt        time.Time       `json:"updated_at"`
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
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (c Chassis) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
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

// // GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
// func (c Chassis) GetReferencedStructs() []jsonapi.MarshalIdentifier {
// 	result := []jsonapi.MarshalIdentifier{}
// 	result = append(result, blade...)
// 	result = append(result, storageBlade...)
// 	result = append(result, nic...)
// 	return result
// }
