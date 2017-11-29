package connectors

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
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

// DumpInvalidPayload is here to help identify unknown or broken payload messages
func DumpInvalidPayload(name string, payload []byte) (err error) {
	// TODO(jumartinez): We need to also add the reference for this payload or it's useless
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

	log.WithFields(log.Fields{"operation": "notify", "callback": callback}).Debug("notifying ServerDB")

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

func standardizeProcessorName(name string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimSuffix(strings.Split(name, "@")[0], "0")))
}
