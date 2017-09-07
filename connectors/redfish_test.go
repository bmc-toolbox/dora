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
			RFCPU:        []byte(``),
			RFBMC:        []byte(``),
			RFBMCNetwork: []byte(``),
		},
		HP: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Systems/Members/$entity","@odata.id":"/redfish/v1/Systems/1/","@odata.type":"#ComputerSystem.1.0.1.ComputerSystem","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset/"}},"AssetTag":"","AvailableActions":[{"Action":"Reset","Capabilities":[{"AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"PropertyName":"ResetType"}]}],"Bios":{"Current":{"VersionString":"I36 v2.40 (02/17/2017)"}},"BiosVersion":"I36 v2.40 (02/17/2017)","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideSupported":["None","Cd","Hdd","Usb","Utilities","Diags","BiosSetup","Pxe","UefiShell","UefiTarget"],"BootSourceOverrideTarget":"None","UefiTargetBootSourceOverride":"None","UefiTargetBootSourceOverrideSupported":["HD.Emb.1.3","Generic.USB.1.1","HD.Emb.1.2","NIC.FlexLOM.1.1.IPv4","NIC.FlexLOM.1.1.IPv6"]},"Description":"Computer System View","HostCorrelation":{"HostMACAddress":["ec:b1:d7:b8:ac:c0","ec:b1:d7:b8:ac:c8"],"HostName":"bbmi","IPAddress":["",""]},"HostName":"bbmi","Id":"1","IndicatorLED":"Off","LogServices":{"@odata.id":"/redfish/v1/Systems/1/LogServices/"},"Manufacturer":"HPE","Memory":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGB":128},"MemorySummary":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGiB":128},"Model":"ProLiant BL460c Gen9","Name":"Computer System","Oem":{"Hp":{"@odata.type":"#HpComputerSystemExt.1.1.2.HpComputerSystemExt","Actions":{"#HpComputerSystemExt.PowerButton":{"PushType@Redfish.AllowableValues":["Press","PressAndHold"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.PowerButton/"},"#HpComputerSystemExt.SystemReset":{"ResetType@Redfish.AllowableValues":["ColdBoot"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.SystemReset/"}},"AvailableActions":[{"Action":"PowerButton","Capabilities":[{"AllowableValues":["Press","PressAndHold"],"PropertyName":"PushType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"ServerSigRecompute","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"SystemReset","Capabilities":[{"AllowableValues":["ColdBoot"],"PropertyName":"ResetType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]}],"Battery":[{"Condition":"Ok","ErrorCode":0,"FirmwareVersion":"1.3","Index":1,"MaxCapWatts":12,"Model":"727261-B21","Present":"Yes","ProductName":"HPE Smart Storage Battery ","SerialNumber":"6EZBP0GB2190JM","Spare":"815984-001"}],"Bios":{"Backup":{"Date":"12/28/2015","Family":"I36","VersionString":"I36 v2.00 (12/28/2015)"},"Current":{"Date":"02/17/2017","Family":"I36","VersionString":"I36 v2.40 (02/17/2017)"},"UefiClass":2},"DeviceDiscoveryComplete":{"AMSDeviceDiscovery":"NoAMS","DeviceDiscovery":"vMainDeviceDiscoveryComplete","SmartArrayDiscovery":"Complete"},"IntelligentProvisioningIndex":3,"IntelligentProvisioningLocation":"System Board","IntelligentProvisioningVersion":"N/A","PostState":"FinishedPost","PowerAllocationLimit":500,"PowerAutoOn":"PowerOn","PowerOnDelay":"Minimum","PowerRegulatorMode":"Max","PowerRegulatorModesSupported":["OSControl","Dynamic","Max","Min"],"TrustedModules":[{"Status":"NotPresent"}],"Type":"HpComputerSystemExt.1.1.2","VirtualProfile":"Inactive","links":{"BIOS":{"href":"/redfish/v1/systems/1/bios/"},"EthernetInterfaces":{"href":"/redfish/v1/Systems/1/EthernetInterfaces/"},"FirmwareInventory":{"href":"/redfish/v1/Systems/1/FirmwareInventory/"},"Memory":{"href":"/redfish/v1/Systems/1/Memory/"},"NetworkAdapters":{"href":"/redfish/v1/Systems/1/NetworkAdapters/"},"PCIDevices":{"href":"/redfish/v1/Systems/1/PCIDevices/"},"PCISlots":{"href":"/redfish/v1/Systems/1/PCISlots/"},"SUT":{"href":"/redfish/v1/systems/1/hpsut/"},"SecureBoot":{"href":"/redfish/v1/Systems/1/SecureBoot/"},"SmartStorage":{"href":"/redfish/v1/Systems/1/SmartStorage/"},"SoftwareInventory":{"href":"/redfish/v1/Systems/1/SoftwareInventory/"}}}},"Power":"On","PowerState":"On","ProcessorSummary":{"Count":2,"Model":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"Processors":{"Count":2,"ProcessorFamily":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"SKU":"813198-B21","SerialNumber":"CZ3629FY8B","Status":{"Health":"OK","State":"Enabled"},"SystemType":"Physical","Type":"ComputerSystem.1.0.1","UUID":"31333138-3839-5A43-3336-323946593842","links":{"Chassis":[{"href":"/redfish/v1/Chassis/1/"}],"Logs":{"href":"/redfish/v1/Systems/1/LogServices/"},"ManagedBy":[{"href":"/redfish/v1/Managers/1/"}],"Processors":{"href":"/redfish/v1/Systems/1/Processors/"},"self":{"href":"/redfish/v1/Systems/1/"}}}`),
			RFPower:      []byte(``),
			RFThermal:    []byte(``),
			RFCPU:        []byte(``),
			RFBMC:        []byte(``),
			RFBMCNetwork: []byte(``),
		},
		Supermicro: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#ComputerSystem.ComputerSystem","@odata.type":"#ComputerSystem.ComputerSystem","@odata.id":"/redfish/v1/Systems/1","Id":"1","Name":"System","Description":"Description of server","Status":{"State":"Enabled","Health":"OK"},"SerialNumber":"","PartNumber":"","SystemType":"Physical","BiosVersion":"2.0","Manufacturer":"Supermicro","Model":"X10DFF-CTG","SKU":"Default string","UUID":"00000000-0000-0000-0000-0CC47AB721C4","ProcessorSummary":{"Count":16,"Model":"Intel(R) Xeon(R) processor","Status":{"State":"Enabled","Health":"OK"}},"MemorySummary":{"TotalSystemMemoryGiB":128,"Status":{"State":"Enabled","Health":"OK"}},"IndicatorLED":"Off","PowerState":"On","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideTarget":"None","BootSourceOverrideTarget@Redfish.AllowableValues":["None","Pxe","Hdd","Diags","Cd","BiosSetup","FloppyRemovableMedia","UsbKey","UsbHdd","UsbFloppy","UsbCd","UefiUsbKey","UefiCd","UefiHdd","UefiUsbHdd","UefiUsbCd"]},"Processors":{"@odata.id":"/redfish/v1/Systems/1/Processors"},"Links":{"Chassis":[{"@odata.id":"/redfish/v1/Chassis/1"}],"ManagedBy":[{"@odata.id":"/redfish/v1/Managers/1"}],"Oem":{}},"Actions":{"#ComputerSystem.Reset":{"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset","ResetType@Redfish.AllowableValues":["On","ForceOff","GracefulRestart","ForceRestart","Nmi","ForceOn"]}}}`),
			RFPower:      []byte(``),
			RFThermal:    []byte(``),
			RFCPU:        []byte(``),
			RFBMC:        []byte(``),
			RFBMCNetwork: []byte(``),
		},
	}
)

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
