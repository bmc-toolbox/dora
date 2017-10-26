package connectors

import "errors"

var (
	// ErrLoginFailed is returned when we fail to login to a bmc
	ErrLoginFailed = errors.New("Failed to login")
	// ErrBiosNotFound is returned when we are not able to find the server bios version
	ErrBiosNotFound = errors.New("Bios version not found")
	// ErrVendorUnknown is returned when we are unable to identify the redfish vendor
	ErrVendorUnknown = errors.New("Unable to identify the vendor")
	// ErrInvalidSerial is returned when the serial number for the device is invalid
	ErrInvalidSerial = errors.New("Unable to find the serial number")
	// ErrPageNotFound is used to inform the http request that we couldn't find the expected page and/or endpoint
	ErrPageNotFound = errors.New("Requested page couldn't be found in the server")
	// ErrRedFishNotSupported is returned when redfish isn't supported by the vendor
	ErrRedFishNotSupported = errors.New("RedFish not supported")
	// ErrRedFishEndPoint500 is retured when we receive 500 in a redfish api call and the bmc dies with the request
	ErrRedFishEndPoint500 = errors.New("We've received 500 calling this endpoint")
	// ErrUnableToReadData is returned when we fail to read data from a chassis or bmc
	ErrUnableToReadData = errors.New("Unable to read data from this device")
)
