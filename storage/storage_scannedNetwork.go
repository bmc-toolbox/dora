package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/filter"
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

// GetAllWithAssociations returns all chassis with their relationships
func (s ScannedNetworkStorage) GetAllWithAssociations(offset string, limit string) (count int, networks []model.ScannedNetwork, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Preload("Hosts").Order("cidr").Find(&networks).Error; err != nil {
			return count, networks, err
		}
		s.db.Model(&model.ScannedNetwork{}).Preload("Hosts").Order("cidr").Count(&count)
	} else {
		if err = s.db.Order("cidr").Preload("Hosts").Find(&networks).Error; err != nil {
			return count, networks, err
		}
	}
	return count, networks, err
}

// GetAllByFilters get all chassis based on the filter
func (s ScannedNetworkStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, networks []model.ScannedNetwork, err error) {
	query, err := filters.BuildQuery(model.ScannedNetwork{})
	if err != nil {
		return count, networks, err
	}

	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Where(query).Find(&networks).Error; err != nil {
			return count, networks, err
		}
		s.db.Model(&model.ScannedNetwork{}).Where(query).Count(&count)
	} else {
		if err = s.db.Where(query).Find(&networks).Error; err != nil {
			return count, networks, err
		}
	}

	return count, networks, err
}

// GetAllByIP retrieve networks by IP
func (s ScannedNetworkStorage) GetAllByIP(offset string, limit string, ips []string) (count int, networks []model.ScannedNetwork, err error) {
	if offset != "" && limit != "" {
		if err = s.db.Limit(limit).Offset(offset).Joins("INNER JOIN scanned_hosts ON scanned_hosts.cidr = scanned_network.cidr").Where("ip in (?)", ips).Find(&networks).Error; err != nil {
			return count, networks, err
		}
		s.db.Model(&model.ScannedNetwork{}).Joins("INNER JOIN scanned_hosts ON scanned_hosts.cidr = scanned_network.cidr").Where("ip in (?)", ips).Count(&count)
	} else {
		if err = s.db.Joins("INNER JOIN scanned_hosts ON scanned_hosts.cidr = scanned_network.cidr").Where("ip in (?)", ips).Find(&networks).Error; err != nil {
			return count, networks, err
		}
	}
	return count, networks, err
}

// GetOne Network
func (s ScannedNetworkStorage) GetOne(cidr string) (scan model.ScannedNetwork, err error) {
	if err := s.db.Where("cidr = ?", cidr).Preload("Hosts").First(&scan).Error; err != nil {
		return scan, err
	}
	return scan, err
}
