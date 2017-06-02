package parsers

type BLADE struct {
	BAY          *BAY   `xml:" BAY,omitempty" json:"BAY,omitempty"`
	BLADEROMVER  string `xml:" BLADEROMVER,omitempty" json:"BLADEROMVER,omitempty"`
	BSN          string `xml:" BSN,omitempty" json:"BSN,omitempty"`
	MANUFACTURER string `xml:" MANUFACTURER,omitempty" json:"MANUFACTURER,omitempty"`
	NAME         string `xml:" NAME,omitempty" json:"NAME,omitempty"`
	POWER        *POWER `xml:" POWER,omitempty" json:"POWER,omitempty"`
	STATUS       string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TEMPS        *TEMPS `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type BAY struct {
	CONNECTION string `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
}

type INFRA2 struct {
	ADDR   string  `xml:" ADDR,omitempty" json:"ADDR,omitempty"`
	BLADES *BLADES `xml:" BLADES,omitempty" json:"BLADES,omitempty"`
	POWER  *POWER  `xml:" POWER,omitempty" json:"POWER,omitempty"`
	STATUS string  `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TEMPS  *TEMP   `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type BLADES struct {
	BLADE []*BLADE `xml:" BLADE,omitempty" json:"BLADE,omitempty"`
}

type POWER struct {
	CAPACITY           string         `xml:" CAPACITY,omitempty" json:"CAPACITY,omitempty"`
	DYNAMICPOWERSAVER  string         `xml:" DYNAMICPOWERSAVER,omitempty" json:"DYNAMICPOWERSAVER,omitempty"`
	NEEDED_PS          string         `xml:" NEEDED_PS,omitempty" json:"NEEDED_PS,omitempty"`
	OUTPUT_POWER       string         `xml:" OUTPUT_POWER,omitempty" json:"OUTPUT_POWER,omitempty"`
	PDU                string         `xml:" PDU,omitempty" json:"PDU,omitempty"`
	POWERMODE          string         `xml:" POWERMODE,omitempty" json:"POWERMODE,omitempty"`
	POWERONFLAG        string         `xml:" POWERONFLAG,omitempty" json:"POWERONFLAG,omitempty"`
	POWERSTATE         string         `xml:" POWERSTATE,omitempty" json:"POWERSTATE,omitempty"`
	POWERSUPPLY        []*POWERSUPPLY `xml:" POWERSUPPLY,omitempty" json:"POWERSUPPLY,omitempty"`
	POWER_CONSUMED     float64        `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
	POWER_OFF_WATTAGE  string         `xml:" POWER_OFF_WATTAGE,omitempty" json:"POWER_OFF_WATTAGE,omitempty"`
	POWER_ON_WATTAGE   string         `xml:" POWER_ON_WATTAGE,omitempty" json:"POWER_ON_WATTAGE,omitempty"`
	REDUNDANCY         string         `xml:" REDUNDANCY,omitempty" json:"REDUNDANCY,omitempty"`
	REDUNDANCYMODE     string         `xml:" REDUNDANCYMODE,omitempty" json:"REDUNDANCYMODE,omitempty"`
	REDUNDANT_CAPACITY string         `xml:" REDUNDANT_CAPACITY,omitempty" json:"REDUNDANT_CAPACITY,omitempty"`
	STATUS             string         `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TYPE               string         `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
	WANTED_PS          string         `xml:" WANTED_PS,omitempty" json:"WANTED_PS,omitempty"`
}

type POWERSUPPLY struct {
	ACINPUT      string `xml:" ACINPUT,omitempty" json:"ACINPUT,omitempty"`
	ACTUALOUTPUT string `xml:" ACTUALOUTPUT,omitempty" json:"ACTUALOUTPUT,omitempty"`
	CAPACITY     string `xml:" CAPACITY,omitempty" json:"CAPACITY,omitempty"`
	STATUS       string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type RIMP struct {
	INFRA2 *INFRA2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
}

type TEMP struct {
	C         string       `xml:" C,omitempty" json:"C,omitempty"`
	DESC      string       `xml:" DESC,omitempty" json:"DESC,omitempty"`
	LOCATION  string       `xml:" LOCATION,omitempty" json:"LOCATION,omitempty"`
	THRESHOLD []*THRESHOLD `xml:" THRESHOLD,omitempty" json:"THRESHOLD,omitempty"`
}

type TEMPS struct {
	TEMP *TEMP `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}

type THRESHOLD struct {
	C      string `xml:" C,omitempty" json:"C,omitempty"`
	DESC   string `xml:" DESC,omitempty" json:"DESC,omitempty"`
	STATUS string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}
