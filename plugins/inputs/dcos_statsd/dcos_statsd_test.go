package dcos_statsd

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	var acc telegraf.Accumulator
	acc = &testutil.Accumulator{}
	ds := DCOSStatsd{Listen: ":0"}
	err := ds.Start(acc)
	assert.Nil(t, err)

	// TODO test that the command API stats
	// TODO test that saved statsd servers are started
}

func TestStop(t *testing.T) {
	ds := DCOSStatsd{Listen: ":0"}
	ds.Stop()

	// TODO test that the command API stops
	// TODO test that all statsd servers are stopped
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	ds := DCOSStatsd{}

	err := acc.GatherError(ds.Gather)
	assert.Nil(t, err)

	// TODO test that statsd metrics are passed in and tagged
}
