package filter_test

import (
	"fmt"

	"github.com/jinzhu/gorm"
	mocket "github.com/selvatico/go-mocket"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/model"
	"github.com/manyminds/api2go"
)

func setupDB() *gorm.DB {
	mocket.Catcher.Register()
	mocket.Catcher.Logging = true
	// GORM
	db, _ := gorm.Open(mocket.DriverName, "connection_string") // Can be any connection string

	return db
}

func ExampleFilters() {
	db := setupDB()

	request := api2go.Request{
		QueryParams: map[string][]string{
			"filter[bmc_type]": {"iLO4"},
		},
	}

	filters, hasFilters := filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	q, err := filters.BuildQuery(model.Blade{}, db)
	if err != nil {
		panic(err)
	}
	fmt.Println(q.QueryExpr())

	request.QueryParams = map[string][]string{
		"filter[bmc_type]!": {"iLO4"},
	}

	filters, hasFilters = filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	q, err = filters.BuildQuery(model.Blade{}, db)
	if err != nil {
		panic(err)
	}
	fmt.Println(q.QueryExpr())

	filters.Clean()
	q, err = filters.BuildQuery(model.Blade{}, db)
	if err != nil {
		panic(err)
	}
	fmt.Println(q.QueryExpr())

	request.QueryParams = map[string][]string{}
	filters, hasFilters = filter.NewFilterSet(&request)
	fmt.Println(hasFilters)

	// Output:
	// `MOCK_FAKE_DRIVER` is not officially supported, running under compatibility mode.
	// true
	// &{SELECT * FROM ""  WHERE ("bmc_type" in (?)) [iLO4]}
	// true
	// &{SELECT * FROM ""  WHERE ("bmc_type" not in (?)) [iLO4]}
	// &{SELECT * FROM ""   []}
	// false
}

func ExampleOffSetAndLimitParse() {
	request := api2go.Request{
		QueryParams: map[string][]string{
			"page[offset]": {"100"},
			"page[limit]":  {"10"},
		},
	}

	offset, limit := filter.OffSetAndLimitParse(&request)
	fmt.Println(offset)
	fmt.Println(limit)

	request.QueryParams = map[string][]string{
		"page[offset]": {"100"},
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
