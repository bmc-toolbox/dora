package storage

import (
	"fmt"
	"strings"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// NewPsuStorage initializes the storage
func NewPsuStorage(db *gorm.DB) *PsuStorage {
	return &PsuStorage{db}
}

// PsuStorage stores all psus used by blades
type PsuStorage struct {
	db *gorm.DB
}

// Count get psus count based on the filter
func (p PsuStorage) Count(filters *filter.Filters) (count int, err error) {
	q, err := filters.BuildQuery(model.Psu{}, p.db)
	if err != nil {
		return count, err
	}

	err = q.Model(&model.Psu{}).Count(&count).Error
	return count, err
}

// GetAll psus
func (p PsuStorage) GetAll(offset string, limit string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Order("serial").Count(&count)
	} else {
		if err = p.db.Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (p PsuStorage) GetAllWithAssociations(offset string, limit string, include []string) (count int, psus []model.Psu, err error) {
	q := p.db.Order("serial asc")
	for _, preload := range include {
		q = q.Preload(strings.Title(preload))
	}

	if offset != "" && limit != "" {
		q = p.db.Limit(limit).Offset(offset)
		p.db.Order("serial asc").Find(&model.Psu{}).Count(&count)
	}

	if err = q.Find(&psus).Error; err != nil {
		if strings.Contains(err.Error(), "can't preload field") {
			return count, psus, api2go.NewHTTPError(nil,
				fmt.Sprintf("invalid include: %s", strings.Split(err.Error(), " ")[3]), 422)
		}
		return count, psus, err
	}

	return count, psus, err
}

// GetAllByChassisID of the psus by ChassisID
func (p PsuStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllByDiscreteID of the psus by DiscreteID
func (p PsuStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = p.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		p.db.Model(&model.Psu{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = p.db.Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetOne psu
func (p PsuStorage) GetOne(serial string) (psu model.Psu, err error) {
	if err := p.db.Where("serial = ?", serial).First(&psu).Error; err != nil {
		return psu, err
	}
	return psu, err
}

// GetAllByFilters get all blades based on the filter
func (p PsuStorage) GetAllByFilters(offset string, limit string, filters *filter.Filters) (count int, psus []model.Psu, err error) {
	q, err := filters.BuildQuery(model.Psu{}, p.db)
	if err != nil {
		return count, psus, err
	}

	if offset != "" && limit != "" {
		if err = q.Limit(limit).Offset(offset).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		q.Model(&model.Psu{}).Count(&count)
	} else {
		if err = q.Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}

	return count, psus, nil
}
