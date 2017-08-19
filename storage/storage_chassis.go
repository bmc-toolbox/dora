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
func (c ChassisStorage) GetAll(offset string, limit string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Order("serial").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Order("serial").Find(&model.Chassis{}).Count(&count)
	} else {
		if err = c.db.Order("serial").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (c ChassisStorage) GetAllWithAssociations(offset string, limit string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Order("serial").Preload("Blades").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Order("serial").Find(&model.Chassis{}).Count(&count)
	} else {
		if err = c.db.Order("serial").Preload("Blades").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetOne Chassis
func (c ChassisStorage) GetOne(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", serial).Preload("Blades").First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetAllByFilters get all chassis detecting the struct members dinamically
func (c ChassisStorage) GetAllByFilters(offset string, limit string, filters map[string][]string) (count int, chassis []model.Chassis, err error) {
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
			return count, chassis, err
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

	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Where(query).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Where(query).Count(&count)
	} else {
		if err = c.db.Where(query).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}

	return count, chassis, err
}

// GetAllByBladesID Chassis
func (c ChassisStorage) GetAllByBladesID(offset string, limit string, serials []string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Joins("INNER JOIN blade ON blade.chassis_serial = chassis.serial").Where("blade.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Joins("INNER JOIN blade ON blade.chassis_serial = chassis.serial").Where("blade.serial in (?)", serials).Count(&count)
	} else {
		if err = c.db.Joins("INNER JOIN blade ON blade.chassis_serial = chassis.serial").Where("blade.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// UpdateOrCreate
func (c *ChassisStorage) UpdateOrCreate(chassis *model.Chassis) (serial string, err error) {
	if err = c.db.Save(&chassis).Error; err != nil {
		return serial, err
	}
	return chassis.Serial, nil
}
