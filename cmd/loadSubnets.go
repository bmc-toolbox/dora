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
	"gitlab.booking.com/infra/dora/scanner"
)

// loadSubnetsCmd represents the loadSubnets command
var loadSubnetsCmd = &cobra.Command{
	Use:   "loadSubnets",
	Short: "Load subnets to the scanner from different sources",
	Long: `Load subnets to the scanner from different sources. 
	
	usage: dora scan loadSubnets kea
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			scanner.LoadSubnets(viper.GetString("scanner.subnet_source"))
		} else {
			if !scanner.SupportedSources(args[0]) {
				fmt.Println("Source not supported")
				os.Exit(1)
			}
			scanner.LoadSubnets(args[0])
		}
	},
}

func init() {
	scanCmd.AddCommand(loadSubnetsCmd)
}
