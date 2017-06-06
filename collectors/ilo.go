package collectors

type Blade struct {
	Bay    *Bay   `xml:" BAY,omitempty" json:"BAY,omitempty"`
	Name   string `xml:" NAME,omitempty" json:"NAME,omitempty"`
	Power  *Power `xml:" POWER,omitempty" json:"POWER,omitempty"`
	Status string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	Temps  *Temps `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type Bay struct {
	Connection int `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
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
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
}

type Rimp struct {
	Infra2 *Infra2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
}

type Temp struct {
	C    string `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty" json:"DESC,omitempty"`
}

type Temps struct {
	Temp *Temp `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}
