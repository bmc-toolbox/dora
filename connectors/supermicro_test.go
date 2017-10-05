package connectors

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

var (
	supermicroAnswers = map[string][]byte{
		"FRU_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <FRU_INFO RES="1">
				<DEVICE ID="0"/>
				<CHASSIS TYPE="1" PART_NUM="CSE-F414IS2-R2K04BP" SERIAL_NUM="CF414AF38N50003"/>
				<BOARD LAN="0" MFG_DATE="1996/01/01 00:00:00" PROD_NAME="X10DRFF-CTG" MFC_NAME="Supermicro" SERIAL_NUM="VM158S009467" PART_NUM="X10DRFF-CTG"/>
				<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="NONE" PART_NUM="SYS-F618H6-FTPTL+" VERSION="NONE" SERIAL_NUM="A19627226A05569" ASSET_TAG="NONE"/>
			  </FRU_INFO>
			</IPMI>`),
		"Get_PlatformCap.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <Platform Cap="8004c039" FanModeSupport="1b" LanModeSupport="7" EnPowerSupplyPage="81" EnStorage="0" EnECExpand="0" EnMultiNode="1" EnX10TwinProMCUUpdate="1" EnPCIeSSD="0" EnAtomHDD="0" EnLANByPassMode="0" EnDP="0" EnSMBIOS="1" SmartCoolCap="0" SmartCooling="0" EnHDDPwrCtrl="0" TwinType="a5" TwinNodeNumber="00" EnBigTwinLCMCCPLDUpdate="0" EnSmartPower="0"/>
			</IPMI>`),
		"GENERIC_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>  <IPMI>  <GENERIC_INFO>  <GENERIC BMC_IP="010.193.171.016" BMC_MAC="0c:c4:7a:b8:22:64" WEB_VERSION="1.1" IPMIFW_TAG="BL_SUPERMICRO_X7SB3_2017-05-23_B" IPMIFW_VERSION="0325" IPMIFW_BLDTIME="05/23/2017" SESSION_TIMEOUT="00" SDR_VERSION="0000" FRU_VERSION="0000" BIOS_VERSION="        " />  <KERNAL VERSION="2.6.28.9 "/>  </GENERIC_INFO>  </IPMI>`),
		"Get_PlatformInfo.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <PLATFORM_INFO MB_MAC_NUM="2" MB_MAC_ADDR1="0c:c4:7a:bc:dc:1a" MB_MAC_ADDR2="0c:c4:7a:bc:dc:1b" BIOS_VERSION="2.0" BIOS_VERSION_EXIST="1" BIOS_BUILD_DATE="12/17/2015" BIOS_BUILD_DATE_EXIST="1" CPLD_VERSION_EXIST="1" CPLD_VERSION="01.a1.02" REDFISH_REV="1.0.1">
				<HOST_AND_USER HOSTNAME="" BMC_IP="010.193.171.016" SESS_USER_NAME="Administrator" USER_ACCESS="04" DHCP6C_DUID="0E 00 00 01 00 01 20 FA 0E 90 0C C4 7A B8 22 64 "/>
			  </PLATFORM_INFO>
			</IPMI>`),
		"CONFIG_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <CONFIG_INFO>
				<TOTAL_NUMBER LAN="1" USER="a"/>
				<LAN BMC_IP="010.193.171.016" BMC_MAC="0c-c4-7a-b8-22-64" BMC_NETMASK="255.255.255.000" GATEWAY_IP="010.193.171.254" GATEWAY_MAC="0c-c4-7a-b8-22-64" VLAN_ID="0000" DHCP_TOUT="0" DHCP_EN="1" RMCP_PORT="026f"/>
				<USER NAME="                " USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="Administrator" USER_ACCESS="04" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<SERVICE DNS_ADDR="000.000.000.000" ALERT_EN="0" SMTP_SERVER=" " SMTP_PORT="587" MAIL_ADDR="0;0;0;0;0;0;0;0;0;0;0;0;0;0;0;0;" MAIL_USR=" " MAIL_PWD=" " SMTP_SSL="0"/>
				<LDAP LDAP_SSL="0" LDAP_IP="000.000.000.000" LDAP_EN="0" Encryption_EN="1" TIMEOUT="00" LDAP_PORT="00000" BASE_DN=" " BINDDN=" "/>
				<DNS DNS_SERVER="10.252.13.2"/>
				<LAN_IF INTERFACE="2"/>
				<HOSTNAME NAME="testserver"/>
				<DHCP6C DUID="0E 00 00 01 00 01 20 FA 0E 90 0C C4 7A B8 22 64 "/>
				<LINK_INFO MII_LINK_CONF="0" MII_AUTO_NEGOTIATION="0" MII_DUPLEX="1" MII_SPEED="2" MII_OPERSTATE="1" NCSI_AUTO_NEGOTIATION="0" NCSI_SPEED_AND_DUPLEX="0" NCSI_OPERSTATE="0" DEV_IF_MODE="2" BOND0_PORT="0"/>
			  </CONFIG_INFO>
			</IPMI>`),
		"SMBIOS_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <BIOS VENDOR="American Megatrends Inc." VER="2.0" REL_DATE="12/17/2015"/>
			  <SYSTEM MANUFACTURER="Supermicro" PN="SYS-F618H6-FTPTL+" SN="A19627226A05569" SKUN="Default string"/>
			  <CPU TYPE="03h" SPEED="2200 MHz" PROC_UPGRADE="2bh" CORE="10" CORE_ENABLED="10" SOCKET="CPU2" MANUFACTURER="Intel" VER="Intel(R) Xeon(R) CPU E5-2630 v4 @ 2.20GHz"/>
			  <CPU TYPE="03h" SPEED="2200 MHz" PROC_UPGRADE="2bh" CORE="10" CORE_ENABLED="10" SOCKET="CPU1" MANUFACTURER="Intel" VER="Intel(R) Xeon(R) CPU E5-2630 v4 @ 2.20GHz"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMH1" SN="10D12481" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel3_Dimm0" ASSET="P2-DIMMH1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMG1" SN="10D12494" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel2_Dimm0" ASSET="P2-DIMMG1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMF1" SN="10D1247D" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel1_Dimm0" ASSET="P2-DIMMF1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMME1" SN="10D12480" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel0_Dimm0" ASSET="P2-DIMME1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMD1" SN="10D12482" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel3_Dimm0" ASSET="P1-DIMMD1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMC1" SN="10D12520" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel2_Dimm0" ASSET="P1-DIMMC1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMB1" SN="10D1247E" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel1_Dimm0" ASSET="P1-DIMMB1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMA1" SN="10D12479" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel0_Dimm0" ASSET="P1-DIMMA1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="2000 Watts" GROUP="2" LOCATION="SLOT 2" SN="P2K4ACG22QT0165" PN="PWS-2K04A-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-2K04A-1R" REV="1.1"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="2000 Watts" GROUP="1" LOCATION="SLOT 1" SN="P2K4ACG22QT0168" PN="PWS-2K04A-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-2K04A-1R" REV="1.1"/>
			</IPMI>`),
	}
)

func smSetup() (r *SupermicroReader, err error) {
	viper.SetDefault("debug", false)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	mux.HandleFunc("/cgi/ipmi.cgi", func(w http.ResponseWriter, r *http.Request) {
		query, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(supermicroAnswers[string(query)])
	})

	mux.HandleFunc("/cgi/login.cgi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("../cgi/url_redirect.cgi?url_name=mainmenu"))
	})

	r, err = NewSupermicroReader(&ip, &username, &password)
	if err != nil {
		return r, err
	}

	return r, err
}

func smTeardown() {
	server.Close()
}

func TestSupermicroLogin(t *testing.T) {
	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	err = bmc.Login()
	if err != nil {
		t.Errorf("Unable to login: %v", err)
	}
}

func TestSupermicroSerial(t *testing.T) {
	expectedAnswer := "CF414AF38N50003@VM158S009467"

	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	answer, err := bmc.Serial()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestSupermicroModel(t *testing.T) {
	expectedAnswer := "X10DRFF-CTG"

	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	answer, err := bmc.Model()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestSupermicroBmcType(t *testing.T) {
	expectedAnswer := "Supermicro"

	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	answer, err := bmc.BmcType()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestSupermicroBmcVersion(t *testing.T) {
	expectedAnswer := "0325"

	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	answer, err := bmc.BmcVersion()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestSupermicroName(t *testing.T) {
	expectedAnswer := "testserver"

	bmc, err := smSetup()
	if err != nil {
		t.Fatalf("Found errors during the test smSetup %v", err)
	}

	answer, err := bmc.Name()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
