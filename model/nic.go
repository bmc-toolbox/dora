package model

import (
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// Nic contains the network information of the cards attached to blades or chassis
type Nic struct {
	MacAddress     string    `json:"mac_address" gorm:"primary_key"`
	Name           string    `json:"name"`
	UpdatedAt      time.Time `json:"updated_at"`
	BladeSerial    string    `json:"-"`
	DiscreteSerial string    `json:"-"`
	ChassisSerial  string    `json:"-"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (n Nic) GetID() string {
	return n.MacAddress
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (n Nic) GetReferences() []jsonapi.Reference {
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
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (n Nic) GetReferencedIDs() []jsonapi.ReferenceID {
	if n.BladeSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           n.BladeSerial,
				Type:         "blades",
				Name:         "blades",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	} else if n.DiscreteSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           n.DiscreteSerial,
				Type:         "discretes",
				Name:         "discretes",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	} else if n.ChassisSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           n.ChassisSerial,
				Type:         "chassis",
				Name:         "chassis",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	}
	return []jsonapi.ReferenceID{}
}

// Diff compare to objects and return list of string with their differences
func (n *Nic) Diff(nic *Nic) (differences []string) {
	for _, diff := range pretty.Diff(n, nic) {
		if !strings.Contains(diff, "UpdatedAt.") {
			differences = append(differences, diff)
		}
	}

	return differences
}

type byMacAddress []*Nic

func (b byMacAddress) Len() int           { return len(b) }
func (b byMacAddress) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMacAddress) Less(i, j int) bool { return b[i].MacAddress < b[j].MacAddress }
