package storage

import (
	"github.com/bmc-toolbox/dora/model"
	"github.com/jinzhu/gorm"

	// Imports for the PostgreSQL database backends
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// Imports for the MySQL database backends
	_ "github.com/jinzhu/gorm/dialects/mysql"
	// Imports for the sqlite database backends
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/spf13/viper"
)

var (
	db   *gorm.DB
	rodb *gorm.DB
	err  error
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
	db.DB().SetMaxIdleConns(viper.GetInt("database_max_connections") / 2)
	db.DB().SetMaxOpenConns(viper.GetInt("database_max_connections"))

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
		&model.Fan{},
	)

	return db
}

// InitRODB creates a new read only db handler
func InitRODB() *gorm.DB {
	if rodb != nil {
		return rodb
	}

	rodb, err = gorm.Open(viper.GetString("database_type"), viper.GetString("ro_database_options"))
	if err != nil {
		panic(err)
	}
	rodb.DB().SetMaxIdleConns(viper.GetInt("database_max_connections") / 2)
	rodb.DB().SetMaxOpenConns(viper.GetInt("database_max_connections"))

	rodb.LogMode(viper.GetBool("debug"))
	rodb.SingularTable(true)

	return rodb
}
