package resource

import (
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
)

// ScannedPortResource for api2go routes
type ScannedPortResource struct {
	ScannedPortStorage *storage.ScannedPortStorage
}

// FindAll Scans
func (s ScannedPortResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, scans, err := s.queryAndCountAllWrapper(r)
	return &Response{Res: scans}, err
}

// FindOne Scanner
func (s ScannedPortResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := s.ScannedPortStorage.GetOne(strings.Replace(ID, "-", "/", -1))
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load Scans in chunks
func (s ScannedPortResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, scans, err := s.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: scans}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (s ScannedPortResource) queryAndCountAllWrapper(r api2go.Request) (count int, scans []model.ScannedPort, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, scans, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, scans, err = s.ScannedPortStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, scans, err
		}
	}

	if !hasFilters {
		count, scans, err = s.ScannedPortStorage.GetAll(offset, limit)
		if err != nil {
			return count, scans, err
		}
	}

	return count, scans, err
}
