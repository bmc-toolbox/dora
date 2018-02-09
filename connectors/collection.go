package connectors

import (
	"fmt"
	"net"
	"sync"

	"gitlab.booking.com/go/bmc/errors"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/bmc/devices"
	"gitlab.booking.com/go/bmc/discover"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/storage"
)

var (
	notifyChange chan string
)

func collect(input <-chan string, source *string, db *gorm.DB) {
	bmcUser := viper.GetString("bmc_user")
	bmcPass := viper.GetString("bmc_pass")

	for host := range input {
		log.WithFields(log.Fields{"operation": "scan", "ip": host}).Debug("collection started")

		conn, err := discover.ScanAndConnect(host, bmcUser, bmcPass)
		if err != nil {
			log.WithFields(log.Fields{"operation": "scan", "ip": host}).Error(err)
			continue
		}

		if bmc, ok := conn.(devices.Bmc); ok {
			err = bmc.Login()
			if err == errors.ErrLoginFailed {
				//username := viper.GetString(fmt.Sprintf("collector.default.%s.username"))
				//password := viper.GetString(fmt.Sprintf("collector.default.%s.password"))
			} else if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": host}).Error(err)
				continue
			}
			err := collectBMC(bmc)
			if err != nil {
				log.WithFields(log.Fields{"operation": "collection", "ip": host}).Error(err)
			}
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
				log.WithFields(log.Fields{"operation": "ServerDB callback", "url": callback}).Error(err)
			}
		}
	}(notifyChange)

	if ips[0] == "all" {
		hosts := []model.ScannedPort{}
		if err := db.Where("port = 443 and protocol = 'tcp' and state = 'open'").Find(&hosts).Error; err != nil {
			log.WithFields(log.Fields{"operation": "retrieving scanned hosts", "ip": "all"}).Error(err)
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
					log.WithFields(log.Fields{"operation": "retrieving scanned hosts", "ip": ip}).Error(err)
					continue
				}
				ip = lookup[0]
			}

			if err := db.Where("ip = ? and port = 443 and protocol = 'tcp' and state = 'open'", ip).Find(&host).Error; err != nil {
				log.WithFields(log.Fields{"operation": "retrieving scanned hosts", "ip": ip}).Error(err)
				continue
			}
			cc <- host.IP
		}
	}

	close(cc)
	wg.Wait()
}

func collectBMC(bmc devices.Bmc) (err error) {
	defer bmc.Logout()

	serial, err := bmc.Serial()
	if err != nil {
		return err
	}

	if serial == "" || serial == "[unknown]" || serial == "0000000000" || serial == "_" {
		return ErrInvalidSerial
	}

	isBlade, err := bmc.IsBlade()
	if err != nil {
		return err
	}

	db := storage.InitDB()

	if isBlade {
		server, err := bmc.ServerSnapshot()
		if err != nil {
			return err
		}

		b, ok := server.(*devices.Blade)
		if !ok {
			return fmt.Errorf("Unable to read devices.Blade")
		}

		blade := model.NewBladeFromDevice(b)
		blade.BmcAuth = true
		blade.BmcWEBReachable = true

		scans := []model.ScannedPort{}
		db.Where("ip = ?", blade.BmcAddress).Find(&scans)
		for _, scan := range scans {
			if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
				blade.BmcSSHReachable = true
			} else if scan.Port == 443 && scan.Protocol == "tcp" && scan.State == "open" {
				blade.BmcWEBReachable = true
			} else if scan.Port == 623 && scan.Protocol == "ipmi" && scan.State == "open" {
				blade.BmcIpmiReachable = true
			}
		}

		bladeStorage := storage.NewBladeStorage(db)
		existingData, err := bladeStorage.GetOne(blade.Serial)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		_, err = bladeStorage.UpdateOrCreate(blade)
		if err != nil {
			return err
		}

		if len(blade.Diff(&existingData)) != 0 {
			notifyChange <- fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "blades", blade.Serial)
		}
	}

	return nil
}
