package collectors

type DellRedFishPower struct {
	PowerControl []struct {
		PowerConsumedWatts int `json:"PowerConsumedWatts"`
	} `json:"PowerControl"`
}