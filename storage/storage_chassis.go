package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/hashicorp/go-multierror"
	"github.com/jinzhu/gorm"
)

// NewChassisStorage initializes the storage
func NewChassisStorage(db *gorm.DB) *ChassisStorage {
	return &ChassisStorage{db}
}

// ChassisStorage stores all Chassis
type ChassisStorage struct {
	db *gorm.DB
}

// Count get chassis count based on the filter
func (c ChassisStorage) Count(filters *filter.Filters) (count int, err error) {
	query, err := filters.BuildQuery(model.Chassis{})
	if err != nil {
		return count, err
	}

	err = c.db.Model(&model.Chassis{}).Where(query).Count(&count).Error
	return count, err
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
		if err = c.db.Order("serial asc").Preload("Blades").Preload("StorageBlades").Preload("Psus").Preload("Nics").Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Order("serial asc").Count(&count)
	} else {
		if err = c.db.Order("serial asc").Preload("Blades").Preload("StorageBlades").Preload("Psus").Preload("Nics").Find(&chassis).Error; err != nil {
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

// GetAllByPsusID retrieve chassis by psusID
func (c ChassisStorage) GetAllByPsusID(offset string, limit string, serials []string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Joins("INNER JOIN psu ON psu.chassis_serial = chassis.serial").Where("psu.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Joins("INNER JOIN psu ON psu.chassis_serial = chassis.serial").Where("psu.serial in (?)", serials).Count(&count)
	} else {
		if err = c.db.Joins("INNER JOIN psu ON psu.chassis_serial = chassis.serial").Where("psu.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
	}
	return count, chassis, err
}

// GetOne Chassis
func (c ChassisStorage) GetOne(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", serial).Preload("Blades").Preload("Blades.Nics").Preload("StorageBlades").Preload("Nics").Preload("Psus").First(&chassis).Error; err != nil {
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

// GetAllByFansID Chassis
func (c ChassisStorage) GetAllByFansID(offset string, limit string, serials []string) (count int, chassis []model.Chassis, err error) {
	if offset != "" && limit != "" {
		if err = c.db.Limit(limit).Offset(offset).Joins("INNER JOIN fan ON fan.chassis_serial = chassis.serial").Where("fan.serial in (?)", serials).Find(&chassis).Error; err != nil {
			return count, chassis, err
		}
		c.db.Model(&model.Chassis{}).Joins("INNER JOIN fan ON fan.chassis_serial = chassis.serial").Where("fan.serial in (?)", serials).Count(&count)
	} else {
		if err = c.db.Joins("INNER JOIN fan ON fan.chassis_serial = chassis.serial").Where("fan.serial in (?)", serials).Find(&chassis).Error; err != nil {
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

// UpdateOrCreate updates or create a new object
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

	if len(chassis.Blades) == 0 {
		if err = c.db.Model(&model.Blade{}).Where("serial is not null and chassis_serial = ?", chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
	} else {
		if err = c.db.Model(&model.Blade{}).Where("serial not in (?) and chassis_serial = ?", connectedSerials, chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
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

	if len(chassis.StorageBlades) == 0 {
		if err = c.db.Model(&model.StorageBlade{}).Where("serial is not null and chassis_serial = ?", chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
	} else {
		if err = c.db.Model(&model.StorageBlade{}).Where("serial not in (?) and chassis_serial = ?", connectedSerials, chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
	}

	if count > 0 {
		if err = c.db.Where("serial in (?) and chassis_serial = ?", serials, chassis.Serial).Delete(model.StorageBlade{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}

// RemoveOldNicRefs deletes all the old references from Nics that used to be inside of the chassis
func (c *ChassisStorage) RemoveOldNicRefs(chassis *model.Chassis) (count int, macAddresses []string, err error) {
	var connectedMacAddresses []string
	for _, nic := range chassis.Nics {
		connectedMacAddresses = append(connectedMacAddresses, nic.MacAddress)
	}

	if len(chassis.Nics) == 0 {
		if err = c.db.Model(&model.Nic{}).Where("mac_address is not null and chassis_serial = ?", chassis.Serial).Pluck("mac_address", &macAddresses).Count(&count).Error; err != nil {
			return count, macAddresses, err
		}
	} else {
		if err = c.db.Model(&model.Nic{}).Where("mac_address not in (?) and chassis_serial = ?", connectedMacAddresses, chassis.Serial).Pluck("mac_address", &macAddresses).Count(&count).Error; err != nil {
			return count, macAddresses, err
		}
	}

	if count > 0 {
		if err = c.db.Where("mac_address in (?) and chassis_serial = ?", macAddresses, chassis.Serial).Delete(model.Nic{}).Error; err != nil {
			return count, macAddresses, err
		}
	}

	return count, macAddresses, err
}

// RemoveOldPsuRefs deletes all the old references from Psus that used to be inside of the chassis
func (c *ChassisStorage) RemoveOldPsuRefs(chassis *model.Chassis) (count int, serials []string, err error) {
	var connectedSerials []string
	for _, psu := range chassis.Psus {
		connectedSerials = append(connectedSerials, psu.Serial)
	}

	if len(chassis.Psus) == 0 {
		if err = c.db.Model(&model.Psu{}).Where("serial is not null and chassis_serial = ?", chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
	} else {
		if err = c.db.Model(&model.Psu{}).Where("serial not in (?) and chassis_serial = ?", connectedSerials, chassis.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
			return count, serials, err
		}
	}

	if count > 0 {
		if err = c.db.Where("serial in (?) and chassis_serial = ?", serials, chassis.Serial).Delete(model.Psu{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}

// RemoveOldRefs deletes all the old references from all attached components
func (c *ChassisStorage) RemoveOldRefs(chassis *model.Chassis) (err error) {
	var merror *multierror.Error
	_, _, err = c.RemoveOldPsuRefs(chassis)
	merror = multierror.Append(merror, err)
	_, _, err = c.RemoveOldStorageBladesRefs(chassis)
	merror = multierror.Append(merror, err)
	_, _, err = c.RemoveOldNicRefs(chassis)
	merror = multierror.Append(merror, err)
	_, _, err = c.RemoveOldBladesRefs(chassis)
	merror = multierror.Append(merror, err)
	return merror.ErrorOrNil()
}
