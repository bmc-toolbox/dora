package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	"github.com/GeertJohan/go.rice"
	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go-adapter/gingonic"
	nats "github.com/nats-io/go-nats"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/resource"
	"gitlab.booking.com/go/dora/scanner"
	"gitlab.booking.com/go/dora/storage"

	"github.com/gin-gonic/gin"
)

type scanRequest struct {
	Networks []string `json:"networks"`
}

type collectionRequest struct {
	Ips []string `json:"ips"`
}

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
	r.SetHTMLTemplate(doc)
	r.StaticFS("/api_static", staticBox.HTTPBox())
	api := api2go.NewAPIWithRouting(
		"v1",
		api2go.NewStaticResolver("/"),
		gingonic.New(r),
	)

	db := storage.InitDB()
	defer db.Close()

	chassisStorage := storage.NewChassisStorage(db)
	bladeStorage := storage.NewBladeStorage(db)
	discreteStorage := storage.NewDiscreteStorage(db)
	nicStorage := storage.NewNicStorage(db)
	storageBladeStorage := storage.NewStorageBladeStorage(db)
	scannedPortStorage := storage.NewScannedPortStorage(db)
	psuStorage := storage.NewPsuStorage(db)
	diskStorage := storage.NewDiskStorage(db)

	api.AddResource(model.Chassis{}, resource.ChassisResource{ChassisStorage: chassisStorage})
	api.AddResource(model.Blade{}, resource.BladeResource{BladeStorage: bladeStorage})
	api.AddResource(model.Discrete{}, resource.DiscreteResource{DiscreteStorage: discreteStorage})
	api.AddResource(model.StorageBlade{}, resource.StorageBladeResource{StorageBladeStorage: storageBladeStorage})
	api.AddResource(model.Nic{}, resource.NicResource{NicStorage: nicStorage})
	api.AddResource(model.ScannedPort{}, resource.ScannedPortResource{ScannedPortStorage: scannedPortStorage})
	api.AddResource(model.Psu{}, resource.PsuResource{PsuStorage: psuStorage})
	api.AddResource(model.Disk{}, resource.DiskResource{DiskStorage: diskStorage})

	r.POST("/api/v1/collect", func(c *gin.Context) {
		subject := "dora::collect"
		jsonPayload := &collectionRequest{}
		var response []gin.H
		if err := c.ShouldBindWith(&jsonPayload, binding.JSON); err == nil {
			nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
			if err != nil {
				c.JSON(http.StatusPreconditionFailed, gin.H{"message": fmt.Sprintf("publisher unable to connect: %v", err)})
				return
			}
			for _, ip := range jsonPayload.Ips {
				if net.ParseIP(ip) == nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid ip: %s", ip)})
					return
				}
				nc.Publish(subject, []byte(ip))
				nc.Flush()
				if err := nc.LastError(); err != nil {
					log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject, "payload": ip}).Error(err)
					response = append(response, gin.H{"ip": ip, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}
				response = append(response, gin.H{"ip": ip, "message": "ok"})
				log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject, "payload": ip}).Info("sent")
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
		return
	})

	r.POST("/api/v1/scan", func(c *gin.Context) {
		subject := "dora::scan"
		jsonPayload := &scanRequest{}
		var response []gin.H
		if err := c.ShouldBindWith(&jsonPayload, binding.JSON); err == nil {
			nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
			if err != nil {
				c.JSON(http.StatusPreconditionFailed, gin.H{"message": fmt.Sprintf("publisher unable to connect: %v", err)})
				return
			}
			for _, network := range jsonPayload.Networks {
				_, _, err := net.ParseCIDR(network)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid network: %s", network)})
					return
				}

				subnets := scanner.LoadSubnets(viper.GetString("scanner.subnet_source"), []string{network}, viper.GetStringSlice("site"))
				subnet := subnets[0]
				s, err := json.Marshal(subnet)
				if err != nil {
					log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject, "operation": "encoding subnet"}).Error(err)
					c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
					return
				}

				nc.Publish(subject, s)
				nc.Flush()
				if err := nc.LastError(); err != nil {
					log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject, "payload": s}).Error(err)
					response = append(response, gin.H{"network": network, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}
				response = append(response, gin.H{"network": network, "message": "ok"})
				log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject, "payload": s}).Info("sent")
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
		return
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "doc.tmpl", gin.H{})
	})

	r.GET("/doc", func(c *gin.Context) {
		c.HTML(200, "doc.tmpl", gin.H{})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/ping_db", func(c *gin.Context) {
		err := db.DB().Ping()
		if err == nil {
			c.String(200, "pong")
		} else {
			c.String(451, "database has gone away")
		}
	})

	r.Run(fmt.Sprintf(":%d", port))
}
