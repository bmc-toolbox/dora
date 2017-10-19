package model

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedPort contains all ports found by the scanner
type ScannedPort struct {
	ID        string    `gorm:"primary_key" json:"-"`
	Site      string    `gorm:"unique_index:scanned_result" json:"site"`
	CIDR      string    `gorm:"unique_index:scanned_result;column:cidr" json:"cidr"`
	IP        string    `gorm:"unique_index:scanned_result" json:"ip"`
	Port      int       `gorm:"unique_index:scanned_result" json:"port"`
	Protocol  string    `gorm:"unique_index:scanned_result" json:"protocol"`
	ScannedBy string    `gorm:"unique_index:scanned_result" json:"scanned_by"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GenID generates the ID based on the date we have
func (s *ScannedPort) GenID() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%d-%s-%s", s.Site, s.CIDR, s.IP, s.Port, s.Protocol, s.ScannedBy))))
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
