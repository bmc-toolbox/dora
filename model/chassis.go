package model

import (
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// Chassis contains all the chassis the information we will expose across diferent vendors
type Chassis struct {
	Serial           string          `json:"serial" gorm:"primary_key"`
	Name             string          `json:"name"`
	BmcAddress       string          `json:"bmc_address"`
	BmcSSHReachable  bool            `json:"bmc_ssh_reachable"`
	BmcWEBReachable  bool            `json:"bmc_web_reachable"`
	BmcAuth          bool            `json:"bmc_auth"`
	Blades           []*Blade        `json:"-" gorm:"ForeignKey:ChassisSerial"`
	StorageBlades    []*StorageBlade `json:"-" gorm:"ForeignKey:ChassisSerial"`
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
	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (c Chassis) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for _, blade := range c.Blades {
		result = append(result, blade)
	}
	for _, storageBlade := range c.StorageBlades {
		result = append(result, storageBlade)
	}
	return result
}
