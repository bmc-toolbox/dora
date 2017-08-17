package storage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewChassisStorage initializes the storage
func NewChassisStorage(db *gorm.DB) *ChassisStorage {
	return &ChassisStorage{db}
}

// ChassisStorage stores all Chassiss
type ChassisStorage struct {
	db *gorm.DB
}

// GetAll returns all chassis
func (c ChassisStorage) GetAll() (chassis []model.Chassis, err error) {
	if err = c.db.Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (c ChassisStorage) GetAllWithAssociations() (chassis []model.Chassis, err error) {
	if err = c.db.Preload("Blades").Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetOne Chassis
func (c ChassisStorage) GetOne(serial string) (chassis model.Chassis, err error) {
	chassis, err = c.getOneWithAssociations(serial)
	return chassis, err
}

// GetAllByFilters get all chassis detecting the struct members dinamically
func (c ChassisStorage) GetAllByFilters(filters map[string][]string) (chassis []model.Chassis, err error) {
	query := ""
	for key, values := range filters {
		if len(values) == 1 && values[0] == "" {
			continue
		}
		ch := model.Chassis{}
		rfct := reflect.ValueOf(ch)
		rfctType := rfct.Type()

		var structMemberName string
		var structJSONMemberName string
		for i := 0; i < rfctType.NumField(); i++ {
			jsondName := rfctType.Field(i).Tag.Get("json")
			if key == jsondName {
				structMemberName = rfctType.Field(i).Name
				structJSONMemberName = jsondName
				break
			}
		}

		if structJSONMemberName == "" || structJSONMemberName == "-" {
			return chassis, err
		}

		ftype := reflect.Indirect(rfct).FieldByName(structMemberName)
		switch ftype.Kind() {
		case reflect.String:
			if query == "" {
				query = fmt.Sprintf("%s in ('%s')", structJSONMemberName, strings.Join(values, "', '"))
			} else {
				query = fmt.Sprintf("%s and %s in ('%s')", query, structJSONMemberName, strings.Join(values, "', '"))
			}
		case reflect.Bool:
			if query == "" {
				query = fmt.Sprintf("%s in (%s)", structJSONMemberName, strings.Join(values, ", "))
			} else {
				query = fmt.Sprintf("%s and %s in (%s)", query, structJSONMemberName, strings.Join(values, ", "))
			}
		case reflect.Int:
			if query == "" {
				query = fmt.Sprintf("%s in (%s)", structJSONMemberName, strings.Join(values, ", "))
			} else {
				query = fmt.Sprintf("%s and %s in (%s)", query, structJSONMemberName, strings.Join(values, ", "))
			}
		}
	}
	c.db.Where(query).Find(&chassis)
	return chassis, nil
}

// GetAllByBladesID Chassis
func (c ChassisStorage) GetAllByBladesID(serials []string) (chassis []model.Chassis, err error) {
	for _, serial := range serials {
		ch, err := c.getByBladeID(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return chassis, err
		}
		chassis = append(chassis, ch)
	}
	return chassis, nil
}

func (c ChassisStorage) getByBladeID(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Joins("INNER JOIN blade ON blade.chassis_serial = chassis.serial").Where("blade.serial = ?", serial).Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

func (c ChassisStorage) getOneWithAssociations(id string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", id).Preload("Blades").First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// UpdateOrCreate
func (c *ChassisStorage) UpdateOrCreate(chassis *model.Chassis) (serial string, err error) {
	if err = c.db.Save(&chassis).Error; err != nil {
		return serial, err
	}
	return chassis.Serial, nil
}
