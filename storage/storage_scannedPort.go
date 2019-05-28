package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
)

// NewScannedPortStorage initializes the storage
func NewScannedPortStorage(db *gorm.DB) *ScannedPortStorage {
	return &ScannedPortStorage{db}
}

// ScannedPortStorage stores all ScannedPorts
type ScannedPortStorage struct {
	db *gorm.DB
}

// Count get ScannedPorts count based on the filter
func (s ScannedPortStorage) Count(filters *filter.Filters) (count int, err error) {
	q, err := filters.BuildQuery(model.ScannedPort{}, s.db)
	if err != nil {
		return count, err
	}

	err = q.Model(&model.ScannedPort{}).Count(&count).Error
	return count, err
}

// GetAll of the ScannedPorts
func (s ScannedPortStorage) GetAll(offset string, limit string) (count int, ports []model.ScannedPort, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Order("cidr").Find(&ports).Error; err != nil {
			return count, ports, err
		}
		s.db.Model(&model.ScannedPort{}).Order("cidr").Count(&count)
	} else {
		if err = s.db.Order("cidr").Find(&ports).Error; err != nil {
			return count, ports, err
		}
	}
	return count, ports, err
}

// GetAllByFilters get all chassis based on the filter
func (s ScannedPortStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, ports []model.ScannedPort, err error) {
	q, err := filters.BuildQuery(model.ScannedPort{}, s.db)
	if err != nil {
		return count, ports, err
	}

	if offset != "" && limit != "" {
		if err = q.Limit(limit).Offset(offset).Find(&ports).Error; err != nil {
			return count, ports, err
		}
		q.Model(&model.ScannedPort{}).Count(&count)
	} else {
		if err = q.Find(&ports).Error; err != nil {
			return count, ports, err
		}
	}

	return count, ports, err
}

// GetOne Host
func (s ScannedPortStorage) GetOne(id string) (scan model.ScannedPort, err error) {
	if err := s.db.Where("id = ?", id).First(&scan).Error; err != nil {
		return scan, err
	}
	return scan, err
}
