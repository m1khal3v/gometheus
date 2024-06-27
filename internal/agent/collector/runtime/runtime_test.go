package runtime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollector_Collect(t *testing.T) {
	collector := New()
	metrics := collector.Collect()
	assert.Len(t, metrics, 28)
	for _, metric := range metrics {
		if metric.Name() == "PollCount" {
			assert.Equal(t, "counter", metric.Type())
		} else {
			assert.Equal(t, "gauge", metric.Type())
		}
	}
}
