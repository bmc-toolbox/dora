package storage

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/model"
)

var (
	db  *gorm.DB
	err error
)

// InitDB creates and migrates the database
func InitDB() *gorm.DB {
	if db != nil && db.DB().Ping() == nil {
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
	db.AutoMigrate(&model.Blade{}, &model.Chassis{}, &model.Nic{}, &model.ScannedHost{}, &model.ScannedPort{})

	return db
}
