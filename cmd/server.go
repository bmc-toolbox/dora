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
	"github.com/bmc-toolbox/dora/internal/metrics"
	"os"
	"time"

	"github.com/bmc-toolbox/dora/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var port int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Dora Api Server",
	Long: `Dora API exposed all the stored information from the database 
via json:api (http://jsonapi.org). To know more check our docs. 

usage: dora server
`,
	Run: func(cmd *cobra.Command, args []string) {
		configItems := []string{
			"database_type",
			"database_options",
		}

		for _, param := range configItems {
			if !viper.IsSet(param) {
				fmt.Printf("Parameter %s is missing in the config file\n", param)
				os.Exit(1)
			}
		}

		if viper.GetBool("metrics.enabled") {
			err := metrics.Setup(
				viper.GetString("metrics.type"),
				viper.GetString("metrics.host"),
				viper.GetInt("metrics.port"),
				viper.GetString("metrics.prefix.server"),
				time.Minute,
			)
			if err != nil {
				fmt.Printf("Failed to set up monitoring: %s", err)
				os.Exit(1)
			}
		}

		web.RunGin(viper.GetInt("api.http_server_port"), viper.GetBool("debug"))
	},
}

func init() {
	RootCmd.Flags().IntVar(&port, "port", 8080, "Port to bind the webserver")
	viper.BindPFlag("api.http_server_port", RootCmd.Flags().Lookup("port"))

	RootCmd.AddCommand(serverCmd)
}
