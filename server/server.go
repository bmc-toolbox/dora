package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

func serverDBQuery(subnet *string, macAddress *string) (lease string, err error) {
	key := fmt.Sprintf("%s-%s", *subnet, *macAddress)
	data, found := clru.Get(key)
	if found {
		lease := data.(string)
		_, err := clru.IncrementInt("CacheHit", 1)
		if err != nil {
			log.Println(err)
		}
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

	clru.Set(key, sdbr.IPAddress, cache.DefaultExpiration)
	_, err = clru.IncrementInt("CacheMiss", 1)
	if err != nil {
		log.Println(err)
	}

	return sdbr.IPAddress, err
}

// cacheStats returns that status of our cache
func cacheStats() (hit int, miss int, ratio float64) {
	if data, found := clru.Get("CacheHit"); found {
		hit = data.(int)
	}
	if data, found := clru.Get("CacheMiss"); found {
		miss = data.(int)
	}
	ratio = float64(hit) / (float64(hit) + float64(miss))
	return hit, miss, ratio
}

// Serve start and build the webservice binding on unix socket
func Serve() {
	os.Remove(viper.GetString("socket_path"))
	var err error
	cacheFile := viper.GetString("cache_file")

	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		err := clru.LoadFile(cacheFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	authHeader = fmt.Sprintf("ApiKey %s:%s", viper.GetString("notify_api_user"), viper.GetString("notify_api_key"))
	serverDBUrl = viper.GetString("notify_url")
	debug := viper.GetBool("debug")

	client, err = buildClient()
	if err != nil {
		log.Fatal(err)
	}

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	go func(clru *cache.Cache, cacheFile *string, debug bool) {
		for {
			time.Sleep(1 * time.Minute)
			err := clru.SaveFile(*cacheFile)
			if err != nil {
				log.Println("Error saving cache file: ", err)
			} else {
				if debug {
					log.Println("Cache file saved: ", *cacheFile)
				}
			}
		}
	}(clru, &cacheFile, debug)

	router := gin.Default()
	clru.Set("CacheHit", 0, cache.NoExpiration)
	clru.Set("CacheMiss", 0, cache.NoExpiration)

	router.GET("/kea/:subnet/:macAddress/", func(c *gin.Context) {
		subnet := c.Param("subnet")
		macAddress := c.Param("macAddress")
		data, err := serverDBQuery(&subnet, &macAddress)
		if err != nil {
			c.String(http.StatusForbidden, fmt.Sprintf("We got an error from ServerDB %s", err))
		}

		c.String(http.StatusOK, data)
		if debug {
			log.Printf("Sending answer from %s - %s: %s", subnet, macAddress, data)
		}
	})

	router.GET("/stats/", func(c *gin.Context) {
		hit, miss, ratio := cacheStats()
		c.JSON(http.StatusOK, gin.H{
			"CacheHit":   hit,
			"CacheMiss":  miss,
			"CacheRatio": ratio,
		})
	})

	router.RunUnix(viper.GetString("socket_path"))
}
