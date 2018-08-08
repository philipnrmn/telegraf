package dcos_metadata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
)

type DCOSMetadata struct {
	MesosAgentUrl string
}

const sampleConfig = `
## The URL of the local mesos agent
mesos_agent_url = "http://$NODE_PRIVATE_IP:5051"
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
		// TODO: does the metric have a containerID tag?
		// TODO: retrieve container ID and tag it with metrics.
		// TODO: was the container ID not found? Get state.

	}
	return in
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_metadata::init")
	processors.Add("dcos_metadata", func() telegraf.Processor {
		return &DCOSMetadata{}
	})
}
