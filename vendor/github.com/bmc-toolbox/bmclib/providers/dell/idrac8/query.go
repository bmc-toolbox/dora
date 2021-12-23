package idrac8

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/internal/helper"
)

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface
func (i *IDrac8) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {
	dialer := &net.Dialer{
		Timeout: time.Duration(10) * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", i.ip+":"+"443", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return []*x509.Certificate{{}}, true, err
	}

	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, true, nil
}

// Screenshot Grab screen preview.
func (i *IDrac8) Screenshot() (response []byte, extension string, err error) {
	err = i.httpLogin()
	if err != nil {
		return nil, "", err
	}

	endpoint := fmt.Sprintf("data?get=consolepreview[auto%%20%d]",
		time.Now().UnixNano()/int64(time.Millisecond))

	extension = "png"

	// here we expect an empty response
	statusCode, response, err := i.get(endpoint, &map[string]string{"idracAutoRefresh": "1"})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return nil, "", err
	}

	if !strings.Contains(string(response), "<status>ok</status>") {
		return nil, "", fmt.Errorf(string(response))
	}

	endpoint = fmt.Sprintf("capconsole/scapture0.png?%d",
		time.Now().UnixNano()/int64(time.Millisecond))

	statusCode, response, err = i.get(endpoint, &map[string]string{})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return nil, "", err
	}

	return response, extension, nil
}

// Queries for current user accounts.
func (i *IDrac8) queryUsers() (usersInfo UsersInfo, err error) {
	usersInfo = make(UsersInfo)

	endpoint := "data?get=user"

	statusCode, response, err := i.get(endpoint, &map[string]string{})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryUsers(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"StatusCode", statusCode,
			"step", helper.WhosCalling(),
		)
		return usersInfo, err
	}

	xmlData := XMLRoot{}
	err = xml.Unmarshal(response, &xmlData)
	if err != nil {
		i.log.V(1).Error(err, "queryUsers(): Unable to unmarshal payload.",
			"step", "queryUserInfo",
			"resource", "User",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return usersInfo, err
	}

	for _, userAccount := range xmlData.XMLUserAccount {
		user := UserInfo{
			UserName:  userAccount.Name,
			Privilege: strconv.Itoa(userAccount.Privileges),
		}

		switch userAccount.Privileges {
		case 511:
			user.IpmiLanPrivilege = "Administrator"
		case 499:
			user.IpmiLanPrivilege = "Operator"
		}

		if userAccount.SolEnabled == 1 {
			user.SolEnable = "Enabled"
		} else {
			user.SolEnable = "disabled"
		}

		if userAccount.Enabled == 1 {
			user.Enable = "Enabled"
		} else {
			user.Enable = "disabled"
		}

		usersInfo[userAccount.ID] = user
	}

	return usersInfo, err
}
