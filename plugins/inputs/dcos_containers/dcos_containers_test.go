package dcos_containers

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	fixture string
	fields  map[string]interface{}
	tags    map[string]string
	ts      int64
}

var (
	TEST_CASES = []testCase{
		testCase{
			fixture: "empty",
			fields:  map[string]interface{}{},
			tags:    map[string]string{},
			ts:      0,
		},
		testCase{
			fixture: "normal",
			fields: map[string]interface{}{
				"cpus_limit":               8.25,
				"cpus_nr_periods":          uint32(769021),
				"cpus_nr_throttled":        uint32(1046),
				"cpus_system_time_secs":    34501.45,
				"cpus_throttled_time_secs": 352.597023453,
				"cpus_user_time_secs":      96348.84,
				"mem_anon_bytes":           uint64(4845449216),
				"mem_file_bytes":           uint64(260165632),
				"mem_limit_bytes":          uint64(7650410496),
				"mem_mapped_file_bytes":    uint64(7159808),
				"mem_rss_bytes":            uint64(5105614848),
			},
			tags: map[string]string{
				"container_id": "abc123",
			},
			ts: 1388534400,
		},
		testCase{
			fixture: "fresh",
			fields: map[string]interface{}{
				"cpus_limit":               8.25,
				"cpus_nr_periods":          uint32(769021),
				"cpus_nr_throttled":        uint32(1046),
				"cpus_system_time_secs":    34501.45,
				"cpus_throttled_time_secs": 352.597023453,
				"cpus_user_time_secs":      96348.84,
				"mem_anon_bytes":           uint64(4845449216),
				"mem_file_bytes":           uint64(260165632),
				"mem_limit_bytes":          uint64(7650410496),
				"mem_mapped_file_bytes":    uint64(7159808),
				"mem_rss_bytes":            uint64(5105614848),
			},
			tags: map[string]string{
				"container_id": "abc123",
			},
			ts: 1388534400,
		},
		testCase{
			fixture: "stale",
			fields:  map[string]interface{}{},
			tags:    map[string]string{},
			ts:      0,
		},
	}
)

func TestGather(t *testing.T) {
	for _, tc := range TEST_CASES {
		t.Run(tc.fixture, func(t *testing.T) {
			var acc testutil.Accumulator

			server, teardown := startTestServer(t, tc.fixture)
			defer teardown()

			dc := DCOSContainers{
				MesosAgentUrl: server.URL,
			}

			err := acc.GatherError(dc.Gather)
			assert.Nil(t, err)
			if len(tc.fields) > 0 {
				// all expected fields are present
				acc.AssertContainsFields(t, "dcos_containers", tc.fields)
				// all expected tags are present
				acc.AssertContainsTaggedFields(t, "dcos_containers", tc.fields, tc.tags)
				// the expected timestamp is present
				assertHasTimestamp(t, acc, "dcos_containers", tc.ts)
			} else {
				acc.AssertDoesNotContainMeasurement(t, "dcos_containers")
			}
		})
	}
}

// assertHasTimestamp checks that the specified measurement has the expected ts
func assertHasTimestamp(t *testing.T, acc testutil.Accumulator, measurement string, ts int64) {
	expected := time.Unix(ts, 0)
	if acc.HasTimestamp(measurement, expected) {
		return
	}
	if m, ok := acc.Get(measurement); ok {
		actual := m.Time
		t.Errorf("%s had a bad timestamp: expected %q; got %q", measurement, expected, actual)
		return
	}
	t.Errorf("%s could not be retrieved while attempting to assert it had timestamp", measurement)
}
