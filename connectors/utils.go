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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/publicsuffix"
)

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

func httpGetDell(hostname *string, username *string, password *string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "ChassisConnections", "hostname": *hostname}).Debug("Requesting data from BMC")

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

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return payload, err
	}

	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: tr,
		Jar:       jar,
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

	resp, err = client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/json?method=groupinfo", *hostname))
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
