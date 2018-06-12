package dcos_containers

import (
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGather(t *testing.T) {
	testCases := []struct {
		fixture string
		err     bool
	}{
		{"empty", false},
		{"healthy", false},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			var acc testutil.Accumulator
			expectedFields := loadResults(t, tc.fixture)
			server, teardown := startTestServer(t, tc.fixture)
			defer teardown()

			dc := DCOSContainers{AgentUrl: server.URL}

			err := acc.GatherError(dc.Gather)
			if tc.err {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			if len(expectedFields) > 0 {
				acc.AssertContainsFields(t, "dcos_containers", expectedFields)
			}
			// TODO test tags
			// TODO test timestamps
		})
	}
}

// startTestServer starts a server and serves the specified fixture's content
// at /monitor/statistics
func startTestServer(t *testing.T, fixture string) (*httptest.Server, func()) {
	content := loadFixture(t, fixture+".json")

	router := http.NewServeMux()
	router.HandleFunc("/monitor/statistics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	server := httptest.NewServer(router)

	return server, server.Close
}

// loadResults reads the specified fixture as a CSV
func loadResults(t *testing.T, fixture string) map[string]interface{} {
	results := make(map[string]interface{})
	body := loadFixture(t, fixture+".csv")
	r := csv.NewReader(bytes.NewReader(body))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		results[record[0]] = parseCsvValue(record[1])
	}
	return results
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

// parseCsvValue casts a string to int or float if possible,
// returning the original string if not
func parseCsvValue(in string) interface{} {
	if v, err := strconv.ParseInt(in, 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(in, 64); err == nil {
		return v
	}
	return in
}
