package memory

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/counter"
	"github.com/m1khal3v/gometheus/internal/metric/gauge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStorage(t *testing.T) {
	assert.Equal(t, New(), &Storage{metrics: make(map[string]_metric.Metric)})
}

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name   string
		preset map[string]_metric.Metric
		metric _metric.Metric
		want   any
	}{
		{
			name:   "set gauge",
			preset: map[string]_metric.Metric{},
			metric: gauge.New("m1", 123.321),
			want:   "123.321",
		},
		{
			name:   "set counter",
			preset: map[string]_metric.Metric{},
			metric: counter.New("m2", 123),
			want:   "123",
		},
		{
			name: "update counter",
			preset: map[string]_metric.Metric{
				"m3": counter.New("m3", 123),
			},
			metric: counter.New("m3", 5),
			want:   "5", // because the storage should not know about business logic
		},
		{
			name: "gauge -> counter",
			preset: map[string]_metric.Metric{
				"m4": gauge.New("m4", 123.321),
			},
			metric: counter.New("m4", 5),
			want:   "5",
		},
		{
			name: "counter -> gauge",
			preset: map[string]_metric.Metric{
				"m5": counter.New("m5", 123),
			},
			metric: gauge.New("m5", 123.321),
			want:   "123.321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			storage.Save(tt.metric)
			metric, ok := storage.metrics[tt.metric.GetName()]
			assert.True(t, ok)
			assert.Equal(t, tt.want, metric.GetStringValue())
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		preset     map[string]_metric.Metric
		metricName string
		want       _metric.Metric
	}{
		{
			name:       "empty storage",
			preset:     map[string]_metric.Metric{},
			metricName: "m1",
			want:       nil,
		},
		{
			name: "defined name",
			preset: map[string]_metric.Metric{
				"m2": counter.New("m2", 1),
			},
			metricName: "m2",
			want:       counter.New("m2", 1),
		},
		{
			name: "undefined name",
			preset: map[string]_metric.Metric{
				"m3": gauge.New("m3", 1.1),
			},
			metricName: "m4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			assert.Equal(t, tt.want, storage.Get(tt.metricName))
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		preset map[string]_metric.Metric
	}{
		{
			name:   "empty storage",
			preset: map[string]_metric.Metric{},
		},
		{
			name: "not empty storage",
			preset: map[string]_metric.Metric{
				"m1": gauge.New("m1", 123.321),
				"m2": counter.New("m2", 123),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			assert.Equal(t, tt.preset, storage.GetAll())
		})
	}
}
