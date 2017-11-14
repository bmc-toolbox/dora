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

	"github.com/google/gops/agent"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dora",
	Short: "Tool to discover, collect data and manage all types of BMCs and Chassis",
	Long: `Tool to discover, collect data and manage all types of BMCs and Chassis:

Dora scan the networks found in the kea.conf from there it discovers 
all types of BMCs and Chassis. Dora can also configure chassis and/or 
make ad-hoc queries to specific devices and/or ips. 

We currently support HP, Dell and Supermicros.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if err := agent.Listen(&agent.Options{}); err != nil {
		log.Fatal(err)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/bmc-toolbox/dora.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetConfigName("dora")
		viper.AddConfigPath("/etc/bmc-toolbox")
		viper.AddConfigPath(fmt.Sprintf("%s/.bmc-toolbox", home))
	}

	viper.SetDefault("site", []string{"all"})
	viper.SetDefault("debug", false)
	viper.SetDefault("noop", false)

	// Collector
	viper.SetDefault("collector.dump_invalid_payloads", false)
	viper.SetDefault("collector.dump_invalid_payload_path", "/tmp/dora/dumps")

	// Api
	viper.SetDefault("api.http_server_port", 8000)

	// Scan
	viper.SetDefault("scanner.kea_domain_name_suffix", ".lom.booking.com")
	viper.SetDefault("scanner.kea_config", "/etc/kea/kea.conf")
	viper.SetDefault("scanner.subnet_source", "kea")
	viper.SetDefault("scanner.nmap", "/usr/bin/nmap")
	viper.SetDefault("scanner.concurrency", 100)

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Unable to find my hostname: %s", err)
	}

	viper.SetDefault("scanner.scanned_by", hostname)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Failed to read config: %s", err)
		os.Exit(1)
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}
