package connectors

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
