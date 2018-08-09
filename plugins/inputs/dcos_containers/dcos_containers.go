package dcos_containers

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

const sampleConfig = `
## The URL of the local mesos agent
mesos_agent_url = "http://$NODE_PRIVATE_IP:5051"
`

// DCOSContainers describes the options available to this plugin
type DCOSContainers struct {
	MesosAgentUrl string
	// TODO configurable timeouts
}

// SampleConfig returns the default configuration
func (dc *DCOSContainers) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_containers
func (dc *DCOSContainers) Description() string {
	return "Plugin for monitoring mesos container resource consumption"
}

// Gather takes in an accumulator and adds the metrics that the plugin gathers.
// It is invoked on a schedule (default every 10s) by the telegraf runtime.
func (dc *DCOSContainers) Gather(acc telegraf.Accumulator) error {
	// TODO: timeout

	uri := dc.MesosAgentUrl + "/api/v1"
	cli := httpagent.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
	ctx := context.Background()

	gc, err := dc.getContainers(ctx, cli)
	if err != nil {
		return err
	}

	for _, c := range gc.Containers {
		acc.AddFields("dcos_containers", cFields(c), cTags(c), cTS(c))
	}

	return nil
}

// getContainers requests a list of containers from the operator API
func (dc *DCOSContainers) getContainers(ctx context.Context, cli calls.Sender) (*agent.Response_GetContainers, error) {
	resp, err := cli.Send(ctx, calls.NonStreaming(calls.GetContainers()))
	if err != nil {
		return nil, err
	}
	r, err := processResponse(resp, agent.Response_GET_CONTAINERS)
	if err != nil {
		return nil, err
	}

	gc := r.GetGetContainers()
	if gc == nil {
		return gc, errors.New("the getContainers response from the mesos agent was empty")
	}

	return gc, nil
}

// processResponse reads the response from a triggered request, verifies its
// type, and returns an agent response
func processResponse(resp mesos.Response, t agent.Response_Type) (agent.Response, error) {
	var r agent.Response
	defer func() {
		if resp != nil {
			resp.Close()
		}
	}()
	for {
		if err := resp.Decode(&r); err != nil {
			if err == io.EOF {
				break
			}
			return r, err
		}
	}
	if r.GetType() == t {
		return r, nil
	} else {
		return r, fmt.Errorf("processResponse expected type %q, got %q", t, r.GetType())
	}
}

// cFields flattens a Container object into a map of metric labels and values
func cFields(c agent.Response_GetContainers_Container) map[string]interface{} {
	results := make(map[string]interface{})
	// TODO account for all possible fields in ResourceStatistics
	rs := c.ResourceStatistics
	results["cpus_limit"] = *rs.CPUsLimit
	results["cpus_nr_periods"] = *rs.CPUsNrPeriods
	results["cpus_nr_throttled"] = *rs.CPUsNrThrottled
	results["cpus_system_time_secs"] = *rs.CPUsSystemTimeSecs
	results["cpus_throttled_time_secs"] = *rs.CPUsThrottledTimeSecs
	results["cpus_user_time_secs"] = *rs.CPUsUserTimeSecs
	results["mem_anon_bytes"] = *rs.MemAnonBytes
	results["mem_file_bytes"] = *rs.MemFileBytes
	results["mem_limit_bytes"] = *rs.MemLimitBytes
	results["mem_mapped_file_bytes"] = *rs.MemMappedFileBytes
	results["mem_rss_bytes"] = *rs.MemRSSBytes
	return results
}

// cTags extracts relevant metadata from a Container object as a map of tags
func cTags(c agent.Response_GetContainers_Container) map[string]string {
	return map[string]string{"container_id": c.ContainerID}
}

// cTS retrieves the timestamp from a Container object as a time rounded to the
// nearest second
func cTS(c agent.Response_GetContainers_Container) time.Time {
	return time.Unix(int64(math.Trunc(c.ResourceStatistics.Timestamp)), 0)
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_containers::init")
	inputs.Add("dcos_containers", func() telegraf.Input {
		return &DCOSContainers{}
	})
}
