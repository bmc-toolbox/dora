package filter_test

import (
	"fmt"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/manyminds/api2go"
)

func ExampleFilters() {
	request := api2go.Request{
		QueryParams: map[string][]string{
			"filter[bmc_type]": []string{"iLO4"},
		},
	}

	filters, hasFilters := filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	query, err := filters.BuildQuery(model.Blade{})
	if err != nil {
		panic(err)
	}
	fmt.Println(query)

	request.QueryParams = map[string][]string{
		"filter[bmc_type]!": []string{"iLO4"},
	}

	filters, hasFilters = filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	query, err = filters.BuildQuery(model.Blade{})
	if err != nil {
		panic(err)
	}
	fmt.Println(query)

	filters.Clean()
	query, err = filters.BuildQuery(model.Blade{})
	if err != nil {
		panic(err)
	}
	fmt.Println(query)

	request.QueryParams = map[string][]string{}
	filters, hasFilters = filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	// Output:
	// true
	// bmc_type in ('iLO4')
	// true
	// bmc_type not in ('iLO4')
	//
	// false
}

func ExampleOffSetAndLimitParse() {
	request := api2go.Request{
		QueryParams: map[string][]string{
			"page[offset]": []string{"100"},
			"page[limit]":  []string{"10"},
		},
	}

	offset, limit := filter.OffSetAndLimitParse(&request)
	fmt.Println(offset)
	fmt.Println(limit)

	request.QueryParams = map[string][]string{
		"page[offset]": []string{"100"},
	}

	offset, limit = filter.OffSetAndLimitParse(&request)
	fmt.Println(offset)
	fmt.Println(limit)

	// Output:
	// 100
	// 10
	// 100
	// 100
}
