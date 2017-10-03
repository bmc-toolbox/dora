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
			RFPower:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Power.Power","@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Power","@odata.type":"#Power.v1_0_2.Power","Description":"Power","Id":"Power","Name":"Power","PowerControl":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Power/PowerControl","MemberID":"PowerControl","Name":"System Power Control","PowerAllocatedWatts":252,"PowerAvailableWatts":0,"PowerCapacityWatts":304,"PowerConsumedWatts":121,"PowerLimit":{"CorrectionInMs":0,"LimitException":"HardPowerOff","LimitInWatts":440},"PowerMetrics":{"AverageConsumedWatts":125,"IntervalInMin":60,"MaxConsumedWatts":151,"MinConsumedWatts":121},"PowerRequestedWatts":440,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"},{"@odata.id":"/redfish/v1/Systems/System.Embedded.1"}],"RelatedItem@odata.count":2}],"PowerControl@odata.count":1,"PowerSupplies":[],"PowerSupplies@odata.count":0,"Redundancy":[],"Redundancy@odata.count":0,"Voltages":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1VCOREPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU1VCOREPG","MinReadingRange":44,"Name":"CPU1 VCORE PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":35,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2VCOREPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU2VCOREPG","MinReadingRange":0,"Name":"CPU2 VCORE PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":36,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard3.3VPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoard3.3VPG","MinReadingRange":44,"Name":"System Board 3.3V PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":25,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard12VPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoard12VPG","MinReadingRange":44,"Name":"System Board 12V PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":17,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard5VAUXPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#SystemBoard5VAUXPG","MinReadingRange":0,"Name":"System Board 5V AUX PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":26,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M23VPPPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2M23VPPPG","MinReadingRange":44,"Name":"CPU2 M23 VPP PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":34,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M23VPPPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU1M23VPPPG","MinReadingRange":0,"Name":"CPU1 M23 VPP PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":37,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard1.5VPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoard1.5VPG","MinReadingRange":44,"Name":"System Board 1.5V PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":40,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard1.5VAUXPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoard1.5VAUXPG","MinReadingRange":44,"Name":"System Board 1.5V AUX PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":249,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard1.05VPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#SystemBoard1.05VPG","MinReadingRange":0,"Name":"System Board 1.05V PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":39,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M01VTTPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU1M01VTTPG","MinReadingRange":0,"Name":"CPU1 M01 VTT PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":20,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M23VDDQPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU1M23VDDQPG","MinReadingRange":0,"Name":"CPU1 M23 VDDQ PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":21,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M23VTTPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU1M23VTTPG","MinReadingRange":44,"Name":"CPU1 M23 VTT PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":22,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoardDIMMPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoardDIMMPG","MinReadingRange":44,"Name":"System Board DIMM PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":41,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoardVCCIOPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#SystemBoardVCCIOPG","MinReadingRange":0,"Name":"System Board VCCIO PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":43,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M01VDDQPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2M01VDDQPG","MinReadingRange":44,"Name":"CPU2 M01 VDDQ PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":27,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M01VDDQPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU1M01VDDQPG","MinReadingRange":0,"Name":"CPU1 M01 VDDQ PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":30,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M23VTTPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2M23VTTPG","MinReadingRange":44,"Name":"CPU2 M23 VTT PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":46,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M01VTTPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2M01VTTPG","MinReadingRange":44,"Name":"CPU2 M01 VTT PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":28,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23NDCPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#NDCPG","MinReadingRange":0,"Name":"NDC PG","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":47,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M01VPPPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2M01VPPPG","MinReadingRange":44,"Name":"CPU2 M01 VPP PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":31,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1M01VPPPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU1M01VPPPG","MinReadingRange":0,"Name":"CPU1 M01 VPP PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":32,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2FIVRPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU2FIVRPG","MinReadingRange":44,"Name":"CPU2 FIVR PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":252,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1FIVRPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#CPU1FIVRPG","MinReadingRange":44,"Name":"CPU1 FIVR PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":251,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU2M23VDDQPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#CPU2M23VDDQPG","MinReadingRange":0,"Name":"CPU2 M23 VDDQ PG","PhysicalContext":"CPU","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":29,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23MEZZBPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#MEZZBPG","MinReadingRange":44,"Name":"MEZZB PG","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":18,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23PERC1PG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":0,"MemberID":"iDRAC.Embedded.1#PERC1PG","MinReadingRange":0,"Name":"PERC1 PG","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":42,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23MEZZCPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#MEZZCPG","MinReadingRange":44,"Name":"MEZZC PG","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":19,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23SystemBoard2.5VAUXPG","LowerThresholdCritical":null,"LowerThresholdFatal":null,"LowerThresholdNonCritical":null,"MaxReadingRange":48,"MemberID":"iDRAC.Embedded.1#SystemBoard2.5VAUXPG","MinReadingRange":44,"Name":"System Board 2.5V AUX PG","PhysicalContext":"SystemBoard","ReadingVolts":1,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":38,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":null,"UpperThresholdFatal":null,"UpperThresholdNonCritical":null}],"Voltages@odata.count":29}`),
			RFThermal:    []byte(`{"@odata.context":"/redfish/v1/$metadata#Thermal.Thermal","@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Thermal","@odata.type":"#Thermal.v1_0_2.Thermal","Description":"Represents the properties for Temperature and Cooling","Fans":[],"Fans@odata.count":0,"Id":"Thermal","Name":"Thermal","Redundancy":[],"Redundancy@odata.count":0,"Temperatures":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Temperatures/iDRAC.Embedded.1%23SystemBoardInletTemp","LowerThresholdCritical":-7,"LowerThresholdFatal":-7,"LowerThresholdNonCritical":3,"MaxReadingRangeTemp":19,"MemberID":"iDRAC.Embedded.1#SystemBoardInletTemp","MinReadingRangeTemp":13,"Name":"System Board Inlet Temp","PhysicalContext":"Intake","ReadingCelsius":17,"RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"RelatedItem@odata.count":1,"SensorNumber":1,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":47,"UpperThresholdFatal":47,"UpperThresholdNonCritical":42},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Temperatures/iDRAC.Embedded.1%23CPU1Temp","LowerThresholdCritical":3,"LowerThresholdFatal":3,"LowerThresholdNonCritical":8,"MaxReadingRangeTemp":19,"MemberID":"iDRAC.Embedded.1#CPU1Temp","MinReadingRangeTemp":13,"Name":"CPU1 Temp","PhysicalContext":"CPU","ReadingCelsius":47,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"}],"RelatedItem@odata.count":1,"SensorNumber":12,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":85,"UpperThresholdFatal":85,"UpperThresholdNonCritical":80},{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1/Sensors/Temperatures/iDRAC.Embedded.1%23CPU2Temp","LowerThresholdCritical":3,"LowerThresholdFatal":3,"LowerThresholdNonCritical":8,"MaxReadingRangeTemp":19,"MemberID":"iDRAC.Embedded.1#CPU2Temp","MinReadingRangeTemp":13,"Name":"CPU2 Temp","PhysicalContext":"CPU","ReadingCelsius":43,"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"RelatedItem@odata.count":1,"SensorNumber":13,"Status":{"Health":"OK","State":"Enabled"},"UpperThresholdCritical":85,"UpperThresholdFatal":85,"UpperThresholdNonCritical":80}],"Temperatures@odata.count":3}`),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Processor.Processor","@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1","@odata.type":"#Processor.v1_0_2.Processor","Description":"Represents the properties of a Processor attached to this System","Id":"CPU.Socket.1","InstructionSet":[{"Member":"x86-64"}],"Manufacturer":"Intel","MaxSpeedMHz":4000,"Model":"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","Name":"CPU 1","ProcessorArchitecture":[{"Member":"x86"}],"ProcessorId":{"EffectiveFamily":"6","EffectiveModel":"63","IdentificationRegisters":"0x000306F2","MicrocodeInfo":"0x39","Step":"2","VendorID":"GenuineIntel"},"ProcessorType":"CPU","Socket":"CPU.Socket.1","Status":{"Health":"OK","State":"Enabled"},"TotalCores":6,"TotalThreads":12}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Manager.Manager","@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1","@odata.type":"#Manager.v1_0_2.Manager","Actions":{"#Manager.Reset":{"ResetType@Redfish.AllowableValues":["GracefulRestart"],"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset"},"Oem":{"OemManager.v1_0_0#OemManager.ExportSystemConfiguration":{"ExportFormat@Redfish.AllowableValues":["XML"],"ExportUse@Redfish.AllowableValues":["Default","Clone","Replace"],"IncludeInExport@Redfish.AllowableValues":["Default","IncludeReadOnly","IncludePasswordHashValues"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL","IDRAC","BIOS","NIC","RAID"]},"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration"},"OemManager.v1_0_0#OemManager.ImportSystemConfiguration":{"HostPowerState@Redfish.AllowableValues":["On","Off"],"ImportSystemConfiguration@Redfish.AllowableValues":["TimeToWait","ImportBuffer"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL","IDRAC","BIOS","NIC","RAID"]},"ShutdownType@Redfish.AllowableValues":["Graceful","Forced","NoReboot"],"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"},"OemManager.v1_0_0#OemManager.ImportSystemConfigurationPreview":{"ImportSystemConfigurationPreview@Redfish.AllowableValues":["ImportBuffer"],"ShareParameters":{"ShareParameters@Redfish.AllowableValues":["IPAddress","ShareName","FileName","UserName","Password","Workgroup"],"ShareType@Redfish.AllowableValues":["NFS","CIFS"],"Target@Redfish.AllowableValues":["ALL"]},"target":"/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview"}}},"CommandShell":{"ConnectTypesSupported":["SSH","Telnet","IPMI"],"ConnectTypesSupported@odata.count":3,"MaxConcurrentSessions":5,"ServiceEnabled":true},"DateTime":"2017-09-07T16:48:14-05:00","DateTimeLocalOffset":"-05:00","Description":"BMC","EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces"},"FirmwareVersion":"2.41.40.40","GraphicalConsole":{"ConnectTypesSupported":["KVMIP"],"ConnectTypesSupported@odata.count":1,"MaxConcurrentSessions":6,"ServiceEnabled":true},"Id":"iDRAC.Embedded.1","Links":{"ManagerForChassis":[{"@odata.id":"/redfish/v1/Chassis/System.Embedded.1"}],"ManagerForChassis@odata.count":1,"ManagerForServers":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1"}],"ManagerForServers@odata.count":1},"LogServices":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices"},"ManagerType":"BMC","Model":"13G Modular","Name":"Manager","NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/NetworkProtocol"},"Redundancy":[],"Redundancy@odata.count":0,"RedundancySet":[],"RedundancySet@odata.count":0,"SerialConsole":{"ConnectTypesSupported":[],"ConnectTypesSupported@odata.count":0,"MaxConcurrentSessions":0,"ServiceEnabled":false},"SerialInterfaces":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/SerialInterfaces"},"Status":{"Health":"Ok","State":"Enabled"},"UUID":"3243434f-c0b9-4480-3510-00304c4c4544","VirtualMedia":{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia"}}`),
			RFBMCNetwork: []byte(``),
		},
		HP: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Systems/Members/$entity","@odata.id":"/redfish/v1/Systems/1/","@odata.type":"#ComputerSystem.1.0.1.ComputerSystem","Actions":{"#ComputerSystem.Reset":{"ResetType@Redfish.AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset/"}},"AssetTag":"","AvailableActions":[{"Action":"Reset","Capabilities":[{"AllowableValues":["On","ForceOff","ForceRestart","Nmi","PushPowerButton"],"PropertyName":"ResetType"}]}],"Bios":{"Current":{"VersionString":"I36 v2.40 (02/17/2017)"}},"BiosVersion":"I36 v2.40 (02/17/2017)","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideSupported":["None","Cd","Hdd","Usb","Utilities","Diags","BiosSetup","Pxe","UefiShell","UefiTarget"],"BootSourceOverrideTarget":"None","UefiTargetBootSourceOverride":"None","UefiTargetBootSourceOverrideSupported":["HD.Emb.1.3","Generic.USB.1.1","HD.Emb.1.2","NIC.FlexLOM.1.1.IPv4","NIC.FlexLOM.1.1.IPv6"]},"Description":"Computer System View","HostCorrelation":{"HostMACAddress":["ec:b1:d7:b8:ac:c0","ec:b1:d7:b8:ac:c8"],"HostName":"bbmi","IPAddress":["",""]},"HostName":"bbmi","Id":"1","IndicatorLED":"Off","LogServices":{"@odata.id":"/redfish/v1/Systems/1/LogServices/"},"Manufacturer":"HPE","Memory":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGB":128},"MemorySummary":{"Status":{"HealthRollUp":"OK"},"TotalSystemMemoryGiB":128},"Model":"ProLiant BL460c Gen9","Name":"Computer System","Oem":{"Hp":{"@odata.type":"#HpComputerSystemExt.1.1.2.HpComputerSystemExt","Actions":{"#HpComputerSystemExt.PowerButton":{"PushType@Redfish.AllowableValues":["Press","PressAndHold"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.PowerButton/"},"#HpComputerSystemExt.SystemReset":{"ResetType@Redfish.AllowableValues":["ColdBoot"],"target":"/redfish/v1/Systems/1/Actions/Oem/Hp/ComputerSystemExt.SystemReset/"}},"AvailableActions":[{"Action":"PowerButton","Capabilities":[{"AllowableValues":["Press","PressAndHold"],"PropertyName":"PushType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"ServerSigRecompute","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"SystemReset","Capabilities":[{"AllowableValues":["ColdBoot"],"PropertyName":"ResetType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]}],"Battery":[{"Condition":"Ok","ErrorCode":0,"FirmwareVersion":"1.3","Index":1,"MaxCapWatts":12,"Model":"727261-B21","Present":"Yes","ProductName":"HPE Smart Storage Battery ","SerialNumber":"6EZBP0GB2190JM","Spare":"815984-001"}],"Bios":{"Backup":{"Date":"12/28/2015","Family":"I36","VersionString":"I36 v2.00 (12/28/2015)"},"Current":{"Date":"02/17/2017","Family":"I36","VersionString":"I36 v2.40 (02/17/2017)"},"UefiClass":2},"DeviceDiscoveryComplete":{"AMSDeviceDiscovery":"NoAMS","DeviceDiscovery":"vMainDeviceDiscoveryComplete","SmartArrayDiscovery":"Complete"},"IntelligentProvisioningIndex":3,"IntelligentProvisioningLocation":"System Board","IntelligentProvisioningVersion":"N/A","PostState":"FinishedPost","PowerAllocationLimit":500,"PowerAutoOn":"PowerOn","PowerOnDelay":"Minimum","PowerRegulatorMode":"Max","PowerRegulatorModesSupported":["OSControl","Dynamic","Max","Min"],"TrustedModules":[{"Status":"NotPresent"}],"Type":"HpComputerSystemExt.1.1.2","VirtualProfile":"Inactive","links":{"BIOS":{"href":"/redfish/v1/systems/1/bios/"},"EthernetInterfaces":{"href":"/redfish/v1/Systems/1/EthernetInterfaces/"},"FirmwareInventory":{"href":"/redfish/v1/Systems/1/FirmwareInventory/"},"Memory":{"href":"/redfish/v1/Systems/1/Memory/"},"NetworkAdapters":{"href":"/redfish/v1/Systems/1/NetworkAdapters/"},"PCIDevices":{"href":"/redfish/v1/Systems/1/PCIDevices/"},"PCISlots":{"href":"/redfish/v1/Systems/1/PCISlots/"},"SUT":{"href":"/redfish/v1/systems/1/hpsut/"},"SecureBoot":{"href":"/redfish/v1/Systems/1/SecureBoot/"},"SmartStorage":{"href":"/redfish/v1/Systems/1/SmartStorage/"},"SoftwareInventory":{"href":"/redfish/v1/Systems/1/SoftwareInventory/"}}}},"Power":"On","PowerState":"On","ProcessorSummary":{"Count":2,"Model":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"Processors":{"Count":2,"ProcessorFamily":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Status":{"HealthRollUp":"OK"}},"SKU":"813198-B21","SerialNumber":"CZ3629FY8B","Status":{"Health":"OK","State":"Enabled"},"SystemType":"Physical","Type":"ComputerSystem.1.0.1","UUID":"31333138-3839-5A43-3336-323946593842","links":{"Chassis":[{"href":"/redfish/v1/Chassis/1/"}],"Logs":{"href":"/redfish/v1/Systems/1/LogServices/"},"ManagedBy":[{"href":"/redfish/v1/Managers/1/"}],"Processors":{"href":"/redfish/v1/Systems/1/Processors/"},"self":{"href":"/redfish/v1/Systems/1/"}}}`),
			RFPower:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Chassis/Members/1/Power$entity","@odata.id":"/redfish/v1/Chassis/1/Power/","@odata.type":"#Power.1.0.1.Power","Id":"Power","Name":"PowerMetrics","Oem":{"Hp":{"@odata.type":"#HpPowerMetricsExt.1.2.0.HpPowerMetricsExt","SNMPPowerThresholdAlert":{"DurationInMin":0,"ThresholdWatts":0,"Trigger":"Disabled"},"Type":"HpPowerMetricsExt.1.2.0","links":{"FastPowerMeter":{"href":"/redfish/v1/Chassis/1/Power/FastPowerMeter/"},"FederatedGroupCapping":{"href":"/redfish/v1/Chassis/1/Power/FederatedGroupCapping/"},"PowerMeter":{"href":"/redfish/v1/Chassis/1/Power/PowerMeter/"}}}},"PowerAllocatedWatts":189,"PowerAvailableWatts":311,"PowerCapacityWatts":500,"PowerConsumedWatts":73,"PowerControl":[{"PowerAllocatedWatts":189,"PowerAvailableWatts":311,"PowerCapacityWatts":500,"PowerConsumedWatts":73,"PowerLimit":{"LimitInWatts":null},"PowerMetrics":{"AverageConsumedWatts":87,"IntervalInMin":20,"MaxConsumedWatts":126,"MinConsumedWatts":65},"PowerRequestedWatts":189}],"PowerLimit":{"LimitInWatts":null},"PowerMetrics":{"AverageConsumedWatts":87,"IntervalInMin":20,"MaxConsumedWatts":126,"MinConsumedWatts":65},"PowerRequestedWatts":189,"Type":"PowerMetrics.0.11.0","links":{"self":{"href":"/redfish/v1/Chassis/1/Power/"}}}`),
			RFThermal:    []byte(`{"@odata.context":"/redfish/v1/$metadata#Chassis/Members/1/Thermal$entity","@odata.id":"/redfish/v1/Chassis/1/Thermal/","@odata.type":"#Thermal.1.1.0.Thermal","Fans":[{"CurrentReading":81,"FanName":"Fan 1","Oem":{"Hp":{"@odata.type":"#HpServerFan.1.0.0.HpServerFan","Location":"Virtual","Type":"HpServerFan.1.0.0"}},"Status":{"Health":"OK","State":"Enabled"},"Units":"Percent"}],"Id":"Thermal","Name":"Thermal","Temperatures":[{"CurrentReading":18,"LowerThresholdCritical":46,"LowerThresholdNonCritical":42,"Name":"01-Inlet Ambient","Number":1,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":9,"LocationYmm":0,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"Intake","ReadingCelsius":18,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":40,"LowerThresholdCritical":0,"LowerThresholdNonCritical":70,"Name":"02-CPU 1","Number":2,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":10,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"CPU","ReadingCelsius":40,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":40,"LowerThresholdCritical":0,"LowerThresholdNonCritical":70,"Name":"03-CPU 2","Number":3,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":6,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"CPU","ReadingCelsius":40,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":23,"LowerThresholdCritical":0,"LowerThresholdNonCritical":89,"Name":"04-P1 DIMM 1-4","Number":4,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":2,"LocationYmm":9,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":23,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"05-P1 DIMM 5-8","Number":5,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":2,"LocationYmm":5,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":22,"LowerThresholdCritical":0,"LowerThresholdNonCritical":89,"Name":"06-P2 DIMM 1-4","Number":6,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":13,"LocationYmm":5,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":22,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"07-P2 DIMM 5-8","Number":7,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":13,"LocationYmm":10,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":22,"LowerThresholdCritical":95,"LowerThresholdNonCritical":90,"Name":"08-P1 Mem Zone","Number":8,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":1,"LocationYmm":7,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":22,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":22,"LowerThresholdCritical":95,"LowerThresholdNonCritical":90,"Name":"09-P2 Mem Zone","Number":9,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":14,"LocationYmm":7,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":22,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":40,"LowerThresholdCritical":0,"LowerThresholdNonCritical":60,"Name":"10-HD Max","Number":10,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":2,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":40,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"11-Exp Bay Drive","Number":11,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":2,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":26,"LowerThresholdCritical":0,"LowerThresholdNonCritical":105,"Name":"12-Chipset","Number":12,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":2,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":26,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":26,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"13-VR P1","Number":13,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":9,"LocationYmm":12,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":26,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":26,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"14-VR P2","Number":14,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":8,"LocationYmm":3,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":26,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":22,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"15-VR P1 Mem-1","Number":15,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":4,"LocationYmm":12,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":22,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":20,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"16-VR P1 Mem-2","Number":16,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":14,"LocationYmm":13,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":20,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":18,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"17-VR P2 Mem-1","Number":17,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":11,"LocationYmm":2,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":18,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":21,"LowerThresholdCritical":0,"LowerThresholdNonCritical":115,"Name":"18-VR P2 Mem-2","Number":18,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":1,"LocationYmm":1,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":21,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":20,"LowerThresholdCritical":0,"LowerThresholdNonCritical":65,"Name":"19-Storage Batt","Number":19,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":13,"LocationYmm":10,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":20,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":47,"LowerThresholdCritical":0,"LowerThresholdNonCritical":100,"Name":"20-HD Controller","Number":20,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":8,"LocationYmm":6,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":47,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":21,"LowerThresholdCritical":75,"LowerThresholdNonCritical":70,"Name":"21-HDCntlr Inlet","Number":21,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":7,"LocationYmm":4,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":21,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"22-Mezz 1","Number":22,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":8,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"23-Mezz 1 Inlet","Number":23,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":8,"LocationYmm":13,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"24-Mezz 2","Number":24,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":5,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"25-Mezz 2 Inlet","Number":25,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":5,"LocationYmm":13,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":40,"LowerThresholdCritical":0,"LowerThresholdNonCritical":100,"Name":"26-LOM Card","Number":26,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":10,"LocationYmm":13,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":40,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":0,"LowerThresholdCritical":0,"LowerThresholdNonCritical":0,"Name":"27-LOMCard Inlet","Number":27,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":10,"LocationYmm":15,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":0,"Status":{"State":"Absent"},"Units":"Celsius"},{"CurrentReading":23,"LowerThresholdCritical":95,"LowerThresholdNonCritical":90,"Name":"28-I/O Zone-1","Number":28,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":14,"LocationYmm":15,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":23,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":23,"LowerThresholdCritical":95,"LowerThresholdNonCritical":90,"Name":"29-I/O Zone-2","Number":29,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":10,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":23,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":23,"LowerThresholdCritical":95,"LowerThresholdNonCritical":90,"Name":"30-I/O Zone-3","Number":30,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":11,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"SystemBoard","ReadingCelsius":23,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":25,"LowerThresholdCritical":85,"LowerThresholdNonCritical":80,"Name":"31-Sys Exhaust-1","Number":31,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":10,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"Exhaust","ReadingCelsius":25,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"},{"CurrentReading":34,"LowerThresholdCritical":85,"LowerThresholdNonCritical":80,"Name":"32-Sys Exhaust-2","Number":32,"Oem":{"Hp":{"@odata.type":"#HpSeaOfSensors.1.0.0.HpSeaOfSensors","LocationXmm":1,"LocationYmm":14,"Type":"HpSeaOfSensors.1.0.0"}},"PhysicalContext":"Exhaust","ReadingCelsius":34,"Status":{"Health":"OK","State":"Enabled"},"Units":"Celsius"}],"Type":"ThermalMetrics.0.10.0","links":{"self":{"href":"/redfish/v1/Chassis/1/Thermal/"}}}`),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Systems/Members/1/Processors/Members/$entity","@odata.id":"/redfish/v1/Systems/1/Processors/1/","@odata.type":"#Processor.1.0.0.Processor","Id":"1","InstructionSet":"x86-64","Manufacturer":"Intel(R) Corporation","MaxSpeedMHz":4000,"Model":"Intel(R) Xeon(R) CPU E5-2620 v4 @ 2.10GHz","Name":"Processors","Oem":{"Hp":{"@odata.type":"#HpProcessorExt.1.0.0.HpProcessorExt","AssetTag":"UNKNOWN","Cache":[{"Associativity":"8waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":512,"Location":"Internal","MaximumSizeKB":512,"Name":"L1-Cache","Policy":"WriteBack","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"},{"Associativity":"8waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":2048,"Location":"Internal","MaximumSizeKB":2048,"Name":"L2-Cache","Policy":"Varies","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"},{"Associativity":"20waySetAssociative","CacheSpeedns":0,"CurrentSRAMType":["Synchronous"],"EccType":"SingleBitECC","InstalledSizeKB":20480,"Location":"Internal","MaximumSizeKB":20480,"Name":"L3-Cache","Policy":"Varies","Socketed":false,"SupportedSRAMType":["Synchronous"],"SystemCacheType":"Unified"}],"Characteristics":["64Bit","MultiCore","HwThread","ExecuteProtection","EnhancedVirtualization","PowerPerfControl"],"ConfigStatus":{"Populated":true,"State":"Enabled"},"CoresEnabled":8,"ExternalClockMHz":100,"MicrocodePatches":[{"CpuId":"0x000306F2","Date":"2016-10-07T00:00:00Z","PatchId":"0x00000039"},{"CpuId":"0x000406F1","Date":"2016-10-07T00:00:00Z","PatchId":"0x0B00001F"}],"PartNumber":"","RatedSpeedMHz":2100,"SerialNumber":"","Type":"HpProcessorExt.1.0.0","VoltageVoltsX10":18}},"ProcessorArchitecture":"x86","ProcessorId":{"EffectiveFamily":"179","EffectiveModel":"15","IdentificationRegisters":"0x06f10004fbffbfeb","MicrocodeInfo":null,"Step":"1","VendorId":"Intel"},"ProcessorType":"CPU","Socket":"Proc 1","Status":{"Health":"OK"},"TotalCores":8,"TotalThreads":16,"Type":"Processor.1.0.0","links":{"self":{"href":"/redfish/v1/Systems/1/Processors/1/"}}}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Managers/Members/$entity","@odata.id":"/redfish/v1/Managers/1/","@odata.type":"#Manager.1.0.0.Manager","Actions":{"#Manager.Reset":{"target":"/redfish/v1/Managers/1/Actions/Manager.Reset/"}},"AvailableActions":[{"Action":"Reset"}],"CommandShell":{"ConnectTypesSupported":["SSH","Oem"],"Enabled":true,"MaxConcurrentSessions":9,"ServiceEnabled":true},"Description":"Manager View","EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/1/EthernetInterfaces/"},"Firmware":{"Current":{"VersionString":"iLO 4 v2.54"}},"FirmwareVersion":"iLO 4 v2.54","GraphicalConsole":{"ConnectTypesSupported":["KVMIP"],"Enabled":true,"MaxConcurrentSessions":10,"ServiceEnabled":true},"Id":"1","LogServices":{"@odata.id":"/redfish/v1/Managers/1/LogServices/"},"ManagerType":"BMC","Name":"Manager","NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/1/NetworkService/"},"Oem":{"Hp":{"@odata.type":"#HpiLO.1.1.0.HpiLO","Actions":{"#HpiLO.ClearRestApiState":{"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.ClearRestApiState/"},"#HpiLO.ResetToFactoryDefaults":{"ResetType@Redfish.AllowableValues":["Default"],"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.ResetToFactoryDefaults/"},"#HpiLO.iLOFunctionality":{"target":"/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.iLOFunctionality/"}},"AvailableActions":[{"Action":"ClearRestApiState","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"ResetToFactoryDefaults","Capabilities":[{"AllowableValues":["Default"],"PropertyName":"ResetType"},{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]},{"Action":"iLOFunctionality","Capabilities":[{"AllowableValues":["/Oem/Hp"],"PropertyName":"Target"}]}],"ClearRestApiStatus":"DataPresent","FederationConfig":{"IPv6MulticastScope":"Site","MulticastAnnouncementInterval":600,"MulticastDiscovery":"Enabled","MulticastTimeToLive":5,"iLOFederationManagement":"Enabled"},"Firmware":{"Current":{"Date":"Jun 15 2017","DebugBuild":false,"MajorVersion":2,"MinorVersion":54,"Time":"","VersionString":"iLO 4 v2.54"}},"License":{"LicenseKey":"xxxxx-xxxxx-xxxxx-xxxxx-3DXMM","LicenseString":"iLO Advanced","LicenseType":"Perpetual"},"RequiredLoginForiLORBSU":false,"SerialCLISpeed":9600,"SerialCLIStatus":"EnabledAuthReq","Type":"HpiLO.1.1.0","VSPLogDownloadEnabled":false,"iLOSelfTestResults":[{"Notes":"","SelfTestName":"NVRAMData","Status":"OK"},{"Notes":"","SelfTestName":"NVRAMSpace","Status":"OK"},{"Notes":"Controller firmware revision  2.10.00  ","SelfTestName":"EmbeddedFlash/SDCard","Status":"OK"},{"Notes":"","SelfTestName":"EEPROM","Status":"OK"},{"Notes":"","SelfTestName":"HostRom","Status":"OK"},{"Notes":"","SelfTestName":"SupportedHost","Status":"OK"},{"Notes":"Version 1.0.9","SelfTestName":"PowerManagementController","Status":"Informational"},{"Notes":"ProLiant BL460c Gen9 System Programmable Logic Device version 0x17","SelfTestName":"CPLDPAL0","Status":"Informational"},{"Notes":"ProLiant BL460c Gen9 SAS Programmable Logic Device version 0x02","SelfTestName":"CPLDPAL1","Status":"Informational"}],"links":{"ActiveHealthSystem":{"href":"/redfish/v1/Managers/1/ActiveHealthSystem/"},"DateTimeService":{"href":"/redfish/v1/Managers/1/DateTime/"},"EmbeddedMediaService":{"href":"/redfish/v1/Managers/1/EmbeddedMedia/"},"FederationDispatch":{"extref":"/dispatch/"},"FederationGroups":{"href":"/redfish/v1/Managers/1/FederationGroups/"},"FederationPeers":{"href":"/redfish/v1/Managers/1/FederationPeers/"},"LicenseService":{"href":"/redfish/v1/Managers/1/LicenseService/"},"SecurityService":{"href":"/redfish/v1/Managers/1/SecurityService/"},"UpdateService":{"href":"/redfish/v1/Managers/1/UpdateService/"},"VSPLogLocation":{"extref":"/sol.log.gz/"}}}},"SerialConsole":{"ConnectTypesSupported":["SSH","IPMI","Oem"],"Enabled":true,"MaxConcurrentSessions":13,"ServiceEnabled":true},"Status":{"State":"Enabled"},"Type":"Manager.1.0.0","UUID":"1cf36323-33b6-50e8-a7b3-f58c1fea3f58","VirtualMedia":{"@odata.id":"/redfish/v1/Managers/1/VirtualMedia/"},"links":{"EthernetNICs":{"href":"/redfish/v1/Managers/1/EthernetInterfaces/"},"Logs":{"href":"/redfish/v1/Managers/1/LogServices/"},"ManagerForChassis":[{"href":"/redfish/v1/Chassis/1/"}],"ManagerForServers":[{"href":"/redfish/v1/Systems/1/"}],"NetworkService":{"href":"/redfish/v1/Managers/1/NetworkService/"},"VirtualMedia":{"href":"/redfish/v1/Managers/1/VirtualMedia/"},"self":{"href":"/redfish/v1/Managers/1/"}}}`),
			RFBMCNetwork: []byte(``),
		},
		Supermicro: map[string][]byte{
			RFEntry:      []byte(`{"@odata.context":"/redfish/v1/$metadata#ComputerSystem.ComputerSystem","@odata.type":"#ComputerSystem.ComputerSystem","@odata.id":"/redfish/v1/Systems/1","Id":"1","Name":"System","Description":"Description of server","Status":{"State":"Enabled","Health":"OK"},"SerialNumber":"","PartNumber":"","SystemType":"Physical","BiosVersion":"2.0","Manufacturer":"Supermicro","Model":"X10DFF-CTG","SKU":"Default string","UUID":"00000000-0000-0000-0000-0CC47AB721C4","ProcessorSummary":{"Count":16,"Model":"Intel(R) Xeon(R) processor","Status":{"State":"Enabled","Health":"OK"}},"MemorySummary":{"TotalSystemMemoryGiB":128,"Status":{"State":"Enabled","Health":"OK"}},"IndicatorLED":"Off","PowerState":"On","Boot":{"BootSourceOverrideEnabled":"Disabled","BootSourceOverrideTarget":"None","BootSourceOverrideTarget@Redfish.AllowableValues":["None","Pxe","Hdd","Diags","Cd","BiosSetup","FloppyRemovableMedia","UsbKey","UsbHdd","UsbFloppy","UsbCd","UefiUsbKey","UefiCd","UefiHdd","UefiUsbHdd","UefiUsbCd"]},"Processors":{"@odata.id":"/redfish/v1/Systems/1/Processors"},"Links":{"Chassis":[{"@odata.id":"/redfish/v1/Chassis/1"}],"ManagedBy":[{"@odata.id":"/redfish/v1/Managers/1"}],"Oem":{}},"Actions":{"#ComputerSystem.Reset":{"target":"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset","ResetType@Redfish.AllowableValues":["On","ForceOff","GracefulRestart","ForceRestart","Nmi","ForceOn"]}}}`),
			RFPower:      []byte(`{"@odata.context":"/redfish/v1/$metadata#Power.Power","@odata.type":"#Power.Power","@odata.id":"/redfish/v1/Chassis/1/Power","Id":"Power","Name":"Power","Voltages":[{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/0","Name":"12V","MemberID":"0","SensorNumber":48,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":12.06,"UpperThresholdNonCritical":12.94,"UpperThresholdCritical":13.26,"UpperThresholdFatal":13.390000000000001,"LowerThresholdNonCritical":11.369999999999999,"LowerThresholdCritical":11.18,"LowerThresholdFatal":10.99,"MinReadingRange":0.16,"MaxReadingRange":16.219999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/1","Name":"5VCC","MemberID":"1","SensorNumber":49,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":4.9699999999999998,"UpperThresholdNonCritical":5.3899999999999997,"UpperThresholdCritical":5.5499999999999998,"UpperThresholdFatal":5.5999999999999996,"LowerThresholdNonCritical":4.4800000000000004,"LowerThresholdCritical":4.2999999999999998,"LowerThresholdFatal":4.25,"MinReadingRange":0.16,"MaxReadingRange":6.79,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/2","Name":"3.3VCC","MemberID":"2","SensorNumber":50,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":3.3199999999999998,"UpperThresholdNonCritical":3.5499999999999998,"UpperThresholdCritical":3.6600000000000001,"UpperThresholdFatal":3.6899999999999999,"LowerThresholdNonCritical":2.96,"LowerThresholdCritical":2.8199999999999998,"LowerThresholdFatal":2.79,"MinReadingRange":0.089999999999999997,"MaxReadingRange":4.4199999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/3","Name":"VBAT","MemberID":"3","SensorNumber":51,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":3.1899999999999999,"UpperThresholdNonCritical":3.5099999999999998,"UpperThresholdCritical":3.6000000000000001,"UpperThresholdFatal":3.71,"LowerThresholdNonCritical":2.6099999999999999,"LowerThresholdCritical":2.4900000000000002,"LowerThresholdFatal":2.4100000000000001,"MinReadingRange":0,"MaxReadingRange":7.3899999999999997,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/4","Name":"Vcpu1","MemberID":"4","SensorNumber":52,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.8100000000000001,"UpperThresholdNonCritical":1.8999999999999999,"UpperThresholdCritical":2.0899999999999999,"UpperThresholdFatal":2.1099999999999999,"LowerThresholdNonCritical":1.3899999999999999,"LowerThresholdCritical":1.26,"LowerThresholdFatal":1.24,"MinReadingRange":0.14000000000000001,"MaxReadingRange":2.4399999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/5","Name":"Vcpu2","MemberID":"5","SensorNumber":54,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.8200000000000001,"UpperThresholdNonCritical":1.8999999999999999,"UpperThresholdCritical":2.0899999999999999,"UpperThresholdFatal":2.1099999999999999,"LowerThresholdNonCritical":1.3899999999999999,"LowerThresholdCritical":1.26,"LowerThresholdFatal":1.24,"MinReadingRange":0.14000000000000001,"MaxReadingRange":2.4399999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/6","Name":"VDIMMAB","MemberID":"6","SensorNumber":53,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.1899999999999999,"UpperThresholdNonCritical":1.3400000000000001,"UpperThresholdCritical":1.4199999999999999,"UpperThresholdFatal":1.4399999999999999,"LowerThresholdNonCritical":1.05,"LowerThresholdCritical":0.97999999999999998,"LowerThresholdFatal":0.94999999999999996,"MinReadingRange":0.080000000000000002,"MaxReadingRange":2.3700000000000001,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/7","Name":"VDIMMCD","MemberID":"7","SensorNumber":55,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.2,"UpperThresholdNonCritical":1.3400000000000001,"UpperThresholdCritical":1.4199999999999999,"UpperThresholdFatal":1.4399999999999999,"LowerThresholdNonCritical":1.05,"LowerThresholdCritical":0.97999999999999998,"LowerThresholdFatal":0.94999999999999996,"MinReadingRange":0.080000000000000002,"MaxReadingRange":2.3700000000000001,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/8","Name":"VDIMMEF","MemberID":"8","SensorNumber":58,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.1899999999999999,"UpperThresholdNonCritical":1.3400000000000001,"UpperThresholdCritical":1.4199999999999999,"UpperThresholdFatal":1.4399999999999999,"LowerThresholdNonCritical":1.05,"LowerThresholdCritical":0.97999999999999998,"LowerThresholdFatal":0.94999999999999996,"MinReadingRange":0.080000000000000002,"MaxReadingRange":2.3700000000000001,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/9","Name":"VDIMMGH","MemberID":"9","SensorNumber":59,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.1899999999999999,"UpperThresholdNonCritical":1.3400000000000001,"UpperThresholdCritical":1.4199999999999999,"UpperThresholdFatal":1.4399999999999999,"LowerThresholdNonCritical":1.05,"LowerThresholdCritical":0.97999999999999998,"LowerThresholdFatal":0.94999999999999996,"MinReadingRange":0.080000000000000002,"MaxReadingRange":2.3700000000000001,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/10","Name":"5VSB","MemberID":"10","SensorNumber":56,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":4.9500000000000002,"UpperThresholdNonCritical":5.3899999999999997,"UpperThresholdCritical":5.5499999999999998,"UpperThresholdFatal":5.5999999999999996,"LowerThresholdNonCritical":4.4800000000000004,"LowerThresholdCritical":4.2999999999999998,"LowerThresholdFatal":4.25,"MinReadingRange":0.14000000000000001,"MaxReadingRange":6.7699999999999996,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/11","Name":"3.3VSB","MemberID":"11","SensorNumber":57,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":3.21,"UpperThresholdNonCritical":3.5499999999999998,"UpperThresholdCritical":3.6600000000000001,"UpperThresholdFatal":3.6899999999999999,"LowerThresholdNonCritical":2.96,"LowerThresholdCritical":2.8199999999999998,"LowerThresholdFatal":2.79,"MinReadingRange":0.02,"MaxReadingRange":4.3499999999999996,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/12","Name":"1.5V PCH","MemberID":"12","SensorNumber":60,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.5,"UpperThresholdNonCritical":1.6399999999999999,"UpperThresholdCritical":1.6699999999999999,"UpperThresholdFatal":1.7,"LowerThresholdNonCritical":1.3999999999999999,"LowerThresholdCritical":1.3500000000000001,"LowerThresholdFatal":1.3200000000000001,"MinReadingRange":0.12,"MaxReadingRange":2.4199999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/13","Name":"1.2V BMC","MemberID":"13","SensorNumber":61,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.21,"UpperThresholdNonCritical":1.3400000000000001,"UpperThresholdCritical":1.3700000000000001,"UpperThresholdFatal":1.3999999999999999,"LowerThresholdNonCritical":1.0900000000000001,"LowerThresholdCritical":1.05,"LowerThresholdFatal":1.02,"MinReadingRange":0.050000000000000003,"MaxReadingRange":2.3399999999999999,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/Voltages/14","Name":"1.05V PCH","MemberID":"14","SensorNumber":62,"Status":{"State":"Enabled","Health":"OK"},"ReadingVolts":1.04,"UpperThresholdNonCritical":1.1899999999999999,"UpperThresholdCritical":1.22,"UpperThresholdFatal":1.25,"LowerThresholdNonCritical":0.93999999999999995,"LowerThresholdCritical":0.90000000000000002,"LowerThresholdFatal":0.87,"MinReadingRange":0.080000000000000002,"MaxReadingRange":2.3700000000000001,"PhysicalContext":"VoltageRegulator","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]}],"PowerControl":[{"@odata.id":"/redfish/v1/Chassis/1/Power#/PowerControl/0","Name":"System Power Control","MemberID":"0","PowerConsumedWatts":176,"PowerAvailableWatts":755,"PowerAllocatedWatts":755,"PowerMetrics":{"IntervalInMin":5,"MinConsumedWatts":176,"MaxConsumedWatts":180,"AverageConsumedWatts":178},"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1/Processors/1"},{"@odata.id":"/redfish/v1/Systems/1/Processors/2"}],"Status":{"State":"Enabled","Health":"OK"},"Oem":{}}],"PowerSupplies":[{"@odata.id":"/redfish/v1/Chassis/1/Power#/PowerSupplies/0","MemberID":"0","Name":"Power Supply Bay 1","Status":{"State":"Enabled","Health":"OK"},"Oem":{},"PowerSupplyType":"AC","LineInputVoltageType":"ACMidLine","LineInputVoltage":228,"LastPowerOutputWatts":400,"Model":"PWS-2K04A-1R","FirmwareVersion":"REV1.1","SerialNumber":"P2K4ACF35MB0473","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"}],"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Power#/Redundancy/0"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/PowerSupplies/1","MemberID":"1","Name":"Power Supply Bay 2","Status":{"State":"Enabled","Health":"OK"},"Oem":{},"PowerSupplyType":"AC","LineInputVoltageType":"ACMidLine","LineInputVoltage":229,"LastPowerOutputWatts":355,"Model":"PWS-2K04A-1R","FirmwareVersion":"REV1.1","SerialNumber":"P2K4ACF35MB0471","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"}],"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Power#/Redundancy/0"}]},{"@odata.id":"/redfish/v1/Chassis/1/Power#/PowerSupplies/2","MemberID":"2","Name":"Power Supply Bay 3","Status":{"State":"Absent"},"Oem":{}},{"@odata.id":"/redfish/v1/Chassis/1/Power#/PowerSupplies/3","MemberID":"3","Name":"Power Supply Bay 4","Status":{"State":"Absent"},"Oem":{}}],"Oem":{}}`),
			RFThermal:    []byte(`{"@odata.context":"/redfish/v1/$metadata#Thermal.Thermal","@odata.type":"#Thermal.Thermal","@odata.id":"/redfish/v1/Chassis/1/Thermal","Id":"Thermal","Name":"Thermal","Temperatures":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/0","MemberID":"0","Name":"CPU1 Temp","SensorNumber":1,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":35,"UpperThresholdNonCritical":82,"UpperThresholdCritical":87,"UpperThresholdFatal":87,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":-2,"MaxReadingRangeTemp":89,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/0"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/1","MemberID":"1","Name":"CPU2 Temp","SensorNumber":2,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":34,"UpperThresholdNonCritical":82,"UpperThresholdCritical":87,"UpperThresholdFatal":87,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":-2,"MaxReadingRangeTemp":89,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/2","MemberID":"2","Name":"PCH Temp","SensorNumber":10,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":25,"UpperThresholdNonCritical":90,"UpperThresholdCritical":95,"UpperThresholdFatal":100,"LowerThresholdNonCritical":10,"LowerThresholdCritical":5,"LowerThresholdFatal":0,"MinReadingRangeTemp":-2,"MaxReadingRangeTemp":102,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/3","MemberID":"3","Name":"System Temp","SensorNumber":11,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":19,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":0,"LowerThresholdCritical":-5,"LowerThresholdFatal":-10,"MinReadingRangeTemp":-12,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/4","MemberID":"4","Name":"Peripheral Temp","SensorNumber":12,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":21,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":0,"LowerThresholdCritical":-5,"LowerThresholdFatal":-10,"MinReadingRangeTemp":-12,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/5","MemberID":"5","Name":"10GLAN Temp","SensorNumber":13,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":29,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/6","MemberID":"6","Name":"Vcpu1VRM Temp","SensorNumber":16,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":31,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/7","MemberID":"7","Name":"Vcpu2VRM Temp","SensorNumber":17,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":31,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/8","MemberID":"8","Name":"VDIMMABVRM Temp","SensorNumber":18,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":31,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/9","MemberID":"9","Name":"VDIMMCDVRM Temp","SensorNumber":19,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":24,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/10","MemberID":"10","Name":"VDIMMEFVRM Temp","SensorNumber":20,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":29,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/11","MemberID":"11","Name":"VDIMMGHVRM Temp","SensorNumber":21,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":30,"UpperThresholdNonCritical":95,"UpperThresholdCritical":100,"UpperThresholdFatal":105,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":107,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Chassis/1"},{"@odata.id":"/redfish/v1/Systems/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/12","MemberID":"12","Name":"P1-DIMMA1 Temp","SensorNumber":176,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":24,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/13","MemberID":"13","Name":"P1-DIMMA2 Temp","SensorNumber":177,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/14","MemberID":"14","Name":"P1-DIMMB1 Temp","SensorNumber":180,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":23,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/15","MemberID":"15","Name":"P1-DIMMB2 Temp","SensorNumber":181,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/16","MemberID":"16","Name":"P1-DIMMC1 Temp","SensorNumber":184,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":24,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/17","MemberID":"17","Name":"P1-DIMMC2 Temp","SensorNumber":185,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/18","MemberID":"18","Name":"P1-DIMMD1 Temp","SensorNumber":188,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":23,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/19","MemberID":"19","Name":"P1-DIMMD2 Temp","SensorNumber":189,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/20","MemberID":"20","Name":"P2-DIMME1 Temp","SensorNumber":208,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":24,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/21","MemberID":"21","Name":"P2-DIMME2 Temp","SensorNumber":209,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/22","MemberID":"22","Name":"P2-DIMMF1 Temp","SensorNumber":212,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":24,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/23","MemberID":"23","Name":"P2-DIMMF2 Temp","SensorNumber":213,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/24","MemberID":"24","Name":"P2-DIMMG1 Temp","SensorNumber":216,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":26,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/25","MemberID":"25","Name":"P2-DIMMG2 Temp","SensorNumber":217,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/26","MemberID":"26","Name":"P2-DIMMH1 Temp","SensorNumber":220,"Status":{"State":"Enabled","Health":"OK"},"ReadingCelsius":28,"UpperThresholdNonCritical":80,"UpperThresholdCritical":85,"UpperThresholdFatal":90,"LowerThresholdNonCritical":5,"LowerThresholdCritical":0,"LowerThresholdFatal":-5,"MinReadingRangeTemp":-7,"MaxReadingRangeTemp":92,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Temperatures/27","MemberID":"27","Name":"P2-DIMMH2 Temp","SensorNumber":221,"Status":{"State":"Absent"},"ReadingCelsius":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRangeTemp":0,"MaxReadingRangeTemp":0,"PhysicalContext":"CPU","RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1#/Processors/2"}]}],"Fans":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/0","MemberID":"0","FanName":"FAN1","PhysicalContext":"Backplane","Status":{"State":"Enabled","Health":"OK"},"ReadingUnits":"RPM","Reading":4600,"UpperThresholdNonCritical":25300,"UpperThresholdCritical":25400,"UpperThresholdFatal":25500,"LowerThresholdNonCritical":700,"LowerThresholdCritical":500,"LowerThresholdFatal":300,"MinReadingRange":200,"MaxReadingRange":25600,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/1","MemberID":"1","FanName":"FAN2","PhysicalContext":"Backplane","Status":{"State":"Enabled","Health":"OK"},"ReadingUnits":"RPM","Reading":4600,"UpperThresholdNonCritical":25300,"UpperThresholdCritical":25400,"UpperThresholdFatal":25500,"LowerThresholdNonCritical":700,"LowerThresholdCritical":500,"LowerThresholdFatal":300,"MinReadingRange":200,"MaxReadingRange":25600,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/2","MemberID":"2","FanName":"FAN3","PhysicalContext":"Backplane","Status":{"State":"Enabled","Health":"OK"},"ReadingUnits":"RPM","Reading":4600,"UpperThresholdNonCritical":25300,"UpperThresholdCritical":25400,"UpperThresholdFatal":25500,"LowerThresholdNonCritical":700,"LowerThresholdCritical":500,"LowerThresholdFatal":300,"MinReadingRange":200,"MaxReadingRange":25600,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/3","MemberID":"3","FanName":"FAN4","PhysicalContext":"Backplane","Status":{"State":"Enabled","Health":"OK"},"ReadingUnits":"RPM","Reading":4600,"UpperThresholdNonCritical":25300,"UpperThresholdCritical":25400,"UpperThresholdFatal":25500,"LowerThresholdNonCritical":700,"LowerThresholdCritical":500,"LowerThresholdFatal":300,"MinReadingRange":200,"MaxReadingRange":25600,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/4","MemberID":"4","FanName":"FAN5","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/5","MemberID":"5","FanName":"FAN6","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/6","MemberID":"6","FanName":"FAN7","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/7","MemberID":"7","FanName":"FAN8","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/8","MemberID":"8","FanName":"FAN9","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]},{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Fans/9","MemberID":"9","FanName":"FAN10","PhysicalContext":"Backplane","Status":{"State":"Absent"},"ReadingUnits":"RPM","Reading":0,"UpperThresholdNonCritical":0,"UpperThresholdCritical":0,"UpperThresholdFatal":0,"LowerThresholdNonCritical":0,"LowerThresholdCritical":0,"LowerThresholdFatal":0,"MinReadingRange":0,"MaxReadingRange":0,"Redundancy":[{"@odata.id":"/redfish/v1/Chassis/1/Thermal#/Redundancy/0"}],"RelatedItem":[{"@odata.id":"/redfish/v1/Systems/1"},{"@odata.id":"/redfish/v1/Chassis/1"}]}]}`),
			RFCPU:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Processor.Processor","@odata.type":"#Processor.Processor","@odata.id":"/redfish/v1/Systems/1/Processors/1","Id":"1","Name":"Processor","Description":"Processor","Socket":"CPU1","Manufacturer":"Intel","Model":"Intel(R) Xeon(R) CPU E5-2630 v3 @ 2.40GHz","MaxSpeedMHz":4000,"TotalCores":8,"TotalThreads":16,"ProcessorType":"CPU","ProcessorArchitecture":"x86","InstructionSet":"x86-64","ProcessorId":{"VendorId":"GenuineIntel","IdentificationRegisters":"0xBFEBFBFF000306F2","EffectiveFamily":"0x6","EffectiveModel":"0x3F","Step":"0x2"},"Status":{"State":"Enabled","Health":"OK"}}`),
			RFCPUEntry:   []byte(`{"@odata.context":"/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection","@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors","@odata.type":"#ProcessorCollection.ProcessorCollection","Description":"Collection of Processors for this System","Members":[{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"},{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"}],"Members@odata.count":2,"Name":"ProcessorsCollection"}`),
			RFBMC:        []byte(`{"@odata.context":"/redfish/v1/$metadata#Manager.Manager","@odata.type":"#Manager.Manager","@odata.id":"/redfish/v1/Managers/1","Id":"1","Name":"Manager","Description":"BMC","ManagerType":"BMC","UUID":"00000000-0000-0000-0000-0CC47AB982F7","Model":"ASPEED","FirmwareVersion":"3.25","DateTime":"2017-09-07T14:49:19+00:00","DateTimeLocalOffset":"+00:00","Status":{"State":"Enabled","Health":"OK"},"GraphicalConsole":{"ServiceEnabled":true,"MaxConcurrentSessions":4,"ConnectTypesSupported":["KVMIP"]},"SerialConsole":{"ServiceEnabled":true,"MaxConcurrentSessions":1,"ConnectTypesSupported":["SSH","IPMI"]},"CommandShell":{"ServiceEnabled":true,"MaxConcurrentSessions":0,"ConnectTypesSupported":["SSH"]},"EthernetInterfaces":{"@odata.id":"/redfish/v1/Managers/1/EthernetInterfaces"},"SerialInterfaces":{"@odata.id":"/redfish/v1/Managers/1/SerialInterfaces"},"NetworkProtocol":{"@odata.id":"/redfish/v1/Managers/1/NetworkProtocol"},"LogServices":{"@odata.id":"/redfish/v1/Managers/1/LogServices"},"VirtualMedia":{"@odata.id":"/redfish/v1/Managers/1/VM1"},"Links":{"ManagerForServers":[{"@odata.id":"/redfish/v1/Systems/1"}],"ManagerForChassis":[{"@odata.id":"/redfish/v1/Chassis/1"}],"Oem":{}},"Actions":{"Oem":{"#ManagerConfig.Reset":{"target":"/redfish/v1/Managers/1/Actions/Oem/ManagerConfig.Reset"}},"#Manager.Reset":{"target":"/redfish/v1/Managers/1/Actions/Manager.Reset"}},"Oem":{"ActiveDirectory":{"@odata.id":"/redfish/v1/Managers/1/ActiveDirectory"},"SMTP":{"@odata.id":"/redfish/v1/Managers/1/SMTP"},"RADIUS":{"@odata.id":"/redfish/v1/Managers/1/RADIUS"},"MouseMode":{"@odata.id":"/redfish/v1/Managers/1/MouseMode"},"NTP":{"@odata.id":"/redfish/v1/Managers/1/NTP"},"LDAP":{"@odata.id":"/redfish/v1/Managers/1/LDAP"},"IPAccessControl":{"@odata.id":"/redfish/v1/Managers/1/IPAccessControl"},"SMCRAKP":{"@odata.id":"/redfish/v1/Managers/1/SMCRAKP"},"SNMP":{"@odata.id":"/redfish/v1/Managers/1/SNMP"},"Syslog":{"@odata.id":"/redfish/v1/Managers/1/Syslog"},"Snooping":{"@odata.id":"/redfish/v1/Managers/1/Snooping"},"FanMode":{"@odata.id":"/redfish/v1/Managers/1/FanMode"}}}`),
			RFBMCNetwork: []byte(``),
		},
	}
)

// Serial() (string, error)

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

func TestRedfishStatus(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"Status",
			HP,
			RFEntry,
			"OK",
			"iLO",
		},
		{
			"Status",
			Dell,
			RFEntry,
			"OK",
			"iDRAC",
		},
		{
			"Status",
			Supermicro,
			RFEntry,
			"OK",
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

func TestRedfishModel(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"Model",
			HP,
			RFEntry,
			"ProLiant BL460c Gen9",
			"iLO",
		},
		{
			"Model",
			Dell,
			RFEntry,
			"PowerEdge M630",
			"iDRAC",
		},
		{
			"Model",
			Supermicro,
			RFEntry,
			"X10DFF-CTG",
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

func TestRedfishName(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  string
		detectionString string
	}{
		{
			"Name",
			HP,
			RFEntry,
			"bbmi",
			"iLO",
		},
		{
			"Name",
			Dell,
			RFEntry,
			"machine.example.com",
			"iDRAC",
		},
		{
			"Name",
			Supermicro,
			RFEntry,
			"",
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

func TestRedfishPowerKw(t *testing.T) {
	tt := []struct {
		testType        string
		vendor          string
		redfishendpoint string
		expectedAnswer  float64
		detectionString string
	}{
		{
			"PowerKw",
			HP,
			RFPower,
			0.073,
			"iLO",
		},
		{
			"PowerKw",
			Dell,
			RFPower,
			0.121,
			"iDRAC",
		},
		{
			"PowerKw",
			Supermicro,
			RFPower,
			0.176,
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
