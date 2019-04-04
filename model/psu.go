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

// Psu contains the network information of the cards attached to blades or chassis
type Psu struct {
	Serial         string    `json:"serial" gorm:"primary_key"`
	CapacityKw     float64   `json:"capacity_kw"`
	PowerKw        float64   `json:"power_kw"`
	Status         string    `json:"status"`
	PartNumber     string    `json:"part_number"`
	UpdatedAt      time.Time `json:"updated_at"`
	DiscreteSerial string    `json:"-"`
	ChassisSerial  string    `json:"-"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (p Psu) GetID() string {
	return p.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (p Psu) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "discretes",
			Name:         "discretes",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (p Psu) GetReferencedIDs() []jsonapi.ReferenceID {
	if p.DiscreteSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           p.DiscreteSerial,
				Type:         "discretes",
				Name:         "discretes",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	} else if p.ChassisSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           p.ChassisSerial,
				Type:         "chassis",
				Name:         "chassis",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	}
	return []jsonapi.ReferenceID{}
}

// Diff compare to objects and return list of string with their differences
func (p *Psu) Diff(psu *Psu) (differences []string) {
	for _, diff := range pretty.Diff(p, psu) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "TempC") {
			differences = append(differences, diff)
		}
	}

	return differences
}

type byPsuSerial []*Psu

func (b byPsuSerial) Len() int           { return len(b) }
func (b byPsuSerial) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byPsuSerial) Less(i, j int) bool { return b[i].Serial < b[j].Serial }
