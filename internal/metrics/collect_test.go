package metrics

import (
	"github.com/bmc-toolbox/dora/storage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

func TestNewAndUpdateCandle(t *testing.T) {
	s := Stats{StartTime: time.Time{}}

	oldUptime := s.Uptime
	s.UpdateUptime()
	assert.NotEqual(t, oldUptime, s.Uptime,
		"uptime updated")

	dbInit, mock, _ := sqlmock.New()
	db, _ := gorm.Open("postgres", dbInit)
	defer db.Close()

	resultRows := sqlmock.NewRows([]string{"count(*)"}).
		AddRow(10).
		AddRow(5).
		AddRow(4).
		AddRow(3)

	mock.ExpectQuery("^SELECT count\\(\\*\\) FROM \"chasses\"$").WillReturnRows(resultRows)
	mock.ExpectQuery("^SELECT count\\(\\*\\) FROM \"chasses\" WHERE \"updated_at\".*").WillReturnRows(resultRows)
	mock.ExpectQuery("^SELECT count\\(\\*\\) FROM \"chasses\" WHERE \\(vendor in \\('hp'\\)\\)$").WillReturnRows(resultRows)
	mock.ExpectQuery("^SELECT count\\(\\*\\) FROM \"chasses\" WHERE \\(vendor in \\('hp'\\)\\) AND \"updated_at\".*$").WillReturnRows(resultRows)
	// to prevent errors like "all expectations were already fulfilled" in logs
	zeroRow := sqlmock.NewRows([]string{"count(*)"})
	for i := 0; i <= 100; i++ {
		zeroRow.AddRow(0)
		mock.ExpectQuery("^SELECT count\\(\\*\\) FROM \".+\".*").WillReturnRows(zeroRow)
	}

	s.GatherDBStats(
		storage.NewChassisStorage(db),
		storage.NewBladeStorage(db),
		storage.NewDiscreteStorage(db),
		storage.NewNicStorage(db),
		storage.NewStorageBladeStorage(db),
		storage.NewScannedPortStorage(db),
		storage.NewPsuStorage(db),
		storage.NewDiskStorage(db),
		storage.NewFanStorage(db),
	)

	assert.EqualValues(t, 10, s.Chassis.Total,
		"total count of chassis is right")
	assert.EqualValues(t, 5, s.Chassis.Updated24hAgo,
		"total count of updated 24h ago chassis is right")
	assert.EqualValues(t, 4, s.Chassis.Vendors["hp"].Total,
		"total value of hp chassis is right")
	assert.EqualValues(t, 3, s.Chassis.Vendors["hp"].Updated24hAgo,
		"total count of updated 24h ago hp chassis is right")
}
