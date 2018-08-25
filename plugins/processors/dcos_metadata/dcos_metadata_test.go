package dcos_metadata

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	fixture string
	metrics []telegraf.Metric
	tags    map[string]string
	// cachedContainers prepopulates the plugin with container info
	cachedContainers map[string]containerInfo
	// containers is how the dm.containers map should look after
	// metrics are retrieved
	containers map[string]containerInfo
}

var (
	TEST_CASES = []testCase{
		testCase{
			fixture: "empty",
			metrics: []telegraf.Metric{},
			tags:    map[string]string{},
		},
		testCase{
			fixture: "normal",
			metrics: []telegraf.Metric{},
			tags: map[string]string{
				"service_name":  "framework",
				"executor_name": "executor",
				"task_name":     "task",
			},
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework", map[string]string{}},
			},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework", map[string]string{}},
			},
		},
		testCase{
			fixture: "fresh",
			metrics: []telegraf.Metric{},
			tags: map[string]string{
				"service_name":  "framework",
				"executor_name": "executor",
				"task_name":     "task",
			},
			cachedContainers: map[string]containerInfo{},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework", map[string]string{}},
			},
		},
		testCase{
			fixture: "stale",
			metrics: []telegraf.Metric{},
			tags:    map[string]string{},
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework", map[string]string{}},
			},
			containers: map[string]containerInfo{},
		},
	}
)

func TestApply(t *testing.T) {

	for _, tc := range TEST_CASES {
		t.Run(tc.fixture, func(t *testing.T) {
			server, teardown := startTestServer(t, tc.fixture)
			defer teardown()

			input, _ := metric.New("test",
				map[string]string{"container_id": "abc123"},
				map[string]interface{}{"value": int64(1)},
				time.Now(),
			)

			dm := DCOSMetadata{
				MesosAgentUrl: server.URL,
				Timeout:       100 * time.Millisecond,
			}
			metrics := dm.Apply(input)
			assert.Equal(t, 1, len(metrics))

			output := metrics[0]

			expectedTags := map[string]string{
				"container_id": "abc123",
			}
			assert.Equal(t, expectedTags, output.Tags())
		})
	}

}
