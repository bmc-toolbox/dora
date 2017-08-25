package filter

import (
	"strings"

	"github.com/manyminds/api2go"
	log "github.com/sirupsen/logrus"
)

// Filter is meant to store the filters of requested via api
type Filter struct {
	Filter    map[string][]string
	Exclusion bool
}

// Filters is is the collection of filters received on the api call
type Filters struct {
	filters []*Filter
}

// NewFilterSet returns an empty new filter structure
func NewFilterSet(r *api2go.Request) (f *Filters, hasFilters bool) {
	f = &Filters{}
	for key, values := range r.QueryParams {
		if strings.HasPrefix(key, "filter") {
			hasFilters = true
			exclusion := false
			if strings.HasSuffix(key, "!") {
				exclusion = true
				key = key[:len(key)-1]
			}
			filter := strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]")
			f.Add(filter, values, exclusion)
			log.WithFields(log.Fields{"step": "request filter", "filter": filter, "values": values}).Debug("Dora web request with filters")
		}
	}

	return f, hasFilters
}

// Add adds a new filter to the filter map
func (f *Filters) Add(name string, values []string, exclusion bool) {
	ft := &Filter{
		Filter:    map[string][]string{name: values},
		Exclusion: exclusion,
	}

	f.filters = append(f.filters, ft)
}

// Get retrieve all filters
func (f *Filters) Get() []*Filter {
	return f.filters
}

// Clean cleanup the current filter list
func (f *Filters) Clean() {
	f.filters = make([]*Filter, 0)
}

// OffSetAndLimitParse parsers the limit and offset of the requests
func OffSetAndLimitParse(r *api2go.Request) (offset string, limit string) {
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
