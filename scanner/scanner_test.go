package scanner

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	tt := []struct {
		content  []byte
		networks []string
	}{
		{
			[]byte(`{"Dhcp4": { "subnet4": [{"option-data": [{"data": "adc1.bmc.example.com","name": "domain-name"}], "subnet": "192.168.64.0/24"},
											{"option-data": [{"data": "edc4.ext.example.com","name": "domain-name"}], "subnet": "192.168.68.0/24"},
											{"option-data": [{"data": "edc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.17.0/24"},
											{"option-data": [{"data": "udc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.15.0/24"}]}}`),
			[]string{"192.168.17.0/24", "192.168.15.0/24"},
		},
		{
			[]byte(`{"Dhcp4": { "subnet4": [{"option-data": [{"data": "adc1.bmc.example.com","name": "domain-name"}], "subnet": "192.168.64.0/24"},
											{"option-data": [{"data": "edc4.ext.example.com","name": "domain-name"}], "subnet": "192.168.68.0/24"},
											{"option-data": [{"data": "edc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.17.0/24"},
											{"option-data": [{"data": "udc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.15.0/24"}]}}`),
			[]string{"192.168.64.0/24"},
		},
		{
			[]byte(`{"Dhcp4": { "subnet4": [{"option-data": [{"data": "adc1.bmc.example.com","name": "domain-name"}], "subnet": "192.168.64.0/24"},
											{"option-data": [{"data": "edc4.ext.example.com","name": "domain-name"}], "subnet": "192.168.68.0/24"},
											{"option-data": [{"data": "edc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.17.0/24"},
											{"option-data": [{"data": "udc4.bmc.example.com","name": "domain-name"}], "subnet": "192.168.15.0/24"}]}}`),
			[]string{"192.168.64.0/24", "192.168.17.0/24", "192.168.15.0/24"},
		},
	}

	viper.SetDefault("scanner.kea_domain_name_suffix", ".bmc.example.com")
	for _, tc := range tt {
		networks := LoadSubnetsFromKea(tc.content)
		found := false
		foundNetworks := make([]string, 0)
		for _, network := range networks {
			for _, n := range tc.networks {
				if n == network.CIDR {
					foundNetworks = append(foundNetworks, network.CIDR)
					found = true
					break
				}
			}
		}
		if found == false || len(foundNetworks) != len(tc.networks) {
			t.Errorf("The result of %v for should be %v: found %v", string(tc.content), tc.networks, foundNetworks)
		}
	}
}
