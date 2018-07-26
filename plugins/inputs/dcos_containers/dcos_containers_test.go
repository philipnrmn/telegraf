package dcos_containers

// NOTE: this file relies on protobuf fixtures. These are binary files and
// cannot readily be changed. We therefore provide the go generate step below
// which serializes the contents of json files in the testdata directory to
// protobuf.
//
// You should run 'go generate' every time you change one of the json files in
// the testdata directory, and commit both the changed json file and the
// changed binary file.
//go:generate go run cmd/gen.go

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

// raw protobuf request types:
var (
	GET_CONTAINERS = []byte{8, 10}
	GET_STATE      = []byte{8, 9}
)

type testCase struct {
	fixture string
	fields  map[string]interface{}
	tags    map[string]string
	ts      int64
	// cachedContainers prepopulates the plugin with container info
	cachedContainers map[string]containerInfo
	// containers is how the dc.containers should look after
	// metrics are retrieved
	containers map[string]containerInfo
}

var (
	TEST_CASES = []testCase{
		testCase{
			fixture:          "empty",
			fields:           map[string]interface{}{},
			tags:             map[string]string{},
			ts:               0,
			cachedContainers: map[string]containerInfo{},
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
				"service_name":  "framework",
				"executor_name": "executor",
				"task_name":     "task",
			},
			ts: 1388534400,
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework"},
			},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework"},
			},
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
				"service_name":  "framework",
				"executor_name": "executor",
				"task_name":     "task",
			},
			ts:               1388534400,
			cachedContainers: map[string]containerInfo{},
			containers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework"},
			},
		},
		testCase{
			fixture: "stale",
			fields:  map[string]interface{}{},
			tags:    map[string]string{},
			ts:      0,
			cachedContainers: map[string]containerInfo{
				"abc123": containerInfo{"abc123", "task", "executor", "framework"},
			},
			containers: map[string]containerInfo{},
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
				containers:    tc.cachedContainers,
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

			if tc.containers != nil {
				assertHasContainers(t, dc, tc.containers)
			}
		})
	}
}

// startTestServer starts a server and serves the specified fixture's content
// at /api/v1
func startTestServer(t *testing.T, fixture string) (*httptest.Server, func()) {
	router := http.NewServeMux()
	router.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		if bytes.Equal(body, GET_CONTAINERS) {
			containers := loadFixture(t, filepath.Join(fixture, "containers.bin"))
			w.Write(containers)
			return
		}
		if bytes.Equal(body, GET_STATE) {
			state := loadFixture(t, filepath.Join(fixture, "state.bin"))
			w.Write(state)
			return
		}
		panic("Body contained an unknown request: " + string(body))
	})
	server := httptest.NewServer(router)

	return server, server.Close

}

// loadFixture retrieves data from a file in ./testdata
func loadFixture(t *testing.T, filename string) []byte {
	path := filepath.Join("testdata", filename)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
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

// assertHasContainers checks that DCOSContainers has the expected container info
func assertHasContainers(t *testing.T, dc DCOSContainers, expected map[string]containerInfo) {
	if len(dc.containers) != len(expected) {
		t.Errorf("expected container info cache to be of size %d, but it was %d", len(expected), len(dc.containers))
		return
	}
	for cid, _ := range expected {
		if _, ok := dc.containers[cid]; !ok {
			t.Errorf("expected container %s to be present in cache, but it was not", cid)
			return
		}
	}
}
