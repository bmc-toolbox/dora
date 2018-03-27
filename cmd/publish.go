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
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Dora publish adds a job to one of the dora queues",
	Long: `Dora publish adds a job to one of the dora queues, checking
wheter it's valid for the given queue.

usage: dora publish 192.168.0.1 -q dora -s collect 
	   dora publish 192.168.0.1 -q dora -s collect 
	   dora publish all -q dora -s collect 
	   dora publish all -q dora -s collect 
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		nc, err := nats.Connect(viper.GetString("collector.worker.server"), nats.UserInfo(viper.GetString("collector.worker.username"), viper.GetString("collector.worker.password")))
		if err != nil {
			log.Fatalf("Subscriber unable to connect: %v\n", err)
		}

		switch subject {
		case "scan":
			subject = "dora::scan"
		case "collect":
			subject = "dora::collect"
		default:
			log.WithFields(log.Fields{"queue": queue, "subject": subject}).Error("unknown subject: %s", subject)
		}

		for _, payload := range args {
			nc.Publish(subject, []byte(payload))
			nc.Flush()
			if err := nc.LastError(); err != nil {
				log.WithFields(log.Fields{"queue": queue, "subject": subject, "payload": payload}).Error(err)
			} else {
				log.WithFields(log.Fields{"queue": queue, "subject": subject, "payload": payload}).Info("sent")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringVarP(&subject, "subject", "s", "", "subject to be used for the queue")
	publishCmd.Flags().StringVarP(&queue, "queue", "q", "", "queue where we will publish the message")
}
