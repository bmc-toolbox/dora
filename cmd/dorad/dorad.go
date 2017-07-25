package main

import (
	"os"

	"gitlab.booking.com/infra/dora/web"

	"github.com/google/gops/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

}

func main() {
	if err := agent.Listen(nil); err != nil {
		log.Fatal("Couldn't start gops agent", err)
	}
	viper.SetConfigName("dora")
	viper.AddConfigPath("/etc/bmc-toolbox")
	viper.AddConfigPath("$HOME/.bmc-toolbox")
	viper.SetDefault("debug", false)
	viper.SetDefault("http_server_port", 8000)

	configItems := []string{
		"database_type",
		"database_options",
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Exiting because I couldn't find the configuration file...")
	}

	for _, item := range configItems {
		if !viper.IsSet(item) {
			log.Fatalf("Parameter %s is missing in the config file\n", item)
		}
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	web.RunGin(viper.GetInt("http_server_port"))
}
