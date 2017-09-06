package connectors

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"golang.org/x/net/publicsuffix"
)

var (
	// ErrLoginFailed is returned when we fail to login to a bmc
	ErrLoginFailed = errors.New("Failed to login")
	// ErrBiosNotFound is returned when we are not able to find the server bios version
	ErrBiosNotFound = errors.New("Bios version not found")
	// ErrVendorUnknown is returned when we are unable to identify the redfish vendor
	ErrVendorUnknown = errors.New("Unabled to identify the vendor")
	// ErrPageNotFound is used to inform the http request that we couldn't find the expected page and/or endpoint
	ErrPageNotFound = errors.New("Requested page couldn't be found in the server")
	// ErrRedFishNotSupported is returned when redfish isn't supported by the vendor
	ErrRedFishNotSupported = errors.New("RedFish not supported")
	// ErrRedFishEndPoint500 is retured when we receive 500 in a redfish api call and the bmc dies with the request
	ErrRedFishEndPoint500 = errors.New("We've received 500 calling this endpoint")
	// ErrUnabletoReadData is returned when we fail to read data from a chassis or bmc
	ErrUnabletoReadData = errors.New("Unable to read data from this device")
)

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func httpGet(url string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "url": url}).Debug("Requesting data from BMC")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payload, err
	}
	req.SetBasicAuth(*username, *password)
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

func buildClient() (client *http.Client, err error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 30 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return client, err
	}

	client = &http.Client{
		Timeout:   time.Second * 60,
		Transport: tr,
		Jar:       jar,
	}

	return client, err
}

func httpGetDell(hostname *string, endpoint string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections Dell", "hostname": *hostname}).Debug("Requesting data from BMC")

	form := url.Values{}
	form.Add("user", *username)
	form.Add("password", *password)

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", *hostname))
	if err != nil {
		return payload, err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return payload, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client, err := buildClient()
	if err != nil {
		return payload, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	auth, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	if strings.Contains(string(auth), "Try Again") {
		return nil, ErrLoginFailed
	}

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", *hostname, endpoint))
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", *hostname))
	if err != nil {
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	// Dell has a really shitty consistency of the data type returned, here we fix what's possible
	payload = bytes.Replace(payload, []byte("\"bladeTemperature\":-1"), []byte("\"bladeTemperature\":\"0\""), -1)
	payload = bytes.Replace(payload, []byte("\"nic\": [],"), []byte("\"nic\": {},"), -1)
	payload = bytes.Replace(payload, []byte("N\\/A"), []byte("0"), -1)

	return payload, err
}

func httpGetHP(hostname *string, endpoint string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections HP", "hostname": *hostname}).Debug("Requesting data from BMC")

	data := []byte(fmt.Sprintf("{\"method\":\"login\", \"user_login\":\"%s\", \"password\":\"%s\" }", *username, *password))
	u, err := url.Parse(fmt.Sprintf("https://%s/json/login_session", *hostname))
	if err != nil {
		return payload, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(data))
	if err != nil {
		return payload, err
	}
	req.Header.Set("Content-Type", "application/json")

	client, err := buildClient()
	if err != nil {
		return payload, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return payload, err
	}

	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	resp, err = client.Get(fmt.Sprintf("https://%s/%s", *hostname, endpoint))
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	data = []byte(`{"method":"logout"}`)

	req, err = http.NewRequest("POST", u.String(), bytes.NewBuffer(data))
	if err != nil {
		return payload, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		return payload, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return payload, ErrPageNotFound
	}

	return payload, err
}

func detectVendorAndType(hostname *string) (vendor string, deviceType string, err error) {
	log.WithFields(log.Fields{"step": "onnection", "hostname": *hostname}).Info("Detecting vendor")

	client, err := buildClient()
	if err != nil {
		return vendor, deviceType, err
	}

	resp, err := client.Get(fmt.Sprintf("https://%s/xmldata?item=all", *hostname))
	if err != nil {
		return vendor, deviceType, err
	}

	if resp.StatusCode == 200 {
		payload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return vendor, deviceType, err
		}

		iloXML := &HpRimpBlade{}
		err = xml.Unmarshal(payload, iloXML)
		if err != nil {
			fmt.Println(err)
		}

		if iloXML.HpBladeBlade != nil {
			return HP, Blade, err
		}
	}

	return vendor, deviceType, err
}

// DumpInvalidPayload is here to help identify unknown or broken payload messages
func DumpInvalidPayload(name string, payload []byte) (err error) {
	if viper.GetBool("dump_invalid_payloads") {
		log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name}).Info("Dump invalid payload")

		t := time.Now()
		timeStamp := t.Format("20060102150405")

		dumpPath := viper.GetString("dump_invalid_payload_path")
		err = os.MkdirAll(path.Join(dumpPath, name), 0755)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(path.Join(dumpPath, name, timeStamp), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name, "error": err}).Error("Dump invalid payload")
			return err
		}

		_, err = file.Write(payload)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name, "error": err}).Error("Dump invalid payload")
			return err
		}
		file.Sync()
		file.Close()
	}

	return err
}
