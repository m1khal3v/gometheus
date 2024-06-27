package storage

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestStorage_Append(t *testing.T) {
	tests := []struct {
		name    string
		metrics []metric.Metric
	}{
		{
			"empty metrics",
			[]metric.Metric{},
		},
		{
			"append metrics",
			[]metric.Metric{
				counter.New("counter", 123),
				gauge.New("gauge", 123.456),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := New()
			storage.Append(tt.metrics)
			assert.Equal(t, tt.metrics, storage.metrics)
			for index, metric := range tt.metrics {
				assert.NotSame(t, metric, storage.metrics[index])
			}
		})
	}
}

func TestStorage_Remove(t *testing.T) {
	metrics := []metric.Metric{
		counter.New("counter", 123),
		gauge.New("gauge", 123.456),
	}
	tests := []struct {
		name    string
		filter  filter
		metrics []metric.Metric
	}{
		{
			name: "remove counters",
			filter: func(metric metric.Metric) bool {
				return metric.Type() == counter.MetricType
			},
			metrics: []metric.Metric{
				gauge.New("gauge", 123.456),
			},
		},
		{
			name: "remove value > 123",
			filter: func(metric metric.Metric) bool {
				value, _ := strconv.ParseFloat(metric.StringValue(), 64)
				return value > 123
			},
			metrics: []metric.Metric{
				counter.New("counter", 123),
			},
		},
		{
			name: "remove name contains u",
			filter: func(metric metric.Metric) bool {
				return strings.Contains(metric.Name(), "u")
			},
			metrics: []metric.Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := New()
			storage.Append(metrics)
			storage.Remove(tt.filter)
			assert.Equal(t, tt.metrics, storage.metrics)
		})
	}
}
