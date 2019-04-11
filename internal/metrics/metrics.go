// Copyright © 2018 Joel Rebello <joel.rebello@booking.com>
// Copyright © 2018 Juliano Martinez <juliano.martinez@booking.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cyberdelia/go-metrics-graphite"
	gometrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

var (
	emm *emitter
)

// emitter struct holds attributes for the metrics emitter.
// we can convert int64 to float64, but not other way around
// because of that we store the metrics data in float64
type emitter struct {
	registry    gometrics.Registry
	metricsChan chan metric
	metricsData map[string]map[string]float64
}

// metric struct holds attributes for a metric.
type metric struct {
	Type  string   //counter/gauge
	Key   []string //metric key
	Value float64  //metric value
}

// Setup sets up external and internal metric sinks.
func Setup(clientType string, host string, port int, prefix string, flushInterval time.Duration) (err error) {
	if emm != nil {
		return err
	}

	emm = &emitter{
		registry:    gometrics.NewRegistry(),
		metricsChan: make(chan metric),
		metricsData: make(map[string]map[string]float64),
	}

	//setup metrics client based on config
	switch clientType {
	case "graphite":
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return fmt.Errorf("error resolving tcp addr -> %s", err.Error())
		}

		go graphite.Graphite(gometrics.DefaultRegistry, flushInterval, prefix, addr)
	default:
		return fmt.Errorf("no supported metrics client declared in config")
	}

	//go routine that records metricsData
	go emm.store()

	return err
}

//- writes/updates metric key/vals to metricsData
//- register and write metrics to the go-metrics registries.
func (e *emitter) store() {
	//A map of metric names to go-metrics registry
	//the keys to this map could be of type metrics.Counter/metrics.Gauge
	goMetricsRegistry := make(map[string]interface{})

	for {
		data, ok := <-e.metricsChan
		if !ok {
			return
		}

		identifier := data.Key[0]
		key := strings.Join(data.Key, ".")

		_, keyExists := e.metricsData[identifier]
		if !keyExists {
			e.metricsData[identifier] = make(map[string]float64)
		}

		//register the metric with go-metrics,
		//the metric key is used as the identifier.
		_, registryExists := goMetricsRegistry[identifier]
		if !registryExists {
			switch data.Type {
			case "counter":
				c := gometrics.NewCounter()
				gometrics.Register(key, c)
				goMetricsRegistry[key] = c
			case "gauge":
				g := gometrics.NewGauge()
				gometrics.Register(key, g)
				goMetricsRegistry[key] = g
			case "timer":
				g := gometrics.NewTimer()
				gometrics.Register(key, g)
				goMetricsRegistry[key] = g
			case "histogram":
				g := gometrics.NewHistogram(gometrics.NewExpDecaySample(1028, 0.015))
				gometrics.Register(key, g)
				goMetricsRegistry[key] = g
			}
		}

		//based on the metric type, update the store/registry.
		switch data.Type {
		case "counter":
			e.metricsData[identifier][key] += data.Value

			//type assert metrics registry to its type - metrics.Counter
			//type cast float64 metric value type to int64
			goMetricsRegistry[key].(gometrics.Counter).Inc(
				int64(e.metricsData[identifier][key]))
		case "gauge":
			e.metricsData[identifier][key] = data.Value

			//type assert metrics registry to its type - metrics.Gauge
			//type cast float64 metric value type to int64
			goMetricsRegistry[key].(gometrics.Gauge).Update(
				int64(e.metricsData[identifier][key]))
		case "timer":
			e.metricsData[identifier][key] = data.Value

			//type assert metrics registry to its type - metrics.Timer
			//type cast float64 metric value type to int64
			goMetricsRegistry[key].(gometrics.Timer).Update(
				time.Duration(e.metricsData[identifier][key]))
		case "histogram":
			e.metricsData[identifier][key] = data.Value

			//type assert metrics registry to its type - metrics.Histogram
			//type cast float64 metric value type to int64
			goMetricsRegistry[key].(gometrics.Histogram).Update(
				int64(e.metricsData[identifier][key]))
		}
	}
}

//Logs current metrics
func (e *emitter) dumpStats() {
	for source, metricsTmp := range e.metricsData {
		var metric string
		for k, v := range metricsTmp {
			metric += fmt.Sprintf("%s: %f ", k, v)
		}
		log.WithFields(log.Fields{"data": metric, "source": source}).Info("metric")
	}
}

// IncrCounter sets up metric attributes and passes them to the metricsChan.
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = float64 metric value
func IncrCounter(key []string, value int64) {
	d := metric{
		Type:  "counter",
		Key:   key,
		Value: float64(value),
	}

	emm.metricsChan <- d
}

// UpdateGauge sets up the Gauge metric and passes them to the metricsChan.
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = float64 metric value
func UpdateGauge(key []string, value int64) {
	d := metric{
		Type:  "gauge",
		Key:   key,
		Value: float64(value),
	}

	emm.metricsChan <- d
}

// UpdateTimer sets up the Timer metric and passes them to the metricsChan.
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = time.Time metric value
func UpdateTimer(key []string, value time.Duration) {
	d := metric{
		Type:  "timer",
		Key:   key,
		Value: float64(value.Nanoseconds()),
	}

	emm.metricsChan <- d
}

// UpdateHistogram sets up the Histogram metric and passes them to the metricsChan.
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = int64 metric value
func UpdateHistogram(key []string, value int64) {
	d := metric{
		Type:  "histogram",
		Key:   key,
		Value: float64(value),
	}

	emm.metricsChan <- d
}

// MeasureRuntime measures time elapsed since invocation
func MeasureRuntime(key []string, start time.Time) {
	//convert time.Duration to milliseconds
	elapsed := int64(time.Since(start) / time.Millisecond)
	UpdateGauge(key, elapsed)
}

// Close runs cleanup actions
func Close(printStats bool) {
	close(emm.metricsChan)

	if printStats {
		emm.dumpStats()
	}
}
