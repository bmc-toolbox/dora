package connectors

import (
	"bytes"
	"crypto/tls"
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

// DumpInvalidPayload is here to help identify unknown or broken payload messages
func DumpInvalidPayload(name string, payload []byte) (err error) {
	// TODO: We need to also add the reference for this payload or it's useless
	if viper.GetBool("collector.dump_invalid_payloads") {
		log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name}).Info("dump invalid payload")

		t := time.Now()
		timeStamp := t.Format("20060102150405")

		dumpPath := viper.GetString("collector.dump_invalid_payload_path")
		err = os.MkdirAll(path.Join(dumpPath, name), 0755)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(path.Join(dumpPath, name, timeStamp), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name, "error": err}).Error("dump invalid payload")
			return err
		}

		_, err = file.Write(payload)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "name": name, "error": err}).Error("dump invalid payload")
			return err
		}
		file.Sync()
		file.Close()
	}

	return err
}

func assetNotify(callback string) (err error) {
	client, err := buildClient()
	if err != nil {
		return err
	}

	authHeader := fmt.Sprintf("ApiKey %s:%s", viper.GetString("notify_api_user"), viper.GetString("notify_api_key"))
	serverDBUrl := viper.GetString("notify_url")
	payload := []byte(fmt.Sprintf(`{"callback": "%s"}`, callback))

	url := fmt.Sprintf("%s/api/v1/server/dora/dora_update_or_create/", serverDBUrl)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authHeader)

	log.WithFields(log.Fields{"operation": "notify", "callback": callback}).Info("notifying ServerDB")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error code: %d, message: %s", resp.StatusCode, string(response))
	}

	return err
}
