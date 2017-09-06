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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/dora/connectors"
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects hosts found by the scanner or collect a given list of hosts",
	Long: `Collects hosts found by the scanner or collect a given list of hosts, we will
use the data from the scanner and will only try to collect hosts that have the required 
ports opened. 

usage: dora collect 
       dora collect 192.168.0.1
`,
	Run: func(cmd *cobra.Command, args []string) {
		configItems := []string{
			"bmc_pass",
			"bmc_user",
			"site",
			"database_type",
			"database_options",
		}

		for _, item := range configItems {
			if !viper.IsSet(item) {
				fmt.Printf("Parameter %s is missing in the config file\n", item)
				os.Exit(1)
			}
		}

		if len(args) == 0 {
			connectors.DataCollection([]string{"all"})
		} else {
			connectors.DataCollection(args)
		}
	},
}

func init() {
	RootCmd.AddCommand(collectCmd)
}
