package resource

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// StorageBladeResource for api2go routes
type StorageBladeResource struct {
	StorageBladeStorage *storage.StorageBladeStorage
}

// FindAll StorageBlades
func (s StorageBladeResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, storageblades, err := s.queryAndCountAllWrapper(r)
	return &Response{Res: storageblades}, err
}

// FindOne StorageBlade
func (s StorageBladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := s.StorageBladeStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load Storageblades in chunks
func (s StorageBladeResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, storageblades, err := s.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: storageblades}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a stardard way
func (s StorageBladeResource) queryAndCountAllWrapper(r api2go.Request) (count int, storageblades []model.StorageBlade, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, storageblades, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, storageblades, err = s.StorageBladeStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, storageblades, err
		}
	}

	chassisID, hasChassis := r.QueryParams["chassisID"]
	if hasChassis {
		count, storageblades, err = s.StorageBladeStorage.GetAllByChassisID(offset, limit, chassisID)
		if err != nil {
			return count, storageblades, err
		}
	}

	bladesID, hasBlade := r.QueryParams["bladesID"]
	if hasBlade {
		count, storageblades, err = s.StorageBladeStorage.GetAllByBladeID(offset, limit, bladesID)
		if err != nil {
			return count, storageblades, err
		}
	}

	if !hasFilters && !hasChassis && !hasBlade {
		count, storageblades, err = s.StorageBladeStorage.GetAll(offset, limit)
		if err != nil {
			return count, storageblades, err
		}
	}

	return count, storageblades, err
}
