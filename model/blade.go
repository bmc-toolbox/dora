package model

import (
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database collumn name

*/

// Blade contains all the blade information we will expose across diferent vendors
type Blade struct {
	Serial               string       `json:"serial" gorm:"primary_key"`
	Name                 string       `json:"name"`
	BiosVersion          string       `json:"bios_version"`
	BmcType              string       `json:"bmc_type"`
	BmcAddress           string       `json:"bmc_address"`
	BmcVersion           string       `json:"bmc_version"`
	BmcSSHReachable      bool         `json:"bmc_ssh_reachable"`
	BmcWEBReachable      bool         `json:"bmc_web_reachable"`
	BmcIpmiReachable     bool         `json:"bmc_ipmi_reachable"`
	BmcLicenceType       string       `json:"bmc_licence_type"`
	BmcLicenceStatus     string       `json:"bmc_licence_status"`
	BmcAuth              bool         `json:"bmc_auth"`
	Nics                 []*Nic       `json:"-" gorm:"ForeignKey:BladeSerial"`
	BladePosition        int          `json:"blade_position"`
	Model                string       `json:"model"`
	TempC                int          `json:"temp_c"`
	PowerKw              float64      `json:"power_kw"`
	Status               string       `json:"status"`
	Vendor               string       `json:"vendor"`
	ChassisSerial        string       `json:"-"`
	Processor            string       `json:"processor"`
	ProcessorCount       int          `json:"processor_count"`
	ProcessorCoreCount   int          `json:"processor_core_count"`
	ProcessorThreadCount int          `json:"processor_thread_count"`
	StorageBlade         StorageBlade `json:"-" gorm:"ForeignKey:BladeSerial"`
	Memory               int          `json:"memory_in_gb"`
	UpdatedAt            time.Time    `json:"updated_at"`
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
		{
			Type:         "storage_blades",
			Name:         "storage_blades",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (b Blade) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}

	if b.ChassisSerial != "" {
		result = append(result, jsonapi.ReferenceID{
			ID:           b.ChassisSerial,
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		})
	}

	if b.StorageBlade.Serial != "" {
		result = append(result, jsonapi.ReferenceID{
			ID:           b.StorageBlade.Serial,
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		})
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

// Diff compare to objects and return list of string with their differences
func (b *Blade) Diff(blade *Blade) (differences []string) {
	if len(b.Nics) != len(blade.Nics) {
		return []string{"Number of Nics is different"}
	}

	sort.Slice(b.Nics, func(i, j int) bool {
		switch strings.Compare(b.Nics[i].MacAddress, b.Nics[j].MacAddress) {
		case -1:
			return true
		case 1:
			return false
		}
		return b.Nics[i].MacAddress > b.Nics[j].MacAddress
	})

	sort.Slice(blade.Nics, func(i, j int) bool {
		switch strings.Compare(blade.Nics[i].MacAddress, blade.Nics[j].MacAddress) {
		case -1:
			return true
		case 1:
			return false
		}
		return blade.Nics[i].MacAddress > blade.Nics[j].MacAddress
	})

	for _, diff := range pretty.Diff(b, blade) {
		differences = append(differences, diff)
	}

	return differences
}
