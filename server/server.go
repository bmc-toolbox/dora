package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

var (
	clru        = cache.New(60*time.Minute, 90*time.Minute)
	client      *http.Client
	authHeader  string
	serverDBUrl string
	// ErrPageNotFound is used to inform the http request that we couldn't find the expected page and/or endpoint
	ErrPageNotFound = errors.New("Requested page couldn't be found in the server")
)

type serverDBResponse struct {
	IPAddress string `json:"ip_address"`
	MacAddres string `json:"mac_address"`
}

func buildClient() (client *http.Client, err error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
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

func serverDBQuery(subnet *string, macAddress *string) (lease *string, err error) {
	key := fmt.Sprintf("%s-%s", *subnet, *macAddress)
	data, found := clru.Get(key)
	if found {
		lease := data.(*string)
		return lease, err
	}

	url := fmt.Sprintf("%s/api/v1/interface/dhcp_request/%s/%s/", serverDBUrl, *subnet, *macAddress)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return lease, err
	}
	req.Header.Add("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return lease, err
	}

	if resp.StatusCode == 404 {
		return lease, ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return lease, err
	}
	defer resp.Body.Close()

	sdbr := &serverDBResponse{}
	err = json.Unmarshal(payload, sdbr)
	if err != nil {
		return lease, err
	}

	clru.Set(key, &sdbr.IPAddress, cache.DefaultExpiration)
	return &sdbr.IPAddress, err
}

func Serve() {
	os.Remove(viper.GetString("socket_path"))

	authHeader = fmt.Sprintf("ApiKey %s:%s", viper.GetString("notify_api_user"), viper.GetString("notify_api_key"))
	serverDBUrl = viper.GetString("notify_url")

	var err error
	client, err = buildClient()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.GET("/kea/:subnet/:macAddress/", func(c *gin.Context) {
		subnet := c.Param("subnet")
		macAddress := c.Param("macAddress")
		data, err := serverDBQuery(&subnet, &macAddress)
		if err != nil {
			c.String(http.StatusForbidden, fmt.Sprintf("We got an error from ServerDB %s", err))
		}
		c.String(http.StatusOK, *data)
	})

	router.RunUnix(viper.GetString("socket_path"))
}
