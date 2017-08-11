package connectors

// DellCMC is the entry of the json exposed by dell
// We don't need to use an maps[string] with DellChassis, because we don't have clusters
type DellCMC struct {
	DellChassis *DellChassis `json:"0"`
}

// DellChassis groups all the interresting stuff we will ready from the chassis
type DellChassis struct {
	DellChassisGroupMemberHealthBlob *DellChassisGroupMemberHealthBlob `json:"ChassisGroupMemberHealthBlob"`
}

// DellChassisGroupMemberHealthBlob has a collection of metrics from the chassis, psu and blades
type DellChassisGroupMemberHealthBlob struct {
	DellBlades        map[string]*DellBlade `json:"blades_status"`
	DellPsuStatus     *DellPsuStatus        `json:"psu_status"`
	DellChassisStatus *DellChassisStatus    `json:"chassis_status"`
	DellCMCStatus     *DellCMCStatus        `json:"cmc_status"`
	// TODO: active_alerts
}

// DellChassisStatus expose the basic information that identify the chassis
type DellChassisStatus struct {
	ROCmcFwVersionString string `json:"RO_cmc_fw_version_string"`
	ROChassisServiceTag  string `json:"RO_chassis_service_tag"`
	ROChassisProductname string `json:"RO_chassis_productname"`
	CHASSISName          string `json:"CHASSIS_name"`
}

// DellCMCStatus brings the information about the cmc status itself we will use it to know if the chassis has errors
type DellCMCStatus struct {
	CMCActiveError string `json:"cmcActiveError"`
}

// DellNic is the nic we have on a servers
type DellNic struct {
	BladeNicName string `json:"bladeNicName"`
	BladeNicVer  string `json:"bladeNicVer"`
}

// DellBlade contains all the blade information
type DellBlade struct {
	BladeTemperature    string              `json:"bladeTemperature"`
	BladePresent        int                 `json:"bladePresent"`
	IdracURL            string              `json:"idracURL"`
	BladeLogDescription string              `json:"bladeLogDescription"`
	StorageNumDrives    int                 `json:"storageNumDrives"`
	Nics                map[string]*DellNic `json:"nic"`
	BladeMasterSlot     int                 `json:"bladeMasterSlot"`
	BladeUSCVer         string              `json:"bladeUSCVer"`
	BladeSvcTag         string              `json:"bladeSvcTag"`
	BladeBIOSver        string              `json:"bladeBIOSver"`
	ActualPwrConsump    int                 `json:"actualPwrConsump"`
	IsStorageBlade      int                 `json:"isStorageBlade"`
	BladeName           string              `json:"bladeName"`
	BladeSerialNum      string              `json:"bladeSerialNum"`
}

// DellPsuStatus contains the information and power usage of the pdus
type DellPsuStatus struct {
	AcPower  string `json:"acPower"`
	PsuCount string `json:"psuCount"`
}
