package resource

import (
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// BladeResource for api2go routes
type BladeResource struct {
	BladeStorage   *storage.BladeStorage
	ChassisStorage *storage.ChassisStorage
	NicStorage     *storage.NicStorage
}

// FindAll Blades
func (b BladeResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, blades, err := b.queryAndCountAllWrapper(r)
	return &Response{Res: blades}, err
}

func (b BladeResource) queryAndCountAllWrapper(r api2go.Request) (count int, blades []model.Blade, err error) {
	filters := NewFilter()
	hasFilters := false
	var offset string
	var limit string

	include, hasInclude := r.QueryParams["include"]
	chassisID, hasChassis := r.QueryParams["chassisID"]
	nicsID, hasNIC := r.QueryParams["nicsID"]
	offsetQuery, hasOffset := r.QueryParams["page[offset]"]
	if hasOffset {
		offset = offsetQuery[0]
	}

	limitQuery, hasLimit := r.QueryParams["page[limit]"]
	if hasLimit {
		limit = limitQuery[0]
	}

	for key, values := range r.QueryParams {
		if strings.HasPrefix(key, "filter") {
			hasFilters = true
			filter := strings.TrimRight(strings.TrimLeft(key, "filter["), "]")
			filters.Add(filter, values)
		}
	}

	if hasFilters {
		count, blades, err = b.BladeStorage.GetAllByFilters(offset, limit, filters.Get())
		filters.Clean()
		if err != nil {
			return count, blades, err
		}
	}

	if hasInclude && include[0] == "nics" {
		if len(blades) == 0 {
			count, blades, err = b.BladeStorage.GetAllWithAssociations(offset, limit)
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

	if hasChassis {
		count, blades, err = b.BladeStorage.GetAllByChassisID(offset, limit, chassisID)
		if err != nil {
			return count, blades, err
		}
	}

	if hasNIC {
		count, blades, err = b.BladeStorage.GetAllByNicsID(nicsID)
		if err != nil {
			return count, blades, err
		}
	}

	if !hasFilters && !hasChassis && !hasInclude && !hasNIC {
		count, blades, err = b.BladeStorage.GetAll(offset, limit)
		if err != nil {
			return count, blades, err
		}
	}

	return count, blades, err
}

// PaginatedFindAll can be used to load blades in chunks
func (b BladeResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, blades, err := b.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: blades}, err
}

// FindOne Blade
func (b BladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := b.BladeStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
