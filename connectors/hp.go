package connectors

type HpBlade struct {
	HpBay       *HpBay   `xml:" BAY,omitempty" json:"BAY,omitempty"`
	Bsn         string   `xml:" BSN,omitempty" json:"BSN,omitempty"`
	MgmtIPAddr  string   `xml:" MGMTIPADDR,omitempty" json:"MGMTIPADDR,omitempty"`
	MgmtType    string   `xml:" MGMTPN,omitempty" json:"MGMTPN,omitempty"`
	MgmtVersion string   `xml:" MGMTFWVERSION,omitempty" json:"MGMTFWVERSION,omitempty"`
	Name        string   `xml:" NAME,omitempty" json:"NAME,omitempty"`
	HpPower     *HpPower `xml:" POWER,omitempty" json:"POWER,omitempty"`
	Status      string   `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	Spn         string   `xml:" SPN,omitempty" json:"SPN,omitempty"`
	HpTemps     *HpTemps `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
}

type HpBay struct {
	Connection int `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
}

type HpInfra2 struct {
	Addr           string          `xml:" ADDR,omitempty" json:"ADDR,omitempty"`
	HpBlades       *HpBlades       `xml:" BLADES,omitempty" json:"BLADES,omitempty"`
	HpSwitches     *HpSwitches     `xml:" SWITCHES,omitempty" json:"SWITCHES,omitempty"`
	HpChassisPower *HpChassisPower `xml:" POWER,omitempty" json:"POWER,omitempty"`
	Status         string          `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	HpTemps        *HpTemps        `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
	EnclSn         string          `xml:" ENCL_SN,omitempty" json:"ENCL_SN,omitempty"`
	Pn             string          `xml:" PN,omitempty" json:"PN,omitempty"`
	Encl           string          `xml:" ENCL,omitempty" json:"ENCL,omitempty"`
	Rack           string          `xml:" RACK,omitempty" json:"RACK,omitempty"`
}

type HpMP struct {
	Sn   string `xml:" SN,omitempty" json:"SN,omitempty"`
	Fwri string `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
}

type HpSwitches struct {
	HpSwitch []*HpSwitch `xml:" SWITCH,omitempty" json:"BLADE,omitempty"`
}

type HpSwitch struct {
	Spn string `xml:" SPN,omitempty" json:"SPN,omitempty"`
}

type HpBlades struct {
	HpBlade []*HpBlade `xml:" BLADE,omitempty" json:"BLADE,omitempty"`
}

type HpPower struct {
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
}

type HpChassisPower struct {
	PowerConsumed float64          `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
	HpPowersupply []*HpPowersupply `xml:" POWERSUPPLY,omitempty" json:"POWERSUPPLY,omitempty"`
}

type HpRimp struct {
	HpInfra2 *HpInfra2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
	HpMP     *HpMP     `xml:" MP,omitempty" json:"MP,omitempty"`
}

type HpPowersupply struct {
	Status string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type HpTemp struct {
	C    int    `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty" json:"DESC,omitempty"`
}

type HpTemps struct {
	HpTemp *HpTemp `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}
