package dcos_metadata

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type DCOSMetadata struct {
	MesosAgentUrl string
	Timeout       time.Duration
}

const sampleConfig = `
## The URL of the local mesos agent
mesos_agent_url = "http://$NODE_PRIVATE_IP:5051"
## The period after which requests to mesos agent should time out
timeout = "10s"
`

// SampleConfig returns the default configuration
func (dm *DCOSMetadata) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_metadata
func (dm *DCOSMetadata) Description() string {
	return "Plugin for adding metadata to dcos-specific metrics"
}

// Apply the filter to the given metrics
func (dm *DCOSMetadata) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		log.Println(metric)
		// TODO: does the metric have a containerID tag?
		// TODO: retrieve container ID and tag it with metrics.
		// TODO: was the container ID not found? Get state.

	}
	return in
}

// init is called once when telegraf starts
func init() {
	processors.Add("dcos_metadata", func() telegraf.Processor {
		return &DCOSMetadata{
			Timeout: 10 * time.Second,
		}
	})
}
