package resource

import (
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"net/http"
)

// DiscoverHintResource for api2go routes
type DiscoverHintResource struct {
	DiscoverHintStorage *storage.DiscoverHintStorage
}

// FindAll Discover Hints
func (s DiscoverHintResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, discoverHints, err := s.queryAndCountAllWrapper(r)
	return &Response{Res: discoverHints}, err
}

// FindOne Discover Hint
func (s DiscoverHintResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	discoverHint, err := s.DiscoverHintStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: discoverHint}, err
}

// PaginatedFindAll can be used to load Discover Hints in chunks
func (s DiscoverHintResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, discoverHints, err := s.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: discoverHints}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (s DiscoverHintResource) queryAndCountAllWrapper(r api2go.Request) (count int, discoverHints []model.DiscoverHint, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, discoverHints, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, discoverHints, err = s.DiscoverHintStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, discoverHints, err
		}
	} else {
		count, discoverHints, err = s.DiscoverHintStorage.GetAll(offset, limit)
		if err != nil {
			return count, discoverHints, err
		}
	}

	return count, discoverHints, err
}
