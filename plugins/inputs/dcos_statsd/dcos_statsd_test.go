package dcos_statsd

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	acc := &testutil.Accumulator{}
	// TODO ephemeral free port instead of 8088
	ds := DCOSStatsd{Listen: ":8088"}
	err := ds.Start(acc)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	defer ds.Stop()

	resp, err := http.Get("http://localhost:8088/")
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	// TODO meaningful index response
	assert.Equal(t, "Hello World!", string(body))
}

// TODO TestGather
