package connectors

import (
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	metrics "github.com/bmc-toolbox/gin-go-metrics"

	"github.com/bmc-toolbox/dora/internal/notification"
	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
)

func collect(input <-chan string, source *string, db *gorm.DB) {
	bmcUser := viper.GetString("bmc_user")
	bmcPass := ""
	if viper.IsSet("bmc_pass") {
		bmcPass = viper.GetString("bmc_pass")
	} else {
		bmcPassFile := viper.GetString("bmc_pass_file")
		bmcPassBytes, err := ioutil.ReadFile(bmcPassFile)
		if err != nil {
			log.WithFields(log.Fields{"operation": "scan"}).Error(err)
		}
		bmcPass = string(bmcPassBytes)
	}

	for host := range input {
		log.WithFields(log.Fields{"operation": "scan", "ip": host}).Debug("collection started")

		graphiteKey := "collect.collected_successfully"
		hintOpts := hintOptsInit(host, db)
		updateSacMetricFn := scanAndConnectMetricInit()

		conn, err := discover.ScanAndConnect(host, bmcUser, bmcPass, hintOpts...)
		if err != nil {
			log.WithFields(log.Fields{"operation": "scan", "ip": host}).Error(err)
			graphiteKey = "collect.bmc_scan_failed"
			if viper.GetBool("metrics.enabled") {
				metrics.IncrCounter([]string{graphiteKey}, 1)
			}
			continue
		}

		updateSacMetricFn()

		if bmc, ok := conn.(devices.Bmc); ok {
			err = bmc.CheckCredentials()
			if err == errors.ErrLoginFailed {
				bmc.UpdateCredentials(
					viper.GetString(fmt.Sprintf("collector.default.%s.username", bmc.Vendor())),
					viper.GetString(fmt.Sprintf("collector.default.%s.password", bmc.Vendor())),
				)
				err = bmc.CheckCredentials()
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": host}).Error(err)
					graphiteKey = "collect.bmc_wrong_credentials"
					if viper.GetBool("metrics.enabled") {
						metrics.IncrCounter([]string{graphiteKey}, 1)
					}
					continue
				}
			} else if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": host}).Error(err)
				graphiteKey = "collect.bmc_connection_failed"
				if viper.GetBool("metrics.enabled") {
					metrics.IncrCounter([]string{graphiteKey}, 1)
				}
				continue
			}

			isBlade, err := bmc.IsBlade()
			if err != nil {
				log.WithFields(log.Fields{"operation": "collection", "ip": host}).Error(err)
				graphiteKey = "collect.bmc_is_blade_detection_failed"
				if viper.GetBool("metrics.enabled") {
					metrics.IncrCounter([]string{graphiteKey}, 1)
				}
				continue
			}

			if isBlade && *source != "cli-with-force" {
				chassisSerial, err := bmc.ChassisSerial()
				if err != nil {
					log.WithFields(log.Fields{"operation": "collection", "ip": host}).Error(err)
					continue
				}

				chassis := model.Chassis{}
				db.Where("serial = ?", chassisSerial).First(&chassis)

				if chassis.Managed {
					log.WithFields(log.Fields{"operation": "detection", "ip": host}).Debug("we don't want to scan blades directly since the chassis does it for us")
					continue
				}
			}

			err = collectBmc(bmc)
			if err != nil {
				log.WithFields(log.Fields{"operation": "collection", "ip": host}).Error(err)
				graphiteKey = "collect.bmc_collection_failed"
			}

			log.WithFields(log.Fields{"operation": "collection", "ip": host}).Info("success")
		} else if bmc, ok := conn.(devices.Cmc); ok {
			err = bmc.CheckCredentials()
			if err == errors.ErrLoginFailed {
				bmc.UpdateCredentials(
					viper.GetString(fmt.Sprintf("collector.default.%s.username", bmc.Vendor())),
					viper.GetString(fmt.Sprintf("collector.default.%s.password", bmc.Vendor())),
				)
				err = bmc.CheckCredentials()
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": host}).Error(err)
					graphiteKey = "collect.cmc_wrong_credentials"
					if viper.GetBool("metrics.enabled") {
						metrics.IncrCounter([]string{graphiteKey}, 1)
					}
					continue
				}
			} else if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": host}).Error(err)
				graphiteKey = "collect.cmc_connection_failed"
				if viper.GetBool("metrics.enabled") {
					metrics.IncrCounter([]string{graphiteKey}, 1)
				}
				continue
			}

			err := collectCmc(bmc, bmcUser, bmcPass)
			if err != nil {
				log.WithFields(log.Fields{"operation": "collection", "ip": host}).Error(err)
				graphiteKey = "collect.cmc_collection_failed"
			}

			log.WithFields(log.Fields{"operation": "collection", "ip": host}).Info("success")
		} else {
			log.WithFields(log.Fields{"operation": "collection", "ip": host}).Debug("unknown hardware skipping")
			graphiteKey = "collect.unknown_device"
		}
		// send metric which is not protected by "continue"
		if viper.GetBool("metrics.enabled") {
			metrics.IncrCounter([]string{graphiteKey}, 1)
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

	if ips[0] == "all" {
		var hosts []model.ScannedPort
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

// DataCollectionWorker collects the data of all given ips
func DataCollectionWorker() {
	nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
	if err != nil {
		log.Fatalf("Subscriber unable to connect: %v\n", err)
	}

	concurrency := viper.GetInt("collector.concurrency")
	cc := make(chan string, concurrency)
	db := storage.InitDB()
	source := "worker"

	for i := 0; i < concurrency; i++ {
		go func(input <-chan string, source *string, db *gorm.DB) {
			collect(input, source, db)
		}(cc, &source, db)
	}

	_, err = nc.QueueSubscribe("dora::collect", viper.GetString("collector.worker.queue"), func(msg *nats.Msg) {
		ip := string(msg.Data)
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			lookup, err := net.LookupHost(ip)
			if err != nil {
				log.WithFields(log.Fields{"operation": "retrieving scanned hosts", "ip": ip}).Error(err)
				return
			}
			ip = lookup[0]
		}
		cc <- ip
	})
	if err != nil {
		log.WithFields(log.Fields{"operation": "error subscribing to the queue"}).Fatal(err)
	}
	if err := nc.LastError(); err != nil {
		log.WithFields(log.Fields{"operation": "registering worker"}).Fatal(err)
	}

	log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": "dora::collect"}).Info("subscribed to queue")
}

func collectBmc(bmc devices.Bmc) (err error) {
	defer bmc.Close()

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
			return fmt.Errorf("unable to read devices.Blade")
		}

		blade := model.NewBladeFromDevice(b)
		blade.BmcAuth = true
		blade.BmcWEBReachable = true

		db.Where(model.Chassis{Serial: blade.ChassisSerial}).FirstOrCreate(&model.Chassis{})

		var scans []model.ScannedPort
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
			url := fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "blades", blade.Serial)
			notification.NotifyChange(url)
		}

		err = bladeStorage.RemoveOldRefs(blade)
		if err != nil {
			return err
		}
	} else {
		server, err := bmc.ServerSnapshot()
		if err != nil {
			return err
		}

		b, ok := server.(*devices.Discrete)
		if !ok {
			return fmt.Errorf("unable to read devices.Discrete")
		}

		discrete := model.NewDiscreteFromDevice(b)
		discrete.BmcAuth = true
		discrete.BmcWEBReachable = true

		var scans []model.ScannedPort
		db.Where("ip = ?", discrete.BmcAddress).Find(&scans)
		for _, scan := range scans {
			if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
				discrete.BmcSSHReachable = true
			} else if scan.Port == 443 && scan.Protocol == "tcp" && scan.State == "open" {
				discrete.BmcWEBReachable = true
			} else if scan.Port == 623 && scan.Protocol == "ipmi" && scan.State == "open" {
				discrete.BmcIpmiReachable = true
			}
		}

		discreteStorage := storage.NewDiscreteStorage(db)
		existingData, err := discreteStorage.GetOne(discrete.Serial)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		_, err = discreteStorage.UpdateOrCreate(discrete)
		if err != nil {
			return err
		}

		if len(discrete.Diff(&existingData)) != 0 {
			url := fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "discretes", discrete.Serial)
			notification.NotifyChange(url)
		}

		err = discreteStorage.RemoveOldRefs(discrete)
		if err != nil {
			return err
		}
	}

	return nil
}

func collectCmc(bmc devices.Cmc, bmcUser string, bmcPass string) (err error) {
	defer bmc.Close()

	if !bmc.IsActive() {
		return err
	}

	db := storage.InitDB()

	ch, err := bmc.ChassisSnapshot()
	if err != nil {
		return err
	}

	chassis := model.NewChassisFromDevice(ch)
	chassis.BmcAuth = true
	chassis.Managed = true
	var scans []model.ScannedPort
	db.Where("ip = ?", chassis.BmcAddress).Find(&scans)
	for _, scan := range scans {
		if scan.Port == 443 && scan.Protocol == "tcp" && scan.State == "open" {
			chassis.BmcWEBReachable = true
		} else if scan.Port == 22 && scan.Protocol == "tcp" && scan.State == "open" {
			chassis.BmcSSHReachable = true
		}
	}

	for _, blade := range chassis.Blades {
		hintOpts := hintOptsInit(blade.BmcAddress, db)
		updateSacMetricFn := scanAndConnectMetricInit()

		if conn, err := discover.ScanAndConnect(
			blade.BmcAddress,
			bmcUser,
			bmcPass,
			hintOpts...,
		); err == nil {

			updateSacMetricFn()
			if b, ok := conn.(devices.Bmc); ok {
				err = b.CheckCredentials()
				if err == errors.ErrLoginFailed {
					b.UpdateCredentials(
						viper.GetString(fmt.Sprintf("collector.default.%s.username", b.Vendor())),
						viper.GetString(fmt.Sprintf("collector.default.%s.password", b.Vendor())),
					)
					err = b.CheckCredentials()
					if err != nil {
						log.WithFields(log.Fields{"operation": "connection", "ip": blade.BmcAddress}).Error(err)
						continue
					}
				} else if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": blade.BmcAddress}).Error(err)
					continue
				}

				blade.BmcAuth = true
				blade.BmcWEBReachable = true

				var scans []model.ScannedPort
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

				blade.BmcType = b.HardwareType()

				blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = b.CPU()
				if err != nil {
					log.WithFields(log.Fields{"operation": "reading cpu data", "ip": blade.BmcAddress, "name": blade.Name, "serial": blade.Serial, "type": "chassis"}).Warning(err)
				}

				blade.Memory, err = b.Memory()
				if err != nil {
					log.WithFields(log.Fields{"operation": "reading memory data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
				}

				blade.BmcLicenceType, blade.BmcLicenceStatus, err = b.License()
				if err != nil {
					log.WithFields(log.Fields{"operation": "reading license data", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
				}

				if len(blade.Nics) == 0 {
					nics, err := b.Nics()
					if err != nil {
						log.WithFields(log.Fields{"operation": "reading nice", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
					} else {
						for _, nic := range nics {
							blade.Nics = append(blade.Nics, &model.Nic{
								MacAddress:  nic.MacAddress,
								Name:        nic.Name,
								BladeSerial: blade.Serial,
							})
						}
					}
				}

				if len(blade.Disks) == 0 {
					disks, err := b.Disks()
					if err != nil {
						log.WithFields(log.Fields{"operation": "reading disks", "ip": blade.BmcAddress, "serial": blade.Serial, "type": "chassis"}).Warning(err)
					} else {
						for pos, disk := range disks {
							if disk.Serial == "" {
								disk.Serial = fmt.Sprintf("%s-failed-%d", blade.Serial, pos)
							}

							blade.Disks = append(blade.Disks, &model.Disk{
								Serial:      disk.Serial,
								Size:        disk.Size,
								Status:      disk.Status,
								Model:       disk.Model,
								Location:    disk.Location,
								Type:        disk.Type,
								FwVersion:   disk.FwVersion,
								BladeSerial: blade.Serial,
							})
						}
					}
				}

				if err == nil {
					log.WithFields(log.Fields{"operation": "collection", "ip": blade.BmcAddress}).Info("success")
				}
			}
		}
	}

	chassisStorage := storage.NewChassisStorage(db)
	existingData, err := chassisStorage.GetOne(chassis.Serial)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	_, err = chassisStorage.UpdateOrCreate(chassis)
	if err != nil {
		return err
	}

	if len(chassis.Diff(&existingData)) != 0 {
		url := fmt.Sprintf("%s/%s/%s", viper.GetString("url"), "chassis", chassis.Serial)
		notification.NotifyChange(url)
	}

	var merror *multierror.Error

	bladeStorage := storage.NewBladeStorage(db)
	for _, blade := range chassis.Blades {
		merror = multierror.Append(merror, bladeStorage.RemoveOldRefs(blade))
	}

	err = chassisStorage.RemoveOldRefs(chassis)
	if err != nil {
		merror = multierror.Append(merror, err)
	}

	if err != nil {
		return merror.ErrorOrNil()
	}

	return nil
}

func hintOptsInit(host string, db *gorm.DB) []discover.Option {
	if !viper.GetBool("collector.use_discover_hints") {
		return nil
	}

	hint := hintFromDB(host, db)

	hintCallBack := func(newHint string) error {
		if newHint == hint {
			return nil
		}

		hintToDB(host, newHint, db)

		return nil
	}

	return []discover.Option{
		discover.WithProbeHint(hint),
		discover.WithHintCallBack(hintCallBack),
	}
}

func hintFromDB(host string, db *gorm.DB) string {
	var h model.DiscoverHint

	if err := db.First(&h, "ip = ?", host).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithFields(
				log.Fields{
					"operation": "collection",
					"ip":        host,
				},
			).Warnf("error loading discover hint to DB: %v", err)
		}

		return ""
	}

	return h.Hint
}

func hintToDB(host, hint string, db *gorm.DB) {
	h := model.DiscoverHint{
		IP:   host,
		Hint: hint,
	}

	err := db.Save(&h).Error
	if err != nil {
		log.WithFields(log.Fields{"operation": "collection", "ip": host}).Warnf("error writing discover hint to DB: %v", err)
	}
}

func scanAndConnectMetricInit() func() {
	if !viper.GetBool("metrics.enabled") {
		return func() {}
	}

	var (
		keyCommon = []string{"collect", "scan_and_connect", "successful"}
		keyCount  = append(keyCommon, "count")
		keyTime   = append(keyCommon, "time_milliseconds")
	)

	start := time.Now()

	return func() {
		elapsed := time.Since(start)
		metrics.IncrCounter(keyTime, elapsed.Milliseconds())
		metrics.IncrCounter(keyCount, 1)
	}
}
