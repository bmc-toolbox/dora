package scanner

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/model"
	"gitlab.booking.com/go/dora/storage"
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

// OptionData contains the options send to the clients during the dhcp resquest
type OptionData struct {
	Data string `json:"data"`
	Name string `json:"name"`
}

// ToScan payload message to scan a network
type ToScan struct {
	CIDR string
	Site string
}

type scanOption struct {
	Protocol string
	Port     int
}

var scanProfiles = []scanOption{
	scanOption{
		Protocol: "tcp",
		Port:     22,
	},
	scanOption{
		Protocol: "tcp",
		Port:     443,
	},
	scanOption{
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

		ips, err := ipsWithinASubnet(subnet.CIDR)
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet expansion", "subnet": subnet.CIDR}).Error(err)
			continue
		}

		for _, s := range scanProfiles {
			for _, ip := range ips {
				probeStatus, err := Probe(s.Protocol, ip, s.Port)
				if err != nil {
					log.WithFields(log.Fields{"operation": "scanning host", "subnet": subnet.CIDR, "host": ip, "port": s.Port}).Error(err)
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
				}
			}
		}
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
