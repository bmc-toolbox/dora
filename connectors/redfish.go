package connectors

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrRedFishNotSupported = errors.New("RedFish not supported")
	redfishVendorEndPoints = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
		},
		HP: map[string]string{
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
		Supermicro: map[string]string{
			RFPower:   "redfish/v1/Chassis/1/Power",
			RFThermal: "redfish/v1/Chassis/1/Thermal",
		},
	}
	redfishVendorLabels = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Board Inlet Temp",
		},
		HP: map[string]string{
			//			RFPower:   "PowerMetrics",
			RFThermal: "30-System Board",
		},
		Supermicro: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Temp",
		},
	}
	bmcAddressBuild = regexp.MustCompile(".(prod|corp|dqs).")
)

type RedFishPower struct {
	PowerControl []struct {
		Name               string  `json:"Name"`
		PowerConsumedWatts float64 `json:"PowerConsumedWatts"`
	} `json:"PowerControl"`
}

type RedFishThermal struct {
	Temperatures []struct {
		Name           string `json:"Name"`
		ReadingCelsius int    `json:"ReadingCelsius"`
	} `json:"Temperatures"`
}

type RedFishConnection struct {
	username string
	password string
}

func (c *RedFishConnection) read(ip *string, collectType string, vendor string) (payload []byte, err error) {
	payload, err = c.httpGet(fmt.Sprintf("https://%s/%s", *ip, redfishVendorEndPoints[collectType][vendor]))
	if err == ErrPageNotFound {
		return payload, ErrRedFishNotSupported
	}
	return payload, err
}
