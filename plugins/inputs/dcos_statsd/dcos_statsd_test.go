package dcos_statsd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/dcos_statsd/containers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	t.Run("Server with no saved state", func(t *testing.T) {
		ds := DCOSStatsd{containers: []containers.Container{}}
		// startTestServer runs a /health request test
		addr := startTestServer(t, &ds)
		defer ds.Stop()

		// Check that no containers were created
		resp, err := http.Get(addr + "/containers")
		assertResponseWas(t, resp, err, "[]")
	})

	t.Run("Server with a single container saved", func(t *testing.T) {
		// Create a temp dir:
		dir, err := ioutil.TempDir("", "containers")
		if err != nil {
			assert.Fail(t, fmt.Sprintf("Could not create temp dir: %s", err))
		}
		defer os.RemoveAll(dir)

		// Create JSON in memory:
		ctrport := findFreePort()
		ctrjson := fmt.Sprintf(
			`{"container_id":"abc123","statsd_host":"127.0.0.1","statsd_port":%d}`,
			ctrport)

		// Write JSON to disk:
		err = ioutil.WriteFile(dir+"/abc123", []byte(ctrjson), 0666)
		if err != nil {
			assert.Fail(t, fmt.Sprintf("Could not write container state: %s", err))
		}

		// Finally run DCOSStatsd.Start():
		ds := DCOSStatsd{ContainersDir: dir}
		addr := startTestServer(t, &ds)
		defer ds.Stop()

		// Ensure that container shows up in output:
		resp, err := http.Get(addr + "/containers")
		// encoding/json respects alphabetical order, so this is safe
		assertResponseWas(t, resp, err, fmt.Sprintf("[%s]", ctrjson))
	})

}

func TestStop(t *testing.T) {
	ds := DCOSStatsd{}
	addr := startTestServer(t, &ds)
	ds.Stop()

	// Test that the server has stopped
	resp, err := http.Get(addr + "/health")
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	// TODO test that all statsd servers are stopped
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	ds := DCOSStatsd{}

	err := acc.GatherError(ds.Gather)
	assert.Nil(t, err)

	// TODO test that statsd metrics are passed in and tagged
}

// startTestServer starts a server on the specified DCOSStatsd on a randomly
// selected port and returns the address on which it will be served. It also
// runs a test against the /health endpoint to ensure that the command API is
// ready.
func startTestServer(t *testing.T, ds *DCOSStatsd) string {
	port := findFreePort()
	ds.Listen = fmt.Sprintf(":%d", port)
	addr := fmt.Sprintf("http://localhost:%d", port)

	var acc telegraf.Accumulator
	acc = &testutil.Accumulator{}

	err := ds.Start(acc)
	assert.Nil(t, err)

	// Ensure that the command API is ready
	_, err = http.Get(addr + "/health")
	assert.Nil(t, err)

	return addr
}

// findFreePort momentarily listens on :0, then closes the connection and
// returns the port assigned
func findFreePort() int {
	ln, _ := net.Listen("tcp", ":0")
	ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port
}

// assertResponseWas is a convenience method for testing http request responses
func assertResponseWas(t *testing.T, r *http.Response, err error, expected string) {
	assert.Nil(t, err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(body))
}
