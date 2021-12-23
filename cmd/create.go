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
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sampleConfig = []byte(`---
debug: true
site:
  - all
bmc_user: Priest
bmc_pass: Wololo
url: http://service.example.com/v1
database_type: postgres
database_options: host=0.0.0.0 user=postgres port=32768 dbname=postgres password=mysecretpassword
database_max_connections: 10

api:
  http_server_port: 8000

notification:
  enabled: false
  script: /usr/local/bin/notify-on-dora-change

metrics:
  enabled: false
  type: graphite
  host: localhost
  port: 2003
  prefix:
    publish: dora.publish
    server: dora.server
    worker: dora.worker

collector:
  concurrency: 60

  worker:
    enabled: false
    server: nats://172.17.0.3:4222
    queue: dora
    username:
    password:

  try_default_credentials: false
  default:
    dell:
      username: Priest
      password: Wololo

    supermicro:
      username: Priest
      password: Wololo

    hp:
      username: Priest
      password: Wololo

scanner:
  scanned_by: anomalia
  concurrency: 100
  kea_config: /etc/kea/kea-dhcp4.conf
  subnet_source: kea
  kea_domain_name_suffix: bmc.example.com
`)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates for you a sample config",
	Long:  `Creates for you a sample config file in your $HOME/.bmc-toolbox if you don't have one yet`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		configDir := fmt.Sprintf("%s/.bmc-toolbox", home)
		configFile := fmt.Sprintf("%s/dora.yaml", configDir)
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0o755)
			if err != nil {
				fmt.Printf("Failed to create the config directory %s: %s\n", configDir, err)
				os.Exit(1)
			}
			err = ioutil.WriteFile(configFile, sampleConfig, 0o755)
			if err != nil {
				fmt.Printf("Failed to create the temp config %s: %s\n", configFile, err)
				os.Exit(1)
			}
		} else {
			log.Info(configFile, " already exists...")
		}
	},
}

func init() {
	configCmd.AddCommand(createCmd)
}
