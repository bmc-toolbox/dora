// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/connectors"
	"gitlab.booking.com/go/dora/scanner"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Dora worker service",
	Long: `Dora worker is responsible to get the jobs 
from the queue and process, at this point it's only possible to 
define the queues via config file.

usage: dora worker
`,
	Run: func(cmd *cobra.Command, args []string) {
		scanner.ScanNetworksWorker()
		connectors.DataCollectionWorker()
		runtime.Goexit()
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
	workerCmd.Flags().StringVarP(&queue, "queue", "q", "", "queue where we will listen for messages")
	viper.BindPFlag("collector.worker.queue", workerCmd.Flags().Lookup("queue"))
}
