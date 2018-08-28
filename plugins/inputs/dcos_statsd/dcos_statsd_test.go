package dcos_statsd

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	ds := DCOSStatsd{}

	err := acc.GatherError(ds.Gather)
	assert.Nil(t, err)
}
