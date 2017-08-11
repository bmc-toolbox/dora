package model

import (
	"strconv"
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

// Blade contains all the blade information we will expose across diferent vendors
type Blade struct {
	ID             int64     `json:"-"`
	Serial         string    `json:"serial"`
	Name           string    `json:"name"`
	BiosVersion    string    `json:"bios_version"`
	BmcAddress     string    `json:"bmc_addres"`
	BmcVersion     string    `json:"bmc_version"`
	BmcType        string    `json:"bmc_type"`
	BladePosition  int       `json:"blade_position"`
	Temp           int       `json:"temp_c"`
	Power          float64   `json:"power_kw"`
	Status         string    `json:"status"`
	IsStorageBlade bool      `json:"is_storage_blade"`
	Vendor         string    `json:"vendor"`
	ChassisID      int64     `json:"chassis_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (b Blade) GetID() string {
	return strconv.FormatInt(b.ID, 10)
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (b Blade) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (b Blade) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{
		{
			ID:           strconv.FormatInt(b.ChassisID, 10),
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// Chassis contains all the chassis the information we will expose across diferent vendors
type Chassis struct {
	ID               int64     `json:"-"`
	Serial           string    `json:"serial"`
	Name             string    `json:"name"`
	Rack             string    `json:"rack"`
	Blades           []*Blade  `json:"-"`
	BladesIDS        []int64   `json:"-" sql:"-"`
	Temp             int       `json:"temp_c"`
	PowerSupplyCount int       `json:"power_supply_count"`
	PassThru         string    `json:"pass_thru"`
	Status           string    `json:"status"`
	Power            float64   `json:"power_kw"`
	Model            string    `json:"model"`
	Vendor           string    `json:"vendor"`
	FwVersion        string    `json:"fw_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (c Chassis) GetID() string {
	return strconv.FormatInt(c.ID, 10)
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (c Chassis) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "blades",
			Name:         "blades",
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
	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (c Chassis) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for _, blade := range c.Blades {
		result = append(result, blade)
	}

	return result
}
