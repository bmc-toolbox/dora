package model

import (
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/manyminds/api2go/jsonapi"
)

// Nic contains the network information of the cards attached to blades or chassis
type Nic struct {
	MacAddress  string    `json:"mac_address" gorm:"primary_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	BladeSerial string    `json:"-"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (n Nic) GetID() string {
	return n.MacAddress
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (n Nic) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "blade",
			Name:         "blade",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (n Nic) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{
		{
			ID:           n.BladeSerial,
			Type:         "blade",
			Name:         "blade",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// Blade contains all the blade information we will expose across diferent vendors
type Blade struct {
	Serial         string    `json:"serial" gorm:"primary_key"`
	Name           string    `json:"name"`
	BiosVersion    string    `json:"bios_version"`
	BmcAddress     string    `json:"bmc_address"`
	BmcVersion     string    `json:"bmc_version"`
	BmcSSH         bool      `json:"bmc_ssh_status"`
	BmcWEB         bool      `json:"bmc_web_status"`
	BmcIPMI        bool      `json:"bmc_ipmi_status"`
	BmcAuth        bool      `json:"bmc_auth"`
	Nics           []*Nic    `json:"-" gorm:"ForeignKey:BladeSerial"`
	NicsIDs        []int64   `json:"-" sql:"-"`
	BladePosition  int       `json:"blade_position"`
	Model          string    `json:"model"`
	Temp           int       `json:"temp_c"`
	Power          float64   `json:"power_kw"`
	Status         string    `json:"status"`
	IsStorageBlade bool      `json:"is_storage_blade"`
	Vendor         string    `json:"vendor"`
	ChassisSerial  string    `json:"-"`
	Processor      string    `json:"proc"`
	Memory         int       `json:"memory_in_gb"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TestConnections as the name says, test connections from the bkbuild machines to the bmcs and update the struct data
func (b *Blade) TestConnections() {
	if b.IsStorageBlade == true || b.BmcAddress == "0.0.0.0" {
		return
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.BmcAddress, 443), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test http connection", "ip": b.BmcAddress, "serial": b.Serial, "type": "blade", "error": err, "blade": b.Name, "vendor": b.Vendor}).Error("Auditing blade")
	} else {
		b.BmcWEB = true
		conn.Close()
	}

	conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.BmcAddress, 22), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test ssh connection", "ip": b.BmcAddress, "serial": b.Serial, "type": "blade", "error": err, "blade": b.Name, "vendor": b.Vendor}).Error("Auditing blade")
	} else {
		b.BmcSSH = true
		conn.Close()
	}

	conn, err = net.DialTimeout("udp", fmt.Sprintf("%s:%d", b.BmcAddress, 161), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test ipmi connection", "ip": b.BmcAddress, "serial": b.Serial, "type": "blade", "error": err, "blade": b.Name, "vendor": b.Vendor}).Error("Auditing blade")
	} else {
		b.BmcIPMI = true
		conn.Close()
	}
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (b Blade) GetID() string {
	return b.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (b Blade) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (b Blade) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{
		{
			ID:           b.ChassisSerial,
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
	for _, nic := range b.Nics {
		result = append(result, jsonapi.ReferenceID{
			ID:           nic.GetID(),
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		})
	}

	return result
}

// Chassis contains all the chassis the information we will expose across diferent vendors
type Chassis struct {
	Serial           string    `json:"serial" gorm:"primary_key"`
	Name             string    `json:"name"`
	Rack             string    `json:"rack"`
	BmcAddress       string    `json:"bmc_address"`
	BmcSSH           bool      `json:"bmc_ssh_status"`
	BmcWEB           bool      `json:"bmc_web_status"`
	BmcIPMI          bool      `json:"bmc_ipmi_status"`
	Blades           []*Blade  `json:"-" gorm:"ForeignKey:ChassisSerial"`
	BladesIDS        []int64   `json:"-" sql:"-"`
	Temp             int       `json:"temp_c"`
	PowerSupplyCount int       `json:"power_supply_count"`
	PassThru         string    `json:"pass_thru"`
	Status           string    `json:"status"`
	Power            float64   `json:"power_kw"`
	Model            string    `json:"model"`
	Vendor           string    `json:"vendor"`
	FwVersion        string    `json:"fw_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TestConnections as the name says, test connections from the bkbuild machines to the bmcs and update the struct data
func (c *Chassis) TestConnections() {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.BmcAddress, 443), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test http connection", "ip": c.BmcAddress, "serial": c.Serial, "type": "blade", "error": err, "chassis": c.Name, "vendor": c.Vendor}).Error("Auditing chassis")
	} else {
		c.BmcWEB = true
		conn.Close()
	}

	conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.BmcAddress, 22), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test ssh connection", "ip": c.BmcAddress, "serial": c.Serial, "type": "blade", "error": err, "chassis": c.Name, "vendor": c.Vendor}).Error("Auditing chassis")
	} else {
		c.BmcSSH = true
		conn.Close()
	}

	conn, err = net.DialTimeout("udp", fmt.Sprintf("%s:%d", c.BmcAddress, 161), 15*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"operation": "test ipmi connection", "ip": c.BmcAddress, "serial": c.Serial, "type": "blade", "error": err, "chassis": c.Name, "vendor": c.Vendor}).Error("Auditing chassis")
	} else {
		c.BmcIPMI = true
		conn.Close()
	}
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (c Chassis) GetID() string {
	return c.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (c Chassis) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (c Chassis) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	for _, blade := range c.Blades {
		result = append(result, jsonapi.ReferenceID{
			ID:           blade.GetID(),
			Type:         "blades",
			Name:         "blades",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (c Chassis) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for _, blade := range c.Blades {
		result = append(result, blade)
	}

	return result
}
