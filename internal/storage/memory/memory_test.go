package memory

import (
	storages "github.com/m1khal3v/gometheus/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStorage(t *testing.T) {
	assert.Equal(t, NewStorage(), &Storage{Metrics: make(map[string]*storages.Metric)})
}

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name    string
		preset  map[string]*storages.Metric
		metric  *storages.Metric
		want    any
		wantErr error
	}{
		{
			name:   "set gauge",
			preset: map[string]*storages.Metric{},
			metric: &storages.Metric{
				Type:       "gauge",
				Name:       "m1",
				FloatValue: float64(123.321),
			},
			want: float64(123.321),
		},
		{
			name:   "set counter",
			preset: map[string]*storages.Metric{},
			metric: &storages.Metric{
				Type:     "counter",
				Name:     "m2",
				IntValue: int64(123),
			},
			want: int64(123),
		},
		{
			name: "update counter",
			preset: map[string]*storages.Metric{
				"m3": {
					Type:     "counter",
					Name:     "m3",
					IntValue: int64(123),
				},
			},
			metric: &storages.Metric{
				Type:     "counter",
				Name:     "m3",
				IntValue: int64(5),
			},
			want: int64(128),
		},
		{
			name: "gauge -> counter",
			preset: map[string]*storages.Metric{
				"m4": {
					Type:       "gauge",
					Name:       "m4",
					FloatValue: float64(123.321),
				},
			},
			metric: &storages.Metric{
				Type:     "counter",
				Name:     "m4",
				IntValue: int64(5),
			},
			want: int64(5),
		},
		{
			name:   "invalid metric type",
			preset: map[string]*storages.Metric{},
			metric: &storages.Metric{
				Type:     "invalid",
				Name:     "m5",
				IntValue: int64(5),
			},
			wantErr: storages.UnknownTypeError{
				Type: "invalid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				Metrics: tt.preset,
			}
			err := storage.Save(tt.metric)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.Nil(t, err)
			assert.Contains(t, storage.Metrics, tt.metric.Name)
			assert.Equal(t, tt.want, storage.Metrics[tt.metric.Name].GetValue())
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		preset     map[string]*storages.Metric
		metricName string
		want       *storages.Metric
	}{
		{
			name:       "empty storage",
			preset:     map[string]*storages.Metric{},
			metricName: "m1",
			want:       nil,
		},
		{
			name: "defined name",
			preset: map[string]*storages.Metric{
				"m2": {
					Type:     storages.MetricTypeCounter,
					Name:     "m2",
					IntValue: int64(1),
				},
			},
			metricName: "m2",
			want: &storages.Metric{
				Type:     storages.MetricTypeCounter,
				Name:     "m2",
				IntValue: int64(1),
			},
		},
		{
			name: "undefined name",
			preset: map[string]*storages.Metric{
				"m3": {
					Type:       storages.MetricTypeGauge,
					Name:       "m3",
					FloatValue: float64(1.1),
				},
			},
			metricName: "m4",
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &Storage{
				Metrics: tt.preset,
			}
			assert.Equal(t, tt.want, storage.Get(tt.metricName))
		})
	}
}
