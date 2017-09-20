// Copyright © 2017 Juliano Martinez <juliano.martinez@booking.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/scanner"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "scan networks found in kea config or a list of given networks",
	Long: `scan networks found in kea config or a list of given networks and search 
for the required tcp and udp ports for the hardware discovery. It will build a list of 
discoverable assets to be later used by dora collector

usage: dora scan  
	   dora scan 192.168.0.1/24
	   dora scan list
	   dora scan loadSubnets <subnetSource>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 && args[0] != "all" {
			subnets := []string{}
			for _, subnet := range args {
				_, _, err := net.ParseCIDR(subnet)
				if err != nil {
					fmt.Println(err)
					continue
				}
				subnets = append(subnets, subnet)
			}
			scanner.ScanNetworks(subnets, viper.GetStringSlice("site"))
		} else {
			scanner.ScanNetworks([]string{"all"}, viper.GetStringSlice("site"))
		}
	},
}

func init() {
	viper.SetDefault("scanner.kea_domain_name_suffix", ".lom.booking.com")
	viper.SetDefault("scanner.kea_config", "/etc/kea/kea.conf")
	viper.SetDefault("scanner.subnet_source", "kea")
	viper.SetDefault("scanner.nmap", "/usr/bin/nmap")
	viper.SetDefault("scanner.concurrency", 100)

	RootCmd.AddCommand(scanCmd)
}
