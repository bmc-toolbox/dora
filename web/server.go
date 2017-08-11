package web

import (
	"fmt"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go-adapter/gingonic"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/resource"
	"gitlab.booking.com/infra/dora/storage"
	gin "gopkg.in/gin-gonic/gin.v1"
)

// RunGin is responsible to spin up the gin webservice
func RunGin(port int, debug bool) {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	api := api2go.NewAPIWithRouting(
		"v0",
		api2go.NewStaticResolver("/"),
		gingonic.New(r),
	)

	db, err := storage.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	chassisStorage := storage.NewChassisStorage(db)
	bladeStorage := storage.NewBladeStorage(db)
	api.AddResource(model.Chassis{}, resource.ChassisResource{BladeStorage: bladeStorage, ChassisStorage: chassisStorage})
	api.AddResource(model.Blade{}, resource.BladeResource{BladeStorage: bladeStorage, ChassisStorage: chassisStorage})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/ping_db", func(c *gin.Context) {
		if db.HasTable("chassis") {
			c.String(200, "pong")
		} else {
			c.String(451, "database has gone away")
		}
	})

	r.Run(fmt.Sprintf(":%d", port))
}
