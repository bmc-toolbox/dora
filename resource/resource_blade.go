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

type byInt64Slice []int

func (b byInt64Slice) Len() int           { return len(b) }
func (b byInt64Slice) Swap(x, y int)      { b[x], b[y] = b[y], b[x] }
func (b byInt64Slice) Less(x, y int) bool { return b[x] < b[y] }

// PaginatedFindAll can be used to load blades in chunks
func (b BladeResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	var (
		result                      []model.Blade
		number, size, offset, limit string
		keys                        []int
	)

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
			return 0, &Response{Res: blades}, err
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
					return 0, &Response{}, err
				}
				bladesWithInclude = append(bladesWithInclude, blWithInclude)
			}
			blades = bladesWithInclude
		}
	}

	if hasChassis {
		blades, err = b.BladeStorage.GetAllByChassisID(chassisID)
		if err != nil {
			return 0, &Response{Res: blades}, err
		}
	}

	if hasNIC {
		blades, err = b.BladeStorage.GetAllByNicsID(nicsID)
		if err != nil {
			return 0, &Response{Res: blades}, err
		}
	}

	if !hasFilters && !hasChassis && !hasInclude && !hasNIC {
		blades, err = b.BladeStorage.GetAll()
		if err != nil {
			return 0, &Response{Res: blades}, err
		}
	}

	for k := range blades {
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
			if i >= uint64(len(blades)) {
				break
			}
			result = append(result, blades[keys[i]])
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
			if i >= uint64(len(blades)) {
				break
			}
			result = append(result, blades[keys[i]])
		}
	}

	return uint(len(blades)), &Response{Res: result}, nil
}

// FindOne Blade
func (b BladeResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := b.BladeStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}
