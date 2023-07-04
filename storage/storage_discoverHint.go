package storage

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
)

// NewDiscoverHintStorage initializes the storage
func NewDiscoverHintStorage(db *gorm.DB) *DiscoverHintStorage {
	return &DiscoverHintStorage{db}
}

// DiscoverHintStorage stores all DiscoverHints
type DiscoverHintStorage struct {
	db *gorm.DB
}

// Count gets DiscoverHints count based on the filter
func (s DiscoverHintStorage) Count(filters *filter.Filters) (count int, err error) {
	q, err := filters.BuildQuery(model.DiscoverHint{}, s.db)
	if err != nil {
		return count, err
	}

	err = q.Model(&model.DiscoverHint{}).Count(&count).Error
	return count, err
}

// GetAll of the DiscoverHints
func (s DiscoverHintStorage) GetAll(offset string, limit string) (count int, ips []model.DiscoverHint, err error) {

	query := s.db.Offset(offset).Limit(limit)

	if err = query.Order("ip").Find(&ips).Error; err != nil {
		return count, ips, err
	}
	query.Model(&model.DiscoverHint{}).Order("ip").Count(&count)
	return count, ips, err
}

// GetAllByFilters get all Discover Hints based on the filter
func (s DiscoverHintStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, ips []model.DiscoverHint, err error) {
	query, err := filters.BuildQuery(model.DiscoverHint{}, s.db)
	if err != nil {
		return count, ips, err
	}

	query = query.Offset(offset).Limit(limit)

	if err = query.Order("ip").Find(&ips).Error; err != nil {
		return count, ips, err
	}
	query.Model(&model.DiscoverHint{}).Order("ip").Count(&count)

	return count, ips, err
}

// GetOne Host
func (s DiscoverHintStorage) GetOne(id string) (scan model.DiscoverHint, err error) {
	if err := s.db.Where("ip = ?", id).First(&scan).Error; err != nil {
		return scan, err
	}
	return scan, err
}
