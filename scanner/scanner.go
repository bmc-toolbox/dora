package scanner

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

const (
	nmapTCPPorts = "22,443"
	nmapUDPPorts = "623"
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

type ToScan struct {
	CIDR     string
	Ports    string
	Protocol string
	Site     string
}

// Verify supported methods to load subnet
func SupportedSources(source string) bool {
	switch source {
	case "kea":
		return true
	default:
		return false
	}
}

// LoadSubnets from kea.cfg
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
						log.WithFields(log.Fields{"operation": "subnet parsing", "error": err}).Warn("Scanning networks")
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

func scan(input <-chan ToScan, db *gorm.DB) {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{"operation": "scanning", "error": err}).Warn("Scanning networks")
	}

	for subnet := range input {
		scanType := ""

		switch subnet.Protocol {
		case "udp":
			scanType = "-sU"
		case "tcp":
			scanType = "-sT"
		default:
			log.WithFields(log.Fields{"operation": "scanning", "error": ErrInvalidProtocol}).Warn("Scanning networks")
			continue
		}

		log.WithFields(log.Fields{"operation": "scanning ip", "subnet": subnet.CIDR, "protocol": subnet.Protocol}).Info("Scanning networks")
		cmd := exec.Command("sudo", "nmap", "-oX", "-", scanType, subnet.CIDR, "--max-parallelism=100", "-p", subnet.Ports)
		content, err := cmd.Output()
		if err != nil {
			log.WithFields(log.Fields{"operation": "scanning", "error": err}).Error("Scanning networks")
		}

		nmap, err := nmapParse(content)
		if err != nil {
			log.WithFields(log.Fields{"operation": "scanning", "error": err}).Error("Scanning networks")
			continue
		}
		for _, host := range nmap.Hosts {
			sh := model.ScannedHost{}
			ip := ""
			for _, address := range host.Addresses {
				ip = address.Addr
				break
			}

			if err = db.FirstOrCreate(&sh, model.ScannedHost{IP: ip, CIDR: subnet.CIDR}).Error; err != nil {
				log.WithFields(log.Fields{"operation": "scanning", "error": err, "hosts": sh.IP}).Error("Scanning networks")
			}
			sh.State = host.Status.State

			for _, port := range host.Ports {
				sp := model.ScannedPort{}
				sp.Port = port.PortID
				sp.State = port.State.State
				sp.Protocol = port.Protocol
				sp.ScannedBy = hostname
				sh.Ports = append(sh.Ports, &sp)
			}

			if err = db.Save(&sh).Error; err != nil {
				log.WithFields(log.Fields{"operation": "scanning ip", "error": err, "hosts": sh.IP}).Error("Scanning networks")
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

func LoadSubnets(source string) {
	db := storage.InitDB()

	content, err := ReadKeaConfig()
	if err != nil {
		log.WithFields(log.Fields{"operation": "loading subnets", "error": err}).Error("Scanning networks")
		os.Exit(1)
	}

	if source == "kea" {
		for _, entry := range LoadSubnetsFromKea(content) {
			subnet := model.ScannedNetwork{}
			if err = db.FirstOrCreate(&subnet, model.ScannedNetwork{CIDR: entry.CIDR, Site: entry.Site}).Error; err != nil {
				log.WithFields(log.Fields{"operation": "loading subnets", "error": err, "subnet": entry.CIDR}).Error("Scanning networks")
			}
		}
	}
}

// ListSubnets all or a list of given subnets
func ListSubnets(subnetsToQuery []string) (subnets []model.ScannedNetwork) {
	db := storage.InitDB()

	if len(subnetsToQuery) == 0 {
		if err := db.Find(&subnets).Error; err != nil {
			log.WithFields(log.Fields{"operation": "listing subnets", "error": err}).Error("Scanning networks")
			os.Exit(1)
		}
	} else {
		if err := db.Where("CIDR in (?)", subnetsToQuery).Find(&subnets).Error; err != nil {
			log.WithFields(log.Fields{"operation": "listing subnets", "error": err}).Error("Scanning networks")
			os.Exit(1)
		}
	}
	return subnets
}

// ScanNetworks scan specific or all networks and try to find chassis, blades and servers
func ScanNetworks(subnetsToScan []string, site []string) {
	concurrency := viper.GetInt("scanner.concurrency")

	cc := make(chan ToScan, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan ToScan, db *gorm.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			scan(input, db)
		}(cc, db, &wg)
	}

	subnets := []model.ScannedNetwork{}
	if subnetsToScan[0] == "all" {
		if site[0] == "all" {
			if err := db.Find(&subnets).Error; err != nil {
				log.WithFields(log.Fields{"operation": "scanning ip", "error": err}).Error("Scanning networks")
				os.Exit(1)
			}
		} else {
			if err := db.Where("site in (?)", site).Find(&subnets).Error; err != nil {
				log.WithFields(log.Fields{"operation": "scanning ip", "error": err}).Error("Scanning networks")
				os.Exit(1)
			}
		}
	} else {
		if err := db.Where("cidr in (?)", subnetsToScan).Find(&subnets).Error; err != nil {
			log.WithFields(log.Fields{"operation": "scanning ip", "error": err}).Error("Scanning networks")
			os.Exit(1)
		}
	}

	for _, subnet := range subnets {
		t := ToScan{
			CIDR:     subnet.CIDR,
			Site:     subnet.Site,
			Ports:    nmapTCPPorts,
			Protocol: "tcp",
		}
		cc <- t
	}

	for _, subnet := range subnets {
		t := ToScan{
			CIDR:     subnet.CIDR,
			Site:     subnet.Site,
			Ports:    nmapUDPPorts,
			Protocol: "udp",
		}
		cc <- t
	}

	close(cc)
	wg.Wait()
}
