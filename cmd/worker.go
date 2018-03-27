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
	"fmt"
	"runtime"

	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.booking.com/go/dora/connectors"
)

func work(subject string) {
	nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
	if err != nil {
		log.Fatalf("Subscriber unable to connect: %v\n", err)
	}

	nc.QueueSubscribe(subject, viper.GetString("collector.worker.queue"), func(msg *nats.Msg) {
		fmt.Printf("%v", msg)
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{"queue": viper.GetString("collector.worker.queue"), "subject": subject}).Info("Subscribed to queue")
}

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
		work("dora::scan")
		connectors.DataCollectionWorker()
		runtime.Goexit()
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
	workerCmd.Flags().StringVarP(&queue, "queue", "q", "", "queue where we will listen for messages")
	viper.BindPFlag("collector.worker.queue", workerCmd.Flags().Lookup("queue"))
}
