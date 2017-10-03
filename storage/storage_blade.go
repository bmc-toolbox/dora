package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/filter"
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
func (b BladeStorage) GetAll(offset string, limit string) (count int, blades []model.Blade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Order("serial asc").Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Model(&model.Blade{}).Order("serial asc").Count(&count)
	} else {
		if err = b.db.Order("serial").Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}
	return count, blades, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (b BladeStorage) GetAllWithAssociations(offset string, limit string) (count int, blades []model.Blade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Order("serial asc").Preload("Nics").Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Order("serial asc").Find(&model.Blade{}).Count(&count)
	} else {
		if err = b.db.Order("serial").Preload("Nics").Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}
	return count, blades, err
}

// GetAllByFilters get all blades based on the filter
func (b BladeStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, blades []model.Blade, err error) {
	query, err := filters.BuildQuery(model.Blade{})
	if err != nil {
		return count, blades, err
	}

	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Where(query).Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Model(&model.Blade{}).Where(query).Count(&count)
	} else {
		if err = b.db.Where(query).Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}

	return count, blades, nil
}

// GetAllByChassisID retrieve Blades by chassisID
func (b BladeStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, blades []model.Blade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Preload("Nics").Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Model(&model.Blade{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = b.db.Where("chassis_serial in (?)", serials).Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}
	return count, blades, err
}

// GetAllByNicsID retrieve Blades by nicsID
func (b BladeStorage) GetAllByNicsID(offset string, limit string, macAddresses []string) (count int, blades []model.Blade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Joins("INNER JOIN nic ON nic.blade_serial = blade.serial").Where("nic.mac_address in (?)", macAddresses).Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Model(&model.Blade{}).Joins("INNER JOIN nic ON nic.blade_serial = blade.serial").Where("nic.mac_address in (?)", macAddresses).Count(&count)
	} else {
		if err = b.db.Joins("INNER JOIN nic ON nic.blade_serial = blade.serial").Where("nic.mac_address in (?)", macAddresses).Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}
	return count, blades, err
}

// GetAllByStorageBladesID retrieves Blades by StorageBladesID
func (b BladeStorage) GetAllByStorageBladesID(offset string, limit string, serials []string) (count int, blades []model.Blade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Joins("INNER JOIN storage_blade ON storage_blade.blade_serial = blade.serial").Where("storage_blade.serial in (?)", serials).Find(&blades).Error; err != nil {
			return count, blades, err
		}
		b.db.Model(&model.Blade{}).Joins("INNER JOIN storage_blade ON storage_blade.blade_serial = blade.serial").Where("storage_blade.serial in (?)", serials).Count(&count)
	} else {
		if err = b.db.Joins("INNER JOIN storage_blade ON storage_blade.blade_serial = blade.serial").Where("storage_blade.serial in (?)", serials).Find(&blades).Error; err != nil {
			return count, blades, err
		}
	}
	return count, blades, err
}

// GetOne  Blade
func (b BladeStorage) GetOne(serial string) (blade model.Blade, err error) {
	if err := b.db.Preload("Nics").Where("serial = ?", serial).First(&blade).Error; err != nil {
		return blade, err
	}
	return blade, err
}

// UpdateOrCreate
func (b *BladeStorage) UpdateOrCreate(blade *model.Blade) (serial string, err error) {
	if err = b.db.Save(&blade).Error; err != nil {
		return serial, err
	}
	return blade.Serial, nil
}
