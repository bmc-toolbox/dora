package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
)

// NewPsuStorage initializes the storage
func NewPsuStorage(db *gorm.DB) *PsuStorage {
	return &PsuStorage{db}
}

// PsuStorage stores all psus used by blades
type PsuStorage struct {
	db *gorm.DB
}

// Count get psus count based on the filter
func (p PsuStorage) Count(filters *filter.Filters) (count int, err error) {
	query, err := filters.BuildQuery(model.Psu{})
	if err != nil {
		return count, err
	}

	err = p.db.Model(&model.Psu{}).Where(query).Count(&count).Error
	return count, err
}

// GetAll psus
func (p PsuStorage) GetAll(offset string, limit string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Order("serial").Count(&count)
	} else {
		if err = p.db.Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllByChassisID of the psus by ChassisID
func (p PsuStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllByDiscreteID of the psus by DiscreteID
func (p PsuStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetOne psu
func (p PsuStorage) GetOne(serial string) (psu model.Psu, err error) {
	if err := p.db.Where("serial = ?", serial).First(&psu).Error; err != nil {
		return psu, err
	}
	return psu, err
}

// GetAllByFilters get all blades based on the filter
func (p PsuStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, psus []model.Psu, err error) {
	query, err := filters.BuildQuery(model.Psu{})
	if err != nil {
		return count, psus, err
	}

	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where(query).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Where(query).Count(&count)
	} else {
		if err = p.db.Where(query).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}

	return count, psus, nil
}
