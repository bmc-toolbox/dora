package storage

import (
	"github.com/jinzhu/gorm"
	"gitlab.booking.com/infra/dora/model"
)

// NewNicStorage initializes the storage
func NewNicStorage(db *gorm.DB) *NicStorage {
	return &NicStorage{db}
}

// NicStorage stores all nics used by blades
type NicStorage struct {
	db *gorm.DB
}

// GetAll of the Nic
func (n NicStorage) GetAll() (nics []model.Nic, err error) {
	if err = n.db.Order("mac_address").Find(&nics).Error; err != nil {
		return nics, err
	}
	return nics, err
}

// GetAllByBladeID of the Blades by BladeID
func (n NicStorage) GetAllByBladeID(serials []string) (nics []model.Nic, err error) {
	for _, serial := range serials {
		nc, err := n.getByBladeID(serial)
		if err == gorm.ErrRecordNotFound {
			continue
		} else if err != nil {
			return nics, err
		}
		nics = append(nics, nc)
	}
	return nics, nil
}

func (n NicStorage) getByBladeID(serial string) (nic model.Nic, err error) {
	if err = n.db.Where("blade_serial = ?", serial).First(&nic).Error; err != nil {
		return nic, err
	}
	return nic, err
}

// GetOne  Blade
func (n NicStorage) GetOne(macAddress string) (nic model.Nic, err error) {
	if err := n.db.Where("mac_address = ?", macAddress).First(&nic).Error; err != nil {
		return nic, err
	}
	return nic, err
}
