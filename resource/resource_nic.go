package resource

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// BladeResource for api2go routes
type NicResource struct {
	BladeStorage *storage.BladeStorage
	NicStorage   *storage.NicStorage
}

// FindAll Nics
func (n NicResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var nics []model.Nic
	var err error
	_, hasFilters := r.QueryParams["filter[mac]"]
	bladeID, hasBlade := r.QueryParams["bladeID"]

	if hasBlade {
		nics, err = n.NicStorage.GetAllByBladeID(bladeID)
	}

	if !hasFilters && !hasBlade {
		nics, err = n.NicStorage.GetAll()
		if err != nil {
			return &Response{Res: nics}, err
		}
	}

	return &Response{Res: nics}, nil
}

// FindOne Nics
func (n NicResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := n.NicStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
