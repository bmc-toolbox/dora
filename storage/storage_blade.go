package storage

import (
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
