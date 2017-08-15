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

type HpRimpBlade struct {
	HpHSI *HpHSI `xml:" HSI,omitempty" json:"HSI,omitempty"`
}

type HpHSI struct {
	HpNICS *HpNICS `xml:" NICS,omitempty" json:"NICS,omitempty"`
}

type HpNICS struct {
	HpNIC []*HpNIC `xml:" NIC,omitempty" json:"NICS,omitempty"`
}

type HpNIC struct {
	Description string `xml:" DESCRIPTION,omitempty" json:"DESCRIPTION,omitempty"`
	MacAddr     string `xml:" MACADDR,omitempty" json:"MACADDR,omitempty"`
	Status      string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

// HpFirmware is the struct used to render the data from https://$ip/json/fw_info
type HpFirmware struct {
	HostpwrState     string `json:"hostpwr_state"`
	InPost           int    `json:"in_post"`
	AmsReady         string `json:"ams_ready"`
	DataStateNetwork string `json:"data_state_network"`
	DataStateStorage string `json:"data_state_storage"`
	Firmware         []struct {
		FwIndex   int    `json:"fw_index"`
		FwName    string `json:"fw_name"`
		FwVersion string `json:"fw_version"`
		Location  string `json:"Location"`
	} `json:"firmware"`
}
