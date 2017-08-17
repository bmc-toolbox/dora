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
	var blades []model.Blade
	var err error
	filters := NewFilter()
	hasFilters := false
	include, hasInclude := r.QueryParams["include"]
	chassisID, hasChassis := r.QueryParams["chassisID"]
	nicsID, hasNIC := r.QueryParams["nicsID"]

	for key, values := range r.QueryParams {
		if strings.HasPrefix(key, "filter") {
			hasFilters = true
			filter := strings.TrimRight(strings.TrimLeft(key, "filter["), "]")
			filters.Add(filter, values)
		}
	}

	if hasFilters {
		blades, err = b.BladeStorage.GetAllByFilters(filters.Get())
		if err != nil {
			return &Response{Res: blades}, err
		}
	}

	if hasInclude && include[0] == "nics" {
		if len(blades) == 0 {
			blades, err = b.BladeStorage.GetAllWithAssociations()
		} else {
			var bladesWithInclude []model.Blade
			for _, bl := range blades {
				blWithInclude, err := b.BladeStorage.GetOne(bl.Serial)
				if err != nil {
					return &Response{}, err
				}
				bladesWithInclude = append(bladesWithInclude, blWithInclude)
			}
			blades = bladesWithInclude
		}
	}

	if hasChassis {
		blades, err = b.BladeStorage.GetAllByChassisID(chassisID)
		if err != nil {
			return &Response{Res: blades}, err
		}
	}

	if hasNIC {
		blades, err = b.BladeStorage.GetAllByNicsID(nicsID)
		if err != nil {
			return &Response{Res: blades}, err
		}
	}

	if !hasFilters && !hasChassis && !hasInclude && !hasNIC {
		blades, err = b.BladeStorage.GetAll()
		if err != nil {
			return &Response{Res: blades}, err
		}
	}

	return &Response{Res: blades}, nil
}

// FindOne Blade
func (b BladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := b.BladeStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
