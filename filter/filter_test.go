package filter_test

import (
	"fmt"

	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
)

func ExampleFilters() {
	queryParams := map[string][]string{
		"filter[bmc_type]": []string{"iLO4"},
	}

	filters, hasFilters := filter.NewFilterSet(&queryParams)
	fmt.Println(hasFilters)

	query, err := filters.BuildQuery(model.Blade{})
	if err != nil {
		panic(err)
	}
	fmt.Println(query)

	queryParams = map[string][]string{
		"filter[bmc_type]!": []string{"iLO4"},
	}

	filters, hasFilters = filter.NewFilterSet(&queryParams)
	fmt.Println(hasFilters)

	query, err = filters.BuildQuery(model.Blade{})
	if err != nil {
		panic(err)
	}
	fmt.Println(query)

	queryParams = map[string][]string{}
	filters, hasFilters = filter.NewFilterSet(&queryParams)
	fmt.Println(hasFilters)

	// Output:
	// true
	// bmc_type in ('iLO4')
	// true
	// bmc_type not in ('iLO4')
	// false
}
