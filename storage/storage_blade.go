package storage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewBladeStorage initializes the storage
func NewBladeStorage(db *gorm.DB) *BladeStorage {
	return &BladeStorage{db}
}

// BladeStorage stores all of the tasty Blade, needs to be injected into
// Chassis and Blade Resource. In the real world, you would use a database for that.
type BladeStorage struct {
	db *gorm.DB
}

// GetAll of the Blades
func (b BladeStorage) GetAll() (blades []model.Blade, err error) {
	if err = b.db.Order("serial").Preload("Nics").Find(&blades).Error; err != nil {
		return blades, err
	}
	return blades, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (b BladeStorage) GetAllWithAssociations() (blades []model.Blade, err error) {
	if err = b.db.Preload("Nics").Find(&blades).Error; err != nil {
		return blades, err
	}
	return blades, err
}

// GetAllByFilters of the Blades by chassisID
func (b BladeStorage) GetAllByFilters(filters map[string][]string) (blades []model.Blade, err error) {
	query := ""
	for key, values := range filters {
		if len(values) == 1 && values[0] == "" {
			continue
		}
		blade := model.Blade{}
		rfct := reflect.ValueOf(blade)
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
			return blades, err
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
	b.db.Where(query).Find(&blades)
	return blades, nil
}

// GetAllByChassisID of the Blades by chassisID
func (b BladeStorage) GetAllByChassisID(serials []string) (blades []model.Blade, err error) {
	for _, serial := range serials {
		bls, err := b.getByChassisID(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return blades, err
		}
		blades = append(blades, bls...)
	}
	return blades, nil
}

func (b BladeStorage) getByChassisID(serial string) (blades []model.Blade, err error) {
	if err = b.db.Where("chassis_serial = ?", serial).Preload("Nics").Find(&blades).Error; err != nil {
		return blades, err
	}
	return blades, err
}

// GetAllByNicsID of the Blades by chassisID
func (b BladeStorage) GetAllByNicsID(macAddresses []string) (blades []model.Blade, err error) {
	for _, macAddress := range macAddresses {
		bls, err := b.getByCNicID(macAddress)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return blades, err
		}
		blades = append(blades, bls...)
	}
	return blades, nil
}

func (b BladeStorage) getByCNicID(macAddress string) (blades []model.Blade, err error) {
	if err = b.db.Joins("INNER JOIN nic ON nic.blade_serial = blade.serial").Where("nic.mac_address = ?", macAddress).Find(&blades).Error; err != nil {
		return blades, err
	}
	return blades, err
}

// GetOne  Blade
func (b BladeStorage) GetOne(serial string) (blade model.Blade, err error) {
	if err := b.db.Preload("Nics").Where("serial = ?", serial).First(&blade).Error; err != nil {
		return blade, err
	}
	return blade, err
}
