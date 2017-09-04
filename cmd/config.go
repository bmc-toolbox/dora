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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Shows the path of the current config in use",
	Long: `You can use the config command to create a sample config file 
in your $HOME/.bmc-toolbox directory with all possible config flags.

usage: dora config
       dora config create
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Current config in use: ", viper.ConfigFileUsed())
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
