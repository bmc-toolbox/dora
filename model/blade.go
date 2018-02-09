package model

import (
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
	"gitlab.booking.com/go/bmc/devices"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// NewBladeFromDevice will create a new object comming from the bmc blade devices
func NewBladeFromDevice(b *devices.Blade) (blade *Blade) {
	blade = &Blade{}
	blade.Name = b.Name
	blade.Serial = b.Serial
	blade.BiosVersion = b.BiosVersion
	blade.BmcType = b.BmcType
	blade.BmcAddress = b.BmcAddress
	blade.BmcVersion = b.BmcVersion
	blade.BmcLicenceType = b.BmcLicenceType
	blade.BmcLicenceStatus = b.BmcLicenceStatus
	for _, nic := range b.Nics {
		blade.Nics = make([]*Nic, 0)
		blade.Nics = append(blade.Nics, &Nic{
			MacAddress:  nic.MacAddress,
			Name:        nic.Name,
			BladeSerial: b.Serial,
		})
	}
	blade.BladePosition = b.BladePosition
	blade.Model = b.Model
	blade.TempC = b.TempC
	blade.PowerKw = b.PowerKw
	blade.Status = b.Status
	blade.Vendor = b.Vendor
	blade.ChassisSerial = b.ChassisSerial
	blade.Processor = b.Processor
	blade.ProcessorCount = b.ProcessorCount
	blade.ProcessorCoreCount = b.ProcessorCoreCount
	blade.ProcessorThreadCount = b.ProcessorThreadCount
	blade.Memory = b.Memory

	return blade
}

// Blade contains all the blade information we will expose across different vendors
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

	sort.Sort(byMacAddress(b.Nics))
	sort.Sort(byMacAddress(blade.Nics))

	for _, diff := range pretty.Diff(b, blade) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "TempC") {
			differences = append(differences, diff)
		}
	}

	return differences
}

type byBladeSerial []*Blade

func (b byBladeSerial) Len() int           { return len(b) }
func (b byBladeSerial) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byBladeSerial) Less(i, j int) bool { return b[i].Serial < b[j].Serial }
