package resource

import (
	"net/http"
	"sort"
	"strconv"
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
		filters.Clean()
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

// PaginatedFindAll can be used to load chassis in chunks
func (c ChassisResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	var (
		result                      []model.Chassis
		number, size, offset, limit string
		keys                        []int
	)

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
		filters.Clean()
		if err != nil {
			return 0, &Response{Res: chassis}, err
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
					return 0, &Response{}, err
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
			return 0, &Response{}, err
		}
	}

	for k := range chassis {
		keys = append(keys, k)
	}
	sort.Sort(byInt64Slice(keys))

	numberQuery, ok := r.QueryParams["page[number]"]
	if ok {
		number = numberQuery[0]
	}
	sizeQuery, ok := r.QueryParams["page[size]"]
	if ok {
		size = sizeQuery[0]
	}
	offsetQuery, ok := r.QueryParams["page[offset]"]
	if ok {
		offset = offsetQuery[0]
	}
	limitQuery, ok := r.QueryParams["page[limit]"]
	if ok {
		limit = limitQuery[0]
	}

	if size != "" {
		sizeI, err := strconv.ParseUint(size, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		numberI, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		start := sizeI * (numberI - 1)
		for i := start; i < start+sizeI; i++ {
			if i >= uint64(len(chassis)) {
				break
			}
			result = append(result, chassis[keys[i]])
		}
	} else {
		limitI, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		offsetI, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		for i := offsetI; i < offsetI+limitI; i++ {
			if i >= uint64(len(chassis)) {
				break
			}
			result = append(result, chassis[keys[i]])
		}
	}

	return uint(len(chassis)), &Response{Res: result}, nil
}

// FindOne Chassis
func (c ChassisResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := c.ChassisStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
