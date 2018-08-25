package dcos_metadata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"

	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
)

type DCOSMetadata struct {
	MesosAgentUrl string
	Timeout       time.Duration
	containers    map[string]containerInfo
}

// containerInfo is a tuple of metadata which we use to map a container ID to
// information about the task, executor and framework.
type containerInfo struct {
	containerID   string
	taskName      string
	executorName  string
	frameworkName string
	taskLabels    map[string]string
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

	// stale tracks whether our container cache is stale
	stale := false
	for _, metric := range in {

		// TODO: does the metric have a containerID tag?
		if _, ok := metric.Tags()["container_id"]; ok {
			// TODO try to get associated task, executor and framework
		}
		// TODO: retrieve container ID and tag it with metrics.
		// TODO: was the container ID not found? Get state.
	}

	// TODO rate-limit calls to getState
	if stale {
		uri := dm.MesosAgentUrl + "/api/v1"
		cli := httpagent.NewSender(httpcli.New(httpcli.Endpoint(uri)).Send)
		ctx, _ := context.WithTimeout(context.Background(), dm.Timeout)
		go dm.getState(ctx, cli)
	}
	return in
}

// getState requests state from the operator API
func (dm *DCOSMetadata) getState(ctx context.Context, cli calls.Sender) (*agent.Response_GetState, error) {
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
func (dm *DCOSMetadata) prune(gc *agent.Response_GetContainers) {
	for cid, _ := range dm.containers {
		found := false
		for _, c := range gc.Containers {
			if cid == c.ContainerID.Value {
				found = true
				break
			}
		}
		if !found {
			delete(dm.containers, cid)
		}
	}
	return
}

// isConsistent returns true if container info is available for all containers
func (dm *DCOSMetadata) isConsistent(gc *agent.Response_GetContainers) bool {
	if gc.Containers == nil && len(dm.containers) == 0 {
		return true
	}
	if len(gc.Containers) != len(dm.containers) {
		return false
	}
	for _, c := range gc.Containers {
		if _, ok := dm.containers[c.ContainerID.Value]; !ok {
			return false
		}
	}
	return true
}

// reconcile adds newly discovered container info to container info
func (dm *DCOSMetadata) reconcile(gc *agent.Response_GetContainers, gs *agent.Response_GetState) {
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
		if _, ok := dm.containers[cid]; !ok {

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
			dm.containers[cid] = containerInfo{
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

// init is called once when telegraf starts
func init() {
	processors.Add("dcos_metadata", func() telegraf.Processor {
		return &DCOSMetadata{
			Timeout: 10 * time.Second,
		}
	})
}
