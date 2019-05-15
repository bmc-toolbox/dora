package storage

import (
	"fmt"
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"strings"
)

// NewNicStorage initializes the storage
func NewNicStorage(db *gorm.DB) *NicStorage {
	return &NicStorage{db}
}

// NicStorage stores all nics used by blades
type NicStorage struct {
	db *gorm.DB
}

// Count get nics count based on the filter
func (n NicStorage) Count(filters *filter.Filters) (count int, err error) {
	query, err := filters.BuildQuery(model.Nic{})
	if err != nil {
		return count, err
	}

	err = n.db.Model(&model.Nic{}).Where(query).Count(&count).Error
	return count, err
}

// GetAll nics
func (n NicStorage) GetAll(offset string, limit string) (count int, nics []model.Nic, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Order("mac_address").Find(&nics).Error; err != nil {
			return count, nics, err
		}
		n.db.Model(&model.Nic{}).Order("mac_address").Count(&count)
	} else {
		if err = n.db.Order("mac_address").Find(&nics).Error; err != nil {
			return count, nics, err
		}
	}
	return count, nics, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (n NicStorage) GetAllWithAssociations(offset string, limit string, include []string) (count int, nics []model.Nic, err error) {
	q := n.db.Order("mac_address")
	for _, preload := range include {
		q = q.Preload(strings.Title(preload))
	}

	if offset != "" && limit != "" {
		q = n.db.Limit(limit).Offset(offset)
		n.db.Order("mac_address").Find(&model.Nic{}).Count(&count)
	}

	if err = q.Find(&nics).Error; err != nil {
		if strings.Contains(err.Error(), "can't preload field") {
			return count, nics, api2go.NewHTTPError(nil,
				fmt.Sprintf("invalid include: %s", strings.Split(err.Error(), " ")[3]) , 422)
		}
		return count, nics, err
	}

	return count, nics, err
}

// GetAllByBladeID of the nics by BladeID
func (n NicStorage) GetAllByBladeID(offset string, limit string, serials []string) (count int, nics []model.Nic, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("blade_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
		n.db.Model(&model.Nic{}).Where("blade_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("blade_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
	}
	return count, nics, err
}

// GetAllByChassisID of the nics by ChassisID
func (n NicStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, nics []model.Nic, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
		n.db.Model(&model.Nic{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("chassis_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
	}
	return count, nics, err
}

// GetAllByDiscreteID of the nics by DiscreteID
func (n NicStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, nics []model.Nic, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
		n.db.Model(&model.Nic{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("discrete_serial in (?)", serials).Find(&nics).Error; err != nil {
			return count, nics, err
		}
	}
	return count, nics, err
}

// GetOne nic
func (n NicStorage) GetOne(macAddress string) (nic model.Nic, err error) {
	if err := n.db.Where("mac_address = ?", macAddress).First(&nic).Error; err != nil {
		return nic, err
	}
	return nic, err
}

// GetAllByFilters get all blades based on the filter
func (n NicStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, nics []model.Nic, err error) {
	query, err := filters.BuildQuery(model.Nic{})
	if err != nil {
		return count, nics, err
	}

	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where(query).Find(&nics).Error; err != nil {
			return count, nics, err
		}
		n.db.Model(&model.Nic{}).Where(query).Count(&count)
	} else {
		if err = n.db.Where(query).Find(&nics).Error; err != nil {
			return count, nics, err
		}
	}

	return count, nics, nil
}
