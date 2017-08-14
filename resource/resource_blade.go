package resource

import (
	"fmt"
	"net/http"

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
	_, hasFilters := r.QueryParams["filter[serial]"]
	include, hasInclude := r.QueryParams["include"]
	chassisID, hasChassis := r.QueryParams["chassisID"]

	for key, values := range r.QueryParams {
		fmt.Println(key, values)
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
	}

	if !hasFilters && !hasChassis && !hasInclude {
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
