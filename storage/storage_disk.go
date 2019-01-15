package storage

import (
	"github.com/jinzhu/gorm"
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
)

// NewDiskStorage initializes the storage
func NewDiskStorage(db *gorm.DB) *DiskStorage {
	return &DiskStorage{db}
}

// DiskStorage stores all disks used by blades
type DiskStorage struct {
	db *gorm.DB
}

// GetAll disks
func (d DiskStorage) GetAll(offset string, limit string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Order("serial").Find(&disks).Error; err != nil {
			return count, disks, err
		}
		d.db.Model(&model.Disk{}).Order("serial").Count(&count)
	} else {
		if err = d.db.Order("serial").Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetAllByFilters get all blades based on the filter
func (d DiskStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, disks []model.Disk, err error) {
	query, err := filters.BuildQuery(model.Disk{})
	if err != nil {
		return count, disks, err
	}

	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Where(query).Find(&disks).Error; err != nil {
			return count, disks, err
		}
		d.db.Model(&model.Disk{}).Where(query).Count(&count)
	} else {
		if err = d.db.Where(query).Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}

	return count, disks, nil
}

// GetAllByBladeID of the disks by BladeID
func (d DiskStorage) GetAllByBladeID(offset string, limit string, serials []string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Where("blade_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
		d.db.Model(&model.Disk{}).Where("blade_serial in (?)", serials).Count(&count)
	} else {
		if err = d.db.Where("blade_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetAllByDiscreteID of the disks by DiscreteID
func (d DiskStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = d.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
		d.db.Model(&model.Disk{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = d.db.Where("discrete_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetOne z
func (d DiskStorage) GetOne(serial string) (Disk model.Disk, err error) {
	if err := d.db.Where("serial = ?", serial).First(&Disk).Error; err != nil {
		return Disk, err
	}
	return Disk, err
}
