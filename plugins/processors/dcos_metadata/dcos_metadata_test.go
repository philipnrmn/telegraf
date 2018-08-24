package dcos_metadata

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	//server, teardown := startTestServer(t, "empty")
	// defer teardown()

	metric, _ := metric.New("test",
		map[string]string{"container_id": "abc123"},
		map[string]interface{}{"value": int64(1)},
		time.Now(),
	)

	dm := DCOSMetadata{
		Timeout: 100 * time.Millisecond,
	}
	metrics := dm.Apply(metric)
	assert.Equal(t, 1, len(metrics))

}
