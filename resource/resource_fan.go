package resource

import (
	"net/http"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// PsuResource for api2go routes
type PsuResource struct {
	PsuStorage *storage.PsuStorage
}

// FindAll Psus
func (p PsuResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, psus, err := p.queryAndCountAllWrapper(r)
	return &Response{Res: psus}, err
}

// FindOne Psu
func (p PsuResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := p.PsuStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load psus in chunks
func (p PsuResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, psus, err := p.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: psus}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (p PsuResource) queryAndCountAllWrapper(r api2go.Request) (count int, psus []model.Psu, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, psus, ErrPageSizeAndNumber
		}
	}

	offset, limit := filter.OffSetAndLimitParse(&r)
	filters, hasFilters := filter.NewFilterSet(&r)
	if hasFilters {
		count, psus, err = p.PsuStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, psus, err
		}
	}

	chassisID, hasChassis := r.QueryParams["chassisID"]
	if hasChassis {
		count, psus, err = p.PsuStorage.GetAllByChassisID(offset, limit, chassisID)
		return count, psus, err
	}

	discreteID, hasDiscrete := r.QueryParams["discretesID"]
	if hasDiscrete {
		count, psus, err = p.PsuStorage.GetAllByDiscreteID(offset, limit, discreteID)
		return count, psus, err
	}

	if !hasFilters && !hasChassis && !hasDiscrete {
		count, psus, err = p.PsuStorage.GetAll(offset, limit)
		if err != nil {
			return count, psus, err
		}
	}

	return count, psus, err
}
