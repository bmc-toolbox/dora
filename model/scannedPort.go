package model

import (
	"strconv"
	"time"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedPort contains all ports found by the scanner
type ScannedPort struct {
	ID            int    `gorm:"primary_key"`
	ScannedHostIP string `gorm:"unique_index:scanned_result"`
	Port          int    `gorm:"unique_index:scanned_result"`
	Protocol      string `gorm:"unique_index:scanned_result"`
	ScannedBy     string `gorm:"unique_index:scanned_result"`
	State         string
	UpdatedAt     time.Time
}

// GetName to satisfy jsonapi naming schema
func (s ScannedPort) GetName() string {
	return "scanned_ports"
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s ScannedPort) GetID() string {
	return strconv.Itoa(s.ID)
}
