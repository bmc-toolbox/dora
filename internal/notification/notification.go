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
		// Notification
		viper.SetDefault("notification.enabled", false)
		viper.SetDefault("notification.script", "/usr/local/bin/notify-on-dora-change")
		method := viper.GetString("notification.enabled")
		for endpoint := range notification {
			switch method {
			case "script":
				cmd := exec.Command(viper.GetString("notification.script"), endpoint)
				err := cmd.Run()
				if err != nil {
					log.WithFields(log.Fields{"operation": "notification", "endpoint": endpoint}).Error(err)
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
