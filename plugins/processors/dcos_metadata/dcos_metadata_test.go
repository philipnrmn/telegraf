package dcos_metadata

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	fixture  string
	inputs   []telegraf.Metric
	expected []telegraf.Metric
	// cachedContainers prepopulates the plugin with container info
	cachedContainers map[string]containerInfo
	// containers is how the dm.containers map should look after
	// metrics are retrieved
	containers map[string]containerInfo
}

var (
	TEST_CASES = []testCase{
		// No metrics, no state; nothing to do
		testCase{
			fixture:  "empty",
			inputs:   []telegraf.Metric{},
			expected: []telegraf.Metric{},
		},
		// One metric, cached state; tags are added
		testCase{
			fixture: "normal",
			inputs: []telegraf.Metric{
				newMetric("test",
					map[string]string{"container_id": "abc123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			expected: []telegraf.Metric{
				newMetric("test",
					map[string]string{
						"container_id":  "abc123",
						"service_name":  "framework",
						"executor_name": "executor",
						"task_name":     "task",
						// Generated by mesos task labels:
						"FOO": "bar",
						"BAZ": "qux",
					},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework",
					map[string]string{"FOO": "bar", "BAZ": "qux"}},
			},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework",
					map[string]string{"FOO": "bar", "BAZ": "qux"}},
			},
		},
		// One metric, no cached state; no tags are added but state is updated
		testCase{
			fixture: "fresh",
			inputs: []telegraf.Metric{
				newMetric("test",
					map[string]string{"container_id": "abc123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			expected: []telegraf.Metric{
				newMetric("test",
					// We don't expect tags, since no cache exists
					map[string]string{"container_id": "abc123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			cachedContainers: map[string]containerInfo{},
			// We do expect the cache to be updated when apply is done
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework",
					// Ensure that the tags are picked up from state
					map[string]string{"FOO": "bar", "BAZ": "qux"}},
			},
		},
		// One metric without a container ID; nothing to do
		testCase{
			fixture: "unrelated",
			inputs: []telegraf.Metric{
				newMetric("test",
					map[string]string{}, // no container_id tag
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			expected: []telegraf.Metric{
				newMetric("test",
					map[string]string{},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			cachedContainers: map[string]containerInfo{},
			// We do not expect the cache to be updated
			containers: map[string]containerInfo{},
		},
		// Fetching a nested container ID
		testCase{
			fixture: "nested",
			inputs: []telegraf.Metric{
				newMetric("test",
					map[string]string{"container_id": "xyz123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			expected: []telegraf.Metric{
				newMetric("test",
					map[string]string{"container_id": "xyz123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			cachedContainers: map[string]containerInfo{},
			// We do not expect the cache to be updated
			containers: map[string]containerInfo{
				"xyz123": containerInfo{"xyz123", "task", "executor", "framework",
					map[string]string{}},
			},
		},
		// No executor;
		testCase{
			fixture: "noexecutor",
			inputs: []telegraf.Metric{
				newMetric("test",
					map[string]string{"container_id": "abc123"},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			expected: []telegraf.Metric{
				newMetric("test",
					map[string]string{
						"container_id": "abc123",
						"service_name": "framework",
						// no executor tag at all
						"task_name": "task",
					},
					map[string]interface{}{"value": int64(1)},
					time.Now(),
				),
			},
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "", "framework",
					map[string]string{}},
			},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "", "framework",
					map[string]string{}},
			},
		},
	}
)

func TestApply(t *testing.T) {
	for _, tc := range TEST_CASES {
		t.Run(tc.fixture, func(t *testing.T) {
			server, teardown := startTestServer(t, tc.fixture)
			defer teardown()

			dm := DCOSMetadata{
				MesosAgentUrl: server.URL,
				Timeout:       internal.Duration{Duration: 100 * time.Millisecond},
				RateLimit:     internal.Duration{Duration: 50 * time.Millisecond},
				containers:    tc.cachedContainers,
			}

			outputs := dm.Apply(tc.inputs...)

			// No metrics were dropped
			assert.Equal(t, len(tc.expected), len(outputs))
			// Tags were added as expected
			for i, actual := range outputs {
				expected := tc.expected[i]
				assert.Equal(t, expected.Name(), actual.Name())
				assert.Equal(t, expected.Tags(), actual.Tags())
			}

			waitForContainersToEqual(t, &dm, tc.containers, 100*time.Millisecond)
		})
	}
}

// newMetric is a convenience method which allows us to define test cases at
// package level without doing error handling
func newMetric(name string, tags map[string]string, fields map[string]interface{}, tm time.Time) telegraf.Metric {
	m, err := metric.New(name, tags, fields, tm)
	if err != nil {
		panic(err)
	}
	return m
}

// waitForContainersToEqual waits for the length of the container cache to
// change and asserts that they equal the expected, or times out
func waitForContainersToEqual(t *testing.T, dm *DCOSMetadata, expected map[string]containerInfo, timeout time.Duration) {
	done := make(chan bool)
	go func() {
		for {
			// acquiring the lock here avoids triggering the go race detector
			dm.mu.Lock()
			if len(dm.containers) == len(expected) {
				done <- true
				break
			}
			dm.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		assert.Equal(t, dm.containers, expected)
		return
	case <-time.After(timeout):
		assert.Fail(t, "Timed out waiting for a container update")
		return
	}
}
