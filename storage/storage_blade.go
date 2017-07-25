package storage

import (
	"strconv"

	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewBladeStorage initializes the storage
func NewBladeStorage(db *gorm.DB) *BladeStorage {
	return &BladeStorage{db}
}

// BladeStorage stores all of the tasty Blade, needs to be injected into
// Chassis and Blade Resource. In the real world, you would use a database for that.
type BladeStorage struct {
	db *gorm.DB
}

// GetAll of the Blades
func (b BladeStorage) GetAll() (blades []model.Blade, err error) {
	if err = b.db.Order("id").Find(&blades).Error; err != nil {
		return blades, err
	}
	return blades, err
}

// GetAllByChassisID of the Blades by chassisID
func (b BladeStorage) GetAllByChassisID(IDs []string) (blades []model.Blade, err error) {
	for _, id := range IDs {
		iid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		bl, err := b.getByChassisID(iid)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return blades, err
		}
		blades = append(blades, bl)
	}
	return blades, nil
}

func (b BladeStorage) getByChassisID(id int64) (blade model.Blade, err error) {
	if err = b.db.Where("chassis_id = ?", id).First(&blade).Error; err != nil {
		return blade, err
	}
	return blade, err
}

// GetOne  Blade
func (s BladeStorage) GetOne(id int64) (blade model.Blade, err error) {
	if err := s.db.First(&blade, id).Error; err != nil {
		return blade, err
	}
	return blade, err
}

func (s BladeStorage) getBySerial(serial string) (blade model.Blade, err error) {
	if err = s.db.Where("serial = ?", serial).First(&blade).Error; err != nil {
		return blade, err
	}
	return blade, err
}

// GetBySerial Blade
func (s BladeStorage) GetBySerial(serials []string) (blades []model.Blade, err error) {
	for _, serial := range serials {
		bl, err := s.getBySerial(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return blades, err
		}
		blades = append(blades, bl)
	}
	return blades, nil
}
