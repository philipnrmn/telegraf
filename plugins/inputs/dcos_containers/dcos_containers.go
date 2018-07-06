package dcos_containers

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
)

const sampleConfig = `
## The URL of the local mesos agent
agent_url = "http://127.0.0.1:5051"
`

// DCOSContainers describes the options available to this plugin
type DCOSContainers struct {
	AgentUrl string
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
	// TODO: error handling

	uri := dc.AgentUrl + "/api/v1"
	cli := httpagent.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
	ctx := context.Background()

	resp, err := cli.Send(ctx, calls.NonStreaming(calls.GetContainers()))

	defer func() {
		if resp != nil {
			resp.Close()
		}
	}()
	if err != nil {
		return err
	}
	for {
		var r agent.Response
		if err := resp.Decode(&r); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if t := r.GetType(); t == agent.Response_GET_CONTAINERS {
			gc := r.GetGetContainers()
			for _, c := range gc.Containers {
				acc.AddFields("dcos_containers", cFields(c), cTags(c), cTS(c))
			}
		} else {
			// TODO better error
			fmt.Println("not getcontainers", t, r)
		}
	}

	return nil
}

// cFields flattens a Container object into a map of metric labels and values
func cFields(c agent.Response_GetContainers_Container) map[string]interface{} {
	results := make(map[string]interface{})
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
	results := make(map[string]string)
	// results["service_name"] = *c.FrameworkName
	results["executor_name"] = *c.ExecutorName
	// results["task_name"] = *rs.TaskName
	return results
}

// cTS retrieves the timestamp from a Container object as a time rounded to the
// nearest second
func cTS(c agent.Response_GetContainers_Container) time.Time {
	return time.Unix(int64(math.Trunc(c.ResourceStatistics.Timestamp)), 0)
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_containers::init")
	// TODO: request to mesos to ensure that it's reachable
	inputs.Add("dcos_containers", func() telegraf.Input {
		return &DCOSContainers{}
	})
}
