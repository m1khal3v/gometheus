package memory

import (
	storages "github.com/m1khal3v/gometheus/internal/storage"
	"github.com/m1khal3v/gometheus/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStorage(t *testing.T) {
	assert.Equal(t, NewStorage(), &Storage{metrics: make(map[string]store.Metric)})
}

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name    string
		preset  map[string]store.Metric
		metric  *store.Metric
		want    any
		wantErr error
	}{
		{
			name:   "set gauge",
			preset: map[string]store.Metric{},
			metric: &store.Metric{
				Type:       "gauge",
				Name:       "m1",
				FloatValue: float64(123.321),
			},
			want: float64(123.321),
		},
		{
			name:   "set counter",
			preset: map[string]store.Metric{},
			metric: &store.Metric{
				Type:     "counter",
				Name:     "m2",
				IntValue: int64(123),
			},
			want: int64(123),
		},
		{
			name: "update counter",
			preset: map[string]store.Metric{
				"m3": {
					Type:     "counter",
					Name:     "m3",
					IntValue: int64(123),
				},
			},
			metric: &store.Metric{
				Type:     "counter",
				Name:     "m3",
				IntValue: int64(5),
			},
			want: int64(128),
		},
		{
			name: "gauge -> counter",
			preset: map[string]store.Metric{
				"m4": {
					Type:       "gauge",
					Name:       "m4",
					FloatValue: float64(123.321),
				},
			},
			metric: &store.Metric{
				Type:     "counter",
				Name:     "m4",
				IntValue: int64(5),
			},
			want: int64(5),
		},
		{
			name:   "invalid metric type",
			preset: map[string]store.Metric{},
			metric: &store.Metric{
				Type:     "invalid",
				Name:     "m5",
				IntValue: int64(5),
			},
			wantErr: store.ErrUnknownType{
				Type: "invalid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			err := storage.Save(tt.metric)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.Nil(t, err)
			metric, ok := storage.metrics[tt.metric.Name]
			assert.True(t, ok)
			assert.Equal(t, tt.want, (&metric).GetValue())
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		preset     map[string]store.Metric
		metricName string
		want       *store.Metric
		wantErr    error
	}{
		{
			name:       "empty storage",
			preset:     map[string]store.Metric{},
			metricName: "m1",
			want:       nil,
			wantErr: storages.ErrMetricNotFound{
				Name: "m1",
			},
		},
		{
			name: "defined name",
			preset: map[string]store.Metric{
				"m2": {
					Type:     store.MetricTypeCounter,
					Name:     "m2",
					IntValue: int64(1),
				},
			},
			metricName: "m2",
			want: &store.Metric{
				Type:     store.MetricTypeCounter,
				Name:     "m2",
				IntValue: int64(1),
			},
		},
		{
			name: "undefined name",
			preset: map[string]store.Metric{
				"m3": {
					Type:       store.MetricTypeGauge,
					Name:       "m3",
					FloatValue: float64(1.1),
				},
			},
			metricName: "m4",
			want:       nil,
			wantErr: storages.ErrMetricNotFound{
				Name: "m4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			value, err := storage.Get(tt.metricName)
			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.want, value)
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		preset map[string]store.Metric
	}{
		{
			name:   "empty storage",
			preset: map[string]store.Metric{},
		},
		{
			name: "not empty storage",
			preset: map[string]store.Metric{
				"m1": {
					Name:       "m1",
					Type:       "gauge",
					FloatValue: float64(123.123),
				},
				"m2": {
					Name:     "m2",
					Type:     "counter",
					IntValue: int64(123),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				metrics: tt.preset,
			}
			metrics, err := storage.GetAll()
			assert.Nil(t, err)
			assert.Equal(t, tt.preset, metrics)
		})
	}
}
