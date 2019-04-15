package filter

import (
	"net/url"
	"testing"

	"github.com/bmc-toolbox/dora/model"
	"github.com/manyminds/api2go"
	"github.com/stretchr/testify/assert"
)

var testSet = []struct {
	urlString string
	sqlQuery  string
}{
	{"filter[model]=dell", "model in ('dell')"},
	{"filter[status]!=bad", "status not in ('bad')"},
}

func TestEqualSignAndExclamationMark(t *testing.T) {
	for _, testPair := range testSet {
		queryParams, _ := url.ParseQuery(testPair.urlString)
		request := api2go.Request{
			QueryParams: queryParams,
		}
		filters, hasFilters := NewFilterSet(&request)

		assert.EqualValues(t, true, hasFilters,
			"filter is created")

		query, err := filters.BuildQuery(model.Chassis{})
		if err != nil {
			panic(err)
		}
		assert.Equal(t, testPair.sqlQuery, query)
	}
}
