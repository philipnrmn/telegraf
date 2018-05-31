package dcos_containers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
## The URL of the local mesos agent
agent_url = "http://127.0.0.1:5051"
`

// AgentStats describes the resource consumption of a mesos container
type AgentStats struct {
	// TODO: container_id
	ExecutorID   string `json:"executor_id"`
	ExecutorName string `json:"executor_name"`
	FrameworkID  string `json:"framework_id"`
	Source       string `json:"source"`
	Statistics   struct {
		CpusLimit             float64 `json:"cpus_limit"`
		CpusNrPeriods         int64   `json:"cpus_nr_periods"`
		CpusNrThrottled       int64   `json:"cpus_nr_throttled"`
		CpusSystemTimeSecs    float64 `json:"cpus_system_time_secs"`
		CpusThrottledTimeSecs float64 `json:"cpus_throttled_time_secs"`
		CpusUserTimeSecs      float64 `json:"cpus_user_time_secs"`
		MemAnonBytes          int64   `json:"mem_anon_bytes"`
		MemFileBytes          int64   `json:"mem_file_bytes"`
		MemLimitBytes         int64   `json:"mem_limit_bytes"`
		MemMappedFileBytes    int64   `json:"mem_mapped_file_bytes"`
		MemRssBytes           int64   `json:"mem_rss_bytes"`
		Timestamp             float64 `json:"timestamp"`
	} `json:"statistics"`
	// TODO: blkio
}

// getFields flattens the statistics in AgentStats to a map
func (as *AgentStats) getFields() map[string]interface{} {
	// TODO detect nil results and don't add them to this map
	results := make(map[string]interface{})
	results["cpus_limit"] = as.Statistics.CpusLimit
	results["cpus_nr_periods"] = as.Statistics.CpusNrPeriods
	results["cpus_nr_throttled"] = as.Statistics.CpusNrThrottled
	results["cpus_system_time_secs"] = as.Statistics.CpusSystemTimeSecs
	results["cpus_throttled_time_secs"] = as.Statistics.CpusThrottledTimeSecs
	results["cpus_user_time_secs"] = as.Statistics.CpusUserTimeSecs
	results["mem_anon_bytes"] = as.Statistics.MemAnonBytes
	results["mem_file_bytes"] = as.Statistics.MemFileBytes
	results["mem_limit_bytes"] = as.Statistics.MemLimitBytes
	results["mem_mapped_file_bytes"] = as.Statistics.MemMappedFileBytes
	results["mem_rss_bytes"] = as.Statistics.MemRssBytes

	return results
}

// getTags extracts relevant metadata in AgentStats to a map
func (as *AgentStats) getTags() map[string]string {
	results := make(map[string]string)
	results["executor_id"] = as.ExecutorID
	results["executor_name"] = as.ExecutorName
	results["framework_id"] = as.FrameworkID
	results["source"] = as.Source
	return results
}

// getTimestamp returns the timestamp as a time rounded to the second
func (as *AgentStats) getTimestamp() time.Time {
	return time.Unix(int64(math.Trunc(as.Statistics.Timestamp)), 0)
}

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
	// TODO: use /containers instead
	// TODO: error handling
	resp, _ := http.Get(dc.AgentUrl + "/monitor/statistics")
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	stats := []AgentStats{}
	json.Unmarshal(body, &stats)

	for _, as := range stats {
		acc.AddFields("dcos_containers", as.getFields(), as.getTags(), as.getTimestamp())
	}

	return nil
}

// init is called once when telegraf starts
func init() {
	log.Println("dcos_containers::init")
	// TODO: request to mesos to ensure that it's reachable
	inputs.Add("dcos_containers", func() telegraf.Input {
		return &DCOSContainers{}
	})
}
