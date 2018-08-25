package dcos_statsd

import (
	"fmt"
	// "io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func xTestStart(t *testing.T) {
	ds := DCOSStatsd{}
	addr, err := startTestAPIServer(t, ds)
	defer func() {
		ds.Stop()
	}()

	assert.Nil(t, err)
	fmt.Println(err)
	time.Sleep(1 * time.Second)

	_, err = http.Get(addr)
	assert.Nil(t, err)

	// body, err := ioutil.ReadAll(resp.Body)
	// assert.Nil(t, err)

	// TODO meaningful index response
	// assert.Equal(t, "Hello World!", string(body))
	// fmt.Println("")

	// TODO test that Start() recovers servers persisted to disk
}

func xTestStop(t *testing.T) {
	ds := DCOSStatsd{}
	addr, err := startTestAPIServer(t, ds)
	ds.Stop()
	resp, err := http.Get(addr)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	// TODO test that Stop() stops all the servers

}

// TODO TestGather
func TestGather(t *testing.T) {
	ds := DCOSStatsd{}
	addr, err := startTestAPIServer(t, ds)
	defer ds.Stop()
	assert.Nil(t, err)
	t.Run("With a single server running", func(t *testing.T) {})
	t.Run("With multiple servers running", func(t *testing.T) {})
	t.Run("With no servers running", func(t *testing.T) {})
}

// startAPIServer starts the command server running on an ephemeral port
func startTestAPIServer(t *testing.T, ds DCOSStatsd) (string, error) {
	// Find a free port by momentarily listening in :0
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	err = ln.Close()
	if err != nil {
		return "", err
	}
	addr := ln.Addr().(*net.TCPAddr)
	listen := fmt.Sprintf("localhost:%d", addr.Port)

	ds.Listen = listen
	return fmt.Sprintf("http://%s/", listen), ds.Start(&testutil.Accumulator{})
}
