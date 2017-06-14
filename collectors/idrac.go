package collectors

type CMC struct {
	Chassis *DellChassis          `json:"1"`
	Blades  map[string]*DellBlade `json:"blades_status"`
}

type DellChassis struct {
	TempHealth                 int    `json:"TempHealth"`
	TempUpperCriticalThreshold int    `json:"TempUpperCriticalThreshold"`
	TempSensorID               int    `json:"TempSensorID"`
	TempCurrentValue           int    `json:"TempCurrentValue"`
	TempLowerCriticalThreshold int    `json:"TempLowerCriticalThreshold"`
	TempPresence               int    `json:"TempPresence"`
	TempSensorName             string `json:"TempSensorName"`
}

type DellBlade struct {
	Temperature        string  `json:"bladeTemperature"`
	SystemName         string  `json:"bladeSystemName"`
	Present            int     `json:"bladePresent"`
	IdracURL           string  `json:"idracURL"`
	CurrentConsumption float64 `json:"bladeCurrentConsumption"`
	Slot               string  `json:"bladeSlot"`
	IsStorageBlade     int     `json:"isStorageBlade"`
	Name               string  `json:"bladeName"`
	Bsn                string  `json:"bladeSerialNum"`
}

// type CMC struct {
// 	Chassis map[string]*DellChassis `json:"0"`
// 	Blades  map[string]*DellBlade   `json:"blades_status"`
// }

// type DellChassis struct {
// 	Blades map[string]*DellBlade `json:"blades_status"`
// }

// type DellBlade struct {
// 	Temperature        string `json:"bladeTemperature"`
// 	SystemName         string `json:"bladeSystemName"`
// 	Present            int    `json:"bladePresent"`
// 	IdracURL           string `json:"idracURL"`
// 	CurrentConsumption int    `json:"bladeCurrentConsumption"`
// 	Slot               string `json:"bladeSlot"`
// 	IsStorageBlade     int    `json:"isStorageBlade"`
// 	Name               string `json:"bladeName"`
// 	Bsn                string `json:"bladeSerialNum"`
// }
