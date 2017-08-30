package scanner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

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

// OptionData contains the options send to the clients during the dhcp resquest
type OptionData struct {
	Data string `json:"data"`
	Name string `json:"name"`
}

type ScannableSubnet struct {
	Subnet4 *net.IPNet
	Gateway *string
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

// ScanNetworks scan all of our networks and try to find chassis, blades and servers
func ScanNetworks() {
	keaConfig := viper.GetString("kea_config")
	site := strings.Split(viper.GetString("site"), " ")

	content, err := ioutil.ReadFile(keaConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, subnet := range loadSubnets(content, site) {
		err := scan(subnet.String(), "22,443", "tcp")
		if err != nil {
			log.WithFields(log.Fields{"operation": "subnet scan", "error": err}).Warn("Scanning network")
		}
	}
}
