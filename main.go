package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"./collectors"
	"./parsers"
	"./simpleapi"

	"github.com/spf13/viper"
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

func collect(chassi *simpleapi.Chassi, collector *collectors.Collector) {
	for name, data := range chassi.Interfaces {
		fmt.Println(fmt.Sprintf("Trying to collect data from %s via web %s", data.IPAddress, name))
		result, err := collector.ViaILOXML(data.IPAddress)
		if err != nil {
			fmt.Println(err)
			continue
		}
		thing := &parsers.RIMP{}
		err = xml.Unmarshal(result, thing)
		if err != nil {
			fmt.Println(err)
		}
		for _, cust := range thing.INFRA2.BLADES.BLADE {
			if cust.NAME != nil {
				fmt.Printf("%s - %s VA - %s C - \\o/ \n", cust.NAME.Text, cust.POWER.POWER_CONSUMED.Text, cust.TEMPS.TEMP.C.Text)
				return
			}
		}
	}

	for name, data := range chassi.Interfaces {
		fmt.Println(fmt.Sprintf("Trying to collect data from %s via console %s", data.IPAddress, name))
		result, err := collector.ViaConsole(data.IPAddress)
		if err == nil {
			parseHPPower(result.PowerUsage)
			continue
		} else if err == collectors.ErrIsNotActive {
			continue
		} else {
			fmt.Println(err)
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

	simpleAPI := simpleapi.New(
		viper.GetString("simpleapi_user"),
		viper.GetString("simpleapi_pass"),
		viper.GetString("simpleapi_base_url"),
	)

	chassis, err := simpleAPI.Chassis()
	if err != nil {
		fmt.Println("error simpleapi:", err)
	}

	collector := collectors.New(
		viper.GetString("bmc_user"),
		viper.GetString("bmc_pass"),
	)

	for _, c := range chassis.Chassis {
		fmt.Println(c.Fqdn)
		collect(&c, collector)
	}
}
