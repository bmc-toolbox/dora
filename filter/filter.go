package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/manyminds/api2go"
	log "github.com/sirupsen/logrus"
)

var (
	simpleFiltering   = regexp.MustCompile(`filter\[(.+)\]`)
	extendedFiltering = regexp.MustCompile(`filter\[(.+)\]\[(.+)\]`)
)

func operation(field string, o string) string {
	switch o {
	case "ne":
		return fmt.Sprintf("\"%s\" != ?", field)
	case "gt":
		return fmt.Sprintf("\"%s\" > ?", field)
	case "ge":
		return fmt.Sprintf("\"%s\" >= ?", field)
	case "lt":
		return fmt.Sprintf("\"%s\" < ?", field)
	case "le":
		return fmt.Sprintf("\"%s\" <= ?", field)
	case "eq":
		return fmt.Sprintf("\"%s\" = ?", field)
	default:
		return ""
	}
}

// Filter is meant to store the filters of requested via api
type Filter struct {
	Filter   map[string][]string
	Operator string
}

// Filters is is the collection of filters received on the api call
type Filters struct {
	filters []*Filter
}

// NewFilterSet returns an empty new filter structure
func NewFilterSet(r *api2go.Request) (f *Filters, hasFilters bool) {
	f = &Filters{}
	for key, values := range r.QueryParams {
		filter := extendedFiltering.FindStringSubmatch(key)
		if len(filter) == 0 {
			filter = simpleFiltering.FindStringSubmatch(key)
			if len(filter) != 0 {
				hasFilters = true
				if strings.HasSuffix(key, "!") {
					f.Add(filter[1], values, "ne")
				} else {
					f.Add(filter[1], values, "eq")
				}
				log.WithFields(log.Fields{"step": "request filter", "filter": filter, "values": values}).Debug("request with filters")
			}
		} else {
			hasFilters = true
			f.Add(filter[1], values, filter[2])
		}
	}
	return f, hasFilters
}

// Add adds a new filter to the filter map
func (f *Filters) Add(name string, values []string, operator string) {
	ft := &Filter{
		Filter:   map[string][]string{name: values},
		Operator: operator,
	}

	f.filters = append(f.filters, ft)
}

// Get retrieve all filters
func (f *Filters) Get() []*Filter {
	return f.filters
}

// BuildQuery receive a model as an interface and builds a query out of it
func (f *Filters) BuildQuery(m interface{}, db *gorm.DB) (q *gorm.DB, err error) {
	q = db
	for _, filter := range f.Get() {
		for key, values := range filter.Filter {
			if len(values) == 1 && values[0] == "" {
				continue
			}
			rfct := reflect.ValueOf(m)
			rfctType := rfct.Type()

			var structJSONMemberName string
			for i := 0; i < rfctType.NumField(); i++ {
				jsondName := rfctType.Field(i).Tag.Get("json")
				if key == jsondName {
					structJSONMemberName = jsondName
					break
				}
			}

			if structJSONMemberName == "" || structJSONMemberName == "-" {
				return q, err
			}

			op := operation(structJSONMemberName, filter.Operator)
			if op == "" {
				return nil, api2go.NewHTTPError(nil, fmt.Sprintf("Invalid filter operation: %s", filter.Operator), 400)
			}

			for _, value := range values {
				q = db.Where(op, value)
			}
		}
	}
	return q, err
}

// Clean cleanup the current filter list
func (f *Filters) Clean() {
	f.filters = make([]*Filter, 0)
}

// OffSetAndLimitParse parsers the limit and offset of the requests
func OffSetAndLimitParse(r *api2go.Request) (offset string, limit string) {
	offsetQuery, hasOffset := r.QueryParams["page[offset]"]
	limitQuery, hasLimit := r.QueryParams["page[limit]"]

	if hasOffset {
		offset = offsetQuery[0]
	}

	if hasLimit {
		limit = limitQuery[0]
	}

	if hasOffset && offset == "" {
		offset = "0"
	}

	if (hasLimit && limit == "") || (hasOffset && !hasLimit) {
		limit = "100"
	}

	return offset, limit
}
