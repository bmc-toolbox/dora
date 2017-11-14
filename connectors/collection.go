package connectors

import (
	"fmt"
	"net"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

var (
	notifyChange chan string
)

func collect(input <-chan string, source *string, db *gorm.DB) {
	bmcUser := viper.GetString("bmc_user")
	bmcPass := viper.GetString("bmc_pass")

	for host := range input {
		c, err := NewConnection(bmcUser, bmcPass, host)
		if err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "error": err}).Error("connecting to host")
			continue
		}

		if c.HwType() == Blade && *source != "cli-with-force" {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType()}).Debug("we don't want to scan blades directly since the chassis does it for us")
			continue
		}

		data, err := c.Collect()
		if err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
			continue
		}

		switch data.(type) {
		case *model.Chassis:
			chassis := data.(*model.Chassis)
			if chassis == nil {
				continue
			}

			chassisStorage := storage.NewChassisStorage(db)
			existingData, err := chassisStorage.GetOne(chassis.Serial)
			if err != nil && err != gorm.ErrRecordNotFound {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			_, err = chassisStorage.UpdateOrCreate(chassis)
			if err != nil {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			if len(chassis.Diff(&existingData)) != 0 {
				notifyChange <- fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "chassis", chassis.Serial)
			}
			// for _, line := range chassis.Diff(&existingData) {
			// 	fmt.Println(line)
			// }
		case *model.Blade:
			blade := data.(*model.Blade)
			if blade == nil {
				continue
			}

			bladeStorage := storage.NewBladeStorage(db)
			existingData, err := bladeStorage.GetOne(blade.Serial)
			if err != nil && err != gorm.ErrRecordNotFound {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			_, err = bladeStorage.UpdateOrCreate(blade)
			if err != nil {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			if len(blade.Diff(&existingData)) != 0 {
				notifyChange <- fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "blades", blade.Serial)
			}
			// for _, line := range blade.Diff(&existingData) {
			// 	fmt.Println(line)
			// }
		case *model.Discrete:
			discrete := data.(*model.Discrete)
			if discrete == nil {
				continue
			}

			discreteStorage := storage.NewDiscreteStorage(db)
			existingData, err := discreteStorage.GetOne(discrete.Serial)
			if err != nil && err != gorm.ErrRecordNotFound {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			_, err = discreteStorage.UpdateOrCreate(discrete)
			if err != nil {
				log.WithFields(log.Fields{"operation": "store", "ip": host, "type": c.HwType(), "error": err}).Error("collecting data")
				continue
			}

			if len(discrete.Diff(&existingData)) != 0 {
				notifyChange <- fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "discretes", discrete.Serial)
			}
			// for _, line := range discrete.Diff(&existingData) {
			// 	fmt.Println(line)
			// }
		}
	}
}

// DataCollection collects the data of all given ips
func DataCollection(ips []string, source string) {
	concurrency := viper.GetInt("collector.concurrency")

	cc := make(chan string, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan string, source *string, db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			collect(input, source, db)
		}(cc, &source, db, &wg)
	}

	notifyChange = make(chan string)
	go func(notification <-chan string) {
		for callback := range notification {
			err := assetNotify(callback)
			if err != nil {
				log.WithFields(log.Fields{"operation": "ServerDB callback", "url": callback, "error": err}).Error("sending ServerDB callback")
			}
		}
	}(notifyChange)

	if ips[0] == "all" {
		hosts := []model.ScannedPort{}
		if err := db.Where("port = 443 and protocol = 'tcp' and state = 'open'").Find(&hosts).Error; err != nil {
			log.WithFields(log.Fields{"operation": "connection", "ip": "all", "error": err}).Error(fmt.Sprintf("retrieving scanned hosts"))
		} else {
			for _, host := range hosts {
				cc <- host.IP
			}
		}
	} else {
		for _, ip := range ips {
			host := model.ScannedPort{}
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				lookup, err := net.LookupHost(ip)
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": ip, "error": err}).Error(fmt.Sprintf("retrieving scanned hosts"))
					continue
				}
				ip = lookup[0]
			}

			if err := db.Where("ip = ? and port = 443 and protocol = 'tcp' and state = 'open'", ip).Find(&host).Error; err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": ip, "error": err}).Error(fmt.Sprintf("retrieving scanned hosts"))
				continue
			}
			cc <- host.IP
		}
	}

	wg.Wait()
	close(cc)
}
