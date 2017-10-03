package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
)

// NewScannedHostStorage initializes the storage
func NewScannedHostStorage(db *gorm.DB) *ScannedHostStorage {
	return &ScannedHostStorage{db}
}

// ScannedHostStorage stores all ScannedHosts
type ScannedHostStorage struct {
	db *gorm.DB
}

// GetAll of the ScannedHosts
func (s ScannedHostStorage) GetAll(offset string, limit string) (count int, hosts []model.ScannedHost, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Order("cidr").Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
		s.db.Model(&model.ScannedHost{}).Order("cidr").Count(&count)
	} else {
		if err = s.db.Order("cidr").Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
	}
	return count, hosts, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (s ScannedHostStorage) GetAllWithAssociations(offset string, limit string) (count int, hosts []model.ScannedHost, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Preload("Hosts").Order("cidr").Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
		s.db.Model(&model.ScannedHost{}).Preload("Hosts").Order("cidr").Count(&count)
	} else {
		if err = s.db.Order("cidr").Preload("Hosts").Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
	}
	return count, hosts, err
}

// GetAllByFilters get all chassis based on the filter
func (s ScannedHostStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, hosts []model.ScannedHost, err error) {
	query, err := filters.BuildQuery(model.ScannedHost{})
	if err != nil {
		return count, hosts, err
	}

	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Where(query).Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
		s.db.Model(&model.ScannedHost{}).Where(query).Count(&count)
	} else {
		if err = s.db.Where(query).Find(&hosts).Error; err != nil {
			return count, hosts, err
		}
	}

	return count, hosts, err
}

// GetOne Host
func (s ScannedHostStorage) GetOne(ip string) (scan model.ScannedHost, err error) {
	if err := s.db.Where("ip = ?", ip).Preload("Ports").First(&scan).Error; err != nil {
		return scan, err
	}
	return scan, err
}
