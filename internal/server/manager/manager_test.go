package manager

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManager_Save(t *testing.T) {
	tests := []struct {
		name   string
		preset metric.Metric
		metric metric.Metric
		want   metric.Metric
	}{
		{
			name:   "gauge",
			metric: gauge.New("m1", 123.321),
			want:   gauge.New("m1", 123.321),
		},
		{
			name:   "counter",
			metric: counter.New("m1", 123),
			want:   counter.New("m1", 123),
		},
		{
			name:   "gauge update",
			preset: gauge.New("m1", 321.123),
			metric: gauge.New("m1", 123.321),
			want:   gauge.New("m1", 123.321),
		},
		{
			name:   "counter update",
			preset: counter.New("m1", 77),
			metric: counter.New("m1", 123),
			want:   counter.New("m1", 200),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			if tt.preset != nil {
				storage.Save(tt.preset)
			}
			manager := New(storage)
			manager.Save(tt.metric)
			assert.Equal(t, tt.want, manager.Get(tt.metric.Type(), tt.metric.Name()))
		})
	}
}
