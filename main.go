package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"./collectors"
	"./parsers"
	"./simpleapi"

	"github.com/spf13/viper"
)

const concurrency = 1

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

// func collect(c chan *simpleapi.Chassi) {
// 	for chassi := range c {
// 		for name, data := range chassi.Interfaces {
// 			fmt.Println(chassi.Fqdn, name, data)
// 		}
// 	}
// }

func collect(c <-chan *simpleapi.Chassi) {
	for chassi := range c {
		viaILO := false
		for name, data := range chassi.Interfaces {
			if data.IPAddress == "" {
				continue
			}

			collector := collectors.New(
				viper.GetString("bmc_user"),
				viper.GetString("bmc_pass"),
			)

			fmt.Println(fmt.Sprintf("Trying to collect data from %s[%s] via web %s", chassi.Fqdn, data.IPAddress, name))
			result, err := collector.ViaILOXML(data.IPAddress)
			if err != nil {
				fmt.Println(err)
			}
			thing := &parsers.RIMP{}
			err = xml.Unmarshal(result, thing)
			if err != nil {
				fmt.Println(err)
			}
			for _, blade := range thing.INFRA2.BLADES.BLADE {
				if blade.NAME != nil {
					now := int32(time.Now().Unix())
					fmt.Printf("power_kw,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassi=%s,subdevice=%s value=%b %d\n", "ams4", "Z04", "LL", "LLEven", "LL08", blade.BAY.CONNECTION.Text, chassi.Fqdn, blade.NAME.Text, blade.POWER.POWER_CONSUMED.Text/1000, now)
					fmt.Printf("temp_c,site=%s,zone=%s,pod=%s,row=%s,rack=%s,bay=%s,device=chassis,chassi=%s,subdevice=%s value=%s %d\n", "ams4", "Z04", "LL", "LLEven", "LL08", blade.BAY.CONNECTION.Text, chassi.Fqdn, blade.NAME.Text, blade.TEMPS.TEMP.C.Text, now)
					viaILO = true
				}
			}
			break
		}

		if !viaILO {
			collector := collectors.New(
				viper.GetString("bmc_user"),
				viper.GetString("bmc_pass"),
			)
			for name, data := range chassi.Interfaces {
				fmt.Println(fmt.Sprintf("Trying to collect data from %s[%s] via console %s", chassi.Fqdn, data.IPAddress, name))
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

	cc := make(chan *simpleapi.Chassi, concurrency)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			collect(cc)
		}()
	}

	for _, c := range chassis.Chassis {
		//fmt.Println(c.Fqdn)
		time.Sleep(500 * time.Millisecond)
		cc <- &c
	}
	close(cc)
	wg.Wait()
}
