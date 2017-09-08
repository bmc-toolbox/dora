package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
)

// NewStorageBladeStorage initializes the storage
func NewStorageBladeStorage(db *gorm.DB) *StorageBladeStorage {
	return &StorageBladeStorage{db}
}

// StorageBladeStorage stores all the storage blades we have in the company
type StorageBladeStorage struct {
	db *gorm.DB
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
func (b StorageBladeStorage) GetAllWithAssociations(offset string, limit string) (count int, storageBlades []model.StorageBlade, err error) {
	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Order("serial asc").Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		b.db.Order("serial asc").Find(&model.StorageBlade{}).Count(&count)
	} else {
		if err = b.db.Order("serial").Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}
	return count, storageBlades, err
}

// GetAllByFilters get all blades detecting the struct members dinamically
func (b StorageBladeStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, storageBlades []model.StorageBlade, err error) {
	query, err := filters.BuildQuery(model.StorageBlade{})
	if err != nil {
		return count, storageBlades, err
	}

	if offset != "" && limit != "" {
		if err = b.db.Limit(limit).Offset(offset).Where(query).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
		b.db.Model(&model.StorageBlade{}).Where(query).Count(&count)
	} else {
		if err = b.db.Where(query).Find(&storageBlades).Error; err != nil {
			return count, storageBlades, err
		}
	}

	return count, storageBlades, nil
}

// GetAllByChassisID retreives Blades by chassisID
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

// GetAllByBladeID retreives StorageBlades by bladesID
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

// GetOne storageBlade
func (b StorageBladeStorage) GetOne(serial string) (storageBlade model.StorageBlade, err error) {
	if err := b.db.Where("serial = ?", serial).First(&storageBlade).Error; err != nil {
		return storageBlade, err
	}
	return storageBlade, err
}

// UpdateOrCreate
func (b *StorageBladeStorage) UpdateOrCreate(storageBlade *model.StorageBlade) (serial string, err error) {
	if err = b.db.Save(&storageBlade).Error; err != nil {
		return serial, err
	}
	return storageBlade.Serial, nil
}
