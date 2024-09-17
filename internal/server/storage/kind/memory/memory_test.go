package memory

import (
	"context"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/slice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name   string
		preset map[string]metric.Metric
		metric metric.Metric
		want   any
	}{
		{
			name:   "set gauge",
			preset: map[string]metric.Metric{},
			metric: gauge.New("m1", 123.321),
			want:   "123.321",
		},
		{
			name:   "set counter",
			preset: map[string]metric.Metric{},
			metric: counter.New("m2", 123),
			want:   "123",
		},
		{
			name: "update counter",
			preset: map[string]metric.Metric{
				"m3": counter.New("m3", 123),
			},
			metric: counter.New("m3", 5),
			want:   "5", // because the storage should not know about business logic
		},
		{
			name: "gauge -> counter",
			preset: map[string]metric.Metric{
				"m4": gauge.New("m4", 123.321),
			},
			metric: counter.New("m4", 5),
			want:   "5",
		},
		{
			name: "counter -> gauge",
			preset: map[string]metric.Metric{
				"m5": counter.New("m5", 123),
			},
			metric: gauge.New("m5", 123.321),
			want:   "123.321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := New()
			for _, metric := range tt.preset {
				storage.Save(ctx, metric)
			}
			storage.Save(ctx, tt.metric)
			metric, err := storage.Get(ctx, tt.metric.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.want, metric.StringValue())
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		preset     map[string]metric.Metric
		metricName string
		want       metric.Metric
	}{
		{
			name:       "empty storage",
			preset:     map[string]metric.Metric{},
			metricName: "m1",
			want:       nil,
		},
		{
			name: "defined name",
			preset: map[string]metric.Metric{
				"m2": counter.New("m2", 1),
			},
			metricName: "m2",
			want:       counter.New("m2", 1),
		},
		{
			name: "undefined name",
			preset: map[string]metric.Metric{
				"m3": gauge.New("m3", 1.1),
			},
			metricName: "m4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := New()
			for _, metric := range tt.preset {
				storage.Save(ctx, metric)
			}
			metric, err := storage.Get(ctx, tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, metric)
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		preset []metric.Metric
	}{
		{
			name:   "empty storage",
			preset: []metric.Metric{},
		},
		{
			name: "not empty storage",
			preset: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := New()
			for _, metric := range tt.preset {
				storage.Save(ctx, metric)
			}
			all, err := storage.GetAll(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.preset, slice.FromChannel(all))
		})
	}
}

func TestStorage_SaveGet(t *testing.T) {
	ctx := context.Background()
	storage := New()
	metric := counter.New("m1", 123)
	storage.Save(ctx, metric)
	get, err := storage.Get(ctx, "m1")
	require.NoError(t, err)
	assert.Equal(t, metric, get)
	assert.NotSame(t, metric, get)
	all, err := storage.GetAll(ctx)
	require.NoError(t, err)
	allSlice := slice.FromChannel(all)
	assert.Equal(t, metric, allSlice[0])
	assert.NotSame(t, metric, allSlice[0])
}
