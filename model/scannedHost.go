package model

import (
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedHost contains all ips and ports found by the scanner
type ScannedHost struct {
	IP        string         `gorm:"primary_key"  json:"ip"`
	CIDR      string         `gorm:"column:cidr"  json:"cidr"`
	State     string         `json:"state"`
	Ports     []*ScannedPort `gorm:"ForeignKey:ScannedHostIP"  json:"-"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GetName to satisfy jsonapi naming schema
func (s ScannedHost) GetName() string {
	return "scanned_hosts"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s ScannedHost) GetID() string {
	return s.IP
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (s ScannedHost) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "scanned_networks",
			Name:         "scanned_networks",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "scanned_ports",
			Name:         "scanned_ports",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (s ScannedHost) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{
		{
			ID:           s.CIDR,
			Type:         "scanned_networks",
			Name:         "scanned_networks",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
	for _, port := range s.Ports {
		result = append(result, jsonapi.ReferenceID{
			ID:           port.GetID(),
			Type:         "scanned_ports",
			Name:         "scanned_ports",
			Relationship: jsonapi.ToManyRelationship,
		})
	}

	return result
}
