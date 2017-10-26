package model

import (
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// StorageBlade contains all the storage blade information we will expose across diferent vendors
type StorageBlade struct {
	Serial        string    `json:"serial" gorm:"primary_key"`
	FwVersion     string    `json:"fw_version"`
	BladePosition int       `json:"blade_position"`
	Model         string    `json:"model"`
	TempC         int       `json:"temp_c"`
	PowerKw       float64   `json:"power_kw"`
	Status        string    `json:"status"`
	Vendor        string    `json:"vendor"`
	ChassisSerial string    `json:"-"`
	BladeSerial   string    `json:"-"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetName to satisfy jsonapi naming schema
func (s StorageBlade) GetName() string {
	return "storage_blades"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s StorageBlade) GetID() string {
	return s.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (s StorageBlade) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (s StorageBlade) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{
		{
			ID:           s.ChassisSerial,
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			ID:           s.BladeSerial,
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// Diff compare to objects and return list of string with their differences
func (s *StorageBlade) Diff(storageBlade *StorageBlade) (differences []string) {
	for _, diff := range pretty.Diff(s, storageBlade) {
		differences = append(differences, diff)
	}

	return differences
}
