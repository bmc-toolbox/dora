package filter

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/jinzhu/gorm"
	mocket "github.com/selvatico/go-mocket"

	"github.com/bmc-toolbox/dora/model"
	"github.com/manyminds/api2go"
	"github.com/stretchr/testify/assert"
)

var testSet = []struct {
	urlString string
	sqlQuery  string
}{
	{"filter[model]=dell", "&{SELECT * FROM \"\"  WHERE (\"model\" = ?) [dell]}"},
	{"filter[status]!=bad", "&{SELECT * FROM \"\"  WHERE (\"status\" != ?) [bad]}"},
	{"filter[model][eq]=dell", "&{SELECT * FROM \"\"  WHERE (\"model\" = ?) [dell]}"},
	{"filter[status][ne]=bad", "&{SELECT * FROM \"\"  WHERE (\"status\" != ?) [bad]}"},
	{"filter[temp_c][le]=3", "&{SELECT * FROM \"\"  WHERE (\"temp_c\" <= ?) [3]}"},
	{"filter[temp_c][lt]=3", "&{SELECT * FROM \"\"  WHERE (\"temp_c\" < ?) [3]}"},
	{"filter[temp_c][ge]=3", "&{SELECT * FROM \"\"  WHERE (\"temp_c\" >= ?) [3]}"},
	{"filter[temp_c][gt]=3", "&{SELECT * FROM \"\"  WHERE (\"temp_c\" > ?) [3]}"},
}

func setupDB() *gorm.DB {
	mocket.Catcher.Register()
	mocket.Catcher.Logging = true
	// GORM
	db, _ := gorm.Open(mocket.DriverName, "connection_string") // Can be any connection string

	return db
}

func TestEqualSignAndExclamationMark(t *testing.T) {
	db := setupDB()
	for _, testPair := range testSet {
		queryParams, _ := url.ParseQuery(testPair.urlString)
		request := api2go.Request{
			QueryParams: queryParams,
		}
		filters, hasFilters := NewFilterSet(&request)

		assert.EqualValues(t, true, hasFilters,
			"filter is created")

		q, err := filters.BuildQuery(model.Chassis{}, db)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, testPair.sqlQuery, fmt.Sprintf("%s", q.QueryExpr()))
	}
}
