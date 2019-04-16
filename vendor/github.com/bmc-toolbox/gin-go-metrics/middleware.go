// MIT License
//
// Copyright (c) 2018 Zachary Sais
// Copyright (c) 2019 Dmitry Verkhoturov <dmitry.verkhoturov@booking.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gin_metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/*
RequestCounterURLLabelMappingFn is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:
func(c *gin.Context) string {
	url := c.Request.URL.String()
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}
which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Metrics contains the metrics gathered by the instance
type Metrics struct {
	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn
}

// NewMetrics generates a new set of metrics
func NewMetrics(expandedParams []string) *Metrics {
	return &Metrics{
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			url := c.Request.URL.EscapedPath() // i.e. by default do nothing, i.e. return URL as is
			for _, p := range c.Params {
				if contains(expandedParams, p.Key) {
					continue
				}
				url = strings.Replace(url, p.Value, ":"+p.Key, 1)
			}
			return url
		},
	}
}

func contains(slice []string, s string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}
	return false
}

func (m *Metrics) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ignore mechanism for particular endpoint
		//if c.Request.URL.String() == metricsPath {
		//	c.Next()
		//	return
		//}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		elapsed := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		resSz := int64(c.Writer.Size())
		url := m.ReqCntURLLabelMappingFn(c)
		// drop non-existent urls to prevent creating a metric for each such url
		if status == "404" {
			url = "all"
		}
		// replace slashes with underscores as they will be replaced by dots in Graphite otherwise
		url = strings.Replace(url, "/", "_", -1)

		UpdateTimer([]string{method, status, url, "requests"}, elapsed)
		UpdateHistogram([]string{method, status, url, "req_size"}, reqSz)
		UpdateHistogram([]string{method, status, url, "resp_size"}, resSz)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int64 {
	s := 0
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return int64(s)
}
