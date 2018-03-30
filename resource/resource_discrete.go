package resource

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/go/dora/filter"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/storage"
)

// DiscreteResource for api2go routes
type DiscreteResource struct {
	DiscreteStorage *storage.DiscreteStorage
}

// FindAll Discretes
func (d DiscreteResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, discretes, err := d.queryAndCountAllWrapper(r)
	return &Response{Res: discretes}, err
}

// FindOne Discrete
func (d DiscreteResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := d.DiscreteStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load discretes in chunks
func (d DiscreteResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, discretes, err := d.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: discretes}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a stardard way
func (d DiscreteResource) queryAndCountAllWrapper(r api2go.Request) (count int, discretes []model.Discrete, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, discretes, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, discretes, err = d.DiscreteStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, discretes, err
		}
	}

	include, hasInclude := r.QueryParams["include"]
	if hasInclude && include[0] == "nics" {
		if len(discretes) == 0 {
			count, discretes, err = d.DiscreteStorage.GetAllWithAssociations(offset, limit)
		} else {
			var discretesWithInclude []model.Discrete
			for _, bl := range discretes {
				blWithInclude, err := d.DiscreteStorage.GetOne(bl.Serial)
				if err != nil {
					return count, discretes, err
				}
				discretesWithInclude = append(discretesWithInclude, blWithInclude)
			}
			discretes = discretesWithInclude
		}
	}

	nicsID, hasNIC := r.QueryParams["nicsID"]
	if hasNIC {
		count, discretes, err = d.DiscreteStorage.GetAllByNicsID(offset, limit, nicsID)
		if err != nil {
			return count, discretes, err
		}
	}

	psusID, hasPSU := r.QueryParams["psusID"]
	if hasPSU {
		count, discretes, err = d.DiscreteStorage.GetAllByPsusID(offset, limit, psusID)
		if err != nil {
			return count, discretes, err
		}
	}

	disksID, hasDisk := r.QueryParams["disksID"]
	if hasDisk {
		count, discretes, err = d.DiscreteStorage.GetAllByDisksID(offset, limit, disksID)
		if err != nil {
			return count, discretes, err
		}
	}

	if !hasFilters && !hasInclude && !hasNIC && !hasDisk && !hasPSU {
		count, discretes, err = d.DiscreteStorage.GetAll(offset, limit)
		if err != nil {
			return count, discretes, err
		}
	}

	return count, discretes, err
}
