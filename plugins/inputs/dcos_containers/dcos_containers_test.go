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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGather(t *testing.T) {
	testCases := []struct {
		fixture string
		fields  map[string]interface{}
		tags    map[string]string
	}{
		{"empty", map[string]interface{}{}, map[string]string{}},
		{
			"healthy",
			map[string]interface{}{
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
			map[string]string{
				"executor_name": "executor",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			var acc testutil.Accumulator

			server, teardown := startTestServer(t, tc.fixture)
			defer teardown()

			dc := DCOSContainers{AgentUrl: server.URL}

			err := acc.GatherError(dc.Gather)
			assert.Nil(t, err)
			// Test that all expected fields are present
			if len(tc.fields) > 0 {
				acc.AssertContainsTaggedFields(t, "dcos_containers", tc.fields, tc.tags)
			}
			// TODO test timestamps
		})
	}
}

// startTestServer starts a server and serves the specified fixture's content
// at /api/v1
func startTestServer(t *testing.T, fixture string) (*httptest.Server, func()) {
	content := loadFixture(t, fixture+".bin")

	router := http.NewServeMux()
	router.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
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
