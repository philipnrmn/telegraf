package dcos_metadata

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// TODO: doublecheck we can't get the operator api from the agent
type DCOSMetadata struct {
	LeaderUrl string
}

const sampleConfig = `
## The URL of the leading mesos master
leader_url = "leader.mesos:5050"
`

// SampleConfig returns the default configuration
func (dm *DCOSMetadata) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_metadata
func (dm *DCOSMetadata) Description() string {
	return "Plugin for adding metadata to dcos-specific metrics"
}

// Apply the filter to the given metric
func (dm *DCOSMetadata) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		log.Println(metric)
	}
	return in
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_metadata::init")
	// TODO: get state.json once
	// TODO: connect to mesos master for operator updates
	processors.Add("dcos_metadata", func() telegraf.Processor {
		return &DCOSMetadata{}
	})
}
