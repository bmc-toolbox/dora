package main

import (
	"os"

	"gitlab.booking.com/infra/dora/scanner"

	"github.com/google/gops/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	site        string
	concurrency int
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
	viper.SetDefault("site", "all")
	viper.SetDefault("concurrency", 20)
	viper.SetDefault("debug", false)
	viper.SetDefault("noop", false)
	viper.SetDefault("kea_config", "/etc/kea/kea.conf")
	viper.SetDefault("nmap_xml_dir", "/tmp/dora/scans")
	viper.SetDefault("nmap", "/bin/nmap")
	viper.SetDefault("nmap_tcp_ports", "22,443")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Exiting because I couldn't find the configuration file...")
	}

	scanner.ScanNetworks()

	// if viper.GetBool("disable_discrete") == false {
	// 	discreteStep()
	// }
}
