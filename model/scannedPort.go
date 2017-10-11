package model

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedPort contains all ports found by the scanner
type ScannedPort struct {
	ID            string `gorm:"primary_key"`
	ScannedHostIP string `gorm:"unique_index:scanned_result"`
	Port          int    `gorm:"unique_index:scanned_result"`
	Protocol      string `gorm:"unique_index:scanned_result"`
	ScannedBy     string `gorm:"unique_index:scanned_result"`
	State         string
	UpdatedAt     time.Time
}

// GenID generates the ID based on the date we have
func (s *ScannedPort) GenID() string {
	fmt.Println(fmt.Sprintf("%s-%d-%s-%s", s.ScannedHostIP, s.Port, s.Protocol, s.ScannedBy))
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d-%s-%s", s.ScannedHostIP, s.Port, s.Protocol, s.ScannedBy))))
}

// BeforeCreate run all operations before creating the object
func (s *ScannedPort) BeforeCreate(scope *gorm.Scope) (err error) {
	scope.SetColumn("ID", s.GenID())
	return nil
}

// GetName to satisfy jsonapi naming schema
func (s ScannedPort) GetName() string {
	return "scanned_ports"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s ScannedPort) GetID() string {
	return s.ID
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (s ScannedPort) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "scanned_hosts",
			Name:         "scanned_hosts",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (s ScannedPort) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{
		{
			ID:           s.ScannedHostIP,
			Type:         "scanned_networks",
			Name:         "scanned_networks",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
	return result
}
