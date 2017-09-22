package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewScannedNetworkStorage initializes the storage
func NewScannedNetworkStorage(db *gorm.DB) *ScannedNetworkStorage {
	return &ScannedNetworkStorage{db}
}

// ScannedNetworkStorage stores all ScannedNetworks
type ScannedNetworkStorage struct {
	db *gorm.DB
}

// GetAll of the ScannedNetworks
func (s ScannedNetworkStorage) GetAll(offset string, limit string) (count int, networks []model.ScannedNetwork, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Order("cidr").Find(&networks).Error; err != nil {
			return count, networks, err
		}
		s.db.Model(&model.ScannedNetwork{}).Order("cidr").Count(&count)
	} else {
		if err = s.db.Order("cidr").Find(&networks).Error; err != nil {
			return count, networks, err
		}
	}
	return count, networks, err
}

// GetOne  Blade
func (s ScannedNetworkStorage) GetOne(cidr string) (scan model.ScannedNetwork, err error) {
	if err := s.db.Where("cidr = ?", cidr).First(&scan).Error; err != nil {
		return scan, err
	}
	return scan, err
}
