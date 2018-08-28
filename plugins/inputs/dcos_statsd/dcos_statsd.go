package dcos_statsd

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
## The address on which the command API should listen
listen = ":8888"
## The directory in which container information is stored
containers_dir = "/run/dcos/mesos/isolators/com_mesosphere_MetricsIsolatorModule/containers"
`

type DCOSStatsd struct {
	// Listen is the address on which the command API listens
	Listen string
	// ContainersDir is the directory in which container information is stored
	ContainersDir string
}

// SampleConfig returns the default configuration
func (ds *DCOSStatsd) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_containers
func (ds *DCOSStatsd) Description() string {
	return "Plugin for monitoring statsd metrics from mesos tasks"
}

// Start is called when the service plugin is ready to start working
func (ds *DCOSStatsd) Start(_ telegraf.Accumulator) error {
	// TODO start the command API
	// TODO load containers from disc
	// TODO start servers
	return nil
}

// Gather takes in an accumulator and adds the metrics that the plugin gathers.
// It is invoked on a schedule (default every 10s) by the telegraf runtime.
func (ds *DCOSStatsd) Gather(_ telegraf.Accumulator) error {
	// TODO instantiate a custom accumulator for each plugin
	// TODO wait for all plugins to accumulate their metrics
	// TODO add container_id tags to each metric
	// TODO pass all metrics into the telegraf accumulator
	return nil
}

// Stop is called when the service plugin needs to stop working
func (ds *DCOSStatsd) Stop() {
	// TODO stop the command API
	// TODO stop servers
}

func init() {
	inputs.Add("dcos_statsd", func() telegraf.Input {
		return &DCOSStatsd{
			ContainersDir: "/run/dcos/mesos/isolators/com_mesosphere_MetricsIsolatorModule/containers",
		}
	})
}
