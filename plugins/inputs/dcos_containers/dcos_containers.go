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
mesos_agent_url = "http://127.0.0.1:5051"
`

// containerInfo is a tuple of metadata which we use to map a container ID to
// information about the task, executor and framework.
type containerInfo struct {
	containerID   string
	taskName      string
	executorName  string
	frameworkName string
	// TODO add taskLabels back in
	// taskLabels    map[string]string
}

// DCOSContainers describes the options available to this plugin
type DCOSContainers struct {
	MesosAgentUrl string
	// containers maps container ID to related metadata obtained from state
	containers map[string]containerInfo
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
	dc.prune(gc)
	if !dc.isConsistent(gc) {
		// new containers were found
		state, err := dc.getState(ctx, cli)
		if err != nil {
			return err
		}
		dc.reconcile(gc, state)
	}

	for _, c := range gc.Containers {
		if info, ok := dc.containers[c.ContainerID.Value]; ok {
			acc.AddFields("dcos_containers", cFields(c), cTags(info), cTS(c))
		} else {
			// TODO better warning
			fmt.Println("could not record metrics for", c.ContainerID.Value, "as no metadata was found in /state")
		}
	}

	return nil
}

// getContainers requests a list of containers from the operator API
func (dc *DCOSContainers) getContainers(ctx context.Context, cli calls.Sender) (*agent.Response_GetContainers, error) {
	// TODO error handling
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

// getState requests state from the operator API
func (dc *DCOSContainers) getState(ctx context.Context, cli calls.Sender) (*agent.Response_GetState, error) {
	// TODO error handling
	resp, err := cli.Send(ctx, calls.NonStreaming(calls.GetState()))
	if err != nil {
		return nil, err
	}
	r, err := processResponse(resp, agent.Response_GET_STATE)
	if err != nil {
		return nil, err
	}

	gs := r.GetGetState()
	if gs == nil {
		return gs, errors.New("the getState response from the mesos agent was empty")
	}
	return gs, nil
}

// prune removes container info for stale containers
func (dc *DCOSContainers) prune(gc *agent.Response_GetContainers) {
	for cid, _ := range dc.containers {
		found := false
		for _, c := range gc.Containers {
			if cid == c.ContainerID.Value {
				found = true
				break
			}
		}
		if !found {
			delete(dc.containers, cid)
		}
	}
	return
}

// isConsistent returns true if container info is available for all containers
func (dc *DCOSContainers) isConsistent(gc *agent.Response_GetContainers) bool {
	if gc.Containers == nil && len(dc.containers) == 0 {
		return true
	}
	if len(gc.Containers) != len(dc.containers) {
		return false
	}
	for _, c := range gc.Containers {
		if _, ok := dc.containers[c.ContainerID.Value]; !ok {
			return false
		}
	}
	return true
}

// reconcile adds newly discovered container info to container info
func (dc *DCOSContainers) reconcile(gc *agent.Response_GetContainers, gs *agent.Response_GetState) {
	gt := gs.GetGetTasks()
	tasks := gt.GetLaunchedTasks()
	gf := gs.GetGetFrameworks()
	frameworks := gf.Frameworks

	for _, c := range gc.Containers {
		cid := c.ContainerID.Value
		fid := c.FrameworkID.Value

		var task mesos.Task
		var framework mesos.FrameworkInfo

		// TODO break these into separate methods
		if _, ok := dc.containers[cid]; !ok {

			// find task:
			for _, t := range tasks {
				if len(t.Statuses) == 0 {
					continue
				}
				s := t.Statuses[0]
				// TODO exercise this code in a test
				if s.ContainerStatus.ContainerID.Parent != nil {
					if s.ContainerStatus.ContainerID.Parent.Value == cid {
						task = t
						break
					}
				}
				if s.ContainerStatus.ContainerID.Value == cid {
					task = t
					break
				}
			}

			// TODO find task labels

			// find framework:
			for _, f := range frameworks {
				if f.FrameworkInfo.ID.Value == fid {
					framework = f.FrameworkInfo
					break
				}
			}

			// TODO exercise this code in a test
			eName := ""
			// executor name can be missing
			if c.ExecutorName != nil {
				eName = *c.ExecutorName
			}
			dc.containers[cid] = containerInfo{
				containerID:   cid,
				executorName:  eName,
				frameworkName: framework.Name,
				taskName:      task.Name,
			}
		}
	}
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
func cTags(info containerInfo) map[string]string {
	results := make(map[string]string)
	results["service_name"] = info.frameworkName
	results["executor_name"] = info.executorName
	results["task_name"] = info.taskName
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
		return &DCOSContainers{
			containers: make(map[string]containerInfo),
		}
	})
}
