package connectors

import (
	"errors"
	"regexp"
)

/*

HP RedFish Root

{
  "@odata.context": "/redfish/v1/$metadata#ServiceRoot",
  "@odata.id": "/redfish/v1/",
  "@odata.type": "#ServiceRoot.1.0.0.ServiceRoot",
  "AccountService": {
    "@odata.id": "/redfish/v1/AccountService/"
  },
  "Chassis": {
    "@odata.id": "/redfish/v1/Chassis/"
  },
  "EventService": {
    "@odata.id": "/redfish/v1/EventService/"
  },
  "Id": "v1",
  "JsonSchemas": {
    "@odata.id": "/redfish/v1/Schemas/"
  },
  "Managers": {
    "@odata.id": "/redfish/v1/Managers/"
  },
  "Name": "HP RESTful Root Service",
  "Oem": {
    "Hp": {
      "@odata.type": "#HpiLOServiceExt.1.0.0.HpiLOServiceExt",
      "Manager": [
        {
          "Blade": {
            "BayNumber": "Bay 12",
            "EnclosureName": "spare-2sn70305f2",
            "RackName": "UnnamedRack"
          },
          "DefaultLanguage": "en",
          "FQDN": "bkbuild-901.las3.lom.booking.com",
          "HostName": "bkbuild-901",
          "Languages": [
            {
              "Language": "en",
              "TranslationName": "English",
              "Version": ""
            }
          ],
          "ManagerFirmwareVersion": "2.54",
          "ManagerType": "iLO 4"
        }
      ],
      "Sessions": {
        "CertCommonName": "bkbuild-901.las3.lom.booking.com",
        "KerberosEnabled": false,
        "LDAPAuthLicenced": true,
        "LDAPEnabled": false,
        "LocalLoginEnabled": true,
        "LoginFailureDelay": 0,
        "LoginHint": {
          "Hint": "POST to /Sessions to login using the following JSON object:",
          "HintPOSTData": {
            "Password": "password",
            "UserName": "username"
          }
        },
        "SecurityOverride": false,
        "ServerName": "bkbuild-901.las3.example.com"
      },
      "Type": "HpiLOServiceExt.1.0.0",
      "links": {
        "ResourceDirectory": {
          "href": "/redfish/v1/ResourceDirectory/"
        }
      }
    }
  },
  "RedfishVersion": "1.0.0",
  "Registries": {
    "@odata.id": "/redfish/v1/Registries/"
  },
  "ServiceVersion": "1.0.0",
  "SessionService": {
    "@odata.id": "/redfish/v1/SessionService/"
  },
  "Systems": {
    "@odata.id": "/redfish/v1/Systems/"
  },
  "Time": "2017-08-09T16:23:37Z",
  "Type": "ServiceRoot.1.0.0",
  "UUID": "c2b6084b-1db5-584e-8bd6-5c7d9f18a699",
  "links": {
    "AccountService": {
      "href": "/redfish/v1/AccountService/"
    },
    "Chassis": {
      "href": "/redfish/v1/Chassis/"
    },
    "EventService": {
      "href": "/redfish/v1/EventService/"
    },
    "Managers": {
      "href": "/redfish/v1/Managers/"
    },
    "Registries": {
      "href": "/redfish/v1/Registries/"
    },
    "Schemas": {
      "href": "/redfish/v1/Schemas/"
    },
    "SessionService": {
      "href": "/redfish/v1/SessionService/"
    },
    "Sessions": {
      "href": "/redfish/v1/SessionService/Sessions/"
    },
    "Systems": {
      "href": "/redfish/v1/Systems/"
    },
    "self": {
      "href": "/redfish/v1/"
    }
  }
}

Dell RedFish Root

{
  "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
  "@odata.id": "/redfish/v1",
  "@odata.type": "#ServiceRoot.v1_0_2.ServiceRoot",
  "AccountService": {
    "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/AccountService"
  },
  "Chassis": {
    "@odata.id": "/redfish/v1/Chassis"
  },
  "Description": "Root Service",
  "EventService": {
    "@odata.id": "/redfish/v1/EventService"
  },
  "Id": "RootService",
  "JsonSchemas": {
    "@odata.id": "/redfish/v1/JSONSchemas"
  },
  "Links": {
    "Sessions": {
      "@odata.id": "/redfish/v1/Sessions"
    }
  },
  "Managers": {
    "@odata.id": "/redfish/v1/Managers"
  },
  "Name": "Root Service",
  "RedfishVersion": "1.0.2",
  "Registries": {
    "@odata.id": "/redfish/v1/Registries"
  },
  "SessionService": {
    "@odata.id": "/redfish/v1/SessionService"
  },
  "Systems": {
    "@odata.id": "/redfish/v1/Systems"
  },
  "Tasks": {
    "@odata.id": "/redfish/v1/TaskService"
  }
}

Supermicro RedFish Root

{
  "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
  "@odata.type": "#ServiceRoot.ServiceRoot",
  "@odata.id": "/redfish/v1",
  "Id": "RootService",
  "Name": "Root Service",
  "RedfishVersion": "1.0.1",
  "UUID": "00000000-0000-0000-0000-0CC47A6DD04E",
  "Systems": {
    "@odata.id": "/redfish/v1/Systems"
  },
  "Chassis": {
    "@odata.id": "/redfish/v1/Chassis"
  },
  "Managers": {
    "@odata.id": "/redfish/v1/Managers"
  },
  "SessionService": {
    "@odata.id": "/redfish/v1/SessionService"
  },
  "AccountService": {
    "@odata.id": "/redfish/v1/AccountService"
  },
  "EventService": {
    "@odata.id": "/redfish/v1/EventService"
  },
  "Registries": {
    "@odata.id": "/redfish/v1/Registries"
  },
  "JsonSchemas": {
    "@odata.id": "/redfish/v1/JsonSchemas"
  },
  "Links": {
    "Sessions": {
      "@odata.id": "/redfish/v1/SessionService/Sessions"
    }
  },
  "Oem": {}
}

*/

var (
	ErrRedFishNotSupported = errors.New("RedFish not supported")
	redfishVendorEndPoints = map[string]map[string]string{
		Common: map[string]string{
			RFEntry: "redfish/v1", 
		}
		Dell: map[string]string{
			RFPower:   "redfish/v1/Chassis/System.Embedded.1/Power",
			RFThermal: "redfish/v1/Chassis/System.Embedded.1/Thermal",
		},
		HP: map[string]string{
			RFPower:   "rest/v1/Chassis/1/Power",
			RFThermal: "rest/v1/Chassis/1/Thermal",
		},
		Supermicro: map[string]string{
			RFPower:   "redfish/v1/Chassis/1/Power",
			RFThermal: "redfish/v1/Chassis/1/Thermal",
		},
	}
	redfishVendorLabels = map[string]map[string]string{
		Dell: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Board Inlet Temp",
		},
		HP: map[string]string{
			//			RFPower:   "PowerMetrics",
			RFThermal: "30-System Board",
		},
		Supermicro: map[string]string{
			RFPower:   "System Power Control",
			RFThermal: "System Temp",
		},
	}
	bmcAddressBuild = regexp.MustCompile(".(prod|corp|dqs).")
)

type RedFishPower struct {
	PowerControl []struct {
		Name               string  `json:"Name"`
		PowerConsumedWatts float64 `json:"PowerConsumedWatts"`
	} `json:"PowerControl"`
}

type RedFishThermal struct {
	Temperatures []struct {
		Name           string `json:"Name"`
		ReadingCelsius int    `json:"ReadingCelsius"`
	} `json:"Temperatures"`
}

type RedFishConnection struct {
	username string
	password string
}

func (c *RedFishConnection) Get(ip *string) (chassis model.Chassis, err error) {


	return 
}


