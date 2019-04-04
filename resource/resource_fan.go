package resource

import (
	"net/http"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// FanResource for api2go routes
type FanResource struct {
	FanStorage *storage.FanStorage
}

// FindAll Fans
func (p FanResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, fans, err := p.queryAndCountAllWrapper(r)
	return &Response{Res: fans}, err
}

// FindOne Fan
func (p FanResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := p.FanStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load fans in chunks
func (p FanResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, fans, err := p.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: fans}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (p FanResource) queryAndCountAllWrapper(r api2go.Request) (count int, fans []model.Fan, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, fans, ErrPageSizeAndNumber
		}
	}

	offset, limit := filter.OffSetAndLimitParse(&r)
	filters, hasFilters := filter.NewFilterSet(&r)
	if hasFilters {
		count, fans, err = p.FanStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, fans, err
		}
	}

	chassisID, hasChassis := r.QueryParams["chassisID"]
	if hasChassis {
		count, fans, err = p.FanStorage.GetAllByChassisID(offset, limit, chassisID)
		return count, fans, err
	}

	if !hasFilters && !hasChassis {
		count, fans, err = p.FanStorage.GetAll(offset, limit)
		if err != nil {
			return count, fans, err
		}
	}

	return count, fans, err
}
