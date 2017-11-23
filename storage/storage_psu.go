package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/go/dora/model"
)

// NewPsuStorage initializes the storage
func NewPsuStorage(db *gorm.DB) *PsuStorage {
	return &PsuStorage{db}
}

// PsuStorage stores all psus used by blades
type PsuStorage struct {
	db *gorm.DB
}

// GetAll psus
func (n PsuStorage) GetAll(offset string, limit string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
		n.db.Model(&model.Psu{}).Order("serial").Count(&count)
	} else {
		if err = n.db.Order("serial").Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllByChassisID of the psus by ChassisID
func (n PsuStorage) GetAllByChassisID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		n.db.Model(&model.Psu{}).Where("chassis_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("chassis_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetAllByDiscreteID of the psus by DiscreteID
func (n PsuStorage) GetAllByDiscreteID(offset string, limit string, serials []string) (count int, psus []model.Psu, err error) {
	if offset != "" && limit != "" {
		if err = n.db.Limit(limit).Offset(offset).Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
		n.db.Model(&model.Psu{}).Where("discrete_serial in (?)", serials).Count(&count)
	} else {
		if err = n.db.Where("discrete_serial in (?)", serials).Find(&psus).Error; err != nil {
			return count, psus, err
		}
	}
	return count, psus, err
}

// GetOne psu
func (n PsuStorage) GetOne(serial string) (psu model.Psu, err error) {
	if err := n.db.Where("serial = ?", serial).First(&psu).Error; err != nil {
		return psu, err
	}
	return psu, err
}
