package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
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

/*
	keaConfig := viper.GetString("kea_config")

	content, err := ioutil.ReadFile(keaConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
*/

func loadSubnets(content []byte, datacenters []string) (subnets []*net.IPNet) {
	keaData := &Kea{}
	err := json.Unmarshal(content, &keaData)
	if err != nil {
		panic(err)
	}

	for _, subnet := range keaData.Dhcp4.Subnet4 {
		oob := false
		for _, option := range subnet.OptionData {
			if option.Name == "domain-name" && strings.HasSuffix(option.Data, ".lom.booking.com") {
				for _, datacenter := range datacenters {
					if strings.HasSuffix(option.Data, fmt.Sprintf("%s.lom.booking.com", datacenter)) {
						oob = true
					}
				}
			}
		}

		if oob {
			_, ipv4Net, err := net.ParseCIDR(subnet.Subnet)
			if err != nil {
				log.WithFields(log.Fields{"operation": "subnet parsing", "error": err}).Warn("Nertwork scanning")
			}
			subnets = append(subnets, ipv4Net)
		}
	}

	return subnets
}
