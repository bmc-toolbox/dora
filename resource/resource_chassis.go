package resource

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// ChassisResource for api2go routes
type ChassisResource struct {
	ChassisStorage *storage.ChassisStorage
}

// FindAll Chassis
func (c ChassisResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, chassis, err := c.queryAndCountAllWrapper(r)
	return &Response{Res: chassis}, err
}

// FindOne Chassis
func (c ChassisResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := c.ChassisStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load chassis in chunks
func (c ChassisResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, chassis, err := c.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: chassis}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a stardard way
func (c ChassisResource) queryAndCountAllWrapper(r api2go.Request) (count int, chassis []model.Chassis, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, chassis, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, chassis, err = c.ChassisStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, chassis, err
		}
	}

	include, hasInclude := r.QueryParams["include"]
	if hasInclude && include[0] == "blades" {
		if len(chassis) == 0 {
			count, chassis, err = c.ChassisStorage.GetAllWithAssociations(offset, limit)
		} else {
			var chassisWithInclude []model.Chassis
			for _, ch := range chassis {
				chWithInclude, err := c.ChassisStorage.GetOne(ch.Serial)
				if err != nil {
					return count, chassis, err
				}
				chassisWithInclude = append(chassisWithInclude, chWithInclude)
			}
			chassis = chassisWithInclude
		}
	}

	bladesID, hasBlade := r.QueryParams["bladesID"]
	if hasBlade {
		count, chassis, err = c.ChassisStorage.GetAllByBladesID(offset, limit, bladesID)
	}

	storageBladesID, hasStorageBlade := r.QueryParams["storage_bladesID"]
	if hasStorageBlade {
		count, chassis, err = c.ChassisStorage.GetAllByStorageBladesID(offset, limit, storageBladesID)
		if err != nil {
			return count, chassis, err
		}
	}

	if !hasFilters && !hasInclude && !hasBlade && !hasStorageBlade {
		count, chassis, err = c.ChassisStorage.GetAll(offset, limit)
		if err != nil {
			return count, chassis, err
		}
	}

	return count, chassis, err
}
