package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/filter"
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
		if err = c.db.Limit(limit).Offset(offset).Order("serial asc").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Order("serial asc").Count(&count)
	} else {
		if err = c.db.Order("serial asc").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetAllWithAssociations returns all Chassis with their relationships
func (c ChassisStorage) GetAllWithAssociations(offset string, limit string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Order("serial asc").Preload("Blades").Preload("StorageBlades").Preload("Nics").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Order("serial asc").Count(&count)
	} else {
		if err = c.db.Order("serial asc").Preload("Blades").Preload("StorageBlades").Preload("Nics").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetAllByNicsID retrieve chassis by nicsID
func (c ChassisStorage) GetAllByNicsID(offset string, limit string, macAddresses []string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Joins("INNER JOIN nic ON nic.chassis_serial = chassis.serial").Where("nic.mac_address in (?)", macAddresses).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Joins("INNER JOIN nic ON nic.chassis_serial = chassis.serial").Where("nic.mac_address in (?)", macAddresses).Count(&count)
	} else {
		if err = c.db.Joins("INNER JOIN nic ON nic.chassis_serial = chassis.serial").Where("nic.mac_address in (?)", macAddresses).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetOne Chassis
func (c ChassisStorage) GetOne(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", serial).Preload("Blades").Preload("Blades.Nics").Preload("StorageBlades").Preload("Nics").First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetAllByFilters get all Chassis based on the filter
func (c ChassisStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, chassis []model.Chassis, err error) {
	query, err := filters.BuildQuery(model.Chassis{})
	if err != nil {
		return count, chassis, err
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

// GetAllByStorageBladesID Chassis
func (c ChassisStorage) GetAllByStorageBladesID(offset string, limit string, serials []string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Joins("INNER JOIN storage_blade ON storage_blade.chassis_serial = chassis.serial").Where("storage_blade.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Joins("INNER JOIN storage_blade ON storage_blade.chassis_serial = chassis.serial").Where("storage_blade.serial in (?)", serials).Count(&count)
	} else {
		if err = c.db.Joins("INNER JOIN storage_blade ON storage_blade.chassis_serial = chassis.serial").Where("storage_blade.serial in (?)", serials).Find(&chassis).Error; err != nil {
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

// RemoveOldBladesRefs deletes all the old references from StorageBlades that used to be inside of the chassis
func (c *ChassisStorage) RemoveOldBladesRefs(chassis *model.Chassis) (count int, serials []string, err error) {
	var connectedSerials []string
	for _, blade := range chassis.Blades {
		connectedSerials = append(connectedSerials, blade.Serial)
	}

	if err = c.db.Model(&model.Blade{}).Where("serial not in (?) and chassis_serial = ?", connectedSerials, chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
		return count, serials, err
	}

	if count > 0 {
		if err = c.db.Where("serial in (?) and chassis_serial = ?", serials, chassis.Serial).Delete(model.Blade{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}

// RemoveOldStorageBladesRefs deletes all the old references from StorageBlades that used to be inside of the chassis
func (c *ChassisStorage) RemoveOldStorageBladesRefs(chassis *model.Chassis) (count int, serials []string, err error) {
	var connectedSerials []string
	for _, blade := range chassis.StorageBlades {
		connectedSerials = append(connectedSerials, blade.Serial)
	}

	if err = c.db.Model(&model.StorageBlade{}).Where("serial not in (?) and chassis_serial = ?", connectedSerials, chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
		return count, serials, err
	}

	if count > 0 {
		if err = c.db.Where("serial in (?) and chassis_serial = ?", serials, chassis.Serial).Delete(model.StorageBlade{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}
