package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"./collectors"
	"./simpleapi"

	"time"

	"github.com/spf13/viper"
)

const concurrency = 2

var (
	simpleAPI *simpleapi.SimpleAPI
	collector *collectors.Collector
)

// TODO: Better error handling for the config
// power_kw,site=AMS4,zone=Z04,pod=JJ,row=JJEven,rack=JJ12,pdu=ams4-bk-pdujj12-01 value=3.100000 1496220541
func parseHPPower(input string) {
	splitInput := strings.Split(input, "\n")
	eboa := 0
	dbps := 0

	for line := range splitInput {
		// We will only try to parse the power block if we are inside of the block
		if eboa != 2 {
			if eboa == 0 && strings.Compare(splitInput[line], "Enclosure Bay Output Allocation:") == 0 {
				eboa = 1
			} else if strings.HasPrefix(strings.TrimSpace(splitInput[line]), "=") {
				fmt.Println(strings.Split(splitInput[line], "                       =  ")[1])
				eboa = 2
			}
		} else if dbps != 2 {
			if dbps == 0 && strings.Compare(splitInput[line], "Device Bay Power Summary:") == 0 {
				dbps = 1
			} else if strings.HasPrefix(strings.TrimSpace(splitInput[line]), "=") {
				dbps = 2
			}
		}
	}
}

func collect(c <-chan *simpleapi.Chassis) {
	for chassis := range c {
		rack, err := simpleAPI.GetRack(chassis.Rack)
		if err != nil {
			fmt.Printf("Received error: %s\n", err)
		}

		fmt.Println(chassis.Fqdn)
		for ifname, ifdata := range chassis.Interfaces {
			if ifdata.IPAddress == "" {
				continue
			}

			err := collector.CollectViaChassi(chassis, &rack, &ifdata.IPAddress, &ifname)
			if err == nil {
				break
			}
		}
	}
}

func main() {
	viper.SetConfigName("thermalnator")
	viper.AddConfigPath("/etc/bmc-toolbox")
	viper.AddConfigPath("$HOME/.bmc-toolbox")

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

	chassis, err := simpleAPI.Chassis()
	if err != nil {
		fmt.Println("error simpleapi:", err)
	}

	cc := make(chan *simpleapi.Chassis, concurrency)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(cc <-chan *simpleapi.Chassis) {
			defer wg.Done()
			collect(cc)
		}(cc)
	}

	for _, c := range chassis.Chassis {
		fmt.Println(c.Fqdn)
		cc <- &c
		time.Sleep(90 * time.Minute)

	}
	close(cc)
	wg.Wait()
}
