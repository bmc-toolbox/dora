package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/go/dora/filter"
	"gitlab.booking.com/go/dora/model"
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
		if err = d.db.Limit(limit).Offset(offset).Order("serial asc").Preload("Nics").Find(&discretes).Error; err != nil {
			return count, discretes, err
		}
		d.db.Order("serial asc").Find(&model.Discrete{}).Count(&count)
	} else {
		if err = d.db.Order("serial").Preload("Nics").Find(&discretes).Error; err != nil {
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

// GetOne Discrete
func (d DiscreteStorage) GetOne(serial string) (discrete model.Discrete, err error) {
	if err := d.db.Preload("Nics").Where("serial = ?", serial).First(&discrete).Error; err != nil {
		return discrete, err
	}
	return discrete, err
}

// UpdateOrCreate
func (d *DiscreteStorage) UpdateOrCreate(discrete *model.Discrete) (serial string, err error) {
	if err = d.db.Save(&discrete).Error; err != nil {
		return serial, err
	}
	return discrete.Serial, nil
}
