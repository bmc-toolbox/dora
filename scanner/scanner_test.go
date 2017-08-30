package scanner

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tt := []struct {
		content     []byte
		datacenters []string
		networks    []string
	}{
		{
			[]byte(`{"Dhcp4": { "subnet4": [{"option-data": [{"data": "hkg1.lom.booking.com","name": "domain-name" }], "subnet": "10.128.64.0/24"},
											{"option-data": [{"data": "ams4.corp.booking.com","name": "domain-name"}], "subnet": "10.196.68.0/24"},
											{"option-data": [{"data": "example.com","name": "domain-name"}], "subnet": "10.196.17.0/24"},
											{"option-data": [{"data": "lhr4.lom.booking.com","name": "domain-name"}], "subnet": "10.189.15.0/24"}]}}`),
			[]string{"lhr4", "ams4"},
			[]string{"10.196.17.0/24", "10.189.15.0/24"},
		},
		{
			[]byte(`{"Dhcp4": { "subnet4": [{"option-data": [{"data": "hkg1.lom.booking.com","name": "domain-name" }], "subnet": "10.128.64.0/24"},
											{"option-data": [{"data": "ams4.corp.booking.com","name": "domain-name"}], "subnet": "10.196.68.0/24"},
											{"option-data": [{"data": "example.com","name": "domain-name"}], "subnet": "10.196.17.0/24"},
											{"option-data": [{"data": "lhr4.lom.booking.com","name": "domain-name"}], "subnet": "10.189.15.0/24"}]}}`),
			[]string{"hkg1"},
			[]string{"10.128.64.0/24"},
		},
	}

	for _, tc := range tt {
		networks := loadSubnets(tc.content, tc.datacenters)
		found := false
		for _, network := range networks {
			for _, n := range tc.networks {
				if n == network.String() {
					found = true
				}
			}
		}
		if found == false || len(networks) != len(tc.networks) {
			t.Errorf("The result of %v for the datacenters %v should be %v: found %v", string(tc.content), tc.datacenters, tc.networks, networks)
		}
	}
}
