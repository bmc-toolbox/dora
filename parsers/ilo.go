package parsers

// Automatically generated with chidley

type AC struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ACINPUT struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ACTUALOUTPUT struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ADDR struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ASSET struct {
}

type BAY struct {
	Attr_NAME  string      `xml:" NAME,attr"  json:",omitempty"`
	CONNECTION *CONNECTION `xml:" CONNECTION,omitempty" json:"CONNECTION,omitempty"`
	MmDepth    *MmDepth    `xml:" mmDepth,omitempty" json:"mmDepth,omitempty"`
	MmHeight   *MmHeight   `xml:" mmHeight,omitempty" json:"mmHeight,omitempty"`
	MmWidth    *MmWidth    `xml:" mmWidth,omitempty" json:"mmWidth,omitempty"`
	MmXOffset  *MmXOffset  `xml:" mmXOffset,omitempty" json:"mmXOffset,omitempty"`
	MmYOffset  *MmYOffset  `xml:" mmYOffset,omitempty" json:"mmYOffset,omitempty"`
	SIDE       *SIDE       `xml:" SIDE,omitempty" json:"SIDE,omitempty"`
}

type BAYS struct {
	BAY []*BAY `xml:" BAY,omitempty" json:"BAY,omitempty"`
}

type BLADE struct {
	BAY             *BAY             `xml:" BAY,omitempty" json:"BAY,omitempty"`
	BLADEROMVER     *BLADEROMVER     `xml:" BLADEROMVER,omitempty" json:"BLADEROMVER,omitempty"`
	BSN             *BSN             `xml:" BSN,omitempty" json:"BSN,omitempty"`
	CONJOINABLE     *CONJOINABLE     `xml:" CONJOINABLE,omitempty" json:"CONJOINABLE,omitempty"`
	CUUID           *CUUID           `xml:" cUUID,omitempty" json:"cUUID,omitempty"`
	DIAG            *DIAG            `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	MANUFACTURER    *MANUFACTURER    `xml:" MANUFACTURER,omitempty" json:"MANUFACTURER,omitempty"`
	MGMTDNSNAME     *MGMTDNSNAME     `xml:" MGMTDNSNAME,omitempty" json:"MGMTDNSNAME,omitempty"`
	MGMTFWVERSION   *MGMTFWVERSION   `xml:" MGMTFWVERSION,omitempty" json:"MGMTFWVERSION,omitempty"`
	MGMTIPADDR      *MGMTIPADDR      `xml:" MGMTIPADDR,omitempty" json:"MGMTIPADDR,omitempty"`
	MGMTIPV6ADDR_LL *MGMTIPV6ADDR_LL `xml:" MGMTIPV6ADDR_LL,omitempty" json:"MGMTIPV6ADDR_LL,omitempty"`
	MGMTPN          *MGMTPN          `xml:" MGMTPN,omitempty" json:"MGMTPN,omitempty"`
	NAME            *NAME            `xml:" NAME,omitempty" json:"NAME,omitempty"`
	PN              *PN              `xml:" PN,omitempty" json:"PN,omitempty"`
	PORTMAP         *PORTMAP         `xml:" PORTMAP,omitempty" json:"PORTMAP,omitempty"`
	POWER           *POWER           `xml:" POWER,omitempty" json:"POWER,omitempty"`
	PWRM            *PWRM            `xml:" PWRM,omitempty" json:"PWRM,omitempty"`
	SPN             *SPN             `xml:" SPN,omitempty" json:"SPN,omitempty"`
	STATUS          *STATUS          `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TEMPS           *TEMPS           `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
	TYPE            *TYPE            `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
	UIDSTATUS       *UIDSTATUS       `xml:" UIDSTATUS,omitempty" json:"UIDSTATUS,omitempty"`
	UUID            *UUID            `xml:" UUID,omitempty" json:"UUID,omitempty"`
	VIRTUAL         *VIRTUAL         `xml:" VIRTUAL,omitempty" json:"VIRTUAL,omitempty"`
	VLAN            *VLAN            `xml:" VLAN,omitempty" json:"VLAN,omitempty"`
	VMSTAT          *VMSTAT          `xml:" VMSTAT,omitempty" json:"VMSTAT,omitempty"`
}

type BLADEBAYNUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type BLADEMEZZNUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type BLADEMEZZPORTNUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type BLADEROMVER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type BLADES struct {
	BAYS  *BAYS    `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	BLADE []*BLADE `xml:" BLADE,omitempty" json:"BLADE,omitempty"`
}

type BSN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type BUTTON_LOCK_ENABLED struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type C struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type CAPACITY struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type CDROMSTAT struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type CDROMURL struct {
}

type CONJOINABLE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type CONNECTION struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Cooling struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type DATETIME struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type DESC struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type DEVICE struct {
	NAME   *NAME   `xml:" NAME,omitempty" json:"NAME,omitempty"`
	PORT   []*PORT `xml:" PORT,omitempty" json:"PORT,omitempty"`
	STATUS *STATUS `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TYPE   *TYPE   `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
}

type DIAG struct {
	AC             *AC             `xml:" AC,omitempty" json:"AC,omitempty"`
	Cooling        *Cooling        `xml:" Cooling,omitempty" json:"Cooling,omitempty"`
	Degraded       *Degraded       `xml:" Degraded,omitempty" json:"Degraded,omitempty"`
	FRU            *FRU            `xml:" FRU,omitempty" json:"FRU,omitempty"`
	Failure        *Failure        `xml:" Failure,omitempty" json:"Failure,omitempty"`
	I2c            *I2c            `xml:" i2c,omitempty" json:"i2c,omitempty"`
	Keying         *Keying         `xml:" Keying,omitempty" json:"Keying,omitempty"`
	Location       *Location       `xml:" Location,omitempty" json:"Location,omitempty"`
	MgmtProc       *MgmtProc       `xml:" MgmtProc,omitempty" json:"MgmtProc,omitempty"`
	OaRedundancy   *OaRedundancy   `xml:" oaRedundancy,omitempty" json:"oaRedundancy,omitempty"`
	Power          *Power          `xml:" Power,omitempty" json:"Power,omitempty"`
	ThermalDanger  *ThermalDanger  `xml:" thermalDanger,omitempty" json:"thermalDanger,omitempty"`
	ThermalWarning *ThermalWarning `xml:" thermalWarning,omitempty" json:"thermalWarning,omitempty"`
}

type DIM struct {
	MmDepth  *MmDepth  `xml:" mmDepth,omitempty" json:"mmDepth,omitempty"`
	MmHeight *MmHeight `xml:" mmHeight,omitempty" json:"mmHeight,omitempty"`
	MmWidth  *MmWidth  `xml:" mmWidth,omitempty" json:"mmWidth,omitempty"`
}

type DVDDRIVE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type DYNAMICPOWERSAVER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Degraded struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ENABLED struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ENCL struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ENCL_SN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type FABRICTYPE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type FAN struct {
	BAY         *BAY         `xml:" BAY,omitempty" json:"BAY,omitempty"`
	PN          *PN          `xml:" PN,omitempty" json:"PN,omitempty"`
	PRODUCTNAME *PRODUCTNAME `xml:" PRODUCTNAME,omitempty" json:"PRODUCTNAME,omitempty"`
	PWR_USED    *PWR_USED    `xml:" PWR_USED,omitempty" json:"PWR_USED,omitempty"`
	RPM_CUR     *RPM_CUR     `xml:" RPM_CUR,omitempty" json:"RPM_CUR,omitempty"`
	RPM_MAX     *RPM_MAX     `xml:" RPM_MAX,omitempty" json:"RPM_MAX,omitempty"`
	RPM_MIN     *RPM_MIN     `xml:" RPM_MIN,omitempty" json:"RPM_MIN,omitempty"`
	STATUS      *STATUS      `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type FANS struct {
	BAYS        *BAYS        `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	FAN         []*FAN       `xml:" FAN,omitempty" json:"FAN,omitempty"`
	NEEDED_FANS *NEEDED_FANS `xml:" NEEDED_FANS,omitempty" json:"NEEDED_FANS,omitempty"`
	REDUNDANCY  *REDUNDANCY  `xml:" REDUNDANCY,omitempty" json:"REDUNDANCY,omitempty"`
	STATUS      *STATUS      `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	WANTED_FANS *WANTED_FANS `xml:" WANTED_FANS,omitempty" json:"WANTED_FANS,omitempty"`
}

type FLOPPYSTAT struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type FLOPPYURL struct {
}

type FRU struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type FUNCTION struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type FWRI struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Failure struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type GUID struct {
	FUNCTION    *FUNCTION    `xml:" FUNCTION,omitempty" json:"FUNCTION,omitempty"`
	GUID_STRING *GUID_STRING `xml:" GUID_STRING,omitempty" json:"GUID_STRING,omitempty"`
	TYPE        *TYPE        `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
}

type GUIDS struct {
	GUID []*GUID `xml:" GUID,omitempty" json:"GUID,omitempty"`
}

type GUID_STRING struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type IMAGE_URL struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type INFRA2 struct {
	ADDR        *ADDR        `xml:" ADDR,omitempty" json:"ADDR,omitempty"`
	ASSET       *ASSET       `xml:" ASSET,omitempty" json:"ASSET,omitempty"`
	BLADES      *BLADES      `xml:" BLADES,omitempty" json:"BLADES,omitempty"`
	DATETIME    *DATETIME    `xml:" DATETIME,omitempty" json:"DATETIME,omitempty"`
	DIAG        *DIAG        `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	DIM         *DIM         `xml:" DIM,omitempty" json:"DIM,omitempty"`
	ENCL        *ENCL        `xml:" ENCL,omitempty" json:"ENCL,omitempty"`
	ENCL_SN     *ENCL_SN     `xml:" ENCL_SN,omitempty" json:"ENCL_SN,omitempty"`
	FANS        *FANS        `xml:" FANS,omitempty" json:"FANS,omitempty"`
	LCDS        *LCDS        `xml:" LCDS,omitempty" json:"LCDS,omitempty"`
	MANAGERS    *MANAGERS    `xml:" MANAGERS,omitempty" json:"MANAGERS,omitempty"`
	PART        *PART        `xml:" PART,omitempty" json:"PART,omitempty"`
	PN          *PN          `xml:" PN,omitempty" json:"PN,omitempty"`
	POWER       *POWER       `xml:" POWER,omitempty" json:"POWER,omitempty"`
	RACK        *RACK        `xml:" RACK,omitempty" json:"RACK,omitempty"`
	SOLUTIONSID *SOLUTIONSID `xml:" SOLUTIONSID,omitempty" json:"SOLUTIONSID,omitempty"`
	STATUS      *STATUS      `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	SWITCHES    *SWITCHES    `xml:" SWITCHES,omitempty" json:"SWITCHES,omitempty"`
	TEMPS       *TEMPS       `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
	TIMEZONE    *TIMEZONE    `xml:" TIMEZONE,omitempty" json:"TIMEZONE,omitempty"`
	UIDSTATUS   *UIDSTATUS   `xml:" UIDSTATUS,omitempty" json:"UIDSTATUS,omitempty"`
	UUID        *UUID        `xml:" UUID,omitempty" json:"UUID,omitempty"`
	VCM         *VCM         `xml:" VCM,omitempty" json:"VCM,omitempty"`
	VM          *VM          `xml:" VM,omitempty" json:"VM,omitempty"`
}

type IPV6STATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Keying struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type LCD struct {
	BAY                 []*BAY               `xml:" BAY,omitempty" json:"BAY,omitempty"`
	BUTTON_LOCK_ENABLED *BUTTON_LOCK_ENABLED `xml:" BUTTON_LOCK_ENABLED,omitempty" json:"BUTTON_LOCK_ENABLED,omitempty"`
	DIAG                *DIAG                `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	FWRI                *FWRI                `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
	IMAGE_URL           *IMAGE_URL           `xml:" IMAGE_URL,omitempty" json:"IMAGE_URL,omitempty"`
	MANUFACTURER        *MANUFACTURER        `xml:" MANUFACTURER,omitempty" json:"MANUFACTURER,omitempty"`
	PIN_ENABLED         *PIN_ENABLED         `xml:" PIN_ENABLED,omitempty" json:"PIN_ENABLED,omitempty"`
	PN                  *PN                  `xml:" PN,omitempty" json:"PN,omitempty"`
	SPN                 *SPN                 `xml:" SPN,omitempty" json:"SPN,omitempty"`
	STATUS              *STATUS              `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	USERNOTES           *USERNOTES           `xml:" USERNOTES,omitempty" json:"USERNOTES,omitempty"`
}

type LCDS struct {
	BAYS *BAYS `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	LCD  *LCD  `xml:" LCD,omitempty" json:"LCD,omitempty"`
}

type LINK_LED_STATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type LOCATION struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Location struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MACADDR struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MANAGER struct {
	BAY          []*BAY        `xml:" BAY,omitempty" json:"BAY,omitempty"`
	BSN          *BSN          `xml:" BSN,omitempty" json:"BSN,omitempty"`
	DIAG         *DIAG         `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	FWRI         *FWRI         `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
	IPV6STATUS   *IPV6STATUS   `xml:" IPV6STATUS,omitempty" json:"IPV6STATUS,omitempty"`
	MACADDR      *MACADDR      `xml:" MACADDR,omitempty" json:"MACADDR,omitempty"`
	MANUFACTURER *MANUFACTURER `xml:" MANUFACTURER,omitempty" json:"MANUFACTURER,omitempty"`
	MGMTIPADDR   *MGMTIPADDR   `xml:" MGMTIPADDR,omitempty" json:"MGMTIPADDR,omitempty"`
	NAME         *NAME         `xml:" NAME,omitempty" json:"NAME,omitempty"`
	POWER        *POWER        `xml:" POWER,omitempty" json:"POWER,omitempty"`
	ROLE         *ROLE         `xml:" ROLE,omitempty" json:"ROLE,omitempty"`
	SPN          *SPN          `xml:" SPN,omitempty" json:"SPN,omitempty"`
	STATUS       *STATUS       `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TEMPS        *TEMPS        `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
	UIDSTATUS    *UIDSTATUS    `xml:" UIDSTATUS,omitempty" json:"UIDSTATUS,omitempty"`
	UUID         *UUID         `xml:" UUID,omitempty" json:"UUID,omitempty"`
	WIZARDSTATUS *WIZARDSTATUS `xml:" WIZARDSTATUS,omitempty" json:"WIZARDSTATUS,omitempty"`
	YOUAREHERE   *YOUAREHERE   `xml:" YOUAREHERE,omitempty" json:"YOUAREHERE,omitempty"`
}

type MANAGERS struct {
	BAYS    *BAYS      `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	MANAGER []*MANAGER `xml:" MANAGER,omitempty" json:"MANAGER,omitempty"`
}

type MANUFACTURER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MEZZ struct {
	DEVICE *DEVICE `xml:" DEVICE,omitempty" json:"DEVICE,omitempty"`
	NUMBER *NUMBER `xml:" NUMBER,omitempty" json:"NUMBER,omitempty"`
	SLOT   *SLOT   `xml:" SLOT,omitempty" json:"SLOT,omitempty"`
}

type MGMTDNSNAME struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MGMTFWVERSION struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MGMTIPADDR struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MGMTIPV6ADDR_LL struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MGMTPN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MGMTURL struct {
}

type MgmtProc struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type NAME struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type NEEDED_FANS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type NEEDED_PS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type NUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type OUTPUT_POWER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PART struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PASSTHRU_MODE_ENABLED struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PDU struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PIN_ENABLED struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PORT struct {
	BLADEBAYNUMBER      *BLADEBAYNUMBER      `xml:" BLADEBAYNUMBER,omitempty" json:"BLADEBAYNUMBER,omitempty"`
	BLADEMEZZNUMBER     *BLADEMEZZNUMBER     `xml:" BLADEMEZZNUMBER,omitempty" json:"BLADEMEZZNUMBER,omitempty"`
	BLADEMEZZPORTNUMBER *BLADEMEZZPORTNUMBER `xml:" BLADEMEZZPORTNUMBER,omitempty" json:"BLADEMEZZPORTNUMBER,omitempty"`
	ENABLED             *ENABLED             `xml:" ENABLED,omitempty" json:"ENABLED,omitempty"`
	GUIDS               *GUIDS               `xml:" GUIDS,omitempty" json:"GUIDS,omitempty"`
	LINK_LED_STATUS     *LINK_LED_STATUS     `xml:" LINK_LED_STATUS,omitempty" json:"LINK_LED_STATUS,omitempty"`
	NUMBER              *NUMBER              `xml:" NUMBER,omitempty" json:"NUMBER,omitempty"`
	STATUS              *STATUS              `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TRAYBAYNUMBER       *TRAYBAYNUMBER       `xml:" TRAYBAYNUMBER,omitempty" json:"TRAYBAYNUMBER,omitempty"`
	TRAYPORTNUMBER      *TRAYPORTNUMBER      `xml:" TRAYPORTNUMBER,omitempty" json:"TRAYPORTNUMBER,omitempty"`
	TYPE                *TYPE                `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
	UID_STATUS          *UID_STATUS          `xml:" UID_STATUS,omitempty" json:"UID_STATUS,omitempty"`
	WWPN                *WWPN                `xml:" WWPN,omitempty" json:"WWPN,omitempty"`
}

type PORTMAP struct {
	MEZZ                  []*MEZZ                `xml:" MEZZ,omitempty" json:"MEZZ,omitempty"`
	PASSTHRU_MODE_ENABLED *PASSTHRU_MODE_ENABLED `xml:" PASSTHRU_MODE_ENABLED,omitempty" json:"PASSTHRU_MODE_ENABLED,omitempty"`
	SLOT                  *SLOT                  `xml:" SLOT,omitempty" json:"SLOT,omitempty"`
	STATUS                *STATUS                `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type POWER struct {
	BAYS               *BAYS               `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	CAPACITY           *CAPACITY           `xml:" CAPACITY,omitempty" json:"CAPACITY,omitempty"`
	DYNAMICPOWERSAVER  *DYNAMICPOWERSAVER  `xml:" DYNAMICPOWERSAVER,omitempty" json:"DYNAMICPOWERSAVER,omitempty"`
	NEEDED_PS          *NEEDED_PS          `xml:" NEEDED_PS,omitempty" json:"NEEDED_PS,omitempty"`
	OUTPUT_POWER       *OUTPUT_POWER       `xml:" OUTPUT_POWER,omitempty" json:"OUTPUT_POWER,omitempty"`
	PDU                *PDU                `xml:" PDU,omitempty" json:"PDU,omitempty"`
	POWERMODE          *POWERMODE          `xml:" POWERMODE,omitempty" json:"POWERMODE,omitempty"`
	POWERONFLAG        *POWERONFLAG        `xml:" POWERONFLAG,omitempty" json:"POWERONFLAG,omitempty"`
	POWERSTATE         *POWERSTATE         `xml:" POWERSTATE,omitempty" json:"POWERSTATE,omitempty"`
	POWERSUPPLY        []*POWERSUPPLY      `xml:" POWERSUPPLY,omitempty" json:"POWERSUPPLY,omitempty"`
	POWER_CONSUMED     *POWER_CONSUMED     `xml:" POWER_CONSUMED,omitempty" json:"POWER_CONSUMED,omitempty"`
	POWER_OFF_WATTAGE  *POWER_OFF_WATTAGE  `xml:" POWER_OFF_WATTAGE,omitempty" json:"POWER_OFF_WATTAGE,omitempty"`
	POWER_ON_WATTAGE   *POWER_ON_WATTAGE   `xml:" POWER_ON_WATTAGE,omitempty" json:"POWER_ON_WATTAGE,omitempty"`
	REDUNDANCY         *REDUNDANCY         `xml:" REDUNDANCY,omitempty" json:"REDUNDANCY,omitempty"`
	REDUNDANCYMODE     *REDUNDANCYMODE     `xml:" REDUNDANCYMODE,omitempty" json:"REDUNDANCYMODE,omitempty"`
	REDUNDANT_CAPACITY *REDUNDANT_CAPACITY `xml:" REDUNDANT_CAPACITY,omitempty" json:"REDUNDANT_CAPACITY,omitempty"`
	STATUS             *STATUS             `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TYPE               *TYPE               `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
	WANTED_PS          *WANTED_PS          `xml:" WANTED_PS,omitempty" json:"WANTED_PS,omitempty"`
}

type POWERMODE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type POWERONFLAG struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type POWERSTATE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type POWERSUPPLY struct {
	ACINPUT      *ACINPUT      `xml:" ACINPUT,omitempty" json:"ACINPUT,omitempty"`
	ACTUALOUTPUT *ACTUALOUTPUT `xml:" ACTUALOUTPUT,omitempty" json:"ACTUALOUTPUT,omitempty"`
	BAY          []*BAY        `xml:" BAY,omitempty" json:"BAY,omitempty"`
	CAPACITY     *CAPACITY     `xml:" CAPACITY,omitempty" json:"CAPACITY,omitempty"`
	DIAG         *DIAG         `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	FWRI         *FWRI         `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
	PN           *PN           `xml:" PN,omitempty" json:"PN,omitempty"`
	SN           *SN           `xml:" SN,omitempty" json:"SN,omitempty"`
	STATUS       *STATUS       `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type POWER_CONSUMED struct {
	Text float64 `xml:",chardata" json:",omitempty"`
}

type POWER_OFF_WATTAGE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type POWER_ON_WATTAGE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PRODUCTNAME struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PWRM struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type PWR_USED struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Power struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type RACK struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type REDUNDANCY struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type REDUNDANCYMODE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type REDUNDANT_CAPACITY struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type RIMP struct {
	INFRA2 *INFRA2 `xml:" INFRA2,omitempty" json:"INFRA2,omitempty"`
}

type ROLE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type RPM_CUR struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type RPM_MAX struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type RPM_MIN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SIDE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SLOT struct {
	NUMBER *NUMBER `xml:" NUMBER,omitempty" json:"NUMBER,omitempty"`
	PORT   []*PORT `xml:" PORT,omitempty" json:"PORT,omitempty"`
	TYPE   *TYPE   `xml:" TYPE,omitempty" json:"TYPE,omitempty"`
}

type SN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SOLUTIONSID struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SPN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type STATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SUPPORT struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type SWITCH struct {
	BAY          []*BAY        `xml:" BAY,omitempty" json:"BAY,omitempty"`
	BSN          *BSN          `xml:" BSN,omitempty" json:"BSN,omitempty"`
	DIAG         *DIAG         `xml:" DIAG,omitempty" json:"DIAG,omitempty"`
	FABRICTYPE   *FABRICTYPE   `xml:" FABRICTYPE,omitempty" json:"FABRICTYPE,omitempty"`
	FWRI         *FWRI         `xml:" FWRI,omitempty" json:"FWRI,omitempty"`
	MANUFACTURER *MANUFACTURER `xml:" MANUFACTURER,omitempty" json:"MANUFACTURER,omitempty"`
	MGMTIPADDR   *MGMTIPADDR   `xml:" MGMTIPADDR,omitempty" json:"MGMTIPADDR,omitempty"`
	MGMTURL      *MGMTURL      `xml:" MGMTURL,omitempty" json:"MGMTURL,omitempty"`
	PN           *PN           `xml:" PN,omitempty" json:"PN,omitempty"`
	PORTMAP      *PORTMAP      `xml:" PORTMAP,omitempty" json:"PORTMAP,omitempty"`
	POWER        *POWER        `xml:" POWER,omitempty" json:"POWER,omitempty"`
	SPN          *SPN          `xml:" SPN,omitempty" json:"SPN,omitempty"`
	STATUS       *STATUS       `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
	TEMPS        *TEMPS        `xml:" TEMPS,omitempty" json:"TEMPS,omitempty"`
	THERMAL      *THERMAL      `xml:" THERMAL,omitempty" json:"THERMAL,omitempty"`
	UIDSTATUS    *UIDSTATUS    `xml:" UIDSTATUS,omitempty" json:"UIDSTATUS,omitempty"`
}

type SWITCHES struct {
	BAYS   *BAYS     `xml:" BAYS,omitempty" json:"BAYS,omitempty"`
	SWITCH []*SWITCH `xml:" SWITCH,omitempty" json:"SWITCH,omitempty"`
}

type TEMP struct {
	C         *C           `xml:" C,omitempty" json:"C,omitempty"`
	DESC      *DESC        `xml:" DESC,omitempty" json:"DESC,omitempty"`
	LOCATION  *LOCATION    `xml:" LOCATION,omitempty" json:"LOCATION,omitempty"`
	THRESHOLD []*THRESHOLD `xml:" THRESHOLD,omitempty" json:"THRESHOLD,omitempty"`
}

type TEMPS struct {
	TEMP *TEMP `xml:" TEMP,omitempty" json:"TEMP,omitempty"`
}

type THERMAL struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type THRESHOLD struct {
	C      *C      `xml:" C,omitempty" json:"C,omitempty"`
	DESC   *DESC   `xml:" DESC,omitempty" json:"DESC,omitempty"`
	STATUS *STATUS `xml:" STATUS,omitempty" json:"STATUS,omitempty"`
}

type TIMEZONE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type TRAYBAYNUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type TRAYPORTNUMBER struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type TYPE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type UIDSTATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type UID_STATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type USERNOTES struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type UUID struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type VCM struct {
	VcmDomainId   *VcmDomainId   `xml:" vcmDomainId,omitempty" json:"vcmDomainId,omitempty"`
	VcmDomainName *VcmDomainName `xml:" vcmDomainName,omitempty" json:"vcmDomainName,omitempty"`
	VcmMode       *VcmMode       `xml:" vcmMode,omitempty" json:"vcmMode,omitempty"`
	VcmUrl        *VcmUrl        `xml:" vcmUrl,omitempty" json:"vcmUrl,omitempty"`
}

type VID struct {
	BSN   *BSN   `xml:" BSN,omitempty" json:"BSN,omitempty"`
	CUUID *CUUID `xml:" cUUID,omitempty" json:"cUUID,omitempty"`
}

type VIRTUAL struct {
	VID *VID `xml:" VID,omitempty" json:"VID,omitempty"`
}

type VLAN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type VM struct {
	DVDDRIVE *DVDDRIVE `xml:" DVDDRIVE,omitempty" json:"DVDDRIVE,omitempty"`
}

type VMSTAT struct {
	CDROMSTAT  *CDROMSTAT  `xml:" CDROMSTAT,omitempty" json:"CDROMSTAT,omitempty"`
	CDROMURL   *CDROMURL   `xml:" CDROMURL,omitempty" json:"CDROMURL,omitempty"`
	FLOPPYSTAT *FLOPPYSTAT `xml:" FLOPPYSTAT,omitempty" json:"FLOPPYSTAT,omitempty"`
	FLOPPYURL  *FLOPPYURL  `xml:" FLOPPYURL,omitempty" json:"FLOPPYURL,omitempty"`
	SUPPORT    *SUPPORT    `xml:" SUPPORT,omitempty" json:"SUPPORT,omitempty"`
}

type WANTED_FANS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type WANTED_PS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type WIZARDSTATUS struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type WWPN struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type YOUAREHERE struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type CUUID struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type I2c struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MmDepth struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MmHeight struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MmWidth struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MmXOffset struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type MmYOffset struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type OaRedundancy struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type Root struct {
	RIMP *RIMP `xml:" RIMP,omitempty" json:"RIMP,omitempty"`
}

type ThermalDanger struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type ThermalWarning struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type VcmDomainId struct {
}

type VcmDomainName struct {
}

type VcmMode struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type VcmUrl struct {
	Text string `xml:",chardata" json:",omitempty"`
}
