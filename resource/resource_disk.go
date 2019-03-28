package resource

import (
	"net/http"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// DiskResource for api2go routes
type DiskResource struct {
	DiskStorage *storage.DiskStorage
}

// FindAll disks
func (d DiskResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, disks, err := d.queryAndCountAllWrapper(r)
	return &Response{Res: disks}, err
}

// FindOne disks
func (d DiskResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := d.DiskStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load disks in chunks
func (d DiskResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, disks, err := d.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: disks}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (d DiskResource) queryAndCountAllWrapper(r api2go.Request) (count int, disks []model.Disk, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, disks, ErrPageSizeAndNumber
		}
	}

	offset, limit := filter.OffSetAndLimitParse(&r)
	filters, hasFilters := filter.NewFilterSet(&r)
	if hasFilters {
		count, disks, err = d.DiskStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, disks, err
		}
	}

	bladeID, hasBlade := r.QueryParams["bladesID"]
	if hasBlade {
		count, disks, err = d.DiskStorage.GetAllByBladeID(offset, limit, bladeID)
		return count, disks, err
	}

	discreteID, hasDiscrete := r.QueryParams["discretesID"]
	if hasDiscrete {
		count, disks, err = d.DiskStorage.GetAllByDiscreteID(offset, limit, discreteID)
		return count, disks, err
	}

	if !hasFilters && !hasBlade && !hasDiscrete {
		count, disks, err = d.DiskStorage.GetAll(offset, limit)
		if err != nil {
			return count, disks, err
		}
	}

	return count, disks, err
}
