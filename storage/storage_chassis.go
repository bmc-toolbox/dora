package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewChassisStorage initializes the storage
func NewChassisStorage(db *gorm.DB) *ChassisStorage {
	return &ChassisStorage{db}
}

// ChassisStorage stores all Chassiss
type ChassisStorage struct {
	db *gorm.DB
}

// GetAll returns all chassis
func (c ChassisStorage) GetAll() (chassis []model.Chassis, err error) {
	if err = c.db.Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetAllWithAssociations returns all chassis with their relationships
func (c ChassisStorage) GetAllWithAssociations() (chassis []model.Chassis, err error) {
	if err = c.db.Preload("Blades").Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetOne Chassis
func (c ChassisStorage) GetOne(serial string) (chassis model.Chassis, err error) {
	chassis, err = c.getOneWithAssociations(serial)
	return chassis, err
}

// GetAllByBladesID Chassis
func (c ChassisStorage) GetAllByBladesID(serials []string) (chassis []model.Chassis, err error) {
	for _, serial := range serials {
		ch, err := c.getByBladeID(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return chassis, err
		}
		chassis = append(chassis, ch)
	}
	return chassis, nil
}

func (c ChassisStorage) getByBladeID(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Joins("INNER JOIN blade ON blade.chassis_serial = chassis.serial").Where("blade.serial = ?", serial).Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

func (c ChassisStorage) getOneWithAssociations(id string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", id).Preload("Blades").First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// UpdateOrCreate
func (c *ChassisStorage) UpdateOrCreate(chassis *model.Chassis) (serial string, err error) {
	if err = c.db.Save(&chassis).Error; err != nil {
		return serial, err
	}
	return chassis.Serial, nil
}
