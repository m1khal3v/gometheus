package queue

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueue(t *testing.T) {
	tests := []struct {
		name    string
		metrics []metric.Metric
	}{
		{
			"empty metrics",
			[]metric.Metric{},
		},
		{
			"not empty metrics",
			[]metric.Metric{
				counter.New("counter", 123),
				gauge.New("gauge", 123.456),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := New(100)
			storage.PushSlice(tt.metrics)
			metrics := make([]metric.Metric, 0, len(tt.metrics))
			for metric := range storage.Pop(uint64(len(tt.metrics))) {
				metrics = append(metrics, metric)
			}
			require.Equal(t, tt.metrics, metrics)
			for index, metric := range tt.metrics {
				assert.NotSame(t, metric, metrics[index])
			}
		})
	}
}
