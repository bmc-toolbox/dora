package storage

import (
	// Imports for the suported database backends
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	// Imports for the suported database backends
	_ "github.com/go-sql-driver/mysql"
	// Imports for the suported database backends
	_ "github.com/lib/pq"
	// Imports for the suported database backends
	_ "github.com/go-sql-driver/mysql"

	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/model"
)

var (
	db  *gorm.DB
	err error
)

// InitDB creates and migrates the database
func InitDB() *gorm.DB {
	if db != nil {
		return db
	}

	db, err = gorm.Open(viper.GetString("database_type"), viper.GetString("database_options"))
	if err != nil {
		panic(err)
	}

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(80)

	db.LogMode(viper.GetBool("debug"))
	db.SingularTable(true)
	db.AutoMigrate(
		&model.Blade{},
		&model.Discrete{},
		&model.Chassis{},
		&model.Nic{},
		&model.StorageBlade{},
		&model.ScannedPort{},
		&model.Psu{},
		&model.Disk{},
	)

	return db
}
