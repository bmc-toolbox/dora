package web

import (
	"fmt"
	"html/template"
	"log"

	"github.com/GeertJohan/go.rice"
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

	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}

	staticBox, err := rice.FindBox("static")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := template.New("doc.tmpl").Parse(templateBox.MustString("doc.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.StaticFS("/staic", staticBox.HTTPBox())
	r.SetHTMLTemplate(doc)
	api := api2go.NewAPIWithRouting(
		"v1",
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
	nicStorage := storage.NewNicStorage(db)
	api.AddResource(model.Chassis{}, resource.ChassisResource{BladeStorage: bladeStorage, ChassisStorage: chassisStorage})
	api.AddResource(model.Blade{}, resource.BladeResource{BladeStorage: bladeStorage, ChassisStorage: chassisStorage, NicStorage: nicStorage})
	api.AddResource(model.Nic{}, resource.NicResource{BladeStorage: bladeStorage, NicStorage: nicStorage})

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "doc.tmpl", gin.H{})
	})

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
