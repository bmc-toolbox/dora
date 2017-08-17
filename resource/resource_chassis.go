package resource

import (
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// ChassisResource for api2go routes
type ChassisResource struct {
	ChassisStorage *storage.ChassisStorage
	BladeStorage   *storage.BladeStorage
}

// FindAll Chassis
func (c ChassisResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var chassis []model.Chassis
	var err error
	hasFilters := false
	filters := NewFilter()
	include, hasInclude := r.QueryParams["include"]
	bladesID, hasBlade := r.QueryParams["bladesID"]

	for key, values := range r.QueryParams {
		if strings.HasPrefix(key, "filter") {
			hasFilters = true
			filter := strings.TrimRight(strings.TrimLeft(key, "filter["), "]")
			filters.Add(filter, values)
		}
	}

	if hasFilters {
		chassis, err = c.ChassisStorage.GetAllByFilters(filters.Get())
		if err != nil {
			return &Response{Res: chassis}, err
		}
	}

	if hasInclude && include[0] == "blades" {
		if len(chassis) == 0 {
			chassis, err = c.ChassisStorage.GetAllWithAssociations()
		} else {
			var chassisWithInclude []model.Chassis
			for _, ch := range chassis {
				chWithInclude, err := c.ChassisStorage.GetOne(ch.Serial)
				if err != nil {
					return &Response{}, err
				}
				chassisWithInclude = append(chassisWithInclude, chWithInclude)
			}
			chassis = chassisWithInclude
		}
	}

	if hasBlade {
		chassis, err = c.ChassisStorage.GetAllByBladesID(bladesID)
	}

	if !hasFilters && !hasInclude && !hasBlade {
		chassis, err = c.ChassisStorage.GetAll()
		if err != nil {
			return &Response{}, err
		}
	}
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
