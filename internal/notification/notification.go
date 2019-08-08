package notification

import (
	"context"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	notifyChange chan string
)

func init() {
	// Creates a channel with a buffer for 600 messages
	notifyChange = make(chan string, 600)
	go func(notification <-chan string) {
		for endpoint := range notification {
			log.WithFields(log.Fields{"operation": "notification", "endpoint": endpoint}).Debug("notification endpoint")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := exec.CommandContext(ctx, viper.GetString("notification.script"), endpoint).Run(); err != nil {
				log.WithFields(log.Fields{"operation": "notification", "endpoint": endpoint}).Error(err)
			}
			cancel()
		}
	}(notifyChange)
}

// NotifyChange will run a script to notify a system on changes to assets
func NotifyChange(asset string) {
	if !viper.GetBool("notification.enabled") {
		return
	}
	notifyChange <- asset
	return
}
