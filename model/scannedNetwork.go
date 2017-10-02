package model

import (
	"strings"
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedHost contains all ips and ports found by the scanner
type ScannedNetwork struct {
	CIDR      string         `gorm:"primary_key;column:cidr" json:"cidr"`
	Hosts     []*ScannedHost `gorm:"ForeignKey:CIDR" json:"-"`
	Site      string         `json:"site"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GetName to satisfy jsonapi naming schema
func (s ScannedNetwork) GetName() string {
	return "scanned_networks"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s ScannedNetwork) GetID() string {
	return strings.Replace(s.CIDR, "/", "-", -1)
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (s ScannedNetwork) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "scanned_hosts",
			Name:         "scanned_hosts",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (s ScannedNetwork) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	for _, host := range s.Hosts {
		result = append(result, jsonapi.ReferenceID{
			ID:           host.GetID(),
			Type:         "scanned_hosts",
			Name:         "scanned_hosts",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	return result
}
