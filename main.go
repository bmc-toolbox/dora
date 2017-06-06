package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"./collectors"
	"./simpleapi"

	"github.com/google/gops/agent"
	"github.com/spf13/viper"
)

var (
	simpleAPI *simpleapi.SimpleAPI
	collector *collectors.Collector
)

// TODO: Better error handling for the config

func collect(c <-chan simpleapi.Chassis) {
	for chassis := range c {
		rack, err := simpleAPI.GetRack(chassis.Rack)
		if err != nil {
			fmt.Printf("Received error: %s\n", err)
		}

		for ifname, ifdata := range chassis.Interfaces {
			if ifdata.IPAddress == "" {
				continue
			}

			err := collector.CollectViaChassi(&chassis, &rack, &ifdata.IPAddress, &ifname)
			if err == nil {
				break
			}
		}
	}
}

func main() {
	if err := agent.Listen(nil); err != nil {
		log.Fatal(err)
	}
	viper.SetConfigName("thermalnator")
	viper.AddConfigPath("/etc/bmc-toolbox")
	viper.AddConfigPath("$HOME/.bmc-toolbox")
	viper.SetDefault("site", "all")
	viper.SetDefault("concurrency", 20)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Exiting because I couldn't find the configuration file...")
	}

	simpleAPI = simpleapi.New(
		viper.GetString("simpleapi_user"),
		viper.GetString("simpleapi_pass"),
		viper.GetString("simpleapi_base_url"),
	)

	collector = collectors.New(
		viper.GetString("bmc_user"),
		viper.GetString("bmc_pass"),
	)

	site := viper.GetString("site")
	concurrency := viper.GetInt("concurrency")

	chassis, err := simpleAPI.Chassis()
	if err != nil {
		fmt.Println("error simpleapi:", err)
	}

	cc := make(chan simpleapi.Chassis, concurrency)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(input <-chan simpleapi.Chassis, wg *sync.WaitGroup, i int) {
			collect(input)
			wg.Done()
		}(cc, &wg, i)
	}

	fmt.Printf("Starting data collection for %s site(s)\n", site)

	for _, c := range chassis.Chassis {
		if strings.Compare(c.Location, site) == 0 || strings.Compare(site, "all") == 0 {
			cc <- *c
		}
	}
	close(cc)
	wg.Wait()
}
