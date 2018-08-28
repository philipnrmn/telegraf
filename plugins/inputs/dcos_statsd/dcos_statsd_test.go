package dcos_statsd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	ds := DCOSStatsd{}
	// startTestServer runs a /health request test
	addr := startTestServer(t, &ds)
	defer ds.Stop()

	// Check that no containers were created
	resp, err := http.Get(addr + "/containers")
	assertResponseWas(t, resp, err, "[]")

	// TODO test that saved statsd servers are started
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
