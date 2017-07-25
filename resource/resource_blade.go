package resource

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// BladeResource for api2go routes
type BladeResource struct {
	BladeStorage   *storage.BladeStorage
	ChassisStorage *storage.ChassisStorage
}

// FindAll Blades
func (b BladeResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var blades []model.Blade
	var err error
	filterSerial, hasFilters := r.QueryParams["filter[serial]"]
	chassisID, hasChassis := r.QueryParams["chassisID"]

	for key, values := range r.QueryParams {
		fmt.Println(key, values)
	}

	if hasFilters {
		// Here it means we want to return all blades matching the given serial numbers
		blades, err = b.BladeStorage.GetBySerial(filterSerial)
		if err != nil {
			return &Response{}, err
		}
		return &Response{Res: blades}, nil
	}

	if hasChassis {
		blades, err = b.BladeStorage.GetAllByChassisID(chassisID)
	}

	if !hasFilters && !hasChassis /* && !hasInclude */ {
		blades, err = b.BladeStorage.GetAll()
		if err != nil {
			return &Response{Res: blades}, err
		}
	}

	return &Response{Res: blades}, nil
}

// FindOne Blade
func (b BladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	id, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(ErrInvalidID, ErrInvalidID.Error(), http.StatusBadRequest)
	}

	res, err := b.BladeStorage.GetOne(id)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
