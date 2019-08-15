package model

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// NewDiscreteFromDevice will create a new object coming from the bmc discrete devices
func NewDiscreteFromDevice(d *devices.Discrete) (discrete *Discrete) {
	discrete = &Discrete{}
	discrete.Name = d.Name
	discrete.Serial = d.Serial
	discrete.BiosVersion = d.BiosVersion
	discrete.BmcType = d.BmcType
	discrete.BmcAddress = d.BmcAddress
	discrete.BmcVersion = d.BmcVersion
	discrete.BmcLicenceType = d.BmcLicenceType
	discrete.BmcLicenceStatus = d.BmcLicenceStatus
	discrete.Nics = make([]*Nic, 0)
	for _, nic := range d.Nics {
		discrete.Nics = append(discrete.Nics, &Nic{
			MacAddress:     nic.MacAddress,
			Name:           nic.Name,
			DiscreteSerial: d.Serial,
			Speed:          nic.Speed,
		})
	}
	discrete.Disks = make([]*Disk, 0)
	for pos, disk := range d.Disks {
		if disk.Serial == "" {
			disk.Serial = fmt.Sprintf("%s-failed-%d", discrete.Serial, pos)
		}
		discrete.Disks = append(discrete.Disks, &Disk{
			Serial:         disk.Serial,
			Size:           disk.Size,
			Status:         disk.Status,
			Model:          disk.Model,
			Location:       disk.Location,
			Type:           disk.Type,
			FwVersion:      disk.FwVersion,
			DiscreteSerial: discrete.Serial,
		})
	}
	discrete.Model = d.Model
	discrete.TempC = d.TempC
	discrete.PowerKw = d.PowerKw
	discrete.Status = d.Status
	discrete.Vendor = d.Vendor
	discrete.PowerState = d.PowerState
	discrete.Processor = d.Processor
	discrete.ProcessorCount = d.ProcessorCount
	discrete.ProcessorCoreCount = d.ProcessorCoreCount
	discrete.ProcessorThreadCount = d.ProcessorThreadCount
	discrete.Memory = d.Memory
	discrete.Psus = make([]*Psu, 0)
	for pos, psu := range d.Psus {
		if psu.Serial == "" {
			psu.Serial = fmt.Sprintf("%s-failed-%d", discrete.Serial, pos)
		}
		discrete.Psus = append(discrete.Psus, &Psu{
			Serial:         psu.Serial,
			CapacityKw:     psu.CapacityKw,
			PowerKw:        psu.PowerKw,
			Status:         psu.Status,
			DiscreteSerial: d.Serial,
		})
	}

	return discrete
}

// Discrete contains all the discrete information we will expose across different vendors
type Discrete struct {
	Serial               string    `json:"serial" gorm:"primary_key"`
	Name                 string    `json:"name"`
	BiosVersion          string    `json:"bios_version"`
	BmcType              string    `json:"bmc_type"`
	BmcAddress           string    `json:"bmc_address"`
	BmcVersion           string    `json:"bmc_version"`
	BmcSSHReachable      bool      `json:"bmc_ssh_reachable"`
	BmcWEBReachable      bool      `json:"bmc_web_reachable"`
	BmcIpmiReachable     bool      `json:"bmc_ipmi_reachable"`
	BmcLicenceType       string    `json:"bmc_licence_type"`
	BmcLicenceStatus     string    `json:"bmc_licence_status"`
	BmcAuth              bool      `json:"bmc_auth"`
	Disks                []*Disk   `json:"-" gorm:"ForeignKey:BladeSerial"`
	Nics                 []*Nic    `json:"-" gorm:"ForeignKey:DiscreteSerial"`
	Psus                 []*Psu    `json:"-" gorm:"ForeignKey:DiscreteSerial"`
	Model                string    `json:"model"`
	TempC                int       `json:"temp_c"`
	PowerKw              float64   `json:"power_kw"`
	PowerState           string    `json:"power_state"`
	Status               string    `json:"status"`
	Vendor               string    `json:"vendor"`
	Processor            string    `json:"processor"`
	ProcessorCount       int       `json:"processor_count"`
	ProcessorCoreCount   int       `json:"processor_core_count"`
	ProcessorThreadCount int       `json:"processor_thread_count"`
	Memory               int       `json:"memory_in_gb"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (d Discrete) GetID() string {
	return d.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (d Discrete) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "disks",
			Name:         "disks",
			Relationship: jsonapi.ToOneRelationship,
		},
		{
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		},
		{
			Type:         "psus",
			Name:         "psus",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (d Discrete) GetReferencedIDs() []jsonapi.ReferenceID {
	var result []jsonapi.ReferenceID
	for _, nic := range d.Nics {
		result = append(result, jsonapi.ReferenceID{
			ID:           nic.GetID(),
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		})
	}

	for _, psu := range d.Psus {
		result = append(result, jsonapi.ReferenceID{
			ID:           psu.GetID(),
			Type:         "psus",
			Name:         "psus",
			Relationship: jsonapi.ToManyRelationship,
		})
	}

	for _, disk := range d.Disks {
		result = append(result, jsonapi.ReferenceID{
			ID:           disk.GetID(),
			Type:         "disks",
			Name:         "disks",
			Relationship: jsonapi.ToManyRelationship,
		})
	}
	return result
}

// Diff compare to objects and return list of string with their differences
func (d *Discrete) Diff(discrete *Discrete) (differences []string) {
	if len(d.Nics) != len(discrete.Nics) {
		return []string{"Number of Nics is different"}
	}

	sort.Sort(byMacAddress(d.Nics))
	sort.Sort(byMacAddress(discrete.Nics))

	sort.Sort(byDiskSerial(d.Disks))
	sort.Sort(byDiskSerial(discrete.Disks))

	for _, diff := range pretty.Diff(d, discrete) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "TempC") {
			differences = append(differences, diff)
		}
	}

	return differences
}

// HasNic checks whether a nic is connected to the discrete
func (d *Discrete) HasNic(macAddress string) bool {
	for _, nic := range d.Nics {
		if nic.MacAddress == macAddress {
			return true
		}
	}
	return false
}

// HasPsu checks whether a psu is connected to the discrete
func (d *Discrete) HasPsu(serial string) bool {
	for _, psu := range d.Psus {
		if psu.Serial == serial {
			return true
		}
	}
	return false
}
