package model

import (
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// Disk represents a disk device
type Disk struct {
	Serial         string    `json:"serial" gorm:"primary_key"`
	Status         string    `json:"status"`
	Type           string    `json:"type"`
	Size           string    `json:"size"`
	Model          string    `json:"model"`
	UpdatedAt      time.Time `json:"updated_at"`
	BladeSerial    string    `json:"-"`
	DiscreteSerial string    `json:"-"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (d Disk) GetID() string {
	return d.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (d Disk) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "discretes",
			Name:         "discretes",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (d Disk) GetReferencedIDs() []jsonapi.ReferenceID {
	if d.BladeSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           d.BladeSerial,
				Type:         "blades",
				Name:         "blades",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	} else if d.DiscreteSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           d.DiscreteSerial,
				Type:         "discretes",
				Name:         "discretes",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	}
	return []jsonapi.ReferenceID{}
}

// Diff compare to objects and return list of string with their differences
func (d *Disk) Diff(disk *Disk) (differences []string) {
	for _, diff := range pretty.Diff(d, disk) {
		if !strings.Contains(diff, "UpdatedAt.") {
			differences = append(differences, diff)
		}
	}

	return differences
}

type byDiskSerial []*Disk

func (b byDiskSerial) Len() int           { return len(b) }
func (b byDiskSerial) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byDiskSerial) Less(i, j int) bool { return b[i].Serial < b[j].Serial }
