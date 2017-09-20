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
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the networks that could be scanned by dora",
	Long: `List all the networks that could be scanned by dora. It also
config if a network is scannable based on the list of arguments passed
to the command

usage: dora scan list
       dora scan list 192.168.0.1
`,
	Run: func(cmd *cobra.Command, args []string) {
		// site := strings.Split(viper.GetString("site"), " ")
		// content, err := scanner.ReadKeaConfig()
		// if err != nil {
		// 	fmt.Printf("Error reading kea config: %s\n", err)
		// 	os.Exit(1)
		// }

		// db := storage.InitDB()
		// for _, subnet := range scanner.LoadSubnets(content, site) {
		// 	if len(args) != 0 {
		// 		for _, n := range args {
		// 			if n == subnet.String() {
		// 				fmt.Println(subnet.String())
		// 			}
		// 		}
		// 	} else {
		// 		fmt.Println(subnet.String())
		// 	}
		// }
	},
}

func init() {
	scanCmd.AddCommand(listCmd)
}
