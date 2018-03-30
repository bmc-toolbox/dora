package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/go/dora/model"
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
func (n DiskStorage) GetAll(offset string, limit string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Order("serial").Find(&disks).Error; err != nil {
			return count, disks, err
		}
		n.db.Model(&model.Disk{}).Order("serial").Count(&count)
	} else {
		if err = n.db.Order("serial").Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetAllByBladeID of the disks by BladeID
func (n DiskStorage) GetAllByBladeID(offset string, limit string, serials []string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("blade_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
		n.db.Model(&model.Disk{}).Where("blade_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("blade_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetAllByDiscreteID of the disks by DiscreteID
func (n DiskStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, disks []model.Disk, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
		n.db.Model(&model.Disk{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("discrete_serial in (?)", serials).Find(&disks).Error; err != nil {
			return count, disks, err
		}
	}
	return count, disks, err
}

// GetOne z
func (n DiskStorage) GetOne(serial string) (Disk model.Disk, err error) {
	if err := n.db.Where("serial = ?", serial).First(&Disk).Error; err != nil {
		return Disk, err
	}
	return Disk, err
}
