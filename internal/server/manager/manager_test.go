package manager

import (
	"context"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/m1khal3v/gometheus/pkg/slice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			ctx := context.Background()
			storage := memory.New()
			if tt.preset != nil {
				require.NoError(t, storage.Save(ctx, tt.preset))
			}
			manager := New(storage)
			saved, err := manager.Save(ctx, tt.metric)
			require.NoError(t, err)
			assert.Equal(t, tt.want, saved)
			got, err := manager.Get(ctx, tt.metric.Type(), tt.metric.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_SaveBatch(t *testing.T) {
	tests := []struct {
		name   string
		preset []metric.Metric
		metric []metric.Metric
		want   []metric.Metric
	}{
		{
			name: "metrics",
			metric: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.321),
			},
			want: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.321),
			},
		},
		{
			name: "replace metrics",
			preset: []metric.Metric{
				gauge.New("m1", 321.123),
				counter.New("m2", 123),
				gauge.New("m3", 123.321),
			},
			metric: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 321.123),
			},
			want: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 246),
				gauge.New("m3", 321.123),
			},
		},
		{
			name: "metrics collision with replace",
			preset: []metric.Metric{
				gauge.New("m1", 321.123),
				counter.New("m2", 123),
				gauge.New("m3", 123.321),
			},
			metric: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 321.123),
				gauge.New("m1", 333.333),
				counter.New("m2", 123),
				gauge.New("m3", 444.444),
			},
			want: []metric.Metric{
				gauge.New("m1", 333.333),
				counter.New("m2", 369),
				gauge.New("m3", 444.444),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := memory.New()
			if tt.preset != nil {
				require.NoError(t, storage.SaveBatch(ctx, tt.preset))
			}
			manager := New(storage)
			saved, err := manager.SaveBatch(ctx, tt.metric)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, saved)
			all, err := manager.GetAll(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, slice.FromChannel(all))
		})
	}
}

func TestManager_Get(t *testing.T) {
	ctx := context.Background()
	storage := memory.New()
	storage.Save(ctx, gauge.New("m1", 123.321))
	storage.Save(ctx, counter.New("m2", 123))
	manager := New(storage)

	tests := []struct {
		name       string
		metricName string
		metricType string
		want       metric.Metric
	}{
		{
			name:       "valid gauge",
			metricName: "m1",
			metricType: gauge.MetricType,
			want:       gauge.New("m1", 123.321),
		},
		{
			name:       "valid counter",
			metricName: "m2",
			metricType: counter.MetricType,
			want:       counter.New("m2", 123),
		},
		{
			name:       "nonexistent gauge",
			metricName: "m3",
			metricType: gauge.MetricType,
			want:       nil,
		},
		{
			name:       "nonexistent counter",
			metricName: "m3",
			metricType: counter.MetricType,
			want:       nil,
		},
		{
			name:       "gauge type mismatch",
			metricName: "m1",
			metricType: counter.MetricType,
			want:       nil,
		},
		{
			name:       "counter type mismatch",
			metricName: "m2",
			metricType: gauge.MetricType,
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Get(ctx, tt.metricType, tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
