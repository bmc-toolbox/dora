package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
)

// NewFanStorage initializes the storage
func NewFanStorage(db *gorm.DB) *FanStorage {
	return &FanStorage{db}
}

// FanStorage stores all fans used by blades
type FanStorage struct {
	db *gorm.DB
}

// Count get fans count based on the filter
func (p FanStorage) Count(filters *filter.Filters) (count int, err error) {
	query, err := filters.BuildQuery(model.Fan{})
	if err != nil {
		return count, err
	}

	err = p.db.Model(&model.Fan{}).Where(query).Count(&count).Error
	return count, err
}

// GetAll fans
func (p FanStorage) GetAll(offset string, limit string) (count int, fans []model.Fan, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Order("serial").Find(&fans).Error; err != nil {
			return count, fans, err
		}
		p.db.Model(&model.Fan{}).Order("serial").Count(&count)
	} else {
		if err = p.db.Order("serial").Find(&fans).Error; err != nil {
			return count, fans, err
		}
	}
	return count, fans, err
}

// GetAllByChassisID of the fans by ChassisID
func (p FanStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, fans []model.Fan, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&fans).Error; err != nil {
			return count, fans, err
		}
		p.db.Model(&model.Fan{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("chassis_serial in (?)", serials).Find(&fans).Error; err != nil {
			return count, fans, err
		}
	}
	return count, fans, err
}

// GetAllByDiscreteID of the fans by DiscreteID
func (p FanStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, fans []model.Fan, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&fans).Error; err != nil {
			return count, fans, err
		}
		p.db.Model(&model.Fan{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("discrete_serial in (?)", serials).Find(&fans).Error; err != nil {
			return count, fans, err
		}
	}
	return count, fans, err
}

// GetOne fan
func (p FanStorage) GetOne(serial string) (fan model.Fan, err error) {
	if err := p.db.Where("serial = ?", serial).First(&fan).Error; err != nil {
		return fan, err
	}
	return fan, err
}

// GetAllByFilters get all blades based on the filter
func (p FanStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, fans []model.Fan, err error) {
	query, err := filters.BuildQuery(model.Fan{})
	if err != nil {
		return count, fans, err
	}

	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where(query).Find(&fans).Error; err != nil {
			return count, fans, err
		}
		p.db.Model(&model.Fan{}).Where(query).Count(&count)
	} else {
		if err = p.db.Where(query).Find(&fans).Error; err != nil {
			return count, fans, err
		}
	}

	return count, fans, nil
}
