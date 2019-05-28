package storage

import (
	"fmt"
	"strings"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// NewStorageBladeStorage initializes the storage
func NewStorageBladeStorage(db *gorm.DB) *StorageBladeStorage {
	return &StorageBladeStorage{db}
}

// StorageBladeStorage stores all the storage blades we have in the company
type StorageBladeStorage struct {
	db *gorm.DB
}

// Count get blades count based on the filter
func (b StorageBladeStorage) Count(filters *filter.Filters) (count int, err error) {
	q, err := filters.BuildQuery(model.StorageBlade{}, b.db)
	if err != nil {
		return count, err
	}

	err = q.Model(&model.StorageBlade{}).Count(&count).Error
	return count, err
}

// GetAll of the Blades
func (b StorageBladeStorage) GetAll(offset string, limit string) (count int, storageBlades []model.StorageBlade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Order("serial asc").Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		b.db.Model(&model.StorageBlade{}).Order("serial asc").Count(&count)
	} else {
		if err = b.db.Order("serial").Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}
	return count, storageBlades, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (b StorageBladeStorage) GetAllWithAssociations(offset string, limit string, include []string) (count int, storageBlades []model.StorageBlade, err error) {
	q := b.db.Order("serial asc")
	for _, preload := range include {
		q = q.Preload(strings.Title(preload))
	}

	if offset != "" && limit != "" {
		q = b.db.Limit(limit).Offset(offset)
		b.db.Order("serial asc").Find(&model.StorageBlade{}).Count(&count)
	}

	if err = q.Find(&storageBlades).Error; err != nil {
		if strings.Contains(err.Error(), "can't preload field") {
			return count, storageBlades, api2go.NewHTTPError(nil,
				fmt.Sprintf("invalid include: %s", strings.Split(err.Error(), " ")[3]), 422)
		}
		return count, storageBlades, err
	}
	return count, storageBlades, err
}

// GetAllByFilters get all StorageBlades based on the filter
func (b StorageBladeStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, storageBlades []model.StorageBlade, err error) {
	q, err := filters.BuildQuery(model.StorageBlade{}, b.db)
	if err != nil {
		return count, storageBlades, err
	}

	if offset != "" && limit != "" {
		if err = q.Limit(limit).Offset(offset).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		q.Model(&model.StorageBlade{}).Count(&count)
	} else {
		if err = q.Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}

	return count, storageBlades, nil
}

// GetAllByChassisID retrieves StorageBlades by chassisID
func (b StorageBladeStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, storageBlades []model.StorageBlade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		b.db.Model(&model.StorageBlade{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = b.db.Where("chassis_serial in (?)", serials).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}
	return count, storageBlades, err
}

// GetAllByBladeID retrieves StorageBlades by bladesID
func (b StorageBladeStorage) GetAllByBladeID(offset string, limit string, serials []string) (count int, storageBlades []model.StorageBlade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Where("blade_serial in (?)", serials).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		b.db.Model(&model.StorageBlade{}).Where("blade_serial in (?)", serials).Count(&count)
	} else {
		if err = b.db.Where("blade_serial in (?)", serials).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}
	return count, storageBlades, err
}

// GetOne StorageBlade
func (b StorageBladeStorage) GetOne(serial string) (storageBlade model.StorageBlade, err error) {
	if err := b.db.Where("serial = ?", serial).First(&storageBlade).Error; err != nil {
		return storageBlade, err
	}
	return storageBlade, err
}

// UpdateOrCreate a StorageBlade
func (b *StorageBladeStorage) UpdateOrCreate(storageBlade *model.StorageBlade) (serial string, err error) {
	if err = b.db.Save(&storageBlade).Error; err != nil {
		return serial, err
	}
	return storageBlade.Serial, nil
}
