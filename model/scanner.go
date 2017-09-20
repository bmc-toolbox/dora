package model

import (
	"time"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// ScannedHost contains all ips and ports found by the scanner
type ScannedNetwork struct {
	CIDR      string        `gorm:"primary_key;column:cidr" json:"cidr"`
	Hosts     []ScannedHost `gorm:"ForeignKey:CIDR"`
	Site      string
	UpdatedAt time.Time
}

// ScannedHost contains all ips and ports found by the scanner
type ScannedHost struct {
	IP        string `gorm:"primary_key"`
	CIDR      string `gorm:"column:cidr"`
	State     string
	Ports     []ScannedPort `gorm:"ForeignKey:ScannedHostIP"`
	UpdatedAt time.Time
}

// ScannedPort contains all ports found by the scanner
type ScannedPort struct {
	ScannedHostIP string `gorm:"primary_key"`
	Port          int    `gorm:"primary_key"`
	Protocol      string `gorm:"primary_key"`
	ScannedBy     string `gorm:"primary_key"`
	State         string
	UpdatedAt     time.Time
}
