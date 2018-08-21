package dcos_statsd

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	api "github.com/influxdata/telegraf/plugins/inputs/dcos_statsd/api"
)

const sampleConfig = `
## The address on which to listen
listen = ":8088"
`

// DCOSStatsd describes the options available to this plugin
type DCOSStatsd struct {
	// Listen is the address on which we listen
	Listen string

	apiServer *http.Server
}

// SampleConfig returns the default configuration
func (ds *DCOSStatsd) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_containers
func (ds *DCOSStatsd) Description() string {
	return "Plugin for monitoring statsd metrics from mesos tasks"
}

// Gather takes in an accumulator and adds the metrics that the plugin gathers.
// It is invoked on a schedule (default every 10s) by the telegraf runtime.
func (ds *DCOSStatsd) Gather(acc telegraf.Accumulator) error {
	log.Println("dcos_statsd::gather")
	// TODO create a waitgroup
	// TODO run Gather() in a goroutine foreach server
	// TODO await
	// TODO pipe the results into the accumulator
	return nil
}

// Start is called when telegraf is ready for the plugin to start gathering
// metrics
func (ds *DCOSStatsd) Start(_ telegraf.Accumulator) error {
	log.Println("dcos_statsd::start")
	router := api.NewRouter()

	srv := &http.Server{
		Handler: router,
		Addr:    ds.Listen,
		// TODO configurable timeouts
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	ds.apiServer = srv
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// Stop is called when telegraf is stopping
func (ds *DCOSStatsd) Stop() {
	log.Println("dcos_statsd::stop")
	// TODO configurable timeouts
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	ds.apiServer.Shutdown(ctx)
	// TODO shut down all statsd instances
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_statsd::init")
	inputs.Add("dcos_statsd", func() telegraf.Input {
		return &DCOSStatsd{}
	})
}
