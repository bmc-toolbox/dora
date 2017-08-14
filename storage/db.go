package storage

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/model"
)

// InitDB creates and migrates the database
func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(viper.GetString("database_type"), viper.GetString("database_options"))
	if err != nil {
		return nil, err
	}

	db.LogMode(viper.GetBool("debug"))
	db.SingularTable(true)
	db.AutoMigrate(&model.Blade{}, &model.Chassis{}, &model.Nic{})

	return db, nil
}
