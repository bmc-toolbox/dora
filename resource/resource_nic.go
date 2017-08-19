package resource

import (
	"errors"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

var (
	// ErrPageSizeAndNumber is returned when page[number] and page[size] are sent on the http request
	ErrPageSizeAndNumber = errors.New("Filters page[number] and page[size] are not supported, please stick to page[offset] and page[limit]")
)

// NicResource for api2go routes
type NicResource struct {
	BladeStorage *storage.BladeStorage
	NicStorage   *storage.NicStorage
}

// FindAll Nics
func (n NicResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, nics, err := n.queryAndCountAllWrapper(r)
	return &Response{Res: nics}, err
}

// FindOne Nics
func (n NicResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := n.NicStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load nics in chunks
func (n NicResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, nics, err := n.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: nics}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a stardard way
func (n NicResource) queryAndCountAllWrapper(r api2go.Request) (count int, nics []model.Nic, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, nics, ErrPageSizeAndNumber
		}
	}

	var offset string
	var limit string

	offsetQuery, hasOffset := r.QueryParams["page[offset]"]
	if hasOffset {
		offset = offsetQuery[0]
	}

	limitQuery, hasLimit := r.QueryParams["page[limit]"]
	if hasLimit {
		limit = limitQuery[0]
	}

	bladeID, hasBlade := r.QueryParams["bladeID"]
	if hasBlade {
		count, nics, err = n.NicStorage.GetAllByBladeID(offset, limit, bladeID)
		return
	}

	if !hasBlade {
		count, nics, err = n.NicStorage.GetAll(offset, limit)
		if err != nil {
			return count, nics, err
		}
	}

	return count, nics, err
}
