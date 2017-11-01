package connectors

// HpBlade contains the unmarshalled data from the hp chassis
type HpBlade struct {
	HpBay           *HpBay   `xml:" BAY,omitempty"`
	Bsn             string   `xml:" BSN,omitempty"`
	MgmtIPAddr      string   `xml:" MGMTIPADDR,omitempty"`
	MgmtType        string   `xml:" MGMTPN,omitempty"`
	MgmtVersion     string   `xml:" MGMTFWVERSION,omitempty"`
	Name            string   `xml:" NAME,omitempty"`
	Type            string   `xml:" TYPE,omitempty"`
	HpPower         *HpPower `xml:" POWER,omitempty"`
	Status          string   `xml:" STATUS,omitempty"`
	Spn             string   `xml:" SPN,omitempty"`
	HpTemp          *HpTemp  `xml:" TEMPS>TEMP,omitempty"`
	BladeRomVer     string   `xml:" BLADEROMVER,omitempty"`
	AssociatedBlade string   `xml:" ASSOCIATEDBLADE,omitempty"`
}

// HpBay contains the position of the blade within the chassis
type HpBay struct {
	Connection int `xml:" CONNECTION,omitempty"`
}

// HpInfra2 is the data retrieved from the chassis xml interface that contains all components
type HpInfra2 struct {
	Addr           string          `xml:" ADDR,omitempty"`
	HpBlades       []*HpBlade      `xml:" BLADES>BLADE,omitempty"`
	HpSwitches     []*HpSwitch     `xml:" SWITCHES>SWITCH,omitempty"`
	HpChassisPower *HpChassisPower `xml:" POWER,omitempty"`
	Status         string          `xml:" STATUS,omitempty"`
	HpTemp         *HpTemp         `xml:" TEMPS>TEMP,omitempty"`
	EnclSn         string          `xml:" ENCL_SN,omitempty"`
	Pn             string          `xml:" PN,omitempty"`
	Encl           string          `xml:" ENCL,omitempty"`
	Rack           string          `xml:" RACK,omitempty"`
	HpManagers     []*HpManager    `xml:" MANAGERS>MANAGER,omitempty"`
}

// HpMP contains the firmware version and the model of the chassis or blade
type HpMP struct {
	Pn   string `xml:" PN,omitempty"`
	Sn   string `xml:" SN,omitempty"`
	Fwri string `xml:" FWRI,omitempty"`
}

// HpSwitch contains the type of the switch
type HpSwitch struct {
	Spn string `xml:" SPN,omitempty"`
}

// HpPower contains the power information of a blade
type HpPower struct {
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty"`
}

// HpChassisPower contains the power information of the chassis
type HpChassisPower struct {
	PowerConsumed float64          `xml:" POWER_CONSUMED,omitempty"`
	HpPowersupply []*HpPowersupply `xml:" POWERSUPPLY,omitempty"`
}

// HpRimp is the entry data structure for the chassis
type HpRimp struct {
	HpInfra2 *HpInfra2 `xml:" INFRA2,omitempty"`
	HpMP     *HpMP     `xml:" MP,omitempty"`
}

// HpManager hold the information of the manager board of the chassis
type HpManager struct {
	MgmtIPAddr string `xml:" MGMTIPADDR,omitempty"`
	MacAddr    string `xml:" MACADDR,omitempty"`
	Status     string `xml:" STATUS,omitempty"`
	Name       string `xml:" NAME,omitempty"`
}

// HpPowersupply contains the data of the power supply of the chassis
type HpPowersupply struct {
	Sn           string `xml:" SN,omitempty"`
	Status       string `xml:" STATUS,omitempty"`
	Capacity     int    `xml:" CAPACITY,omitempty"`
	ActualOutput int    `xml:" ACTUALOUTPUT,omitempty"`
}

// HpTemp contains the thermal data of a chassis or blade
type HpTemp struct {
	C    int    `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty"`
}

// HpRimpBlade is the entry data structure for the blade when queries directly
type HpRimpBlade struct {
	HpMP         *HpMP         `xml:" MP,omitempty"`
	HpHSI        *HpHSI        `xml:" HSI,omitempty"`
	HpBladeBlade *HpBladeBlade `xml:" BLADESYSTEM,omitempty"`
}

// HpBladeBlade blade information from the hprimp of blades
type HpBladeBlade struct {
	Bay int `xml:" BAY,omitempty"`
}

// HpHSI contains the information about the components of the blade
type HpHSI struct {
	HpNICS *HpNICS `xml:" NICS,omitempty"`
	Sbsn   string  `xml:" SBSN,omitempty" json:"SBSN,omitempty"`
	Spn    string  `xml:" SPN,omitempty" json:"SPN,omitempty"`
}

// HpNICS contains the list of nics that a blade has
type HpNICS struct {
	HpNIC []*HpNIC `xml:" NIC,omitempty"`
}

// HpNIC contains the nic information of a blade
type HpNIC struct {
	Description string `xml:" DESCRIPTION,omitempty"`
	MacAddr     string `xml:" MACADDR,omitempty"`
	Status      string `xml:" STATUS,omitempty"`
}

// HpFirmware is the struct used to render the data from https://$ip/json/fw_info, it contains firmware data of the blade
type HpFirmware struct {
	Firmware []struct {
		FwName    string `json:"fw_name"`
		FwVersion string `json:"fw_version"`
	} `json:"firmware"`
}

// HpProcs is the struct used to render the data from https://$ip/json/proc_info, it contains the processor data
type HpProcs struct {
	Processors []struct {
		ProcName       string `json:"proc_name"`
		ProcNumCores   int    `json:"proc_num_cores"`
		ProcNumThreads int    `json:"proc_num_threads"`
	} `json:"processors"`
}

// HpMem is the struct used to render the data from https://$ip/json/mem_info, it contains the ram data
type HpMem struct {
	MemTotalMemSize int          `json:"mem_total_mem_size"`
	Memory          []*HpMemSlot `json:"memory"`
}

// HpMemSlot is part of the payload returned from https://$ip/json/mem_info
type HpMemSlot struct {
	MemDevLoc string `json:"mem_dev_loc"`
	MemSize   int    `json:"mem_size"`
	MemSpeed  int    `json:"mem_speed"`
}

// HpOverview is the struct used to render the data from https://$ip/json/overview, it contains information about bios version, ilo license and a bit more
type HpOverview struct {
	ServerName    string `json:"server_name"`
	ProductName   string `json:"product_name"`
	SerialNum     string `json:"serial_num"`
	SystemRom     string `json:"system_rom"`
	SystemRomDate string `json:"system_rom_date"`
	BackupRomDate string `json:"backup_rom_date"`
	License       string `json:"license"`
	IloFwVersion  string `json:"ilo_fw_version"`
	IPAddress     string `json:"ip_address"`
	SystemHealth  string `json:"system_health"`
	Power         string `json:"power"`
}

// HpPowerSummary is the struct used to render the data from https://$ip/json/power_summary, it contains the basic information about the power usage of the machine
type HpPowerSummary struct {
	HostpwrState          string `json:"hostpwr_state"`
	PowerSupplyInputPower int    `json:"power_supply_input_power"`
}

// HpHelthTemperature is the struct used to render the data from https://$ip/json/health_temperature, it contains the information about the thermal status of the machine
type HpHelthTemperature struct {
	HostpwrState string           `json:"hostpwr_state"`
	InPost       int              `json:"in_post"`
	Temperature  []*HpTemperature `json:"temperature"`
}

// HpTemperature is part of the data rendered from https://$ip/json/health_temperature, it contains the names of each component and their current temp
type HpTemperature struct {
	Label          string `json:"label"`
	Location       string `json:"location"`
	Status         string `json:"status"`
	Currentreading int    `json:"currentreading"`
	TempUnit       string `json:"temp_unit"`
}

// HpIloLicense is the struct used to render the data from https://$ip/json/license, it contains the license information of the ilo
type HpIloLicense struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
