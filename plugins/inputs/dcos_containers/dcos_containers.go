package dcos_containers

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DCOSContainers struct{}

// SampleConfig returns the default configuration
func (dc *DCOSContainers) SampleConfig() string {
	// TODO: sample config
	return "TODO: sample config"
}

// Description returns a one-sentence description of dcos_containers
func (dc *DCOSContainers) Description() string {
	return "Plugin for monitoring mesos container resource consumption"
}

// Gather takes in an accumulator and adds the metrics that the plugin gathers.
func (dc *DCOSContainers) Gather(acc telegraf.Accumulator) error {
	// TODO: gather metrics
	return nil
}

func init() {
	inputs.Add("dcos_containers", func() telegraf.Input {
		return &DCOSContainers{}
	})
}
