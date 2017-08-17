package resource

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// BladeResource for api2go routes
type NicResource struct {
	BladeStorage *storage.BladeStorage
	NicStorage   *storage.NicStorage
}

// FindAll Nics
func (n NicResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var nics []model.Nic
	var err error
	_, hasFilters := r.QueryParams["filter[mac]"]
	bladeID, hasBlade := r.QueryParams["bladeID"]

	if hasBlade {
		nics, err = n.NicStorage.GetAllByBladeID(bladeID)
	}

	if !hasFilters && !hasBlade {
		nics, err = n.NicStorage.GetAll()
		if err != nil {
			return &Response{Res: nics}, err
		}
	}

	return &Response{Res: nics}, nil
}

// PaginatedFindAll can be used to load nics in chunks
func (n NicResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	var (
		result                      []model.Nic
		number, size, offset, limit string
		keys                        []int
	)

	var nics []model.Nic
	var err error
	_, hasFilters := r.QueryParams["filter[mac]"]
	bladeID, hasBlade := r.QueryParams["bladeID"]

	if hasBlade {
		nics, err = n.NicStorage.GetAllByBladeID(bladeID)
	}

	if !hasFilters && !hasBlade {
		nics, err = n.NicStorage.GetAll()
		if err != nil {
			return 0, &Response{Res: nics}, err
		}
	}

	for k := range nics {
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
			if i >= uint64(len(nics)) {
				break
			}
			result = append(result, nics[keys[i]])
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
			if i >= uint64(len(nics)) {
				break
			}
			result = append(result, nics[keys[i]])
		}
	}

	return uint(len(nics)), &Response{Res: result}, nil
}

// FindOne Nics
func (n NicResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := n.NicStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
