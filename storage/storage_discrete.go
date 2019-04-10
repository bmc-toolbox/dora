package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/hashicorp/go-multierror"
	"github.com/jinzhu/gorm"
)

// NewDiscreteStorage initializes the storage
func NewDiscreteStorage(db *gorm.DB) *DiscreteStorage {
	return &DiscreteStorage{db}
}

// DiscreteStorage stores all of the tasty Discrete, needs to be injected into
// Chassis and Discrete Resource. In the real world, you would use a database for that.
type DiscreteStorage struct {
	db *gorm.DB
}

// Count get discretes count based on the filter
func (d DiscreteStorage) Count(filters *filter.Filters) (count int, err error) {
	query, err := filters.BuildQuery(model.Discrete{})
	if err != nil {
		return count, err
	}

	err = d.db.Model(&model.Discrete{}).Where(query).Count(&count).Error
	return count, err
}

// GetAll of the Discretes
func (d DiscreteStorage) GetAll(offset string, limit string) (count int, discretes []model.Discrete, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Order("serial asc").Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
		d.db.Model(&model.Discrete{}).Order("serial asc").Count(&count)
	} else {
		if err = d.db.Order("serial").Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
	}
	return count, discretes, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (d DiscreteStorage) GetAllWithAssociations(offset string, limit string) (count int, discretes []model.Discrete, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Order("serial asc").Preload("Nics").Preload("Disks").Preload("Psus").Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
		d.db.Order("serial asc").Find(&model.Discrete{}).Count(&count)
	} else {
		if err = d.db.Order("serial").Preload("Nics").Preload("Disks").Preload("Psus").Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
	}
	return count, discretes, err
}

// GetAllByFilters get all discretes based on the filter
func (d DiscreteStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, discretes []model.Discrete, err error) {
	query, err := filters.BuildQuery(model.Discrete{})
	if err != nil {
		return count, discretes, err
	}

	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Where(query).Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
		d.db.Model(&model.Discrete{}).Where(query).Count(&count)
	} else {
		if err = d.db.Where(query).Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
	}

	return count, discretes, nil
}

// GetAllByNicsID retrieve Discretes by nicsID
func (d DiscreteStorage) GetAllByNicsID(offset string, limit string, macAddresses []string) (count int, discretes []model.Discrete, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Joins("INNER JOIN nic ON nic.discrete_serial = discrete.serial").Where("nic.mac_address in (?)", macAddresses).Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
		d.db.Model(&model.Discrete{}).Joins("INNER JOIN nic ON nic.discrete_serial = discrete.serial").Where("nic.mac_address in (?)", macAddresses).Count(&count)
	} else {
		if err = d.db.Joins("INNER JOIN nic ON nic.discrete_serial = discrete.serial").Where("nic.mac_address in (?)", macAddresses).Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
	}
	return count, discretes, err
}

// GetAllByPsusID retrieve descretes by psusID
func (d DiscreteStorage) GetAllByPsusID(offset string, limit string, serials []string) (count int, disceretes []model.Discrete, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Joins("INNER JOIN psu ON psu.discerete_serial = discerete.serial").Where("psu.serial in (?)", serials).Find(&disceretes).Error; err != nil {
			return count, disceretes, err
		}
		d.db.Model(&model.Discrete{}).Joins("INNER JOIN psu ON psu.discerete_serial = discerete.serial").Where("psu.serial in (?)", serials).Count(&count)
	} else {
		if err = d.db.Joins("INNER JOIN psu ON psu.discerete_serial = discerete.serial").Where("psu.serial in (?)", serials).Find(&disceretes).Error; err != nil {
			return count, disceretes, err
		}
	}
	return count, disceretes, err
}

// GetAllByDisksID retrieve descretes by disksID
func (d DiscreteStorage) GetAllByDisksID(offset string, limit string, serials []string) (count int, disceretes []model.Discrete, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Joins("INNER JOIN disk ON disk.discerete_serial = discerete.serial").Where("disk.serial in (?)", serials).Find(&disceretes).Error; err != nil {
			return count, disceretes, err
		}
		d.db.Model(&model.Discrete{}).Joins("INNER JOIN disk ON disk.discerete_serial = discerete.serial").Where("disk.serial in (?)", serials).Count(&count)
	} else {
		if err = d.db.Joins("INNER JOIN disk ON disk.discerete_serial = discerete.serial").Where("disk.serial in (?)", serials).Find(&disceretes).Error; err != nil {
			return count, disceretes, err
		}
	}
	return count, disceretes, err
}

// GetOne Discrete
func (d DiscreteStorage) GetOne(serial string) (discrete model.Discrete, err error) {
	if err := d.db.Preload("Nics").Preload("Disks").Preload("Psus").Where("serial = ?", serial).First(&discrete).Error; err != nil {
		return discrete, err
	}
	return discrete, err
}

// UpdateOrCreate updates or create a new object
func (d *DiscreteStorage) UpdateOrCreate(discrete *model.Discrete) (serial string, err error) {
	if err = d.db.Save(&discrete).Error; err != nil {
		return serial, err
	}
	return discrete.Serial, nil
}

// RemoveOldDiskRefs deletes all the old references from Nics that used to be inside of the chassis
func (d *DiscreteStorage) RemoveOldDiskRefs(discrete *model.Discrete) (count int, serials []string, err error) {
	var connectedSerials []string
	for _, disk := range discrete.Disks {
		connectedSerials = append(connectedSerials, disk.Serial)
	}

	if err = d.db.Model(&model.Disk{}).Where("serial not in (?) and discrete_serial = ?", connectedSerials, discrete.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
		return count, serials, err
	}

	if count > 0 {
		if err = d.db.Where("serial in (?) and discrete_serial = ?", serials, discrete.Serial).Delete(model.Disk{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}

// RemoveOldNicRefs deletes all the old references from Nics that used to be inside of the chassis
func (d *DiscreteStorage) RemoveOldNicRefs(discrete *model.Discrete) (count int, macAddresses []string, err error) {
	var connectedMacAddresses []string
	for _, nic := range discrete.Nics {
		connectedMacAddresses = append(connectedMacAddresses, nic.MacAddress)
	}

	if err = d.db.Model(&model.Nic{}).Where("mac_address not in (?) and discrete_serial = ?", connectedMacAddresses, discrete.Serial).Pluck("mac_address", &macAddresses).Count(&count).Error; err != nil {
		return count, macAddresses, err
	}

	if count > 0 {
		if err = d.db.Where("mac_address in (?) and discrete_serial = ?", macAddresses, discrete.Serial).Delete(model.Nic{}).Error; err != nil {
			return count, macAddresses, err
		}
	}

	return count, macAddresses, err
}

// RemoveOldPsuRefs deletes all the old references from Psus that used to be inside of the chassis
func (d *DiscreteStorage) RemoveOldPsuRefs(discrete *model.Discrete) (count int, serials []string, err error) {
	var connectedSerials []string
	for _, psu := range discrete.Psus {
		connectedSerials = append(connectedSerials, psu.Serial)
	}

	if err = d.db.Model(&model.Psu{}).Where("serial not in (?) and discrete_serial = ?", connectedSerials, discrete.Serial).Pluck("serial", &serials).Count(&count).Error; err != nil {
		return count, serials, err
	}

	if count > 0 {
		if err = d.db.Where("serial in (?) and discrete_serial = ?", serials, discrete.Serial).Delete(model.Psu{}).Error; err != nil {
			return count, serials, err
		}
	}

	return count, serials, err
}

// RemoveOldRefs deletes all the old references from all attached components
func (d *DiscreteStorage) RemoveOldRefs(discrete *model.Discrete) (err error) {
	var merror *multierror.Error
	_, _, err = d.RemoveOldPsuRefs(discrete)
	merror = multierror.Append(merror, err)
	_, _, err = d.RemoveOldNicRefs(discrete)
	merror = multierror.Append(merror, err)
	_, _, err = d.RemoveOldDiskRefs(discrete)
	merror = multierror.Append(merror, err)
	return merror.ErrorOrNil()
}
