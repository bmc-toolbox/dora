package scanner

import (
	"encoding/json"
	"fmt"
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

var ()

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

type toScan struct {
	Subnet   string
	Ports    string
	Protocol string
}

func loadSubnets(content []byte, site []string) (subnets []*net.IPNet) {
	keaData := &Kea{}
	err := json.Unmarshal(content, &keaData)
	if err != nil {
		panic(err)
	}

	for _, subnet := range keaData.Dhcp4.Subnet4 {
		oob := false
		for _, option := range subnet.OptionData {
			if option.Name == "domain-name" && strings.HasSuffix(option.Data, ".lom.booking.com") {
				for _, s := range site {
					if strings.HasSuffix(option.Data, fmt.Sprintf("%s.lom.booking.com", s)) {
						oob = true
					} else if s == "all" {
						oob = true
						break
					}
				}
			}
		}

		if oob {
			_, ipv4Net, err := net.ParseCIDR(subnet.Subnet)
			if err != nil {
				log.WithFields(log.Fields{"operation": "subnet parsing", "error": err}).Warn("Scanning networks")
			}
			subnets = append(subnets, ipv4Net)
		}
	}

	return subnets
}

func scan(input <-chan toScan, db *gorm.DB) {
	xmlDir := viper.GetString("nmap_xml_dir")
	err := os.MkdirAll(xmlDir, 0755)
	if err != nil {
		panic(err)
	}

	for subnet := range input {
		scanType := ""

		switch subnet.Protocol {
		case "udp":
			scanType = "-sU"
		case "tcp":
			scanType = "-sT"
		default:
			log.WithFields(log.Fields{"operation": "subnet parsing", "error": ErrInvalidProtocol}).Warn("Scanning networks")
			continue
		}

		cmd := exec.Command("nmap", "-oX", "-", scanType, subnet.Subnet, "--max-parallelism=100", "-p", subnet.Ports, "--unprivileged")
		content, err := cmd.Output()
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet parsing", "error": err}).Error("Scanning networks")
		}

		nmap, err := nmapParse(content)
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet parsing", "error": err}).Error("Scanning networks")
			continue
		}
		for _, host := range nmap.Hosts {
			sh := model.ScannedHost{}
			for _, address := range host.Addresses {
				sh.IP = address.Addr
				break
			}
			sh.State = host.Status.State

			for _, port := range host.Ports {
				sp := model.ScannedPort{}
				sp.Port = port.PortID
				sp.State = port.State.State
				sp.Protocol = port.Protocol
				sh.Ports = append(sh.Ports, sp)
			}
			if err = db.Save(&sh).Error; err != nil {
				log.WithFields(log.Fields{"operation": "scanning ip", "error": err, "hosts": sh.IP}).Error("Scanning networks")
			}
		}
	}
}

// ScanNetworks scan all of our networks and try to find chassis, blades and servers
func ScanNetworks() {
	keaConfig := viper.GetString("kea_config")
	site := strings.Split(viper.GetString("site"), " ")
	tcpPorts := viper.GetString("nmap_tcp_ports")
	concurrency := viper.GetInt("concurrency")

	content, err := ioutil.ReadFile(keaConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cc := make(chan toScan, concurrency)
	wg := sync.WaitGroup{}
	db := storage.InitDB()

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan toScan, db *gorm.DB, wg *sync.WaitGroup) {
			scan(input, db)
			wg.Done()
		}(cc, db, &wg)
	}

	for _, subnet := range loadSubnets(content, site) {
		t := toScan{
			Subnet:   subnet.String(),
			Ports:    tcpPorts,
			Protocol: "tcp",
		}
		cc <- t
	}

	close(cc)
	wg.Wait()
}
