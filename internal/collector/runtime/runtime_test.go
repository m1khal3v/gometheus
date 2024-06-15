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
		if metric.GetName() == "PollCount" {
			assert.Equal(t, "counter", metric.GetType())
		} else {
			assert.Equal(t, "gauge", metric.GetType())
		}
	}
}
