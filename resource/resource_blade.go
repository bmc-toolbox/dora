package resource

import (
	"net/http"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
)

// BladeResource for api2go routes
type BladeResource struct {
	BladeStorage *storage.BladeStorage
}

// FindAll Blades
func (b BladeResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, blades, err := b.queryAndCountAllWrapper(r)
	return &Response{Res: blades}, err
}

// FindOne Blade
func (b BladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := b.BladeStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load blades in chunks
func (b BladeResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, blades, err := b.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: blades}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (b BladeResource) queryAndCountAllWrapper(r api2go.Request) (count int, blades []model.Blade, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, blades, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, blades, err = b.BladeStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, blades, err
		}
	}

	include, hasInclude := r.QueryParams["include"]
	if hasInclude {
		if len(blades) == 0 {
			count, blades, err = b.BladeStorage.GetAllWithAssociations(offset, limit, include)
		} else {
			var bladesWithInclude []model.Blade
			for _, bl := range blades {
				blWithInclude, err := b.BladeStorage.GetOne(bl.Serial)
				if err != nil {
					return count, blades, err
				}
				bladesWithInclude = append(bladesWithInclude, blWithInclude)
			}
			blades = bladesWithInclude
		}
	}

	chassisID, hasChassis := r.QueryParams["chassisID"]
	if hasChassis {
		count, blades, err = b.BladeStorage.GetAllByChassisID(offset, limit, chassisID)
		if err != nil {
			return count, blades, err
		}
	}

	nicsID, hasNIC := r.QueryParams["nicsID"]
	if hasNIC {
		count, blades, err = b.BladeStorage.GetAllByNicsID(offset, limit, nicsID)
		if err != nil {
			return count, blades, err
		}
	}

	storageBladesID, hasStorageBlade := r.QueryParams["storage_bladesID"]
	if hasStorageBlade {
		count, blades, err = b.BladeStorage.GetAllByStorageBladesID(offset, limit, storageBladesID)
		if err != nil {
			return count, blades, err
		}
	}

	disksID, hasDisk := r.QueryParams["disksID"]
	if hasDisk {
		count, blades, err = b.BladeStorage.GetAllByDisksID(offset, limit, disksID)
		if err != nil {
			return count, blades, err
		}
	}

	if !hasFilters && !hasChassis && !hasInclude && !hasNIC && !hasStorageBlade && !hasDisk {
		count, blades, err = b.BladeStorage.GetAll(offset, limit)
		if err != nil {
			return count, blades, err
		}
	}

	return count, blades, err
}
