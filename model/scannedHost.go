package model

import (
	"time"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedHost contains all ips and ports found by the scanner
type ScannedHost struct {
	IP        string `gorm:"primary_key"`
	CIDR      string `gorm:"column:cidr"`
	State     string
	Ports     []*ScannedPort `gorm:"ForeignKey:ScannedHostIP"`
	UpdatedAt time.Time
}

// GetName to satisfy jsonapi naming schema
func (s ScannedHost) GetName() string {
	return "scanned_hosts"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s ScannedHost) GetID() string {
	return s.IP
}
