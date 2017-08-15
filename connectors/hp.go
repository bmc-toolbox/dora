package connectors

// HpBlade contains the unmarshalled data from the hp chassis
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

// HpBay contains the position of the blade within the chassis
type HpBay struct {
	Connection int `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
}

// HpInfra2 is the data retrieved from the chassis xml interface that contains all components
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

// HpMP contains the firmware version and the model of the chassis
type HpMP struct {
	Sn   string `xml:" SN,omitempty" json:"SN,omitempty"`
	Fwri string `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
}

// HpSwitches contains all the switches we have within the chassis
type HpSwitches struct {
	HpSwitch []*HpSwitch `xml:" SWITCH,omitempty" json:"BLADE,omitempty"`
}

// HpSwitch contains the type of the switch
type HpSwitch struct {
	Spn string `xml:" SPN,omitempty" json:"SPN,omitempty"`
}

// HpBlades contains all the blades we have within the chassis
type HpBlades struct {
	HpBlade []*HpBlade `xml:" BLADE,omitempty" json:"BLADE,omitempty"`
}

// HpPower contains the power information of a blade
type HpPower struct {
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
}

// HpChassisPower contains the power information of the chassis
type HpChassisPower struct {
	PowerConsumed float64          `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
	HpPowersupply []*HpPowersupply `xml:" POWERSUPPLY,omitempty" json:"POWERSUPPLY,omitempty"`
}

// HpRimp is the entry data structure for the chassis
type HpRimp struct {
	HpInfra2 *HpInfra2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
	HpMP     *HpMP     `xml:" MP,omitempty" json:"MP,omitempty"`
}

// HpPowersupply contains the data of the power supply of the chassis
type HpPowersupply struct {
	Status string `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

// HpTemp contains the thermal data of a chassis or blade
type HpTemp struct {
	C    int    `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty" json:"DESC,omitempty"`
}

// HpTemps contains the thermal data of a chassis or blade
type HpTemps struct {
	HpTemp *HpTemp `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}

// HpRimpBlade is the entry data structure for the blade when queries directly
type HpRimpBlade struct {
	HpHSI *HpHSI `xml:" HSI,omitempty" json:"HSI,omitempty"`
}

// HpHSI contains the information about the components of the blade
type HpHSI struct {
	HpNICS *HpNICS `xml:" NICS,omitempty" json:"NICS,omitempty"`
}

// HpNICS contains the list of nics that a blade has
type HpNICS struct {
	HpNIC []*HpNIC `xml:" NIC,omitempty" json:"NICS,omitempty"`
}

// HpNIC contains the nic information of a blade
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
