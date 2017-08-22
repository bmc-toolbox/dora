package resource

import (
	"strings"

	"github.com/manyminds/api2go"
)

// Filter is meant to store the filters of requested via api
type Filter struct {
	filters map[string][]string
}

// NewFilter returns an empty new filter structure
func NewFilter(r *api2go.Request) (f *Filter, hasFilters bool) {
	f = &Filter{filters: make(map[string][]string)}
	for key, values := range r.QueryParams {
		if strings.HasPrefix(key, "filter") {
			hasFilters = true
			filter := strings.TrimRight(strings.TrimLeft(key, "filter["), "]")
			f.Add(filter, values)
		}
	}

	return f, hasFilters
}

// Add adds a new filter to the filter map
func (f *Filter) Add(name string, values []string) {
	f.filters[name] = values
}

// Get retrieve all filters
func (f *Filter) Get() map[string][]string {
	return f.filters
}

// Clean cleanup the current filter list
func (f *Filter) Clean() {
	f.filters = make(map[string][]string)
}

func offSetAndLimitParse(r *api2go.Request) (offset string, limit string) {
	offsetQuery, hasOffset := r.QueryParams["page[offset]"]
	if hasOffset {
		offset = offsetQuery[0]
	}

	if hasOffset && offset == "" {
		offset = "0"
	}

	limitQuery, hasLimit := r.QueryParams["page[limit]"]
	if hasLimit {
		limit = limitQuery[0]
	}

	if hasLimit && limit == "" {
		limit = "100"
	}

	return offset, limit
}
