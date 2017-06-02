package collectors

type DellRedFishPower struct {
	PowerControl []struct {
		Name               string  `json:"Name"`
		PowerConsumedWatts float64 `json:"PowerConsumedWatts"`
	} `json:"PowerControl"`
}

type DellRedFishThermal struct {
	Temperatures []struct {
		Name           string `json:"Name"`
		ReadingCelsius int    `json:"ReadingCelsius"`
	} `json:"Temperatures"`
}
