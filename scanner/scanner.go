package scanner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"

	metrics "github.com/bmc-toolbox/gin-go-metrics"

	"github.com/bmc-toolbox/dora/model"
	"github.com/bmc-toolbox/dora/storage"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Kea is the main entry for parsing the kea config file
type Kea struct {
	Dhcp4 *Dhcp4 `json:"Dhcp4"`
}

// Dhcp4 contains the dhcp information for ipv4 networks
type Dhcp4 struct {
	Subnet4 []*Subnet4 `json:"subnet4"`
}

// Subnet4 contains all the subnets managed by Kea
type Subnet4 struct {
	OptionData []*OptionData `json:"option-data"`
	Subnet     string        `json:"subnet"`
}

// OptionData contains the options send to the clients during the dhcp request
type OptionData struct {
	Data string `json:"data"`
	Name string `json:"name"`
}

// ToScan payload message to scan a network
type ToScan struct {
	CIDR string `json:"cidr"`
	Site string `json:"site"`
}

type scanOption struct {
	Protocol string
	Port     int
}

var scanProfiles = []scanOption{
	{
		Protocol: "tcp",
		Port:     22,
	},
	{
		Protocol: "tcp",
		Port:     443,
	},
	{
		Protocol: "ipmi",
		Port:     623,
	},
}

// LoadSubnetsFromKea from kea.cfg
func LoadSubnetsFromKea(content []byte) (subnets []*ToScan) {
	keaData := &Kea{}
	err := json.Unmarshal(content, &keaData)
	if err != nil {
		panic(err)
	}

	keaDomainNameSuffix := viper.GetString("scanner.kea_domain_name_suffix")
	for _, subnet := range keaData.Dhcp4.Subnet4 {
		for _, option := range subnet.OptionData {
			if option.Name == "domain-name" && strings.HasSuffix(option.Data, keaDomainNameSuffix) {
				if strings.HasSuffix(option.Data, keaDomainNameSuffix) {
					_, ipv4Net, err := net.ParseCIDR(subnet.Subnet)
					if err != nil {
						log.WithFields(log.Fields{"operation": "subnet parsing"}).Warn(err)
						continue
					}
					toScan := &ToScan{
						CIDR: ipv4Net.String(),
						Site: strings.TrimSuffix(option.Data, keaDomainNameSuffix),
					}
					subnets = append(subnets, toScan)
				}
			}
		}
	}

	return subnets
}

func nexIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func ipsWithinASubnet(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); nexIP(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func scan(input <-chan *ToScan, db *gorm.DB) {
	ScannedBy := viper.GetString("scanner.scanned_by")
	for subnet := range input {

		log.WithFields(log.Fields{"operation": "subnet expansion", "subnet": subnet.CIDR}).Info("network scan started")

		ips, err := ipsWithinASubnet(subnet.CIDR)
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet expansion", "subnet": subnet.CIDR}).Error(err)
			continue
		}

		for _, s := range scanProfiles {
			for _, ip := range ips {
				graphiteKey := fmt.Sprintf("scan.%v_%v.scanned_successfully", s.Protocol, s.Port)
				probeStatus, err := Probe(s.Protocol, ip, s.Port)
				if err != nil {
					log.WithFields(log.Fields{"operation": "scanning host", "subnet": subnet.CIDR, "host": ip, "port": s.Port}).Error(err)
					// failed scan for particular service is not a problem, we don't want separate metric on that
				}

				sp := model.ScannedPort{
					IP:        ip,
					CIDR:      subnet.CIDR,
					Port:      s.Port,
					State:     probeStatus.String(),
					Site:      subnet.Site,
					Protocol:  s.Protocol,
					ScannedBy: ScannedBy,
				}
				sp.ID = sp.GenID()

				if err = db.Save(&sp).Error; err != nil {
					log.WithFields(log.Fields{"operation": "storing scan", "subnet": subnet.CIDR, "host": ip, "port": s.Port}).Error(err)
					graphiteKey = "scan.db_save_failed"
				}
				if viper.GetBool("metrics.enabled") {
					metrics.IncrCounter([]string{graphiteKey}, 1)
				}
			}
		}

		log.WithFields(log.Fields{"operation": "subnet expansion", "subnet": subnet.CIDR}).Info("network scan finished")
	}
}

// ReadKeaConfig reads the config configuration file and returns its data
func ReadKeaConfig() (content []byte, err error) {
	keaConfig := viper.GetString("scanner.kea_config")
	content, err = ioutil.ReadFile(keaConfig)
	if err != nil {
		return content, err
	}

	return content, err
}

func LoadSubnets(source string, subnetsToScan []string, site []string) (subnets []*ToScan) {
	content, err := ReadKeaConfig()
	if err != nil {
		log.WithFields(log.Fields{"operation": "loading subnets"}).Error(err)
		os.Exit(1)
	}

	if source == "kea" {
		subnets = LoadSubnetsFromKea(content)
	}

	if subnetsToScan[0] == "all" && site[0] == "all" {
		return subnets
	} else if subnetsToScan[0] == "all" && site[0] != "all" {
		filteredSubnets := make([]*ToScan, 0)
		for _, subnet := range subnets {
			for _, s := range site {
				if subnet.Site == s {
					filteredSubnets = append(filteredSubnets, subnet)
				}
			}
		}
		subnets = filteredSubnets
	} else {
		filteredSubnets := make([]*ToScan, 0)
		for _, subnet := range subnets {
			for _, s := range subnetsToScan {
				if s == subnet.CIDR {
					filteredSubnets = append(filteredSubnets, subnet)
				}
			}
		}
		subnets = filteredSubnets
	}

	return subnets
}

// ListSubnets all or a list of given subnets
func ListSubnets(subnetsToQuery []string, site []string) (subnets []*ToScan) {
	return LoadSubnets(viper.GetString("scanner.subnet_source"), subnetsToQuery, site)
}

// ScanNetworks scan specific or all networks and try to find chassis, blades and servers
func ScanNetworks(subnetsToScan []string, site []string) {
	concurrency := viper.GetInt("scanner.concurrency")

	cc := make(chan *ToScan, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan *ToScan, db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			scan(input, db)
		}(cc, db, &wg)
	}

	subnets := LoadSubnets(viper.GetString("scanner.subnet_source"), subnetsToScan, site)

	for idx := range subnets {
		cc <- subnets[idx]
	}

	close(cc)
	wg.Wait()
}

// ScanNetworksWorker scan specific or all networks and try to find chassis, blades and servers
func ScanNetworksWorker() {
	nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
	if err != nil {
		log.Fatalf("Subscriber unable to connect: %v\n", err)
	}

	concurrency := viper.GetInt("scanner.concurrency")
	cc := make(chan *ToScan, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan *ToScan, db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			scan(input, db)
		}(cc, db, &wg)
	}

	nc.QueueSubscribe("dora::scan", viper.GetString("collector.worker.queue"), func(msg *nats.Msg) {
		t := &ToScan{}
		err := json.Unmarshal(msg.Data, t)
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet scan"}).Error(err)
			return
		}
		cc <- t
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.WithFields(log.Fields{"operation": "registering worker"}).Fatal(err)
	}

	log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": "dora::scan"}).Info("Subscribed to queue")
	//	close(cc)
	//	wg.Wait()
}
