package model

import (
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/manyminds/api2go/jsonapi"
)

/* READ THIS BEFORE CHANGING THE SCHEMA

To make the magic of dynamic filtering work, we need to define each json field matching the database column name

*/

// Fan contains the network information of the cards attached to blades or chassis
type Fan struct {
	Serial        string    `json:"serial" gorm:"primary_key"`
	Status        string    `json:"status"`
	Position      int       `json:"position"`
	Model         string    `json:"model"`
	CurrentRPM    int64     `json:"current_rpm"`
	PowerKw       float64   `json:"power_kw"`
	ChassisSerial string    `json:"-"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (p Fan) GetID() string {
	return p.Serial
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (p Fan) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type:         "chassis",
			Name:         "chassis",
			Relationship: jsonapi.ToOneRelationship,
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (p Fan) GetReferencedIDs() []jsonapi.ReferenceID {
	 if p.ChassisSerial != "" {
		return []jsonapi.ReferenceID{
			{
				ID:           p.ChassisSerial,
				Type:         "chassis",
				Name:         "chassis",
				Relationship: jsonapi.ToOneRelationship,
			},
		}
	}
	return []jsonapi.ReferenceID{}
}

// Diff compare to objects and return list of string with their differences
func (p *Fan) Diff(fan *Fan) (differences []string) {
	for _, diff := range pretty.Diff(p, fan) {
		if !strings.Contains(diff, "UpdatedAt.") && !strings.Contains(diff, "PowerKw") && !strings.Contains(diff, "CurrentRPM"){
			differences = append(differences, diff)
		}
	}

	return differences
}

type byFanSerial []*Fan

func (b byFanSerial) Len() int           { return len(b) }
func (b byFanSerial) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byFanSerial) Less(i, j int) bool { return b[i].Serial < b[j].Serial }
