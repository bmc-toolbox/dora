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

package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rcrowley/go-metrics"
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

//HandlerFunc is a function which should be used as middleware to count requests stats
// such as request processing time, request and responce size and store it in rcrowley/go-metrics.DefaultRegistry.
func (m *Metrics) HandlerFunc(metricsPrefix []string, ignoreURLs []string, replaceSlashWithUnderscore bool) gin.HandlerFunc {
	theRegistry := newRegistry()
	return func(c *gin.Context) {
		// ignore mechanism for particular endpoints
		if contains(ignoreURLs, c.Request.URL.String()) {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		elapsed := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		resSz := int64(c.Writer.Size())
		url := m.ReqCntURLLabelMappingFn(c)
		if replaceSlashWithUnderscore {
			// replace slashes with underscores as they will be replaced by dots in Graphite otherwise
			url = strings.Replace(url, "/", "_", -1)
		}

		// drop non-existent urls to prevent creating a metric for each such url
		if status == "404" {
			url = "all"
		}

		var (
			metricsPath    = strings.Join(append(metricsPrefix, []string{method, status, url}...), ".")
			processTimeKey = metricsPath + ".req_process_time"
			reqSzKey       = metricsPath + ".req_size"
			resSzKey       = metricsPath + ".resp_size"
		)
		theRegistry.timer(processTimeKey).Update(elapsed)
		theRegistry.histogram(reqSzKey).Update(reqSz)
		theRegistry.histogram(resSzKey, ).Update(resSz)
	}
}

// registry stores metrics objects in rcrowley/go-metrics.DefaultRegistry. It
// re-implements and thus avoids using
// rcrowley/go-metrics.DefaultRegistry.GetOrRegister(), because that would force
// to re-create metrics objects for each request but that are not actually used.
type registry struct {
	mu      sync.RWMutex
	metrics map[string]interface{}
}

func newRegistry() *registry { return &registry{metrics: make(map[string]interface{})} }

func (r *registry) timer(key string) metrics.Timer {
	return r.getOrRegister(key, "timer").(metrics.Timer)
}

func (r *registry) histogram(key string) metrics.Histogram {
	return r.getOrRegister(key, "histogram").(metrics.Histogram)
}

func (r *registry) getOrRegister(key string, typ string) interface{} {
	// fast path
	r.mu.RLock()
	incumbent, ok := r.metrics[key]
	r.mu.RUnlock()
	if ok {
		return incumbent
	}

	// slow path
	r.mu.Lock()
	defer r.mu.Unlock()

	incumbent, ok = r.metrics[key]
	if ok {
		return incumbent
	}

	var contender interface{}
	switch typ {
	case "timer":
		contender = metrics.NewTimer()
	case "histogram":
		contender = metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015))
	default:
		panic(fmt.Errorf("metric type not implemented: %q", typ))
	}
	r.metrics[key] = contender
	metrics.MustRegister(key, contender)
	return contender
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
