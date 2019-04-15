# gin-go-metrics - [gin-gonic/gin](https://github.com/gin-gonic/gin) middleware to gather and store metrics using [rcrowley/go-metrics](https://github.com/rcrowley/go-metrics)

## How to use

### gin middleware

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
)

func main(){
	err := gin_metrics.Setup(
		"graphite",
		"localhost",
		2003,
		"server",
		time.Minute,
	)
	if err != nil {
		fmt.Printf("Failed to set up monitoring: %s\n", err)
		os.Exit(1)
	}

	r := gin.New()

	// parameter to NewMetrics tells which variables need to be
	// expanded in metrics, more on that by link:
	// https://banzaicloud.com/blog/monitoring-gin-with-prometheus/
	p := gin_metrics.NewMetrics([]string{})
	r.Use(p.HandlerFunc())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.Run(":8000")
}
```

### standalone metrics sending

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bmc-toolbox/gin-go-metrics"
)

func main(){
	err := gin_metrics.Setup(
		"graphite",
		"localhost",
		2003,
		"server",
		time.Minute,
	)
	if err != nil {
		fmt.Printf("Failed to set up monitoring: %s\n", err)
		os.Exit(1)
	}
	// collect data using provided functions with provided arguments once a minute
	go gin_metrics.Scheduler(time.Minute, gin_metrics.GoRuntimeStats, []string{""})
	go gin_metrics.Scheduler(time.Minute, gin_metrics.MeasureRuntime, []string{"uptime"}, time.Now())
}
```

## Provided metrics

Processing time and count stored in [go-metrics.Timer](https://github.com/rcrowley/go-metrics/blob/master/timer.go)

Request size and response size stored in [go-metrics.Histogram](https://github.com/rcrowley/go-metrics/blob/master/histogram.go)

## Data storage

Currently only sending data to Graphite with [cyberdelia/go-metrics-graphite](https://github.com/cyberdelia/go-metrics-graphite)
 is supported, however you can send data using
 [go-metrics.DefaultRegistry](https://github.com/rcrowley/go-metrics/blob/cf894ca225d73a7d5dbb4b3a922f4ae3608bb618/registry.go#L323) anywhere you want.

## Acknowledgment

This library was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification was generalized and published as Open Source on GitHub, for
which the authors would like to express their gratitude.
