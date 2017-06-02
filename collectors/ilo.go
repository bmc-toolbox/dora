package collectors

type Blade struct {
	Bay         *Bay   `xml:" BAY,omitempty" json:"BAY,omitempty"`
	Bladeromver string `xml:" BLADEROMVER,omitempty" json:"BLADEROMVER,omitempty"`
	Name        string `xml:" NAME,omitempty" json:"NAME,omitempty"`
	Power       *Power `xml:" POWER,omitempty" json:"POWER,omitempty"`
	Status      string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	Temps       *Temps `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type Bay struct {
	Connection string `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
}

type Infra2 struct {
	Addr   string  `xml:" ADDR,omitempty" json:"ADDR,omitempty"`
	Blades *Blades `xml:" BLADES,omitempty" json:"BLADES,omitempty"`
	Power  *Power  `xml:" POWER,omitempty" json:"POWER,omitempty"`
	Status string  `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	Temps  *Temp   `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type Blades struct {
	Blade []*Blade `xml:" BLADE,omitempty" json:"BLADE,omitempty"`
}

type Power struct {
	Powermode         string  `xml:" POWERMODE,omitempty" json:"POWERMODE,omitempty"`
	Powerstate        string  `xml:" POWERSTATE,omitempty" json:"POWERSTATE,omitempty"`
	PowerConsumed     float64 `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
	PowerOffWattage   string  `xml:" POWER_OFF_WATTAGE,omitempty" json:"POWER_OFF_WATTAGE,omitempty"`
	PowerOnWattage    string  `xml:" POWER_ON_WATTAGE,omitempty" json:"POWER_ON_WATTAGE,omitempty"`
	Redundancy        string  `xml:" REDUNDANCY,omitempty" json:"REDUNDANCY,omitempty"`
	Redundancymode    string  `xml:" REDUNDANCYMODE,omitempty" json:"REDUNDANCYMODE,omitempty"`
	RedundantCapacity string  `xml:" REDUNDANT_CAPACITY,omitempty" json:"REDUNDANT_CAPACITY,omitempty"`
	Status            string  `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	Type              string  `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
	WantedPS          string  `xml:" WANTED_PS,omitempty" json:"WANTED_PS,omitempty"`
}

type Rimp struct {
	Infra2 *Infra2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
}

type Temp struct {
	C         string       `xml:" C,omitempty" json:"C,omitempty"`
	Desc      string       `xml:" DESC,omitempty" json:"DESC,omitempty"`
	Location  string       `xml:" LOCATION,omitempty" json:"LOCATION,omitempty"`
	Threshold []*Threshold `xml:" THRESHOLD,omitempty" json:"THRESHOLD,omitempty"`
}

type Temps struct {
	Temp *Temp `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}

type Threshold struct {
	C      string `xml:" C,omitempty" json:"C,omitempty"`
	Desc   string `xml:" DESC,omitempty" json:"DESC,omitempty"`
	Status string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}
