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
	rs := c.ResourceStatistics

	// These items are not in alphabetical order; instead we preserve the order
	// in the source of the ResourceStatistics struct to make it easy to update.
	warnIfNotSet(setIfNotNil(results, "processes", rs.GetProcesses))
	warnIfNotSet(setIfNotNil(results, "threads", rs.GetThreads))

	warnIfNotSet(setIfNotNil(results, "cpus_user_time_secs", rs.GetCPUsUserTimeSecs))
	warnIfNotSet(setIfNotNil(results, "cpus_system_time_secs", rs.GetCPUsSystemTimeSecs))
	warnIfNotSet(setIfNotNil(results, "cpus_limit", rs.GetCPUsLimit))
	warnIfNotSet(setIfNotNil(results, "cpus_nr_periods", rs.GetCPUsNrPeriods))
	warnIfNotSet(setIfNotNil(results, "cpus_nr_throttled", rs.GetCPUsNrThrottled))
	warnIfNotSet(setIfNotNil(results, "cpus_throttled_time_secs", rs.GetCPUsThrottledTimeSecs))

	warnIfNotSet(setIfNotNil(results, "mem_total_bytes", rs.GetMemTotalBytes))
	warnIfNotSet(setIfNotNil(results, "mem_total_memsw_bytes", rs.GetMemTotalMemswBytes))
	warnIfNotSet(setIfNotNil(results, "mem_limit_bytes", rs.GetMemLimitBytes))
	warnIfNotSet(setIfNotNil(results, "mem_soft_limit_bytes", rs.GetMemSoftLimitBytes))
	warnIfNotSet(setIfNotNil(results, "mem_file_bytes", rs.GetMemFileBytes))
	warnIfNotSet(setIfNotNil(results, "mem_anon_bytes", rs.GetMemAnonBytes))
	warnIfNotSet(setIfNotNil(results, "mem_cache_bytes", rs.GetMemCacheBytes))
	warnIfNotSet(setIfNotNil(results, "mem_rss_bytes", rs.GetMemRSSBytes))
	warnIfNotSet(setIfNotNil(results, "mem_mapped_file_bytes", rs.GetMemMappedFileBytes))
	warnIfNotSet(setIfNotNil(results, "mem_swap_bytes", rs.GetMemSwapBytes))
	warnIfNotSet(setIfNotNil(results, "mem_unevictable_bytes", rs.GetMemUnevictableBytes))
	warnIfNotSet(setIfNotNil(results, "mem_low_pressure_counter", rs.GetMemLowPressureCounter))
	warnIfNotSet(setIfNotNil(results, "mem_medium_pressure_counter", rs.GetMemMediumPressureCounter))
	warnIfNotSet(setIfNotNil(results, "mem_critical_pressure_counter", rs.GetMemCriticalPressureCounter))

	warnIfNotSet(setIfNotNil(results, "disk_limit_bytes", rs.GetDiskLimitBytes))
	warnIfNotSet(setIfNotNil(results, "disk_used_bytes", rs.GetDiskUsedBytes))
	// TODO: Disk statistics (*rs.DiskStatistics)
	// TODO: Blkio statistics (*rs.BlkioStatistics)
	// TODO: Perf statistics (*rs.Perf)
	warnIfNotSet(setIfNotNil(results, "net_rx_packets", rs.GetNetRxPackets))
	warnIfNotSet(setIfNotNil(results, "net_rx_bytes", rs.GetNetRxBytes))
	warnIfNotSet(setIfNotNil(results, "net_rx_errors", rs.GetNetRxErrors))
	warnIfNotSet(setIfNotNil(results, "net_rx_dropped", rs.GetNetRxDropped))
	warnIfNotSet(setIfNotNil(results, "net_tx_packets", rs.GetNetTxPackets))
	warnIfNotSet(setIfNotNil(results, "net_tx_bytes", rs.GetNetTxBytes))
	warnIfNotSet(setIfNotNil(results, "net_tx_errors", rs.GetNetTxErrors))
	warnIfNotSet(setIfNotNil(results, "net_tx_dropped", rs.GetNetTxDropped))
	warnIfNotSet(setIfNotNil(results, "net_tcp_rtt_microsecs_p50", rs.GetNetTCPRttMicrosecsP50))
	warnIfNotSet(setIfNotNil(results, "net_tcp_rtt_microsecs_p90", rs.GetNetTCPRttMicrosecsP90))
	warnIfNotSet(setIfNotNil(results, "net_tcp_rtt_microsecs_p95", rs.GetNetTCPRttMicrosecsP95))
	warnIfNotSet(setIfNotNil(results, "net_tcp_rtt_microsecs_p99", rs.GetNetTCPRttMicrosecsP99))
	warnIfNotSet(setIfNotNil(results, "net_tcp_active_connections", rs.GetNetTCPActiveConnections))
	warnIfNotSet(setIfNotNil(results, "net_tcp_time_wait_connections", rs.GetNetTCPTimeWaitConnections))
	// TODO: Net traffic control statistics (*rs.NetTrafficControlStatistics)
	// TODO: Net snmp statistics (*rs.NetSNMPStatistics)

	return results
}

// cTags extracts relevant metadata from a Container object as a map of tags
func cTags(c agent.Response_GetContainers_Container) map[string]string {
	return map[string]string{"container_id": c.ContainerID.Value}
}

// cTS retrieves the timestamp from a Container object as a time rounded to the
// nearest second
func cTS(c agent.Response_GetContainers_Container) time.Time {
	return time.Unix(int64(math.Trunc(c.ResourceStatistics.Timestamp)), 0)
}

// setIfNotNil runs get() and adds its value to a map, if not nil
func setIfNotNil(target map[string]interface{}, key string, get interface{}) error {
	var val interface{}
	var zero interface{}

	switch get.(type) {
	case func() uint32:
		val = get.(func() uint32)()
		zero = uint32(0)
		break
	case func() uint64:
		val = get.(func() uint64)()
		zero = uint64(0)
		break
	case func() float64:
		val = get.(func() float64)()
		zero = float64(0)
		break
	default:
		return fmt.Errorf("get function for key %s was not of a recognized type", key)
	}
	// Zero is nil for numeric types
	if val != zero {
		target[key] = val
	}
	return nil
}

// warnIfNotSet is a convenience method to log a warning whenever setIfNotNil
// did not succesfully complete
func warnIfNotSet(err error) {
	if err != nil {
		log.Printf("Warning: %s", err)
	}
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_containers::init")
	inputs.Add("dcos_containers", func() telegraf.Input {
		return &DCOSContainers{}
	})
}
