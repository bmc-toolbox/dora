package notification

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	notifyChange chan string
)

func init() {
	notifyChange = make(chan string)
	go func(notification <-chan string) {
		method := viper.GetString("notification.script_or_http")
		for asset := range notification {
			switch method {
			case "script":
				cmd := exec.Command(viper.GetString("notification.script"), asset)
				err := cmd.Run()
				if err != nil {
					log.WithFields(log.Fields{"operation": "notification", "asset": asset}).Error(err)
					continue
				}
			}

		}
	}(notifyChange)
}

// NotifyChange will run a script to notify a system on changes to assets
func NotifyChange(asset string) (err error) {
	if !viper.GetBool("notification.enabled") {
		return
	}
	return
}
