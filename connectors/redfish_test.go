package connectors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

type rFTest struct {
	Entry map[string][]struct {
		Vendor            string
		Endpoint          string
		ServerPayload     []byte
		ExpectedAnswer    string
		DetectionEndpoint string
		DetectionString   string
	}
}

func setup(vendor string, redfishendpoint string, detectionString string) (r *RedFishReader, err error) {
	viper.SetDefault("debug", false)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")

	mux.HandleFunc(fmt.Sprintf("/%s", redfishVendorEndPoints[vendor][redfishendpoint]), func(w http.ResponseWriter, r *http.Request) {
		w.Write(redfishVendorAnswers[vendor][redfishendpoint])
	})

	mux.HandleFunc("/redfish/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(detectionString))
	})

	r, err = NewRedFishReader(&ip, &username, &password)
	if err != nil {
		return r, err
	}

	return r, err
}

func teardown() {
	server.Close()
}

var (
	username = "super"
	password = "test"

	rft                  rFTest
	mux                  *http.ServeMux
	server               *httptest.Server
	redfishVendorAnswers = map[string]map[string][]byte{
		Dell: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#ComputerSystem.ComputerSystem","@odata.id":"/redfish/v1/Systems/System.Embedded.1","@odata.type":"#ComputerSystem.v1_0_2.ComputerSystem","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["On","ForceOff","GracefulRestart","PushPowerButton","Nmi"],"target":"/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"}},"AssetTag":"","BiosVersion":"2.4.2","Boot":{"BootSourceOverrideEnabled":"Once","BootSourceOverrideTarget":"None","BootSourceOverrideTarget@Redfish.AllowableValues":["None","Pxe","Cd","Floppy","Hdd","BiosSetup","Utilities","UefiTarget","SDCard"],"UefiTargetBootSourceOverride":""},"Description":"Computer System which represents a machine (physical or virtual) and the local resources such as memory, cpu and other devices that can be accessed from that machine.","EthernetInterfaces":{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces"},"HostName":"machine.example.com","Id":"System.Embedded.1","IndicatorLED":"Off","Links":{"Chassis":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"Chassis@odata.count":1,"CooledBy":[],"CooledBy@odata.count":0,"ManagedBy":[{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1"}],"ManagedBy@odata.count":1,"PoweredBy":[],"PoweredBy@odata.count":0},"Manufacturer":"Dell Inc.","MemorySummary":{"Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"TotalSystemMemoryGiB":128.0},"Model":"PowerEdge M630","Name":"System","PartNumber":"0PHY8DA03","PowerState":"On","ProcessorSummary":{"Count":2,"Model":"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"}},"Processors":{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors"},"SKU":"905DCC2","SerialNumber":"CN7016362D00JD","SimpleStorage":{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Storage/Controllers"},"Status":{"Health":"OK","HealthRollUp":"OK","State":"Enabled"},"SystemType":"Physical","UUID":"4c4c4544-0030-3510-8044-b9c04f434332"}`),
			RFPower:      []byte(``),
			RFThermal:    []byte(``),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Processor.Processor","@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1","@odata.type":"#Processor.v1_0_2.Processor","Description":"Represents the properties of a Processor attached to this System","Id":"CPU.Socket.1","InstructionSet":[{"Member":"x86-64"}],"Manufacturer":"Intel","MaxSpeedMHz":4000,"Model":"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","Name":"CPU 1","ProcessorArchitecture":[{"Member":"x86"}],"ProcessorId":{"EffectiveFamily":"6","EffectiveModel":"63","IdentificationRegisters":"0x000306F2","MicrocodeInfo":"0x39","Step":"2","VendorID":"GenuineIntel"},"ProcessorType":"CPU","Socket":"CPU.Socket.1","Status":{"Health":"OK","State":"Enabled"},"TotalCores":6,"TotalThreads":12}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Manager.Manager","@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1","@odata.type":"#Manager.v1_0_2.Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["GracefulRestart"],"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset"},"Oem":{"OemManager.v1_0_0#OemManager.ExportSystemConfiguration":{"ExportFormat@Redfish.AllowableValues":["XML"],"ExportUse@Redfish.AllowableValues":["Default","Clone","Replace"],"IncludeInExport@Redfish.AllowableValues":["Default","IncludeReadOnly","IncludePasswordHashValues"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL","IDRAC","BIOS","NIC","RAID"]},"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration"},"OemManager.v1_0_0#OemManager.ImportSystemConfiguration":{"HostPowerState@Redfish.AllowableValues":["On","Off"],"ImportSystemConfiguration@Redfish.AllowableValues":["TimeToWait","ImportBuffer"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL","IDRAC","BIOS","NIC","RAID"]},"ShutdownType@Redfish.AllowableValues":["Graceful","Forced","NoReboot"],"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"},"OemManager.v1_0_0#OemManager.ImportSystemConfigurationPreview":{"ImportSystemConfigurationPreview@Redfish.AllowableValues":["ImportBuffer"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL"]},"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview"}}},"CommandShell":{"ConnectTypesSupported":["SSH","Telnet","IPMI"],"ConnectTypesSupported@odata.count":3,"MaxConcurrentSessions":5,"ServiceEnabled":true},"DateTime":"2017-09-07T16:48:14-05:00","DateTimeLocalOffset":"-05:00","Description":"BMC","EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces"},"FirmwareVersion":"2.41.40.40","GraphicalConsole":{"ConnectTypesSupported":["KVMIP"],"ConnectTypesSupported@odata.count":1,"MaxConcurrentSessions":6,"ServiceEnabled":true},"Id":"iDRAC.Embedded.1","Links":{"ManagerForChassis":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"ManagerForChassis@odata.count":1,"ManagerForServers":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1"}],"ManagerForServers@odata.count":1},"LogServices":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices"},"ManagerType":"BMC","Model":"13G Modular","Name":"Manager","NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/NetworkProtocol"},"Redundancy":[],"Redundancy@odata.count":0,"RedundancySet":[],"RedundancySet@odata.count":0,"SerialConsole":{"ConnectTypesSupported":[],"ConnectTypesSupported@odata.count":0,"MaxConcurrentSessions":0,"ServiceEnabled":false},"SerialInterfaces":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/SerialInterfaces"},"Status":{"Health":"Ok","State":"Enabled"},"UUID":"3243434f-c0b9-4480-3510-00304c4c4544","VirtualMedia":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia"}}`),
			RFBMCNetwork: []byte(``),
		},
		HP: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Systems/Members/$entity","@odata.id":"/redfish/v1/Systems/1/","@odata.type":"#ComputerSystem.1.0.1.ComputerSystem","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset/"}},"AssetTag":"","AvailableActions":[{"Action":"Reset","Capabilities":[{"AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"PropertyName":"ResetType"}]}],"Bios":{"Current":{"VersionString":"I36 v2.40 (02/17/2017)"}},"BiosVersion":"I36 v2.40 (02/17/2017)","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideSupported":["None","Cd","Hdd","Usb","Utilities","Diags","BiosSetup","Pxe","UefiShell","UefiTarget"],"BootSourceOverrideTarget":"None","UefiTargetBootSourceOverride":"None","UefiTargetBootSourceOverrideSupported":["HD.Emb.1.3","Generic.USB.1.1","HD.Emb.1.2","NIC.FlexLOM.1.1.IPv4","NIC.FlexLOM.1.1.IPv6"]},"Description":"Computer System View","HostCorrelation":{"HostMACAddress":["ec:b1:d7:b8:ac:c0","ec:b1:d7:b8:ac:c8"],"HostName":"bbmi","IPAddress":["",""]},"HostName":"bbmi","Id":"1","IndicatorLED":"Off","LogServices":{"@odata.id":"/redfish/v1/Systems/1/LogServices/"},"Manufacturer":"HPE","Memory":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGB":128},"MemorySummary":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGiB":128},"Model":"ProLiant BL460c Gen9","Name":"Computer System","Oem":{"Hp":{"@odata.type":"#HpComputerSystemExt.1.1.2.HpComputerSystemExt","Actions":{"#HpComputerSystemExt.PowerButton":{"PushType@Redfish.AllowableValues":["Press","PressAndHold"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.PowerButton/"},"#HpComputerSystemExt.SystemReset":{"ResetType@Redfish.AllowableValues":["ColdBoot"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.SystemReset/"}},"AvailableActions":[{"Action":"PowerButton","Capabilities":[{"AllowableValues":["Press","PressAndHold"],"PropertyName":"PushType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"ServerSigRecompute","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"SystemReset","Capabilities":[{"AllowableValues":["ColdBoot"],"PropertyName":"ResetType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]}],"Battery":[{"Condition":"Ok","ErrorCode":0,"FirmwareVersion":"1.3","Index":1,"MaxCapWatts":12,"Model":"727261-B21","Present":"Yes","ProductName":"HPE Smart Storage Battery ","SerialNumber":"6EZBP0GB2190JM","Spare":"815984-001"}],"Bios":{"Backup":{"Date":"12/28/2015","Family":"I36","VersionString":"I36 v2.00 (12/28/2015)"},"Current":{"Date":"02/17/2017","Family":"I36","VersionString":"I36 v2.40 (02/17/2017)"},"UefiClass":2},"DeviceDiscoveryComplete":{"AMSDeviceDiscovery":"NoAMS","DeviceDiscovery":"vMainDeviceDiscoveryComplete","SmartArrayDiscovery":"Complete"},"IntelligentProvisioningIndex":3,"IntelligentProvisioningLocation":"System Board","IntelligentProvisioningVersion":"N/A","PostState":"FinishedPost","PowerAllocationLimit":500,"PowerAutoOn":"PowerOn","PowerOnDelay":"Minimum","PowerRegulatorMode":"Max","PowerRegulatorModesSupported":["OSControl","Dynamic","Max","Min"],"TrustedModules":[{"Status":"NotPresent"}],"Type":"HpComputerSystemExt.1.1.2","VirtualProfile":"Inactive","links":{"BIOS":{"href":"/redfish/v1/systems/1/bios/"},"EthernetInterfaces":{"href":"/redfish/v1/Systems/1/EthernetInterfaces/"},"FirmwareInventory":{"href":"/redfish/v1/Systems/1/FirmwareInventory/"},"Memory":{"href":"/redfish/v1/Systems/1/Memory/"},"NetworkAdapters":{"href":"/redfish/v1/Systems/1/NetworkAdapters/"},"PCIDevices":{"href":"/redfish/v1/Systems/1/PCIDevices/"},"PCISlots":{"href":"/redfish/v1/Systems/1/PCISlots/"},"SUT":{"href":"/redfish/v1/systems/1/hpsut/"},"SecureBoot":{"href":"/redfish/v1/Systems/1/SecureBoot/"},"SmartStorage":{"href":"/redfish/v1/Systems/1/SmartStorage/"},"SoftwareInventory":{"href":"/redfish/v1/Systems/1/SoftwareInventory/"}}}},"Power":"On","PowerState":"On","ProcessorSummary":{"Count":2,"Model":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"Processors":{"Count":2,"ProcessorFamily":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"SKU":"813198-B21","SerialNumber":"CZ3629FY8B","Status":{"Health":"OK","State":"Enabled"},"SystemType":"Physical","Type":"ComputerSystem.1.0.1","UUID":"31333138-3839-5A43-3336-323946593842","links":{"Chassis":[{"href":"/redfish/v1/Chassis/1/"}],"Logs":{"href":"/redfish/v1/Systems/1/LogServices/"},"ManagedBy":[{"href":"/redfish/v1/Managers/1/"}],"Processors":{"href":"/redfish/v1/Systems/1/Processors/"},"self":{"href":"/redfish/v1/Systems/1/"}}}`),
			RFPower:      []byte(``),
			RFThermal:    []byte(``),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Systems/Members/1/Processors/Members/$entity","@odata.id":"/redfish/v1/Systems/1/Processors/1/","@odata.type":"#Processor.1.0.0.Processor","Id":"1","InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Name":"Processors","Oem":{"Hp":{"@odata.type":"#HpProcessorExt.1.0.0.HpProcessorExt","AssetTag":"UNKNOWN","Cache":[{"Associativity":"8waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":512,"Location":"Internal","MaximumSizeKB":512,"Name":"L1-Cache","Policy":"WriteBack","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"},{"Associativity":"8waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":2048,"Location":"Internal","MaximumSizeKB":2048,"Name":"L2-Cache","Policy":"Varies","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"},{"Associativity":"20waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":20480,"Location":"Internal","MaximumSizeKB":20480,"Name":"L3-Cache","Policy":"Varies","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"}],"Characteristics":["64Bit","MultiCore","HwThread","ExecuteProtection","EnhancedVirtualization","PowerPerfControl"],"ConfigStatus":{"Populated":true,"State":"Enabled"},"CoresEnabled":8,"ExternalClockMHz":100,"MicrocodePatches":[{"CpuId":"0x000306F2","Date":"2016-10-07T00:00:00Z","PatchId":"0x00000039"},{"CpuId":"0x000406F1","Date":"2016-10-07T00:00:00Z","PatchId":"0x0B00001F"}],"PartNumber":"","RatedSpeedMHz":2100,"SerialNumber":"","Type":"HpProcessorExt.1.0.0","VoltageVoltsX10":18}},"ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"179","EffectiveModel":"15","IdentificationRegisters":"0x06f10004fbffbfeb","MicrocodeInfo":null,"Step":"1","VendorId":"Intel"},"ProcessorType":"CPU","Socket":"Proc 1","Status":{"Health":"OK"},"TotalCores":8,"TotalThreads":16,"Type":"Processor.1.0.0","links":{"self":{"href":"/redfish/v1/Systems/1/Processors/1/"}}}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Managers/Members/$entity","@odata.id":"/redfish/v1/Managers/1/","@odata.type":"#Manager.1.0.0.Manager","Actions":{"#Manager.Reset":{"target":"/redfish/v1/Managers/1/Actions/Manager.Reset/"}},"AvailableActions":[{"Action":"Reset"}],"CommandShell":{"ConnectTypesSupported":["SSH","Oem"],"Enabled":true,"MaxConcurrentSessions":9,"ServiceEnabled":true},"Description":"Manager View","EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/1/EthernetInterfaces/"},"Firmware":{"Current":{"VersionString":"iLO 4 v2.54"}},"FirmwareVersion":"iLO 4 v2.54","GraphicalConsole":{"ConnectTypesSupported":["KVMIP"],"Enabled":true,"MaxConcurrentSessions":10,"ServiceEnabled":true},"Id":"1","LogServices":{"@odata.id":"/redfish/v1/Managers/1/LogServices/"},"ManagerType":"BMC","Name":"Manager","NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/1/NetworkService/"},"Oem":{"Hp":{"@odata.type":"#HpiLO.1.1.0.HpiLO","Actions":{"#HpiLO.ClearRestApiState":{"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.ClearRestApiState/"},"#HpiLO.ResetToFactoryDefaults":{"ResetType@Redfish.AllowableValues":["Default"],"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.ResetToFactoryDefaults/"},"#HpiLO.iLOFunctionality":{"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.iLOFunctionality/"}},"AvailableActions":[{"Action":"ClearRestApiState","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"ResetToFactoryDefaults","Capabilities":[{"AllowableValues":["Default"],"PropertyName":"ResetType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"iLOFunctionality","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]}],"ClearRestApiStatus":"DataPresent","FederationConfig":{"IPv6MulticastScope":"Site","MulticastAnnouncementInterval":600,"MulticastDiscovery":"Enabled","MulticastTimeToLive":5,"iLOFederationManagement":"Enabled"},"Firmware":{"Current":{"Date":"Jun 15 2017","DebugBuild":false,"MajorVersion":2,"MinorVersion":54,"Time":"","VersionString":"iLO 4 v2.54"}},"License":{"LicenseKey":"xxxxx-xxxxx-xxxxx-xxxxx-3DXMM","LicenseString":"iLO Advanced","LicenseType":"Perpetual"},"RequiredLoginForiLORBSU":false,"SerialCLISpeed":9600,"SerialCLIStatus":"EnabledAuthReq","Type":"HpiLO.1.1.0","VSPLogDownloadEnabled":false,"iLOSelfTestResults":[{"Notes":"","SelfTestName":"NVRAMData","Status":"OK"},{"Notes":"","SelfTestName":"NVRAMSpace","Status":"OK"},{"Notes":"Controller firmware revision  2.10.00  ","SelfTestName":"EmbeddedFlash/SDCard","Status":"OK"},{"Notes":"","SelfTestName":"EEPROM","Status":"OK"},{"Notes":"","SelfTestName":"HostRom","Status":"OK"},{"Notes":"","SelfTestName":"SupportedHost","Status":"OK"},{"Notes":"Version 1.0.9","SelfTestName":"PowerManagementController","Status":"Informational"},{"Notes":"ProLiant BL460c Gen9 System Programmable Logic Device version 0x17","SelfTestName":"CPLDPAL0","Status":"Informational"},{"Notes":"ProLiant BL460c Gen9 SAS Programmable Logic Device version 0x02","SelfTestName":"CPLDPAL1","Status":"Informational"}],"links":{"ActiveHealthSystem":{"href":"/redfish/v1/Managers/1/ActiveHealthSystem/"},"DateTimeService":{"href":"/redfish/v1/Managers/1/DateTime/"},"EmbeddedMediaService":{"href":"/redfish/v1/Managers/1/EmbeddedMedia/"},"FederationDispatch":{"extref":"/dispatch/"},"FederationGroups":{"href":"/redfish/v1/Managers/1/FederationGroups/"},"FederationPeers":{"href":"/redfish/v1/Managers/1/FederationPeers/"},"LicenseService":{"href":"/redfish/v1/Managers/1/LicenseService/"},"SecurityService":{"href":"/redfish/v1/Managers/1/SecurityService/"},"UpdateService":{"href":"/redfish/v1/Managers/1/UpdateService/"},"VSPLogLocation":{"extref":"/sol.log.gz/"}}}},"SerialConsole":{"ConnectTypesSupported":["SSH","IPMI","Oem"],"Enabled":true,"MaxConcurrentSessions":13,"ServiceEnabled":true},"Status":{"State":"Enabled"},"Type":"Manager.1.0.0","UUID":"1cf36323-33b6-50e8-a7b3-f58c1fea3f58","VirtualMedia":{"@odata.id":"/redfish/v1/Managers/1/VirtualMedia/"},"links":{"EthernetNICs":{"href":"/redfish/v1/Managers/1/EthernetInterfaces/"},"Logs":{"href":"/redfish/v1/Managers/1/LogServices/"},"ManagerForChassis":[{"href":"/redfish/v1/Chassis/1/"}],"ManagerForServers":[{"href":"/redfish/v1/Systems/1/"}],"NetworkService":{"href":"/redfish/v1/Managers/1/NetworkService/"},"VirtualMedia":{"href":"/redfish/v1/Managers/1/VirtualMedia/"},"self":{"href":"/redfish/v1/Managers/1/"}}}`),
			RFBMCNetwork: []byte(``),
		},
		Supermicro: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#ComputerSystem.ComputerSystem","@odata.type":"#ComputerSystem.ComputerSystem","@odata.id":"/redfish/v1/Systems/1","Id":"1","Name":"System","Description":"Description of server","Status":{"State":"Enabled","Health":"OK"},"SerialNumber":"","PartNumber":"","SystemType":"Physical","BiosVersion":"2.0","Manufacturer":"Supermicro","Model":"X10DFF-CTG","SKU":"Default string","UUID":"00000000-0000-0000-0000-0CC47AB721C4","ProcessorSummary":{"Count":16,"Model":"Intel(R) Xeon(R) processor","Status":{"State":"Enabled","Health":"OK"}},"MemorySummary":{"TotalSystemMemoryGiB":128,"Status":{"State":"Enabled","Health":"OK"}},"IndicatorLED":"Off","PowerState":"On","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideTarget":"None","BootSourceOverrideTarget@Redfish.AllowableValues":["None","Pxe","Hdd","Diags","Cd","BiosSetup","FloppyRemovableMedia","UsbKey","UsbHdd","UsbFloppy","UsbCd","UefiUsbKey","UefiCd","UefiHdd","UefiUsbHdd","UefiUsbCd"]},"Processors":{"@odata.id":"/redfish/v1/Systems/1/Processors"},"Links":{"Chassis":[{"@odata.id":"/redfish/v1/Chassis/1"}],"ManagedBy":[{"@odata.id":"/redfish/v1/Managers/1"}],"Oem":{}},"Actions":{"#ComputerSystem.Reset":{"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset","ResetType@Redfish.AllowableValues":["On","ForceOff","GracefulRestart","ForceRestart","Nmi","ForceOn"]}}}`),
			RFPower:      []byte(``),
			RFThermal:    []byte(``),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Processor.Processor","@odata.type":"#Processor.Processor","@odata.id":"/redfish/v1/Systems/1/Processors/1","Id":"1","Name":"Processor","Description":"Processor","Socket":"CPU1","Manufacturer":"Intel","Model":"Intel(R) Xeon(R) CPU E5-2630 v3 @ 2.40GHz","MaxSpeedMHz":4000,"TotalCores":8,"TotalThreads":16,"ProcessorType":"CPU","ProcessorArchitecture":"x86","InstructionSet":"x86-64","ProcessorId":{"VendorId":"GenuineIntel","IdentificationRegisters":"0xBFEBFBFF000306F2","EffectiveFamily":"0x6","EffectiveModel":"0x3F","Step":"0x2"},"Status":{"State":"Enabled","Health":"OK"}}`),
			RFCPUEntry:   []byte(`{"@odata.context":"/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection","@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors","@odata.type":"#ProcessorCollection.ProcessorCollection","Description":"Collection of Processors for this System","Members":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"},{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"Members@odata.count":2,"Name":"ProcessorsCollection"}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Manager.Manager","@odata.type":"#Manager.Manager","@odata.id":"/redfish/v1/Managers/1","Id":"1","Name":"Manager","Description":"BMC","ManagerType":"BMC","UUID":"00000000-0000-0000-0000-0CC47AB982F7","Model":"ASPEED","FirmwareVersion":"3.25","DateTime":"2017-09-07T14:49:19+00:00","DateTimeLocalOffset":"+00:00","Status":{"State":"Enabled","Health":"OK"},"GraphicalConsole":{"ServiceEnabled":true,"MaxConcurrentSessions":4,"ConnectTypesSupported":["KVMIP"]},"SerialConsole":{"ServiceEnabled":true,"MaxConcurrentSessions":1,"ConnectTypesSupported":["SSH","IPMI"]},"CommandShell":{"ServiceEnabled":true,"MaxConcurrentSessions":0,"ConnectTypesSupported":["SSH"]},"EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/1/EthernetInterfaces"},"SerialInterfaces":{"@odata.id":"/redfish/v1/Managers/1/SerialInterfaces"},"NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/1/NetworkProtocol"},"LogServices":{"@odata.id":"/redfish/v1/Managers/1/LogServices"},"VirtualMedia":{"@odata.id":"/redfish/v1/Managers/1/VM1"},"Links":{"ManagerForServers":[{"@odata.id":"/redfish/v1/Systems/1"}],"ManagerForChassis":[{"@odata.id":"/redfish/v1/Chassis/1"}],"Oem":{}},"Actions":{"Oem":{"#ManagerConfig.Reset":{"target":"/redfish/v1/Managers/1/Actions/Oem/ManagerConfig.Reset"}},"#Manager.Reset":{"target":"/redfish/v1/Managers/1/Actions/Manager.Reset"}},"Oem":{"ActiveDirectory":{"@odata.id":"/redfish/v1/Managers/1/ActiveDirectory"},"SMTP":{"@odata.id":"/redfish/v1/Managers/1/SMTP"},"RADIUS":{"@odata.id":"/redfish/v1/Managers/1/RADIUS"},"MouseMode":{"@odata.id":"/redfish/v1/Managers/1/MouseMode"},"NTP":{"@odata.id":"/redfish/v1/Managers/1/NTP"},"LDAP":{"@odata.id":"/redfish/v1/Managers/1/LDAP"},"IPAccessControl":{"@odata.id":"/redfish/v1/Managers/1/IPAccessControl"},"SMCRAKP":{"@odata.id":"/redfish/v1/Managers/1/SMCRAKP"},"SNMP":{"@odata.id":"/redfish/v1/Managers/1/SNMP"},"Syslog":{"@odata.id":"/redfish/v1/Managers/1/Syslog"},"Snooping":{"@odata.id":"/redfish/v1/Managers/1/Snooping"},"FanMode":{"@odata.id":"/redfish/v1/Managers/1/FanMode"}}}`),
			RFBMCNetwork: []byte(``),
		},
	}
)

// Serial() (string, error)
// Model() (string, error)
// BmcType() (string, error)
// BmcVersion() (string, error)
// Name() (string, error)
// Status() (string, error)
// Memory() (int, error)
// CPU() (string, int, int, int, error)
// BiosVersion() (string, error)
// PowerKw() (float64, error)
// TempC() (int, error)
// Nics() ([]*model.Nic, error)
// License() (string, string, error)
// Login() error
// Logout() error

func TestRedfishBios(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"BiosVersion",
			HP,
			RFEntry,
			"I36 v2.40 (02/17/2017)",
			"iLO",
		},
		{
			"BiosVersion",
			Dell,
			RFEntry,
			"2.4.2",
			"iDRAC",
		},
		{
			"BiosVersion",
			Supermicro,
			RFEntry,
			"2.0",
			"Supermicro",
		},
	}

	for _, tc := range tt {
		rf, err := setup(tc.vendor, tc.redfishendpoint, tc.detectionString)
		if err == nil {

			method := reflect.ValueOf(rf).MethodByName(tc.testType)
			result := method.Call([]reflect.Value{})
			answer := result[0].Interface()
			nerr := result[1].Interface()
			if nerr != nil {
				t.Errorf("Found errors calling %s: %s", tc.testType, nerr)
			}

			if answer != tc.expectedAnswer {
				t.Errorf("%s from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer, answer)
			}
		} else {
			t.Errorf("Found errors during the test setup %v", err)
		}
		teardown()
	}
}

func TestRedfishBmcType(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"BmcType",
			HP,
			RFBMC,
			"iLO4",
			"iLO",
		},
		{
			"BmcType",
			Dell,
			RFBMC,
			"iDRAC",
			"iDRAC",
		},
		{
			"BmcType",
			Supermicro,
			RFBMC,
			"Supermicro",
			"Supermicro",
		},
	}

	for _, tc := range tt {
		rf, err := setup(tc.vendor, tc.redfishendpoint, tc.detectionString)
		if err == nil {

			method := reflect.ValueOf(rf).MethodByName(tc.testType)
			result := method.Call([]reflect.Value{})
			answer := result[0].Interface()
			nerr := result[1].Interface()
			if nerr != nil {
				t.Errorf("Found errors calling %s: %s", tc.testType, nerr)
			}

			if answer != tc.expectedAnswer {
				t.Errorf("%s from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer, answer)
			}
		} else {
			t.Errorf("Found errors during the test setup %v", err)
		}
		teardown()
	}
}

func TestRedfishBmcVersion(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"BmcVersion",
			HP,
			RFBMC,
			"iLO 4 v2.54",
			"iLO",
		},
		{
			"BmcVersion",
			Dell,
			RFBMC,
			"2.41.40.40",
			"iDRAC",
		},
		{
			"BmcVersion",
			Supermicro,
			RFBMC,
			"3.25",
			"Supermicro",
		},
	}

	for _, tc := range tt {
		rf, err := setup(tc.vendor, tc.redfishendpoint, tc.detectionString)
		if err == nil {

			method := reflect.ValueOf(rf).MethodByName(tc.testType)
			result := method.Call([]reflect.Value{})
			answer := result[0].Interface()
			nerr := result[1].Interface()
			if nerr != nil {
				t.Errorf("Found errors calling %s: %s", tc.testType, nerr)
			}

			if answer != tc.expectedAnswer {
				t.Errorf("%s from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer, answer)
			}
		} else {
			t.Errorf("Found errors during the test setup %v", err)
		}
		teardown()
	}
}

func TestRedfishMemory(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  int
		detectionString string
	}{
		{
			"Memory",
			HP,
			RFEntry,
			128,
			"iLO",
		},
		{
			"Memory",
			Dell,
			RFEntry,
			128,
			"iDRAC",
		},
		{
			"Memory",
			Supermicro,
			RFEntry,
			128,
			"Supermicro",
		},
	}

	for _, tc := range tt {
		rf, err := setup(tc.vendor, tc.redfishendpoint, tc.detectionString)
		if err == nil {

			method := reflect.ValueOf(rf).MethodByName(tc.testType)
			result := method.Call([]reflect.Value{})
			answer := result[0].Interface()
			nerr := result[1].Interface()
			if nerr != nil {
				t.Errorf("Found errors calling %s: %s", tc.testType, nerr)
			}

			if answer != tc.expectedAnswer {
				t.Errorf("%s from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer, answer)
			}
		} else {
			t.Errorf("Found errors during the test setup %v", err)
		}
		teardown()
	}
}

func TestRedfishCPU(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  struct {
			cpu              string
			cpuCount         int
			coreCount        int
			hyperthreadCount int
		}
		detectionString string
	}{
		{
			"CPU",
			HP,
			RFEntry,
			struct {
				cpu              string
				cpuCount         int
				coreCount        int
				hyperthreadCount int
			}{
				"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz",
				2,
				8,
				16,
			},
			"iLO",
		},
		{
			"CPU",
			Dell,
			RFEntry,
			struct {
				cpu              string
				cpuCount         int
				coreCount        int
				hyperthreadCount int
			}{
				"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz",
				2,
				6,
				12,
			},
			"iDRAC",
		},
		{
			"CPU",
			Supermicro,
			RFEntry,
			struct {
				cpu              string
				cpuCount         int
				coreCount        int
				hyperthreadCount int
			}{
				"Intel(R) Xeon(R) processor",
				2,
				8,
				16,
			},
			"Supermicro",
		},
	}

	for _, tc := range tt {
		rf, err := setup(tc.vendor, tc.redfishendpoint, tc.detectionString)
		if err == nil {
			mux.HandleFunc(fmt.Sprintf("/%s", redfishVendorEndPoints[tc.vendor][RFCPU]), func(w http.ResponseWriter, r *http.Request) {
				w.Write(redfishVendorAnswers[tc.vendor][RFCPU])
			})

			// Supermicro doesn't know how to count procs it seems. They are exposing threads
			// over the total proc count, so we need to do one extra call for Supermicro boxes
			if tc.vendor == Supermicro {
				mux.HandleFunc(fmt.Sprintf("/%s", redfishVendorEndPoints[tc.vendor][RFCPUEntry]), func(w http.ResponseWriter, r *http.Request) {
					w.Write(redfishVendorAnswers[tc.vendor][RFCPUEntry])
				})
			}

			cpu, cpuCount, coreCount, hyperthreadCount, err := rf.CPU()
			if err != nil {
				t.Errorf("Found errors calling %s: %s", tc.testType, err)
			}

			if cpu != tc.expectedAnswer.cpu {
				t.Errorf("%s cpu from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer.cpu, cpu)
			}

			if cpuCount != tc.expectedAnswer.cpuCount {
				t.Errorf("%s cpuCount from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer.cpuCount, cpuCount)
			}

			if coreCount != tc.expectedAnswer.coreCount {
				t.Errorf("%s coreCount from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer.cpuCount, coreCount)
			}

			if hyperthreadCount != tc.expectedAnswer.hyperthreadCount {
				t.Errorf("%s hyperthreadCount from vendor %s should answer %v: found %v", tc.testType, tc.vendor, tc.expectedAnswer.cpuCount, hyperthreadCount)
			}
		} else {
			t.Errorf("Found errors during the test setup %v", err)
		}
		teardown()
	}
}
