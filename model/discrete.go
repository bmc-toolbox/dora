package model

import (
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// Discrete contains all the blade information we will expose across different vendors
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
	Nics                 []*Nic    `json:"-" gorm:"ForeignKey:DiscreteSerial"`
	Model                string    `json:"model"`
	TempC                int       `json:"temp_c"`
	PowerKw              float64   `json:"power_kw"`
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
			Type:         "nics",
			Name:         "nics",
			Relationship: jsonapi.ToManyRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (d Discrete) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	for _, nic := range d.Nics {
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
func (d *Discrete) Diff(discrete *Discrete) (differences []string) {
	if len(d.Nics) != len(discrete.Nics) {
		return []string{"Number of Nics is different"}
	}

	sort.Sort(byMacAddress(d.Nics))
	sort.Sort(byMacAddress(discrete.Nics))

	for _, diff := range pretty.Diff(d, discrete) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "TempC") {
			differences = append(differences, diff)
		}
	}

	return differences
}
