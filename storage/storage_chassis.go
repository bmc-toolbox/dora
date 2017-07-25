package storage

import (
	"strconv"

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
func (c ChassisStorage) GetOne(id int64) (chassis model.Chassis, err error) {
	chassis, err = c.getOneWithAssociations(id)
	return chassis, err
}

// GetBySerial Chassis
func (c ChassisStorage) GetBySerial(serials []string) (chassis []model.Chassis, err error) {
	for _, serial := range serials {
		bl, err := c.getBySerial(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return chassis, err
		}
		chassis = append(chassis, bl)
	}
	return chassis, nil
}

func (c ChassisStorage) getBySerial(serial string) (chassis model.Chassis, err error) {
	if err = c.db.Where("serial = ?", serial).First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// GetAllByBladesID Chassis
func (c ChassisStorage) GetAllByBladesID(IDs []string) (chassis []model.Chassis, err error) {
	for _, id := range IDs {
		iid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		ch, err := c.getByBladeID(iid)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return chassis, err
		}
		chassis = append(chassis, ch)
	}
	return chassis, nil
}

func (c ChassisStorage) getByBladeID(id int64) (chassis model.Chassis, err error) {
	if err = c.db.Joins("INNER JOIN blade ON blade.chassis_id = chassis.id").Where("blade.id = ?", id).Find(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

func (c ChassisStorage) getOneWithAssociations(id int64) (chassis model.Chassis, err error) {
	if err = c.db.Where("id = ?", id).Preload("Blades").First(&chassis).Error; err != nil {
		return chassis, err
	}
	return chassis, err
}

// UpdateOrCreate
func (c *ChassisStorage) UpdateOrCreate(chassis *model.Chassis) (id int64, err error) {
	ch := &model.Chassis{}
	if err = c.db.Where("id = ?", chassis.ID).FirstOrCreate(&ch).Error; err != nil {
		return id, err
	}

	chassis.ID = ch.ID
	if err = c.db.Save(&chassis).Error; err != nil {
		return id, err
	}
	return chassis.ID, nil
}
